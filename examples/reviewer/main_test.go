package main

import (
	"context"
	"testing"

	"github.com/deepalert/deepalert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateAlert(t *testing.T) {
	t.Run("As SevSafe with your_alert_rule_id amd YOUR_COMPANY PC", func(tt *testing.T) {
		report := deepalert.Report{
			Alerts: []*deepalert.Alert{{RuleID: "your_alert_rule_id"}},
			Sections: []*deepalert.Section{
				{
					Hosts: []*deepalert.ContentHost{
						{Owner: []string{"YOUR_COMPANY"}},
					},
				},
			},
		}

		result, err := evaluate(context.Background(), report)
		require.NoError(tt, err)
		require.NotNil(tt, result)
		assert.NotEqual(tt, "", result.Reason)
		assert.Equal(tt, deepalert.SevSafe, result.Severity)
	})

	t.Run("Return nil for an alert with your_alert_rule_id but not YOUR_COMPANY PC", func(tt *testing.T) {
		report := deepalert.Report{
			Alerts: []*deepalert.Alert{{RuleID: "your_alert_rule_id"}},
			Sections: []*deepalert.Section{
				{
					Hosts: []*deepalert.ContentHost{},
				},
			},
		}

		result, err := evaluate(context.Background(), report)
		require.NoError(tt, err)
		require.Nil(tt, result)
	})

	t.Run("Return nil for an alert with YOUR_COMPANY PC but not your_alert_rule_id", func(tt *testing.T) {
		report := deepalert.Report{
			Alerts: []*deepalert.Alert{{RuleID: "some_rule"}},
			Sections: []*deepalert.Section{
				{
					Hosts: []*deepalert.ContentHost{},
				},
			},
		}

		result, err := evaluate(context.Background(), report)
		require.NoError(tt, err)
		require.Nil(tt, result)
	})
}
