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

		// Alert
		assert.Equal(tt, 2, len(r2.Alerts))
		assert.Contains(tt, r2.Alerts, r1.Alerts[0])
		assert.Contains(tt, r2.Alerts, r1.Alerts[1])

		// Attribute
		assert.Equal(tt, 2, len(r2.Attributes))
		assert.Contains(tt, r2.Attributes, r1.Attributes[0])
		assert.Contains(tt, r2.Attributes, r1.Attributes[1])

		assert.Equal(tt, r1.Sections[0].OriginAttr, r2.Sections[0].OriginAttr)
		require.Equal(tt, 1, len(r2.Sections[0].Users))
		assert.Equal(tt, 0, len(r2.Sections[0].Hosts))
		assert.Equal(tt, 0, len(r2.Sections[0].Binaries))
		assert.Equal(tt, r1.Sections[0].Users[0], r2.Sections[0].Users[0])

		// Result, status, createdAt
		assert.Equal(tt, r1.Result, r2.Result)
		assert.Equal(tt, r1.Status, r2.Status)
		assert.Equal(tt, r1.CreatedAt.UTC().Unix(), r2.CreatedAt.Unix())
	})
}
