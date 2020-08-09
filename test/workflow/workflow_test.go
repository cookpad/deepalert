package workflow_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/m-mizutani/deepalert"
	gp "github.com/m-mizutani/generalprobe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getReportID(alertID string, table dynamo.Table) (*string, error) {
	var entry struct {
		ReportID string `dynamo:"report_id"`
	}

	pk := "alertmap/" + alertID
	err := table.Get("pk", pk).Range("sk", dynamo.Equal, "Fixed").One(&entry)
	if err != nil {
		return nil, nil
	}

	reportID := entry.ReportID
	logger.WithField("ReportID", reportID).Debug("Got reportID")

	return &reportID, nil
}

type attrCache struct {
	Key   string `dynamo:"attr_key"`
	Value string `dynamo:"attr_value"`
	Type  string `dynamo:"attr_type"`
}
type attrCaches []attrCache

func (x attrCaches) lookup(key string) *attrCache {
	for i := 0; i < len(x); i++ {
		if x[i].Key == key {
			return &x[i]
		}
	}
	return nil
}

func TestNormalWorkflow(t *testing.T) {
	alertKey := uuid.New().String()

	alert := deepalert.Alert{
		Detector:  "test",
		RuleName:  "TestRule",
		RuleID:    "xxx",
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

	var reportID string

	playbook := []gp.Scene{
		// Send request
		gp.PublishSnsMessage(gp.LogicalID("AlertNotification"), alertMsg),
		gp.GetDynamoRecord(gp.LogicalID("CacheTable"), func(table dynamo.Table) bool {
			id, err := getReportID(alert.AlertID(), table)
			require.NoError(t, err)
			if id == nil {
				return false
			}

			reportID = *id
			logger.WithField("ReportID", reportID).Debug("Got reportID")
			return true
		}),
		gp.GetDynamoRecord(gp.LogicalID("CacheTable"), func(table dynamo.Table) bool {
			var caches attrCaches

			pk := "attribute/" + reportID
			if err := table.Get("pk", pk).All(&caches); err != nil {
				return false
			}

			if len(caches) != 2 {
				return false
			}

			var a1, a2 int
			if caches[0].Type == "ipaddr" {
				a1, a2 = 0, 1
			} else {
				a1, a2 = 1, 0
			}

			assert.Equal(t, "192.168.0.1", caches[a1].Value)
			assert.Equal(t, "mizutani", caches[a2].Value)
			assert.Equal(t, "username", caches[a2].Type)
			return true
		}),

		gp.GetDynamoRecord(gp.LogicalID("CacheTable"), func(table dynamo.Table) bool {
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

		gp.Pause(10),

		gp.GetDynamoRecord(gp.Arn(testCfg.ResultTableArn), func(table dynamo.Table) bool {
			var contents []struct {
				Timestamp time.Time `dynamo:"timestamp"`
			}

			pk := "emitter/" + reportID

			if err := table.Get("pk", pk).All(&contents); err != nil {
				require.Equal(t, dynamo.ErrNotFound, err)
				return false
			}

			require.NotEqual(t, 0, len(contents))
			return true
		}),
	}

	err = gp.New(testCfg.Region, testCfg.DeepAlertStackName).Play(playbook)
	require.NoError(t, err)
}

func TestNormalAggregation(t *testing.T) {
	alertKey := uuid.New().String()
	attr1 := uuid.New().String()
	attr2 := uuid.New().String()

	alert := deepalert.Alert{
		Detector:  "test",
		RuleName:  "TestRule",
		RuleID:    "yyy",
		AlertKey:  alertKey,
		Timestamp: time.Now().UTC(),
		Attributes: []deepalert.Attribute{
			{
				Type:    deepalert.TypeUserName,
				Key:     "blue",
				Value:   attr1,
				Context: []deepalert.AttrContext{deepalert.CtxLocal},
			},
		},
	}
	alertMsg1, err := json.Marshal(alert)
	alert.Attributes = []deepalert.Attribute{
		{
			Type:    deepalert.TypeUserName,
			Key:     "orange",
			Value:   attr2,
			Context: []deepalert.AttrContext{deepalert.CtxLocal},
		},
	}
	alertMsg2, err := json.Marshal(alert)

	require.NoError(t, err)

	var reportID string

	playbook := []gp.Scene{
		// Send request
		gp.PublishSnsMessage(gp.LogicalID("AlertNotification"), alertMsg1),
		gp.PublishSnsMessage(gp.LogicalID("AlertNotification"), alertMsg2),
		gp.GetDynamoRecord(gp.LogicalID("CacheTable"), func(table dynamo.Table) bool {
			id, err := getReportID(alert.AlertID(), table)
			require.NoError(t, err)
			if id == nil {
				return false
			}

			reportID = *id
			logger.WithField("ReportID", reportID).Debug("Got reportID")
			return true
		}),

		gp.GetDynamoRecord(gp.LogicalID("CacheTable"), func(table dynamo.Table) bool {
			var caches attrCaches

			pk := "attribute/" + reportID
			if err := table.Get("pk", pk).All(&caches); err != nil {
				return false
			}

			if len(caches) != 3 {
				return false
			}

			blue := caches.lookup("blue")
			orange := caches.lookup("orange")

			assert.Equal(t, deepalert.TypeUserName, blue.Type)
			assert.Equal(t, deepalert.TypeUserName, orange.Type)
			assert.Equal(t, attr1, blue.Value)
			assert.Equal(t, attr2, orange.Value)
			return true
		}),
	}

	err = gp.New(testCfg.Region, testCfg.DeepAlertStackName).Play(playbook)
	require.NoError(t, err)
}
