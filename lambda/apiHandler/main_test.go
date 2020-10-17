package main

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/adaptor"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReceptAlert(t *testing.T) {
	t.Run("Recept single alert", func(tt *testing.T) {
		alert := &deepalert.Alert{
			AlertKey: "5",
			RuleID:   "five",
			RuleName: "fifth",
			Detector: "ao",
		}
		raw, err := json.Marshal(alert)
		require.NoError(tt, err)

		dummySFn, _ := mock.NewSFnClient("")
		dummyRepo := mock.NewRepository("", "")
		args := &handler.Arguments{
			NewRepository: func(string, string) adaptor.Repository { return dummyRepo },
			NewSFn:        func(string) (adaptor.SFnClient, error) { return dummySFn, nil },
			EnvVars: handler.EnvVars{
				InspectorMashine: "arn:aws:states:us-east-1:111122223333:stateMachine:blue",
				ReviewMachine:    "arn:aws:states:us-east-1:111122223333:stateMachine:orange",
			},
			Event: events.APIGatewayProxyRequest{
				HTTPMethod: "POST",
				Path:       "/api/v1/alert",
				Body:       string(raw),
			},
		}

		resp, err := handleRequest(args)
		require.NoError(tt, err)
		assert.NotNil(tt, resp)
		httpResp, ok := resp.(events.APIGatewayProxyResponse)
		assert.Equal(tt, 200, httpResp.StatusCode)

		// Check only execution of StepFunctions. More detailed test are in internal/usecase
		sfn, ok := dummySFn.(*mock.SFnClient)
		require.True(tt, ok)
		require.Equal(tt, 2, len(sfn.Input))
	})
}
