package main

import (
	"encoding/json"
	"time"

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
	messages, err := args.DecapSQSEvent()
	if err != nil {
		return nil, err
	}

	repo, err := args.Repository()
	if err != nil {
		return nil, err
	}
	now := time.Now()

	for _, msg := range messages {
		var ir deepalert.Note
		if err := json.Unmarshal(msg, &ir); err != nil {
			return nil, errors.Wrap(err, "Fail to unmarshal Note from SubmitNotification").With("msg", string(msg))
		}
		logger.WithField("inspectReport", ir).Debug("Handling inspect report")

		if err := repo.SaveNote(ir, now); err != nil {
			return nil, errors.Wrap(err, "Fail to save Note").With("report", ir)
		}
		logger.WithField("section", ir).Info("Saved content")
	}

	return nil, nil
}
