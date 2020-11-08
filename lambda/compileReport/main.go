package main

import (
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/errors"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/logging"
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

	sections, err := svc.FetchSection(report.ID)
	if err != nil {
		return nil, errors.Wrap(err, "FetchSection").With("report", report)
	}

	alerts, err := svc.FetchAlertCache(report.ID)
	if err != nil {
		return nil, errors.Wrap(err, "FetchAlertCache").With("report", report)
	}

	attrs, err := svc.FetchAttributeCache(report.ID)
	if err != nil {
		return nil, errors.Wrap(err, "FetchAttributeCache").With("report", report)
	}

	report.Alerts = alerts
	report.Attributes = attrs
	report.Sections = sections

	logging.Logger.WithField("report", report).Info("Compiled report")

	return &report, nil
}
