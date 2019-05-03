package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/m-mizutani/deepalert"
	f "github.com/m-mizutani/deepalert/functions"
)

func handleRequest(ctx context.Context, event interface{}) error {
	f.SetLoggerContext(ctx, "StepFunctionError", deepalert.ReportID(""))
	f.Logger.WithField("event", event).Info("Catch StepFunction error")
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
