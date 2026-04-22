package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cookpad/deepalert/emitter"
)

func main() {

	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
		reports, err := emitter.SNSEventToReport(event)
		if err != nil {
			return err
		}

		for _, report := range reports {
			log.Println(report.Result.Severity)
			// Or do appropriate action according to report content and severity
		}

		return nil
	})
}
