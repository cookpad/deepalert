package main

import (
	"encoding/json"
	"time"

	"github.com/m-mizutani/golambda"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/handler"
)

var logger = golambda.Logger

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
	snsSvc := args.SNSService()
	repo, err := args.Repository()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	sqsMessages, err := event.DecapSQSBody()
	if err != nil {
		return nil, err
	}

	for _, msg := range sqsMessages {
		var reportedAttr deepalert.ReportAttribute
		if err := json.Unmarshal(msg, &reportedAttr); err != nil {
			return nil, golambda.WrapError(err, "Unmarshal ReportAttribute").With("msg", string(msg))
		}

		logger.With("reportedAttr", reportedAttr).Info("unmarshaled reported attribute")

		for _, attr := range reportedAttr.Attributes {
			sendable, err := repo.PutAttributeCache(reportedAttr.ReportID, *attr, now)
			if err != nil {
				return nil, golambda.WrapError(err, "Fail to manage attribute cache").With("attr", attr)
			}

			logger.With("sendable", sendable).With("attr", attr).Info("attribute")
			if !sendable {
				continue
			}

			task := deepalert.Task{
				ReportID:  reportedAttr.ReportID,
				Attribute: attr,
			}

			if err := snsSvc.Publish(args.TaskTopic, &task); err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}
