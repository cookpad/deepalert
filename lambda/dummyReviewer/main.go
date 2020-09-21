package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/logging"
)

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, event interface{}) (deepalert.ReportResult, error) {
	logging.Logger.WithField("event", event).Info("Start")

	return deepalert.ReportResult{
		Severity: deepalert.SevUnclassified,
		Reason:   "I'm novice",
	}, nil
}
