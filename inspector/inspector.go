package inspector

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/m-mizutani/deepalert"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Handler is a function type of callback of inspector.
type Handler func(ctx context.Context, attr deepalert.Attribute) (*deepalert.TaskResult, error)

var logger = logrus.New()

func init() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
}

type sqsClient interface {
	SendMessage(*sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
}

var newSqsClient = newAwsSqsClient

func newAwsSqsClient(region string) sqsClient {
	ssn := session.New(&aws.Config{Region: aws.String(region)})
	client := sqs.New(ssn)
	return client
}

// Sample: https://sqs.ap-northeast-1.amazonaws.com/123456789xxx/some-queue-name
var regexSqsURL = regexp.MustCompile(`https://sqs.([a-z0-9-]+).amazonaws.com`)

func extractRegionFromURL(url string) (*string, error) {
	if m := regexSqsURL.FindStringSubmatch(url); len(m) == 2 {
		return &m[1], nil
	}
	return nil, fmt.Errorf("Invalid SQS URL foramt: %v", url)
}

func sendSQS(msg interface{}, targetURL string) error {
	region, err := extractRegionFromURL(targetURL)
	if err != nil {
		return err
	}

	client := newSqsClient(*region)

	raw, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrapf(err, "Fail to marshal message: %v", msg)
	}

	input := sqs.SendMessageInput{
		QueueUrl:    &targetURL,
		MessageBody: aws.String(string(raw)),
	}
	resp, err := client.SendMessage(&input)

	if err != nil {
		return errors.Wrapf(err, "Fail to send SQS message: %v", input)
	}

	logger.WithField("resp", resp).Trace("Sent SQS message")

	return nil
}

type reportIDKey struct{}

var contextKey = &reportIDKey{}

// ReportIDFromCtx extracts ReportID from context. The function is available in handler called by Start
func ReportIDFromCtx(ctx context.Context) (*deepalert.ReportID, bool) {
	lc, ok := ctx.Value(contextKey).(*deepalert.ReportID)
	return lc, ok
}

// Start is a wrapper of Inspector.
func Start(handler Handler, author, attrQueueURL, contentQueueURL string) {
	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
		logger.WithField("event", event).Info("Start inspector")

		for _, record := range event.Records {
			var task deepalert.Task
			msg := record.SNS.Message
			if err := json.Unmarshal([]byte(msg), &task); err != nil {
				return errors.Wrapf(err, "Fail to unmarshal task: %s", msg)
			}

			logger.WithField("task", task).Info("run handler")
			newCtx := context.WithValue(ctx, contextKey, &task.ReportID)

			result, err := handler(newCtx, task.Attribute)
			if err != nil {
				return errors.Wrapf(err, "Fail to handle task: %v", task)
			}

			if result == nil {
				continue
			}

			// Sending entities
			for _, entity := range result.Contents {
				section := deepalert.ReportSection{
					ReportID:  task.ReportID,
					Attribute: task.Attribute,
					Author:    author,
					Type:      entity.Type(),
					Content:   entity,
				}

				if err := sendSQS(section, contentQueueURL); err != nil {
					return errors.Wrapf(err, "Fail to publish ReportContent to %s: %v", contentQueueURL, section)
				}
			}

			var newAttrs []deepalert.Attribute
			for _, attr := range result.NewAttributes {
				if attr.Timestamp == nil {
					attr.Timestamp = task.Attribute.Timestamp
				}
				newAttrs = append(newAttrs, attr)
			}

			// Sending new attributes
			if len(result.NewAttributes) > 0 {
				attrReport := deepalert.ReportAttribute{
					ReportID:   task.ReportID,
					Original:   task.Attribute,
					Attributes: newAttrs,
					Author:     author,
				}

				if err := sendSQS(attrReport, attrQueueURL); err != nil {
					return errors.Wrapf(err, "Fail to publish ReportAttribute to %s: %v", attrQueueURL, attrReport)
				}
			}
		}

		return nil
	})
}
