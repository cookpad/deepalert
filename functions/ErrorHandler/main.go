package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	f "github.com/m-mizutani/deepalert/functions"
)

func handleRequest(ctx context.Context, event interface{}) error {
	f.SetLoggerContext(ctx)
	f.Logger.WithField("event", event).Info("Catch error from ErrorNotification topic")
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
