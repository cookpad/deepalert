package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/m-mizutani/deepalert"
	f "github.com/m-mizutani/deepalert/functions"
)

func handleRequest(ctx context.Context, event interface{}) (deepalert.ReportResult, error) {
	f.SetLoggerContext(ctx, deepalert.NullReportID)
	f.Logger.WithField("event", event).Info("Start")

	return deepalert.ReportResult{
		Severity: deepalert.SevUnclassified,
		Reason:   "I'm novice",
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
