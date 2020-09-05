package main

import (
	"github.com/m-mizutani/deepalert"
	"github.com/m-mizutani/deepalert/internal/errors"
	"github.com/m-mizutani/deepalert/internal/handler"
	"github.com/m-mizutani/deepalert/internal/logging"
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
	snsSvc := args.SNSService()

	report.Status = deepalert.StatusPublished
	if report.Result.Severity == "" {
		report.Result.Severity = deepalert.SevUnclassified
	}

	logger.WithField("report", report).Info("Publishing report")

	if err := snsSvc.Publish(args.ReportTopic, &report); err != nil {
		return nil, errors.Wrap(err, "Fail to publish report")
	}

	return nil, nil
}
