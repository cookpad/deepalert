package deepalert

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/pkg/errors"
)

// Task is invoke argument of inspectors
type Task struct {
	ReportID  ReportID  `json:"report_id"`
	Attribute Attribute `json:"attribute"`
}

// TaskHandler is a function type of callback of inspector.
type TaskHandler func(ctx context.Context, attr Attribute) (ReportContentEntity, error)

func publishSNS(topicArn, region string, data interface{}) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Fail to marshal report data")
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	snsService := sns.New(ssn)

	_, err = snsService.Publish(&sns.PublishInput{
		Message:  aws.String(string(msg)),
		TopicArn: aws.String(topicArn),
	})

	if err != nil {
		return errors.Wrap(err, "Fail to publish report")
	}

	return nil
}

// StartInspector is a wrapper of Inspector.
func StartInspector(handler TaskHandler, author string) {
	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
		submitTo := os.Getenv("SUBMIT_TOPIC")
		if submitTo == "" {
			return errors.New("SUBMIT_TOPIC is not set, no destination")
		}
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			return errors.New("AWS_REGION is not set")
		}

		for _, record := range event.Records {
			var task Task
			msg := record.SNS.Message
			if err := json.Unmarshal([]byte(msg), &task); err != nil {
				return errors.Wrapf(err, "Fail to unmarshal task: %s", msg)
			}

			entity, err := handler(ctx, task.Attribute)
			if err != nil {
				return errors.Wrapf(err, "Fail to handle task: %v", task)
			}
			if entity == nil {
				continue // No contents
			}

			content := ReportContent{
				ReportID:  task.ReportID,
				Author:    author,
				Attribute: task.Attribute,
				Type:      entity.Type(),
				Content:   entity,
			}

			if err := publishSNS(submitTo, awsRegion, content); err != nil {
				return errors.Wrapf(err, "Fail to publish ReportContent to %s: %v", submitTo, content)
			}
		}

		return nil
	})
}
