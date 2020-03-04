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
	Report    deepalert.Report
	TableName string
	Region    string
}

func mainHandler(args lambdaArguments) (*deepalert.Report, error) {
	svc := f.NewDataStoreService(args.TableName, args.Region)

	sections, err := svc.FetchReportSection(args.Report.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "Fail to fetch set of ReportContent: %v", args.Report)
	}

	alerts, err := svc.FetchAlertCache(args.Report.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "Fail to fetch alert caches: %v", args.Report)
	}

	args.Report.Alerts = alerts
	args.Report.Sections = sections

	return &args.Report, nil
}

func handleRequest(ctx context.Context, report deepalert.Report) (deepalert.Report, error) {
	f.SetLoggerContext(ctx, report.ID)
	f.Logger.WithField("report", report).Info("Start")

	args := lambdaArguments{
		Report:    report,
		TableName: os.Getenv("CACHE_TABLE"),
		Region:    os.Getenv("AWS_REGION"),
	}

	compiledReport, err := mainHandler(args)
	if err != nil || compiledReport == nil {
		f.Logger.WithError(err).Error("Fail")
		return report, err
	}

	return *compiledReport, nil
}

func main() {
	lambda.Start(handleRequest)
}
