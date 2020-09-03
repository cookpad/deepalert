package handler

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/deepalert/internal/errors"

	"github.com/sirupsen/logrus"
)

// Logger is common logger gateway
var Logger = logrus.New()

// Handler has main logic of the lambda function
type Handler func(Arguments) (interface{}, *errors.Error)

// StartLambda initialize AWS Lambda and invokes handler
func StartLambda(handler Handler) {
	Logger.SetLevel(logrus.InfoLevel)
	Logger.SetFormatter(&logrus.JSONFormatter{})

	lambda.Start(func(ctx context.Context, event interface{}) (interface{}, error) {
		defer errors.Flush()

		var args Arguments
		if err := args.BindEnvVars(); err != nil {
			return nil, err
		}

		/*
			if args.LogLevel != "" {
				internal.SetLogLevel(args.LogLevel)
			}
		*/

		Logger.WithFields(logrus.Fields{"args": args, "event": event}).Debug("Start handler")
		args.Event = event

		out, err := handler(args)
		if err != nil {
			Logger.WithFields(logrus.Fields{
				"args":  args,
				"event": event,
				"error": err,
			}).Error("Failed Handler")
			return nil, err
		}

		return out, nil
	})
}
