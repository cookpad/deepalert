package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportEntry(t *testing.T) {
	t.Run("Import and Export", func(tt *testing.T) {
		r1 := &deepalert.Report{
			ID: "xba123",
			Alerts: []*deepalert.Alert{
				{
					AlertKey: "cxz",
					Detector: "saber",
					RuleID:   "s1",
					Attributes: []deepalert.Attribute{
						{
							Type: deepalert.TypeIPAddr,
							Context: deepalert.AttrContexts{
								deepalert.CtxRemote,
							},
							Key:   "srcAddr",
							Value: "10.1.2.3",
						},
					},
				},
				{
					AlertKey: "bnc",
					Detector: "archer",
					RuleID:   "a1",
					Attributes: []deepalert.Attribute{
						{
							Type: deepalert.TypeIPAddr,
							Context: deepalert.AttrContexts{
								deepalert.CtxLocal,
							},
							Key:   "dstAddr",
							Value: "192.168.2.3",
						},
					},
				},
			},
			Attributes: []*deepalert.Attribute{
				{
					Type: deepalert.TypeIPAddr,
					Context: deepalert.AttrContexts{
						deepalert.CtxRemote,
					},
					Key:   "srcAddr",
					Value: "10.1.2.3",
				},
				{
					Type: deepalert.TypeIPAddr,
					Context: deepalert.AttrContexts{
						deepalert.CtxLocal,
					},
					Key:   "dstAddr",
					Value: "192.168.2.3",
				},
			},
			Sections: []*deepalert.Section{
				{
					OriginAttr: &deepalert.Attribute{
						Type: deepalert.TypeIPAddr,
						Context: deepalert.AttrContexts{
							deepalert.CtxLocal,
						},
						Key:   "dstAddr",
						Value: "192.168.2.3",
					},

					Users: []*deepalert.ContentUser{
						{
							Activities: []deepalert.EntityActivity{
								{
									Action:     "hoge",
									RemoteAddr: "10.5.6.7",
								},
							},
						},
					},
				},
			},

			Result: deepalert.ReportResult{
				Severity: deepalert.SevSafe,
				Reason:   "no reason",
			},
			Status:    deepalert.StatusPublished,
			CreatedAt: time.Now(),
		}

		var entry models.ReportEntry
		err := entry.Import(r1)
		require.NoError(tt, err)
		r2, err := entry.Export()
		require.NoError(tt, err)
		assert.Equal(tt, r1.ID, r2.ID)

		tt.Run("Fetched report does not have Alert, Attribute and Section", func(ttt *testing.T) {
			// Because they should be fetched by FetchAlertCache, FetchAttributeCache and FetchInspectionNote
			assert.Equal(tt, 0, len(r2.Alerts))
			assert.Equal(tt, 0, len(r2.Attributes))
			assert.Equal(tt, 0, len(r2.Sections))
		})

		tt.Run("Result, Status and CreatedAt should be matched with original report", func(ttt *testing.T) {
			// Result, status, createdAt
			assert.Equal(tt, r1.Result, r2.Result)
			assert.Equal(tt, r1.Status, r2.Status)
			assert.Equal(tt, r1.CreatedAt.UTC().Unix(), r2.CreatedAt.Unix())
		})
	})
}

func TestImportDynamoRecord(t *testing.T) {
	sample := `{
	"awsRegion": "ap-northeast-1",
	"dynamodb": {
		"ApproximateCreationDateTime": 1604111356,
		"Keys": {
			"pk": {
				"S": "report/20c62a1d-99a2-45b5-bca1-2f6949b6ee61"
			},
			"sk": {
				"S": "-"
			}
		},
		"NewImage": {
			"created_at": {
				"N": "1604111355"
			},
			"expires_at": {
				"N": "0"
			},
			"id": {
				"S": "20c62a1d-99a2-45b5-bca1-2f6949b6ee61"
			},
			"pk": {
				"S": "report/20c62a1d-99a2-45b5-bca1-2f6949b6ee61"
			},
			"result": {
				"S": "{\"severity\":\"safe\",\"reason\":\"not sane\"}"
			},
			"sk": {
				"S": "-"
			},
			"status": {
				"S": "new"
			}
		},
		"SequenceNumber": "366860700000000000755927238",
		"SizeBytes": 203,
		"StreamViewType": "NEW_IMAGE"
	},
	"eventID": "f4963c472adeda5d90748601e6affbc8",
	"eventName": "INSERT",
	"eventSource": "aws:dynamodb",
	"eventSourceARN": "arn:aws:dynamodb:ap-northeast-1:783957204773:table/DeepAlertTestStack-cacheTable730E8AED-1FEYS10RXIN14/stream/2020-10-12T13:48:05.842",
	"eventVersion": "1.1"
}`
	var record events.DynamoDBEventRecord
	require.NoError(t, json.Unmarshal([]byte(sample), &record))

	t.Run("Normal case", func(t *testing.T) {
		var entry models.ReportEntry
		require.NoError(t, entry.ImportDynamoRecord(&record))
		report, err := entry.Export()
		require.NoError(t, err)
		assert.Equal(t, int64(1604111355), report.CreatedAt.Unix())
		assert.Equal(t, deepalert.ReportID("20c62a1d-99a2-45b5-bca1-2f6949b6ee61"), report.ID)
		assert.Equal(t, deepalert.SevSafe, report.Result.Severity)
		assert.Equal(t, deepalert.StatusNew, report.Status)
	})
}
