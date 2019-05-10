package deepalert

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

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

// InspectFunc is a function type of callback of inspector.
type InspectFunc func(ctx context.Context, attr Attribute) (ReportContentEntity, error)

// SearchFunc is a function type of callback of search.
type SearchFunc func(ctx context.Context, attr Attribute) ([]Attribute, error)

func publishSNS(topicArn string, data interface{}) error {
	// arn
	// aws
	// sns
	// ap-northeast-1
	// 789035092620
	// xxxxxx
	arr := strings.Split(topicArn, ":")
	if len(arr) != 6 {
		return fmt.Errorf("Invalid SNS ARN format: %s", topicArn)
	}
	region := arr[3]

	msg, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Fail to marshal report data")
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	snsService := sns.New(ssn)

	resp, err := snsService.Publish(&sns.PublishInput{
		Message:  aws.String(string(msg)),
		TopicArn: aws.String(topicArn),
	})

	if err != nil {
		return errors.Wrap(err, "Fail to publish report")
	}
	log.Printf("Published SNS %v to %s: %v", data, topicArn, resp)

	return nil
}

// StartInspector is a wrapper of Inspector.
func StartInspector(handler InspectFunc, author, submitTopic string) {
	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
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
				Attribute: task.Attribute,
				Author:    author,
				Type:      entity.Type(),
				Content:   entity,
			}

			if err := publishSNS(submitTopic, content); err != nil {
				return errors.Wrapf(err, "Fail to publish ReportContent to %s: %v", submitTopic, content)
			}
		}

		return nil
	})
}

// StartSearch is a wrapper of Inspector.
func StartSearch(handler SearchFunc, author, attributeTopic string) {
	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
		log.Printf("(Search) SNS event: %v", event)
		for _, record := range event.Records {
			var task Task
			msg := record.SNS.Message
			if err := json.Unmarshal([]byte(msg), &task); err != nil {
				return errors.Wrapf(err, "Fail to unmarshal task: %s", msg)
			}
			log.Printf("(Search) SNS message: %s", msg)

			attributes, err := handler(ctx, task.Attribute)
			if err != nil {
				return errors.Wrapf(err, "Fail to handle task: %v", task)
			}
			if attributes == nil {
				continue // No contents
			}

			attrReport := ReportAttribute{
				ReportID:   task.ReportID,
				Original:   task.Attribute,
				Attributes: attributes,
				Author:     author,
			}

			if err := publishSNS(attributeTopic, attrReport); err != nil {
				return errors.Wrapf(err, "Fail to publish ReportAttribute to %s: %v", attributeTopic, attrReport)
			}
		}

		return nil
	})
}
