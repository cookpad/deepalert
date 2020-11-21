package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/inspector"
)

func lookupHostname(value string) *string {
	response := "resolved.hostname.example.com" // It's jsut example, OK?
	return &response
}

// Handler is exported for main_test
func Handler(ctx context.Context, attr deepalert.Attribute) (*deepalert.TaskResult, error) {
	// Check type of the attribute
	if attr.Type != deepalert.TypeIPAddr {
		return nil, nil
	}

	// Example
	resp := lookupHostname(attr.Value)
	if resp == nil {
		return nil, nil
	}

	result := deepalert.TaskResult{
		Contents: []deepalert.ReportContent{
			&deepalert.ContentHost{
				HostName: []string{*resp},
			},
		},
	}

	return &result, nil
}

func main() {
	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
		tasks, err := inspector.SNSEventToTasks(event)
		if err != nil {
			return err
		}

		inspector.Start(inspector.Arguments{
			Context:         ctx,
			Tasks:           tasks,
			Handler:         Handler,
			Author:          "testInspector",
			ContentQueueURL: os.Getenv("CONTENT_QUEUE"),
			AttrQueueURL:    os.Getenv("ATTRIBUTE_QUEUE"),
		})

		return nil
	})
}
