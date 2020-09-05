package main

import (
	"encoding/json"
	"time"

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
	messages, err := args.DecapSQSEvent()
	if err != nil {
		return nil, err
	}
	repo := args.Repository()
	now := time.Now()

	for _, msg := range messages {
		var section deepalert.ReportSection
		if err := json.Unmarshal(msg, &section); err != nil {
			return nil, errors.Wrapf(err, "Fail to unmarshal ReportContent from SubmitNotification").With("%s", string(msg))
		}

		if err := repo.SaveReportSection(section, now); err != nil {
			return nil, errors.Wrapf(err, "Fail to save ReportContent: %v", section)
		}
		logger.WithField("section", section).Info("Saved content")
	}

	return nil, nil
}
