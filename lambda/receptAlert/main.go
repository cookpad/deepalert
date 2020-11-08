package main

import (
	"time"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/errors"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/logging"
	"github.com/deepalert/deepalert/internal/usecase"
)

var logger = logging.Logger

func main() {
	handler.StartLambda(HandleRequest)
}

// HandleRequest is main logic of ReceptAlert
func HandleRequest(args *handler.Arguments) (handler.Response, error) {
	messages, err := args.DecapSQSEvent()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	for _, msg := range messages {
		var event map[string]interface{}
		if err := msg.Bind(&event); err != nil {
			return nil, errors.Wrap(err, "Failed to bind event").With("msg", msg)
		}

		var data handler.EventRecord
		if v, ok := event["Message"]; ok {
			data = []byte(v.(string))
		} else {
			data = msg
		}

		logger.WithField("message", string(msg)).Debug("Start handle alert")

		var alert deepalert.Alert
		if err := data.Bind(&alert); err != nil {
			return nil, errors.Wrap(err, "Fail to unmarshal alert").With("alert", string(msg))
		}

		_, err = usecase.HandleAlert(args, &alert, now)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}
