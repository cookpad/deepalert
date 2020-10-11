package main

import (
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/errors"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/logging"
)

var logger = logging.Logger

func main() {
	handler.StartLambda(handleRequest)
}

func handleRequest(args *handler.Arguments) (handler.Response, error) {
	var report deepalert.Report
	if err := args.BindEvent(&report); err != nil {
		return nil, err
	}

	report.Status = deepalert.StatusPublished
	if report.Result.Severity == "" {
		report.Result.Severity = deepalert.SevUnclassified
	}

	repo, err := args.Repository()
	if err != nil {
		return nil, err
	}

	logger.WithField("report", report).Info("Publishing report")
	if err := repo.PutReport(&report); err != nil {
		return nil, errors.Wrap(err, "Fail to submit report")
	}

	return nil, nil
}
