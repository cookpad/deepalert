package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/emitter"
)

func handler(ctx context.Context, report deepalert.Report) error {
	log.Println(report.Result.Severity)
	// Or do appropriate action according to report content and severity

	return nil
}

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
