package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"

	"github.com/m-mizutani/deepalert"
	f "github.com/m-mizutani/deepalert/internal"
)

type lambdaArguments struct {
	Event     events.SQSEvent
	TableName string
	Region    string
}

func mainHandler(args lambdaArguments) error {
	svc := f.NewDataStoreService(args.TableName, args.Region)

	for _, msg := range f.SQStoMessage(args.Event) {
		var section deepalert.ReportSection
		if err := json.Unmarshal(msg, &section); err != nil {
			return errors.Wrapf(err, "Fail to unmarshal ReportContent from SubmitNotification: %s", string(msg))
		}

		if err := svc.SaveReportSection(section); err != nil {
			return errors.Wrapf(err, "Fail to save ReportContent: %v", section)
		}
		f.Logger.WithField("section", section).Info("Saved content")
	}

	return nil
}

func handleRequest(ctx context.Context, event events.SQSEvent) error {
	f.SetLoggerContext(ctx, deepalert.ReportID(""))
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
