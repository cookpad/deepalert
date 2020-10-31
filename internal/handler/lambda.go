package handler

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/deepalert/deepalert/internal/errors"
	"github.com/deepalert/deepalert/internal/logging"

	"github.com/sirupsen/logrus"
)

// Logger is common logger gateway
var Logger = logging.Logger

// Handler has main logic of the lambda function
type Handler func(*Arguments) (Response, error)

// Response is return object of Handler required for StepFunctions.
type Response interface{}

// StartLambda initialize AWS Lambda and invokes handler
func StartLambda(handler Handler) {
	Logger.SetLevel(logrus.InfoLevel)
	Logger.SetFormatter(&logrus.JSONFormatter{})

	lambda.Start(func(ctx context.Context, event interface{}) (interface{}, error) {
		defer errors.Flush()

		args := newArguments()
		if err := args.BindEnvVars(); err != nil {
			return nil, err
		}

		if args.LogLevel != "" {
			logging.SetLogLevel(args.LogLevel)
		}

		Logger.WithFields(logrus.Fields{"args": args, "event": event}).Debug("Start handler")
		args.Event = event

		out, err := handler(args)
		if err != nil {
			fields := logrus.Fields{
				"args":  args,
				"event": event,
				"error": err,
				"trace": fmt.Sprintf("%+v", err),
			}

			if daErr, ok := err.(*errors.Error); ok {
				fields["values"] = daErr.Values
			}
			Logger.WithFields(fields).Error("Failed Handler")
			return nil, err
		}

		return out, nil
	})
}
