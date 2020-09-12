package service_test

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/m-mizutani/deepalert"
	"github.com/m-mizutani/deepalert/internal/mock"
	"github.com/m-mizutani/deepalert/internal/repository"
	"github.com/m-mizutani/deepalert/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const commonTTL = int64(10)

func testRepositoryService(t *testing.T, svc *service.RepositoryService) {
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
		assert.Contains(tt, cache, alert1)
		assert.Contains(tt, cache, alert2)
		assert.NotContains(tt, cache, alert3)
	})

	t.Run("Savea and Fetch report section", func(tt *testing.T) {
		id1 := deepalert.ReportID(uuid.New().String())
		id2 := deepalert.ReportID(uuid.New().String())
		now := time.Now()
		s1 := deepalert.ReportSection{
			ReportID: id1,
			Author:   "a1",
			Attribute: deepalert.Attribute{
				Type:  deepalert.TypeIPAddr,
				Value: "10.0.0.1",
			},
		}
		s2 := deepalert.ReportSection{
			ReportID: id1,
			Author:   "a2",
			Attribute: deepalert.Attribute{
				Type:  deepalert.TypeIPAddr,
				Value: "10.0.0.2",
			},
		}
		s3 := deepalert.ReportSection{
			ReportID: id2,
			Author:   "a3",
			Attribute: deepalert.Attribute{
				Type:  deepalert.TypeIPAddr,
				Value: "10.0.0.3",
			},
		}
		require.NoError(tt, svc.SaveReportSection(s1, now))
		require.NoError(tt, svc.SaveReportSection(s2, now))
		require.NoError(tt, svc.SaveReportSection(s3, now))

		sections, err := svc.FetchReportSection(id1)
		require.NoError(tt, err)
		assert.Contains(tt, sections, s1)
		assert.Contains(tt, sections, s2)
		assert.NotContains(tt, sections, s3)
	})

	t.Run("AttributeCache", func(tt *testing.T) {
		testAttributeCache(tt, svc)
	})
}

func testAttributeCache(t *testing.T, svc *service.RepositoryService) {
	t.Run("Put and Fetch attributes", func(tt *testing.T) {
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
		require.NoError(tt, err)
		assert.True(tt, b1)

		b2, err := svc.PutAttributeCache(id1, attr2, now)
		require.NoError(tt, err)
		assert.True(tt, b2)

		b3, err := svc.PutAttributeCache(id2, attr3, now)
		require.NoError(tt, err)
		assert.True(tt, b3)

		attrs, err := svc.FetchAttributeCache(id1)
		require.NoError(tt, err)

		var attrList []deepalert.Attribute
		for _, attr := range attrs {
			a := attr
			a.Timestamp = nil
			attrList = append(attrList, a)
		}
		assert.Contains(tt, attrList, attr1)
		assert.Contains(tt, attrList, attr2)
		assert.NotContains(tt, attrList, attr3)
	})

	t.Run("Duplicated attribute", func(tt *testing.T) {
		id1 := deepalert.ReportID(uuid.New().String())
		now := time.Now()

		attr1 := deepalert.Attribute{
			Type:    deepalert.TypeIPAddr,
			Context: deepalert.AttrContexts{deepalert.CtxRemote},
			Key:     "dst",
			Value:   "10.0.0.2",
		}
		b1, err := svc.PutAttributeCache(id1, attr1, now)
		require.NoError(tt, err)
		assert.True(tt, b1)

		// No error by second PutAttributeCache. But returns false to indicate the attribute already exists
		b1d, err := svc.PutAttributeCache(id1, attr1, now)
		require.NoError(tt, err)
		assert.False(tt, b1d)

		attrs, err := svc.FetchAttributeCache(id1)
		require.NoError(tt, err)
		assert.Equal(tt, 1, len(attrs))
		attrs[0].Timestamp = nil
		assert.Equal(tt, attr1, attrs[0])
	})
}

func TestDynamoDBRepository(t *testing.T) {
	region, tableName := os.Getenv("DEEPALERT_TEST_REGION"), os.Getenv("DEEPALERT_TEST_TABLE")
	if region == "" || tableName == "" {
		t.Skip("Either of DEEPALERT_TEST_REGION and DEEPALERT_TEST_TABLE are not set")
	}

	repo := repository.NewDynamoDB(region, tableName)
	svc := service.NewRepositoryService(repo, commonTTL)

	testRepositoryService(t, svc)
}

func TestMockRepository(t *testing.T) {
	repo := mock.NewRepository("test-region", "test-table")
	svc := service.NewRepositoryService(repo, commonTTL)

	testRepositoryService(t, svc)
}
