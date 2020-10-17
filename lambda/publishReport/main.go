package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/deepalert/deepalert/internal/errors"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/logging"
	"github.com/deepalert/deepalert/internal/models"
	"github.com/deepalert/deepalert/internal/service"
)

var logger = logging.Logger

func main() {
	handler.StartLambda(handleRequest)
}

func handleRequest(args *handler.Arguments) (handler.Response, error) {
	var event events.DynamoDBEvent

	if err := args.BindEvent(&event); err != nil {
		return nil, err
	}

	for _, record := range event.Records {
		logger.WithField("event", event).Info("Recv DynamoDB event")

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

		logger.WithField("report", report).Info("Publishing report")

		if err := args.SNSService().Publish(args.ReportTopic, &report); err != nil {
			return nil, errors.Wrap(err, "Fail to publish report")
		}
	}

	return nil, nil
}
