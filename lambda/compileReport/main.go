package main

import (
	"github.com/m-mizutani/deepalert"
	"github.com/m-mizutani/deepalert/internal/errors"
	"github.com/m-mizutani/deepalert/internal/handler"
	"github.com/m-mizutani/deepalert/internal/logging"
)

func main() {
	handler.StartLambda(handleRequest)
}

func handleRequest(args *handler.Arguments) (handler.Response, error) {
	var report deepalert.Report
	if err := args.BindEvent(&report); err != nil {
		return nil, err
	}

	svc, err := args.Repository()
	if err != nil {
		return nil, err
	}

	sections, err := svc.FetchReportSection(report.ID)
	if err != nil {
		return nil, errors.Wrap(err, "FetchReportSection").With("report", report)
	}

	alerts, err := svc.FetchAlertCache(report.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "FetchAlertCache").With("report", report)
	}

	attrs, err := svc.FetchAttributeCache(report.ID)
	if err != nil {
		return nil, errors.Wrap(err, "FetchAttributeCache").With("report", report)
	}

	report.Alerts = alerts
	report.Sections = sections
	report.Attributes = attrs

	logging.Logger.WithField("report", report).Info("Compiled report")

	return &report, nil
}
