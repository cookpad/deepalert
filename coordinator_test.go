package deepalert

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	TableName string
	Region    string
}

func loadTestConfig() testConfig {
	testConfigPath := "test.json"
	if envvar := os.Getenv("DEEPALERT_TEST_CONFIG"); envvar != "" {
		testConfigPath = envvar
	}

	raw, err := ioutil.ReadFile(testConfigPath)
	if err != nil {
		log.Fatalf("Fail to read testConfigFile: %s, %s", testConfigPath, err)
	}

	var conf testConfig
	if err := json.Unmarshal(raw, &conf); err != nil {
		log.Fatalf("Fail to unmarshal testConfigFile: %s, %s", testConfigPath, err)
	}

	return conf
}

func TestCoordinatorTakeReportID(t *testing.T) {
	cfg := loadTestConfig()

	ts := time.Now()
	alertID := uuid.New().String()

	c := NewReportCoordinator(cfg.TableName, cfg.Region)
	id1, err := TakeReportID(c, alertID, ts)
	require.NoError(t, err)
	assert.NotEqual(t, "", id1)

	id2, err := TakeReportID(c, alertID, ts.Add(time.Hour))
	require.NoError(t, err)
	// Another result of 1 hour later with same alertID should have same ReportID
	assert.Equal(t, id1, id2)

	id3, err := TakeReportID(c, alertID, ts.Add(time.Hour*4))
	require.NoError(t, err)
	// However result over 3 hour later with same alertID should have other ReportID
	assert.NotEqual(t, id1, id3)
}
