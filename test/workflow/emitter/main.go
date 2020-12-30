package main

import (
	"os"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/test/workflow"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		messages, err := event.DecapSNSMessage()
		if err != nil {
			return nil, err
		}

		awsRegion := os.Getenv("AWS_REGION")
		tableName := os.Getenv("RESULT_TABLE")
		repo, err := workflow.NewRepository(awsRegion, tableName)
		if err != nil {
			return nil, golambda.WrapError(err).With("region", awsRegion).With("table", tableName)
		}

		for _, msg := range messages {
			var report deepalert.Report
			if err := msg.Bind(&report); err != nil {
				return nil, err
			}

			logger.With("report", report).Debug("Start emitter")
			if err := repo.PutEmitterResult(&report); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})
}
