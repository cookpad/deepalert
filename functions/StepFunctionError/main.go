package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	da "github.com/m-mizutani/deepalert/functions"
)

func handleRequest(ctx context.Context, event interface{}) error {
	da.SetLoggerContext(ctx)
	da.Logger.WithField("event", event).Info("Start")
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
