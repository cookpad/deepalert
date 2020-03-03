package main

import (
	"context"
	"os"

	"github.com/m-mizutani/deepalert"
	"github.com/m-mizutani/deepalert/remote"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func emitter(ctx context.Context, report deepalert.Report) error {
	logger.WithField("report", report).Debug("Start emitter")

	repo := remote.NewRepository(os.Getenv("AWS_REGION"), os.Getenv("RESULT_TABLE"))

	if err := repo.PutEmitterResult(report.ID); err != nil {
		return err
	}

	return nil
}

func main() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	deepalert.StartEmitter(emitter)
}
