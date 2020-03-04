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
	Report             deepalert.Report
	ReportNotification string
	Region             string
}

func mainHandler(args lambdaArguments) error {
	args.Report.Status = deepalert.StatusPublished
	if err := f.PublishSNS(args.ReportNotification, args.Region, &args.Report); err != nil {
		return errors.Wrap(err, "Fail to publish report")
	}

	return nil
}

func handleRequest(ctx context.Context, report deepalert.Report) error {
	f.SetLoggerContext(ctx, report.ID)
	f.Logger.WithField("report", report).Info("Start")

	args := lambdaArguments{
		Report:             report,
		ReportNotification: os.Getenv("REPORT_NOTIFICATION"),
		Region:             os.Getenv("AWS_REGION"),
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
