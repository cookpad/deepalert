package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/logging"
)

var logger = logging.Logger

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, event interface{}) (deepalert.ReportResult, error) {
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	logger.WithField("event", event).Info("Start")

	return deepalert.ReportResult{
		Severity: deepalert.SevUnclassified,
		Reason:   "I'm novice",
	}, nil
}
