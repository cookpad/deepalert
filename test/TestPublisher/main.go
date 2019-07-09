package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/deepalert"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func handleRequest(ctx context.Context, event events.SNSEvent) error {
	logger.WithField("event", event).Debug("Start")

	for _, record := range event.Records {
		raw := []byte(record.SNS.Message)
		var report deepalert.Report
		if err := json.Unmarshal(raw, &report); err != nil {
			logger.WithError(err).Error("Fail to unmarshal message")
		}

		logger.WithField("report", report).Info("Got report")
	}

	return nil
}

func main() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	lambda.Start(handleRequest)
}
