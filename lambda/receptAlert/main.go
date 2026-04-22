package main

import (
	"encoding/json"
	"time"

	"github.com/cookpad/deepalert"
	"github.com/cookpad/deepalert/internal/handler"
	"github.com/cookpad/deepalert/internal/usecase"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		args := handler.NewArguments()
		if err := args.BindEnvVars(); err != nil {
			return nil, err
		}

		return HandleRequest(args, event)
	})
}

// HandleRequest is main logic of ReceptAlert
func HandleRequest(args *handler.Arguments, event golambda.Event) (interface{}, error) {
	messages, err := event.DecapSQSBody()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	for _, msg := range messages {
		var snsWrapper struct {
			Message string `json:"Message"`
		}

		var data []byte
		if err := msg.Bind(&snsWrapper); err == nil && snsWrapper.Message != "" {
			data = []byte(snsWrapper.Message)
		} else {
			data = msg
		}

		logger.With("data", string(data)).Debug("Start handle alert")

		var alert deepalert.Alert
		if err := json.Unmarshal(data, &alert); err != nil {
			return nil, golambda.WrapError(err, "Fail to unmarshal alert").With("alert", string(msg))
		}

		_, err = usecase.HandleAlert(args, &alert, now)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}
