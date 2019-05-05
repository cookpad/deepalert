package deepalert_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/m-mizutani/deepalert"
	"github.com/m-mizutani/deepalert/test"
	gp "github.com/m-mizutani/generalprobe"
)

func TestNormalWorkFlow(t *testing.T) {
	cfg := test.LoadTestConfig(".")
	alertKey := uuid.New().String()

	alert := deepalert.Alert{
		Detector:  "test",
		RuleName:  "TestRule",
		AlertKey:  alertKey,
		Timestamp: time.Now().UTC(),
		Attributes: []deepalert.Attribute{
			{
				Type:    deepalert.TypeIPAddr,
				Key:     "test value",
				Value:   "192.168.0.1",
				Context: []deepalert.AttrContext{deepalert.CtxLocal},
			},
		},
	}
	alertMsg, err := json.Marshal(alert)
	require.NoError(t, err)

	var reportID string

	g := gp.New(cfg.Region, cfg.StackName)
	g.AddScenes([]gp.Scene{
		// Send request
		g.PublishSnsMessage(g.LogicalID("AlertNotification"), alertMsg),
		g.GetLambdaLogs(g.LogicalID("ReceptAlert"), func(log gp.CloudWatchLog) bool {
			assert.Contains(t, log, alertKey)
			return true
		}).Filter(alertKey),
		g.GetDynamoRecord(g.LogicalID("CacheTable"), func(table dynamo.Table) bool {
			var entry struct {
				ReportID string `dynamo:"report_id"`
			}

			alertID := "alertmap/" + alert.AlertID()
			err := table.Get("pk", alertID).Range("sk", dynamo.Equal, "Fixed").One(&entry)
			if err != nil {
				return false
			}
			require.NotEmpty(t, entry.ReportID)
			reportID = entry.ReportID
			return true
		}),
		g.GetLambdaLogs(g.LogicalID("DispatchInspection"), func(log gp.CloudWatchLog) bool {
			return log.Contains(reportID)
		}),
		g.GetLambdaLogs(g.LogicalID("SubmitReport"), func(log gp.CloudWatchLog) bool {
			return log.Contains(reportID)
		}),

		g.GetDynamoRecord(g.LogicalID("CacheTable"), func(table dynamo.Table) bool {
			var contents []struct {
				Data []byte `dynamo:"data"`
			}

			pk := "content/" + reportID

			if err := table.Get("pk", pk).All(&contents); err != nil {
				return false
			}

			require.True(t, len(contents) > 0)
			require.NotEmpty(t, contents[0].Data)
			return true
		}),

		g.Pause(10),

		g.GetLambdaLogs(g.LogicalID("CompileReport"), func(log gp.CloudWatchLog) bool {
			return log.Contains(reportID)
		}),
		g.GetLambdaLogs(g.LogicalID("PublishReport"), func(log gp.CloudWatchLog) bool {
			return log.Contains(reportID)
		}),
	})

	err = g.Run()
	require.NoError(t, err)
}
