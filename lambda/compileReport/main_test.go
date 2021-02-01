package main

import (
	"testing"
	"time"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/mock"
	"github.com/deepalert/deepalert/internal/service"
	"github.com/google/uuid"
	"github.com/m-mizutani/golambda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleReport(t *testing.T) {
	// Setup dummy repository
	mockRepo, newMockRepo := mock.NewMockRepositorySet()
	repo := service.NewRepositoryService(mockRepo, 10)
	now := time.Now()

	reportID := deepalert.ReportID(uuid.New().String())
	attr := deepalert.Attribute{
		Type:  deepalert.TypeUserName,
		Key:   "username",
		Value: "blue",
	}
	finding := deepalert.Finding{
		ReportID:  reportID,
		Attribute: attr,
		Author:    "tester",
		Type:      deepalert.ContentTypeHost,
		Content: &deepalert.ContentHost{
			HostName: []string{"h1"},
		},
	}

	alert := deepalert.Alert{
		Detector: "tester",
		RuleName: "testRule",
		RuleID:   "testID",
	}

	report := &deepalert.Report{
		ID: reportID,
	}

	require.NoError(t, repo.PutReport(report))
	require.NoError(t, repo.SaveFinding(finding, now))
	require.NoError(t, repo.SaveAlertCache(reportID, alert, now))
	_, err := repo.PutAttributeCache(reportID, attr, now)
	require.NoError(t, err)

	args := &handler.Arguments{
		NewRepository: newMockRepo,
	}

	resp, err := handleRequest(args, golambda.Event{Origin: report})
	require.NoError(t, err)
	updatedReport := resp.(*deepalert.Report)
	require.NotNil(t, updatedReport)
	assert.Equal(t, len(updatedReport.Sections), 1)
	assert.Equal(t, len(updatedReport.Alerts), 1)
	assert.Equal(t, len(updatedReport.Attributes), 1)
}
