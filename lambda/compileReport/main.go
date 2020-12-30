package main

import (
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/m-mizutani/golambda"
)

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		args := handler.NewArguments()
		if err := args.BindEnvVars(); err != nil {
			return nil, err
		}

		return handleRequest(args, event)
	})
}

func handleRequest(args *handler.Arguments, event golambda.Event) (interface{}, error) {
	var report deepalert.Report
	if err := event.Bind(&report); err != nil {
		return nil, err
	}

	svc, err := args.Repository()
	if err != nil {
		return nil, err
	}

	sections, err := svc.FetchSection(report.ID)
	if err != nil {
		return nil, golambda.WrapError(err, "FetchSection").With("report", report)
	}

	alerts, err := svc.FetchAlertCache(report.ID)
	if err != nil {
		return nil, golambda.WrapError(err, "FetchAlertCache").With("report", report)
	}

	attrs, err := svc.FetchAttributeCache(report.ID)
	if err != nil {
		return nil, golambda.WrapError(err, "FetchAttributeCache").With("report", report)
	}

	report.Alerts = alerts
	report.Attributes = attrs
	report.Sections = sections

	golambda.Logger.With("report", report).Info("Compiled report")

	return &report, nil
}
