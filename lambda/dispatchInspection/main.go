package main

import (
	"time"

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
				return nil, golambda.WrapError(err, "Fail to manage attribute cache").With("attr", attr)
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
				return nil, golambda.WrapError(err, "Fail to publihsh task notification").With("task", task)
			}
		}
	}

	return nil, nil
}
