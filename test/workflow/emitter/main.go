package main

import (
	"context"
	"os"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/emitter"
	"github.com/deepalert/deepalert/test/workflow"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func emit(ctx context.Context, report deepalert.Report) error {
	logger.WithField("report", report).Debug("Start emitter")

	repo := workflow.NewRepository(os.Getenv("AWS_REGION"), os.Getenv("RESULT_TABLE"))

	if err := repo.PutEmitterResult(report.ID); err != nil {
		return err
	}

	return nil
}

func main() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	emitter.Start(emit)
}
