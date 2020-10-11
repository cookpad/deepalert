package main

import (
	"time"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/errors"
	"github.com/deepalert/deepalert/internal/handler"
)

func main() {
	handler.StartLambda(handleRequest)
}

func handleRequest(args *handler.Arguments) (handler.Response, error) {
	var report deepalert.Report
	if err := args.BindEvent(&report); err != nil {
		return nil, err
	}

	now := time.Now()
	repo, err := args.Repository()
	if err != nil {
		return nil, err
	}
	snsSvc := args.SNSService()
	for _, alert := range report.Alerts {
		for _, attr := range alert.Attributes {
			sendable, err := repo.PutAttributeCache(report.ID, attr, now)
			if err != nil {
				return nil, errors.Wrapf(err, "Fail to manage attribute cache: %v", attr)
			}

			if !sendable {
				continue
			}

			if attr.Timestamp == nil {
				attr.Timestamp = &alert.Timestamp
			}

			task := deepalert.Task{
				ReportID:  report.ID,
				Attribute: &attr,
			}

			if err := snsSvc.Publish(args.TaskTopic, &task); err != nil {
				return nil, errors.Wrapf(err, "Fail to publihsh task notification: %v", task)
			}
		}
	}

	return nil, nil
}
