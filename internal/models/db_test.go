package models_test

import (
	"testing"
	"time"

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
			Sections: []*deepalert.ReportSection{
				{
					OriginAttr: &deepalert.Attribute{
						Type: deepalert.TypeIPAddr,
						Context: deepalert.AttrContexts{
							deepalert.CtxLocal,
						},
						Key:   "dstAddr",
						Value: "192.168.2.3",
					},

					Users: []*deepalert.ReportUser{
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
			// Because they should be fetched by FetchAlertCache, FetchAttributeCache and FetchInspectReport
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
