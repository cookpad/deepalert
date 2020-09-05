package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/deepalert/internal/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.Logger

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, event events.SNSEvent) error {
	logger.WithField("event", event).Info("Catch error from ErrorNotification topic")

	for _, record := range event.Records {
		var reqEvent interface{}
		if err := json.Unmarshal([]byte(record.SNS.Message), &reqEvent); err != nil {
			logger.WithError(err).WithField("record", record).Fatal("Fail to unmarshal requested event")
			return err
		}

		logger.WithFields(logrus.Fields{
			"ErrorAttributes": record.SNS.MessageAttributes,
			"OriginalRequest": reqEvent,
		}).Error("Error from Lambda")
	}
	return nil
}
