package deepalert_test

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

	da "github.com/m-mizutani/deepalert"
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

	c := da.NewReportCoordinator(cfg.TableName, cfg.Region)
	id1, err := da.TakeReportID(c, alertID, ts)
	require.NoError(t, err)
	assert.NotEqual(t, "", id1)

	id2, err := da.TakeReportID(c, alertID, ts.Add(time.Hour))
	require.NoError(t, err)
	// Another result of 1 hour later with same alertID should have same ReportID
	assert.Equal(t, id1, id2)

	id3, err := da.TakeReportID(c, alertID, ts.Add(time.Hour*4))
	require.NoError(t, err)
	// However result over 3 hour later with same alertID should have other ReportID
	assert.NotEqual(t, id1, id3)
}

func TestCoordinatorAlertCache(t *testing.T) {
	cfg := loadTestConfig()
	c := da.NewReportCoordinator(cfg.TableName, cfg.Region)

	alert1 := da.Alert{
		Detector:  "me",
		RuleName:  "myRule",
		AlertKey:  "blue",
		Timestamp: time.Now(),
	}
	alert2 := da.Alert{
		Detector:  "you",
		RuleName:  "yourRule",
		AlertKey:  "orange",
		Timestamp: time.Now(),
	}
	alert3 := da.Alert{
		Detector:  "someone",
		RuleName:  "addRule",
		AlertKey:  "gray",
		Timestamp: time.Now(),
	}

	var err error
	reportID := da.NewReportID()
	err = da.SaveAlertCache(c, reportID, alert1)
	require.NoError(t, err)
	err = da.SaveAlertCache(c, reportID, alert2)
	require.NoError(t, err)

	anotherReportID := da.NewReportID()
	err = da.SaveAlertCache(c, anotherReportID, alert3)
	require.NoError(t, err)

	alerts, err := da.FetchAlertCache(c, reportID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(alerts))

	assert.True(t, alerts[0].Detector == "me" || alerts[1].Detector == "me")
	assert.True(t, alerts[0].Detector == "you" || alerts[1].Detector == "you")
}

func TestCoordinatorReportContent(t *testing.T) {
	cfg := loadTestConfig()
	c := da.NewReportCoordinator(cfg.TableName, cfg.Region)

	rID1 := da.NewReportID()
	content1 := da.ReportContent{
		ReportID: rID1,
		Author:   "blue",
		Attribute: da.Attribute{
			Type:  da.TypeIPAddr,
			Key:   "Remote host",
			Value: "10.1.2.3",
		},
	}
	content2 := da.ReportContent{
		ReportID: rID1,
		Author:   "orange",
		Attribute: da.Attribute{
			Type:  da.TypeIPAddr,
			Key:   "Remote host",
			Value: "10.1.2.3",
		},
	}

	err := da.SaveReportContent(c, content1)
	require.NoError(t, err)
	err = da.SaveReportContent(c, content2)
	require.NoError(t, err)

	contents, err := da.FetchReportContent(c, rID1)
	require.NoError(t, err)
	assert.Equal(t, 2, len(contents))
}
