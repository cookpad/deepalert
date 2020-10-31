package emitter

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/errors"
)

// Handler is a function type of callback of inspector.
type Handler func(ctx context.Context, report deepalert.Report) error

// Start is a wrapper of Emitter.
func Start(handler Handler) {
	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
		return startWithSNSEvent(ctx, handler, event)
	})
}

func startWithSNSEvent(ctx context.Context, handler Handler, event events.SNSEvent) error {
	for _, record := range event.Records {
		var report deepalert.Report
		msg := record.SNS.Message
		if err := json.Unmarshal([]byte(msg), &report); err != nil {
			return errors.Wrap(err, "Fail to unmarshal report").With("msg", msg)
		}

		if err := handler(ctx, report); err != nil {
			return errors.Wrap(err, "Fail to handle report").With("report", report)
		}
	}

	return nil
}
