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
	_, newMockRepo := mock.NewMockRepositorySet()
	repo := service.NewRepositoryService(newMockRepo("nowhere", "test"), 10)
	now := time.Now()

	reportID := deepalert.ReportID(uuid.New().String())
	attr := deepalert.Attribute{
		Type:  deepalert.TypeUserName,
		Key:   "username",
		Value: "blue",
	}
	r := deepalert.Finding{
		ReportID:  reportID,
		Attribute: attr,
		Author:    "tester",
		Type:      deepalert.ContentTypeHost,
		Content: &deepalert.ContentHost{
			HostName: []string{"h1"},
		},
	}
	require.NoError(t, repo.SaveFinding(r, now))

	args := &handler.Arguments{
		NewRepository: newMockRepo,
	}

	resp, err := handleRequest(args, golambda.Event{Origin: &deepalert.Report{
		ID: reportID,
	}})
	require.NoError(t, err)
	updatedReport := resp.(*deepalert.Report)
	assert.Greater(t, len(updatedReport.Sections), 0)
}
