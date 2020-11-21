package emitter_test

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/emitter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmitter(t *testing.T) {
	t.Run("Valid SNS event", func(tt *testing.T) {
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

		reports, err := emitter.SNSEventToReport(event)
		require.NoError(t, err)
		require.Equal(tt, 1, len(reports))
		assert.Equal(tt, deepalert.ReportID("t1"), reports[0].ID)
	})

	t.Run("Invalid SNS event (report)", func(tt *testing.T) {
		event := events.SNSEvent{
			Records: []events.SNSEventRecord{
				{
					SNS: events.SNSEntity{
						Message: `{"id":"hoge"`, // missing a bracket of end
					},
				},
			},
		}

		reports, err := emitter.SNSEventToReport(event)
		require.Error(t, err)
		assert.Equal(t, 0, len(reports))
	})
}
