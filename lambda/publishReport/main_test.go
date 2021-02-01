package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/mock"
	"github.com/deepalert/deepalert/internal/models"
	"github.com/deepalert/deepalert/internal/service"
	"github.com/google/uuid"
	"github.com/m-mizutani/golambda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleReport(t *testing.T) {
	// Setup dummy repository
	mockRepo, newMockRepo := mock.NewMockRepositorySet()
	mockSNS, newMockSNS := mock.NewMockSNSClientSet()
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

	require.NoError(t, repo.PutReport(&deepalert.Report{
		ID: reportID,
	}))
	require.NoError(t, repo.SaveFinding(finding, now))
	require.NoError(t, repo.SaveAlertCache(reportID, alert, now))
	_, err := repo.PutAttributeCache(reportID, attr, now)
	require.NoError(t, err)

	args := &handler.Arguments{
		NewRepository: newMockRepo,
		NewSNS:        newMockSNS,
		EnvVars: handler.EnvVars{
			ReportTopic: "arn:aws:sns:us-east-1:111122223333:my-topic",
		},
	}

	dynamoEvent := events.DynamoDBEvent{
		Records: []events.DynamoDBEventRecord{
			{
				Change: events.DynamoDBStreamRecord{
					Keys: map[string]events.DynamoDBAttributeValue{
						models.DynamoPKeyName: events.NewStringAttribute("report/" + string(reportID)),
					},
					NewImage: map[string]events.DynamoDBAttributeValue{
						"id":         events.NewStringAttribute(string(reportID)),
						"result":     events.NewStringAttribute(`{}`),
						"status":     events.NewStringAttribute(string(deepalert.StatusPublished)),
						"created_at": events.NewNumberAttribute("1612167325"),
					},
				},
			},
		},
	}

	_, err = handleRequest(args, golambda.Event{Origin: dynamoEvent})
	require.NoError(t, err)

	require.Equal(t, 1, len(mockSNS.Input))
	assert.Equal(t, "arn:aws:sns:us-east-1:111122223333:my-topic", *(mockSNS.Input[0].TopicArn))

	var report deepalert.Report

	require.NoError(t, json.Unmarshal([]byte(*mockSNS.Input[0].Message), &report))
	assert.Equal(t, 1, len(report.Sections))
	assert.Contains(t, report.Alerts, &alert)
	require.Equal(t, 1, len(report.Attributes))
	assert.Equal(t, report.Attributes[0].Value, attr.Value)
}
