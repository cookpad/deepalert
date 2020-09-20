package emitter

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/m-mizutani/deepalert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmitter(t *testing.T) {
	t.Run("Valid SNS event", func(tt *testing.T) {
		var reports []*deepalert.Report
		ctx := context.Background()
		handler := func(ctx context.Context, report deepalert.Report) error {
			reports = append(reports, &report)
			return nil
		}
		report := deepalert.Report{
			ID: deepalert.ReportID("t1"),
			Result: deepalert.ReportResult{
				Severity: deepalert.SevUrgent,
			},
		}

		raw, err := json.Marshal(report)
		require.NoError(tt, err)
		event := events.SNSEvent{
			Records: []events.SNSEventRecord{
				{
					SNS: events.SNSEntity{
						Message: string(raw),
					},
				},
			},
		}

		err = startWithSNSEvent(ctx, handler, event)
		require.NoError(tt, err)
		require.Equal(tt, 1, len(reports))
		assert.Equal(tt, deepalert.ReportID("t1"), reports[0].ID)
	})

	t.Run("Invalid SNS event (report)", func(tt *testing.T) {
		var reports []*deepalert.Report
		ctx := context.Background()
		handler := func(ctx context.Context, report deepalert.Report) error {
			reports = append(reports, &report)
			return nil
		}

		event := events.SNSEvent{
			Records: []events.SNSEventRecord{
				{
					SNS: events.SNSEntity{
						Message: `{"id":"hoge"`, // missing a bracket of end
					},
				},
			},
		}

		err := startWithSNSEvent(ctx, handler, event)
		require.Error(tt, err)
		assert.Equal(tt, 0, len(reports))
	})
}
