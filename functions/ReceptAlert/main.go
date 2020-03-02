package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/m-mizutani/deepalert"
	f "github.com/m-mizutani/deepalert/functions"
)

type lambdaArguments struct {
	Event                 events.SNSEvent
	InspectorDelayMachine string
	ReviewerDelayMachine  string
	CacheTable            string
	ReportNotification    string
	Region                string
}

func mainHandler(args lambdaArguments) error {
	svc := f.NewDataStoreService(args.CacheTable, args.Region)

	for _, msg := range f.SNStoMessages(args.Event) {
		var alert deepalert.Alert
		if err := json.Unmarshal(msg, &alert); err != nil {
			return errors.Wrap(err, "Fail to unmarshal alert from AlertNotification")
		}

		f.Logger.WithField("alert", alert).Info("Taking report")
		report, err := svc.TakeReport(alert)
		if err != nil {
			return errors.Wrapf(err, "Fail to take reportID for alert: %v", alert)
		}
		if report == nil {
			return errors.Wrapf(err, "Fatal error: no report with no error: %v", alert)
		}

		f.SetLoggerReportID(report.ID)

		f.Logger.WithFields(logrus.Fields{
			"ReportID": report.ID,
			"Status":   report.Status,
			"Error":    err,
			"AlertID":  alert.AlertID(),
		}).Info("ReportID has been retrieved")

		report.Alerts = []deepalert.Alert{alert}

		if err := svc.SaveAlertCache(report.ID, alert); err != nil {
			return errors.Wrap(err, "Fail to save alert cache")
		}

		if err := f.ExecDelayMachine(args.InspectorDelayMachine, args.Region, &report); err != nil {
			return errors.Wrap(err, "Fail to execute InspectorDelayMachine")
		}

		if report.IsNew() {
			if err := f.ExecDelayMachine(args.ReviewerDelayMachine, args.Region, &report); err != nil {
				return errors.Wrap(err, "Fail to execute ReviewerDelayMachine")
			}
		}

		if err := f.PublishSNS(args.ReportNotification, args.Region, report); err != nil {
			return errors.Wrap(err, "Fail to publish report")
		}
	}

	return nil
}

func handleRequest(ctx context.Context, event events.SNSEvent) error {
	f.SetLoggerContext(ctx, deepalert.NullReportID)
	f.Logger.WithField("event", event).Info("Start")

	args := lambdaArguments{
		Event:                 event,
		InspectorDelayMachine: os.Getenv("INSPECTOR_MACHINE"),
		ReviewerDelayMachine:  os.Getenv("REVIEW_MACHINE"),
		ReportNotification:    os.Getenv("REPORT_NOTIFICATION"),
		CacheTable:            os.Getenv("CACHE_TABLE"),
		Region:                os.Getenv("AWS_REGION"),
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
