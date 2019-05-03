package functions_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	// "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	da "github.com/m-mizutani/deepalert"
	f "github.com/m-mizutani/deepalert/functions"
)

type testConfig struct {
	TableName string
	Region    string
	LogGroup  string
	LogStream string
}

func loadTestConfig() testConfig {
	type Stack struct {
		StackResources []struct {
			StackId            string
			LogicalResourceId  string
			PhysicalResourceId string
		}
	}

	testConfigPath := "output.json"
	if envvar := os.Getenv("DEEPALERT_TEST_CONFIG"); envvar != "" {
		testConfigPath = envvar
	}

	actualPath := filepath.Join("..", testConfigPath)
	raw, err := ioutil.ReadFile(actualPath)
	if err != nil {
		log.Fatalf("Fail to read testConfigFile: %s, %s", actualPath, err)
	}

	var stack Stack
	if err := json.Unmarshal(raw, &stack); err != nil {
		log.Fatalf("Fail to unmarshal output file: %s, %s", actualPath, err)
	}

	var conf testConfig
	for _, resource := range stack.StackResources {
		switch resource.LogicalResourceId {
		case "CacheTable":
			conf.TableName = resource.PhysicalResourceId
			conf.Region = strings.Split(resource.StackId, ":")[3]

		case "LogStore":
			conf.LogGroup = resource.PhysicalResourceId
		case "LogStream":
			conf.LogStream = resource.PhysicalResourceId
		}
	}

	return conf
}

func TestDataStoreTakeReportID(t *testing.T) {
	cfg := loadTestConfig()
	ts := time.Now().UTC()

	alert1 := da.Alert{
		Detector:  "me",
		RuleName:  "myRule",
		AlertKey:  "blue",
		Timestamp: ts,
	}
	alert2 := da.Alert{
		Detector:  "me",
		RuleName:  "myRule",
		AlertKey:  "blue",
		Timestamp: ts.Add(time.Hour * 1),
	}
	alert3 := da.Alert{
		Detector:  "me",
		RuleName:  "myRule",
		AlertKey:  "orange",
		Timestamp: ts.Add(time.Hour * 4),
	}

	svc := f.NewDataStoreService(cfg.TableName, cfg.Region)

	id1, err := svc.TakeReportID(alert1)
	require.NoError(t, err)
	assert.NotEqual(t, "", id1)

	id2, err := svc.TakeReportID(alert2)
	require.NoError(t, err)
	// Another result of 1 hour later with same alertID should have same ReportID
	assert.Equal(t, id1, id2)

	id3, err := svc.TakeReportID(alert3)
	require.NoError(t, err)
	// However result over 3 hour later with same alertID should have other ReportID
	assert.NotEqual(t, id1, id3)
}

func TestDataStoreAlertCache(t *testing.T) {
	cfg := loadTestConfig()
	svc := f.NewDataStoreService(cfg.TableName, cfg.Region)

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
	reportID := f.NewReportID()
	err = svc.SaveAlertCache(reportID, alert1)
	require.NoError(t, err)
	err = svc.SaveAlertCache(reportID, alert2)
	require.NoError(t, err)

	anotherReportID := f.NewReportID()
	err = svc.SaveAlertCache(anotherReportID, alert3)
	require.NoError(t, err)

	alerts, err := svc.FetchAlertCache(reportID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(alerts))

	assert.True(t, alerts[0].Detector == "me" || alerts[1].Detector == "me")
	assert.True(t, alerts[0].Detector == "you" || alerts[1].Detector == "you")
}

func TestDataStoreReportContent(t *testing.T) {
	cfg := loadTestConfig()
	svc := f.NewDataStoreService(cfg.TableName, cfg.Region)

	rID1 := f.NewReportID()
	rID2 := f.NewReportID()

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
	content3 := da.ReportContent{
		ReportID: rID2,
		Author:   "orange",
		Attribute: da.Attribute{
			Type:  da.TypeIPAddr,
			Key:   "Remote host",
			Value: "10.1.2.3",
		},
	}

	err := svc.SaveReportContent(content1)
	require.NoError(t, err)
	err = svc.SaveReportContent(content2)
	require.NoError(t, err)
	err = svc.SaveReportContent(content3)
	require.NoError(t, err)

	contents, err := svc.FetchReportContent(rID1)
	require.NoError(t, err)
	assert.Equal(t, 2, len(contents))
	assert.Equal(t, rID1, contents[0].ReportID)
	assert.Equal(t, rID1, contents[1].ReportID)
}
