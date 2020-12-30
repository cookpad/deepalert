package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/deepalert/deepalert/internal/errors"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/models"
	"github.com/deepalert/deepalert/internal/service"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		args := handler.NewArguments()
		if err := args.BindEnvVars(); err != nil {
			return nil, err
		}

		return handleRequest(args, event)
	})
}

func handleRequest(args *handler.Arguments, event golambda.Event) (interface{}, error) {
	var dynamoEvent events.DynamoDBEvent

	if err := event.Bind(&dynamoEvent); err != nil {
		return nil, err
	}

	for _, record := range dynamoEvent.Records {
		logger.With("event", event).Info("Recv DynamoDB event")

		if !service.IsReportStreamEvent(&record) {
			continue
		}

		var reportEntry models.ReportEntry
		if err := reportEntry.ImportDynamoRecord(&record); err != nil {
			if err != models.ErrRecordIsNotReport {
				return nil, err
			}
			continue
		}

		report, err := reportEntry.Export()
		if err != nil {
			return nil, err
		}

		logger.With("report", report).Info("Publishing report")

		if err := args.SNSService().Publish(args.ReportTopic, &report); err != nil {
			return nil, errors.Wrap(err, "Fail to publish report")
		}
	}

	return nil, nil
}
