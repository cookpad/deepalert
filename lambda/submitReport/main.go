package main

import (
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		args := handler.NewArguments()
		if err := args.BindEnvVars(); err != nil {
			return nil, err
		}

		if err := handleRequest(args, event); err != nil {
			return nil, err
		}
		return nil, nil
	})
}

func handleRequest(args *handler.Arguments, event golambda.Event) error {
	var report deepalert.Report
	if err := event.Bind(&report); err != nil {
		return err
	}

	report.Status = deepalert.StatusPublished
	if report.Result.Severity == "" {
		report.Result.Severity = deepalert.SevUnclassified
	}

	repo, err := args.Repository()
	if err != nil {
		return err
	}

	logger.With("report", report).Info("Publishing report")
	if err := repo.PutReport(&report); err != nil {
		return golambda.WrapError(err, "Fail to submit report")
	}

	return nil
}
