package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/inspector"
)

func dummyInspector(ctx context.Context, attr deepalert.Attribute) (*deepalert.TaskResult, error) {
	// tableName := os.Getenv("RESULT_TABLE")
	// reportID, _ := deepalert.ReportIDFromCtx(ctx)

	hostReport := deepalert.ContentHost{
		IPAddr: []string{"198.51.100.2"},
		Owner:  []string{"superman"},
	}

	newAttr := deepalert.Attribute{
		Key:   "username",
		Value: "mizutani",
		Type:  deepalert.TypeUserName,
	}

	return &deepalert.TaskResult{
		Contents:      []deepalert.ReportContent{&hostReport},
		NewAttributes: []*deepalert.Attribute{&newAttr},
	}, nil
}

func main() {
	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
		tasks, err := inspector.SNSEventToTasks(event)
		if err != nil {
			return err
		}

		return inspector.Start(inspector.Arguments{
			Context:         ctx,
			Tasks:           tasks,
			Handler:         dummyInspector,
			Author:          "dummyInspector",
			AttrQueueURL:    os.Getenv("ATTRIBUTE_QUEUE"),
			FindingQueueURL: os.Getenv("FINDING_QUEUE"),
		})
	})
}
