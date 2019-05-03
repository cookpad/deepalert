package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"

	"github.com/m-mizutani/deepalert"
	f "github.com/m-mizutani/deepalert/functions"
)

type lambdaArguments struct {
	Report           deepalert.Report
	TaskNotification string
	Region           string
}

func mainHandler(args lambdaArguments) error {
	for _, attr := range args.Report.Alert.Attributes {
		task := deepalert.Task{
			ReportID:  args.Report.ID,
			Attribute: attr,
		}

		if err := f.PublishSNS(args.TaskNotification, args.Region, &task); err != nil {
			return errors.Wrapf(err, "Fail to publihsh task notification: %v", task)
		}
	}

	return nil
}

func handleRequest(ctx context.Context, report deepalert.Report) error {
	f.SetLoggerContext(ctx)
	f.Logger.WithField("report", report).Info("Start")

	args := lambdaArguments{
		Report:           report,
		TaskNotification: os.Getenv("TASK_NOTIFICATION"),
		Region:           os.Getenv("AWS_REGION"),
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
