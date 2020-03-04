package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"

	"github.com/m-mizutani/deepalert"
	f "github.com/m-mizutani/deepalert/internal"
)

type lambdaArguments struct {
	Report           deepalert.Report
	TaskNotification string
	CacheTable       string
	Region           string
}

func mainHandler(args lambdaArguments) error {
	svc := f.NewDataStoreService(args.CacheTable, args.Region)

	for _, alert := range args.Report.Alerts {
		for _, attr := range alert.Attributes {
			sendable, err := svc.PutAttributeCache(args.Report.ID, attr)
			if err != nil {
				return errors.Wrapf(err, "Fail to manage attribute cache: %v", attr)
			}

			if !sendable {
				continue
			}

			if attr.Timestamp == nil {
				attr.Timestamp = &alert.Timestamp
			}

			task := deepalert.Task{
				ReportID:  args.Report.ID,
				Attribute: attr,
			}

			if err := f.PublishSNS(args.TaskNotification, args.Region, &task); err != nil {
				return errors.Wrapf(err, "Fail to publihsh task notification: %v", task)
			}
		}
	}

	return nil
}

func handleRequest(ctx context.Context, report deepalert.Report) error {
	f.SetLoggerContext(ctx, report.ID)
	f.Logger.WithField("report", report).Info("Start")

	args := lambdaArguments{
		Report:           report,
		TaskNotification: os.Getenv("TASK_NOTIFICATION"),
		Region:           os.Getenv("AWS_REGION"),
		CacheTable:       os.Getenv("CACHE_TABLE"),
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
