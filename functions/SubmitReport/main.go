package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"

	"github.com/m-mizutani/deepalert"
	f "github.com/m-mizutani/deepalert/functions"
)

type lambdaArguments struct {
	Event     events.SNSEvent
	TableName string
	Region    string
}

func mainHandler(args lambdaArguments) error {
	svc := f.NewDataStoreService(args.TableName, args.Region)

	for _, msg := range f.SNStoMessages(args.Event) {
		var content deepalert.ReportContent
		if err := json.Unmarshal(msg, &content); err != nil {
			return errors.Wrapf(err, "Fail to unmarshal ReportContent from SubmitNotification: %s", string(msg))
		}

		if err := svc.SaveReportContent(content); err != nil {
			return errors.Wrapf(err, "Fail to save ReportContent: %v", content)
		}
	}

	return nil
}

func handleRequest(ctx context.Context, event events.SNSEvent) error {
	f.SetLoggerContext(ctx, "SubmitReport", deepalert.ReportID(""))
	f.Logger.WithField("event", event).Info("Start")

	args := lambdaArguments{
		Event:     event,
		TableName: os.Getenv("CACHE_TABLE"),
		Region:    os.Getenv("AWS_REGION"),
	}

	if err := mainHandler(args); err != nil {
		f.Logger.WithError(err).Error("Fail")
		return err
	}

	return nil
}

func main() {
	lambda.Start(handleRequest)
}
