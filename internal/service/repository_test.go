package service_test

import (
	"os"
	"testing"
	"time"

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
