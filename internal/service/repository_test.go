package service_test

import (
	"os"
	"testing"
	"time"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/mock"
	"github.com/deepalert/deepalert/internal/repository"
	"github.com/deepalert/deepalert/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const commonTTL = int64(10)

func testRepositoryService(t *testing.T, svc *service.RepositoryService) {
	t.Run("TakeReport", func(tt *testing.T) {
		testTakeReport(tt, svc)
	})
	t.Run("AlertCache", func(tt *testing.T) {
		testAlertCache(tt, svc)
	})
	t.Run("Finding", func(tt *testing.T) {
		testFinding(tt, svc)
	})
	t.Run("AttributeCache", func(tt *testing.T) {
		testAttributeCache(tt, svc)
	})
	t.Run("Report", func(tt *testing.T) {
		testRpoert(tt, svc)
	})
}

func testTakeReport(t *testing.T, svc *service.RepositoryService) {
	now := time.Now()

	t.Run("Take new reports with diffent keys", func(tt *testing.T) {
		r1, err := svc.TakeReport(deepalert.Alert{
			AlertKey:    "x1",
			RuleID:      "r1",
			Description: "d1",
		}, now)
		require.NoError(tt, err)
		require.NotNil(tt, r1)

		r2, err := svc.TakeReport(deepalert.Alert{
			AlertKey:    "x2",
			RuleID:      "r1",
			Description: "d1",
		}, now)
		require.NoError(tt, err)
		require.NotNil(tt, r2)

		assert.NotEqual(tt, r1.ID, r2.ID)
	})

	t.Run("Take reports with same key", func(tt *testing.T) {
		r1, err := svc.TakeReport(deepalert.Alert{
			AlertKey:    "x1",
			RuleID:      "r1",
			Description: "d1",
		}, now)
		require.NoError(tt, err)
		require.NotNil(tt, r1)

		r2, err := svc.TakeReport(deepalert.Alert{
			AlertKey:    "x1",
			RuleID:      "r1",
			Description: "d1",
		}, now.Add(1*time.Second))
		require.NoError(tt, err)
		require.NotNil(tt, r2)

		assert.Equal(tt, r1.ID, r2.ID)
	})
}

func testAlertCache(t *testing.T, svc *service.RepositoryService) {
	t.Run("Save and fetch alert cache", func(tt *testing.T) {
		id1 := deepalert.ReportID(uuid.New().String())
		id2 := deepalert.ReportID(uuid.New().String())

		alert1 := deepalert.Alert{
			AlertKey: "k1",
			RuleID:   "r1",
			RuleName: "n1",
		}
		alert2 := deepalert.Alert{
			AlertKey: "k2",
			RuleID:   "r2",
			RuleName: "n2",
		}
		alert3 := deepalert.Alert{
			AlertKey: "k3",
			RuleID:   "r3",
			RuleName: "n3",
		}
		now := time.Now()
		require.NoError(tt, svc.SaveAlertCache(id1, alert1, now))
		require.NoError(tt, svc.SaveAlertCache(id1, alert2, now))
		require.NoError(tt, svc.SaveAlertCache(id2, alert3, now))

		cache, err := svc.FetchAlertCache(id1)
		require.NoError(tt, err)
		assert.Contains(tt, cache, &alert1)
		assert.Contains(tt, cache, &alert2)
		assert.NotContains(tt, cache, &alert3)
	})
}

func testFinding(t *testing.T, svc *service.RepositoryService) {
	t.Run("Savea and Fetch report section", func(tt *testing.T) {
		id1 := deepalert.ReportID(uuid.New().String())
		id2 := deepalert.ReportID(uuid.New().String())
		now := time.Now()
		s1 := deepalert.Finding{
			ReportID: id1,
			Author:   "a1",
			Attribute: deepalert.Attribute{
				Type:  deepalert.TypeIPAddr,
				Value: "10.0.0.1",
			},
			Type: deepalert.ContentTypeHost,
			Content: deepalert.ContentHost{
				HostName: []string{"h1"},
			},
		}
		s2 := deepalert.Finding{
			ReportID: id1,
			Author:   "a2",
			Attribute: deepalert.Attribute{
				Type:  deepalert.TypeIPAddr,
				Value: "10.0.0.2",
			},
			Type: deepalert.ContentTypeHost,
			Content: deepalert.ContentHost{
				HostName: []string{"h2"},
			},
		}
		s3 := deepalert.Finding{
			ReportID: id2,
			Author:   "a3",
			Attribute: deepalert.Attribute{
				Type:  deepalert.TypeIPAddr,
				Value: "10.0.0.3",
			},
			Type: deepalert.ContentTypeHost,
			Content: deepalert.ContentHost{
				HostName: []string{"h3"},
			},
		}

		attrs := []deepalert.Attribute{s1.Attribute, s2.Attribute, s3.Attribute}
		require.NoError(tt, svc.SaveFinding(s1, now))
		require.NoError(tt, svc.SaveFinding(s2, now))
		require.NoError(tt, svc.SaveFinding(s3, now))

		sections, err := svc.FetchSection(id1)
		require.NoError(tt, err)
		require.Equal(tt, 2, len(sections))
		assert.Contains(tt, attrs, sections[0].Attr)
		assert.Contains(tt, attrs, sections[1].Attr)
	})
}

func testAttributeCache(t *testing.T, svc *service.RepositoryService) {
	t.Run("Put and Fetch attributes", func(t *testing.T) {
		id1 := deepalert.ReportID(uuid.New().String())
		id2 := deepalert.ReportID(uuid.New().String())
		now := time.Now()

		attr1 := deepalert.Attribute{
			Type:    deepalert.TypeIPAddr,
			Context: deepalert.AttrContexts{deepalert.CtxRemote},
			Key:     "dst",
			Value:   "10.0.0.2",
		}
		attr2 := deepalert.Attribute{
			Type:    deepalert.TypeIPAddr,
			Context: deepalert.AttrContexts{deepalert.CtxRemote},
			Key:     "dst",
			Value:   "10.0.0.3",
		}
		attr3 := deepalert.Attribute{
			Type:    deepalert.TypeIPAddr,
			Context: deepalert.AttrContexts{deepalert.CtxRemote},
			Key:     "dst",
			Value:   "10.0.0.4",
		}

		b1, err := svc.PutAttributeCache(id1, attr1, now)
		require.NoError(t, err)
		assert.True(t, b1)

		b2, err := svc.PutAttributeCache(id1, attr2, now)
		require.NoError(t, err)
		assert.True(t, b2)

		b3, err := svc.PutAttributeCache(id2, attr3, now)
		require.NoError(t, err)
		assert.True(t, b3)

		attrs, err := svc.FetchAttributeCache(id1)
		require.NoError(t, err)

		var attrList []*deepalert.Attribute
		for _, attr := range attrs {
			a := attr
			a.Timestamp = nil
			attrList = append(attrList, a)
		}
		assert.Contains(t, attrList, &attr1)
		assert.Contains(t, attrList, &attr2)
		assert.NotContains(t, attrList, &attr3)
	})

	t.Run("Duplicated attribute", func(t *testing.T) {
		id1 := deepalert.ReportID(uuid.New().String())
		now := time.Now()

		attr1 := deepalert.Attribute{
			Type:    deepalert.TypeIPAddr,
			Context: deepalert.AttrContexts{deepalert.CtxRemote},
			Key:     "dst",
			Value:   "10.0.0.2",
		}
		b1, err := svc.PutAttributeCache(id1, attr1, now)
		require.NoError(t, err)
		assert.True(t, b1)

		// No error by second PutAttributeCache. But returns false to indicate the attribute already exists
		b1d, err := svc.PutAttributeCache(id1, attr1, now)
		require.NoError(t, err)
		assert.False(t, b1d)

		attrs, err := svc.FetchAttributeCache(id1)
		require.NoError(t, err)
		assert.Equal(t, 1, len(attrs))
		attrs[0].Timestamp = nil
		assert.Equal(t, attr1, *attrs[0])
	})
}

func testRpoert(t *testing.T, svc *service.RepositoryService) {
	t.Run("Put and Get", func(tt *testing.T) {
		r1 := &deepalert.Report{
			ID: deepalert.ReportID(uuid.New().String()),
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
					Attr: deepalert.Attribute{
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

		err := svc.PutReport(r1)
		require.NoError(tt, err)
		r2, err := svc.GetReport(r1.ID)
		require.NoError(tt, err)
		assert.Equal(tt, r1.ID, r2.ID)
		require.Equal(tt, 0, len(r2.Attributes)) // PutReport does not save attributes
	})

	t.Run("Not found", func(tt *testing.T) {
		id := deepalert.ReportID(uuid.New().String())
		r0, err := svc.GetReport(id)
		require.NoError(tt, err)
		assert.Nil(tt, r0)
	})
}

func TestDynamoDBRepository(t *testing.T) {
	region, tableName := os.Getenv("DEEPALERT_TEST_REGION"), os.Getenv("DEEPALERT_TEST_TABLE")
	if region == "" || tableName == "" {
		t.Skip("Either of DEEPALERT_TEST_REGION and DEEPALERT_TEST_TABLE are not set")
	}

	repo, err := repository.NewDynamoDB(region, tableName)
	require.NoError(t, err)
	svc := service.NewRepositoryService(repo, commonTTL)

	testRepositoryService(t, svc)
}

func TestMockRepository(t *testing.T) {
	repo := mock.NewRepository("test-region", "test-table")
	svc := service.NewRepositoryService(repo, commonTTL)

	testRepositoryService(t, svc)
}
