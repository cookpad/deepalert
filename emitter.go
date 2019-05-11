package deepalert

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
)

// EmitHandler is a function type of callback of inspector.
type EmitHandler func(ctx context.Context, report Report) error

// StartEmitter is a wrapper of Emitter.
func StartEmitter(handler EmitHandler) {
	lambda.Start(func(ctx context.Context, event events.SNSEvent) error {
		for _, record := range event.Records {
			var report Report
			msg := record.SNS.Message
			if err := json.Unmarshal([]byte(msg), &report); err != nil {
				return errors.Wrapf(err, "Fail to unmarshal report: %s", msg)
			}

			err := handler(ctx, report)
			if err != nil {
				return errors.Wrapf(err, "Fail to handle report: %v", report)
			}
		}

		return nil
	})
}
