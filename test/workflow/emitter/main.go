package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/emitter"
	"github.com/deepalert/deepalert/test/workflow"
	"github.com/m-mizutani/golambda"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func emit(ctx context.Context, report deepalert.Report) error {
	logger.WithField("report", report).Debug("Start emitter")

	repo, err := workflow.NewRepository(os.Getenv("AWS_REGION"), os.Getenv("RESULT_TABLE"))
	if err != nil {
		return golambda.WrapError(err, "Failed workflow.NewRepository")
	}

	if err := repo.PutEmitterResult(&report); err != nil {
		return err
	}

	return nil
}

func main() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
		reports, err := emitter.SNSEventToReport(event)
		if err != nil {
			return err
		}

		logger.WithField("report", reports).Debug("Start emitter")

		repo, err := workflow.NewRepository(os.Getenv("AWS_REGION"), os.Getenv("RESULT_TABLE"))
		if err != nil {
			return golambda.WrapError(err, "Failed workflow.NewRepository")
		}

		for _, report := range reports {
			if err := repo.PutEmitterResult(report); err != nil {
				return err
			}
		}

		return nil
	})
}
