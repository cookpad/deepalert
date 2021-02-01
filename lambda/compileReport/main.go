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

	compiledReport, err := svc.GetReport(report.ID)
	if err != nil {
		return nil, err
	}
	golambda.Logger.With("report", compiledReport).Info("Compiled report")

	return compiledReport, nil
}
