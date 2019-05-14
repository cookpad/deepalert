package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"

	"github.com/m-mizutani/deepalert"
	f "github.com/m-mizutani/deepalert/functions"
)

func handleRequest(ctx context.Context, event events.SNSEvent) error {
	f.SetLoggerContext(ctx, "ErrorHandler", deepalert.NullReportID)
	f.Logger.WithField("event", event).Info("Catch error from ErrorNotification topic")

	for _, record := range event.Records {
		var reqEvent interface{}
		if err := json.Unmarshal([]byte(record.SNS.Message), &reqEvent); err != nil {
			f.Logger.WithError(err).WithField("record", record).Fatal("Fail to unmarshal requested event")
			return err
		}

		f.Logger.WithFields(logrus.Fields{
			"ErrorAttributes": record.SNS.MessageAttributes,
			"OriginalRequest": reqEvent,
		}).Error("Error from Lambda")
	}
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
