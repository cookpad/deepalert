package main_test

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/adaptor"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/deepalert/deepalert/lambda/receptAlert"
)

func TestReceptAlert(t *testing.T) {
	t.Run("Recept single alert via SQS", func(tt *testing.T) {
		alert := &deepalert.Alert{
			AlertKey: uuid.New().String(),
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
			Event: events.SQSEvent{
				Records: []events.SQSMessage{
					{
						Body: string(raw),
					},
				},
			},
		}

		resp, err := main.HandleRequest(args)
		require.NoError(tt, err)
		assert.Nil(tt, resp)

		// Check only execution of StepFunctions. More detailed test are in internal/usecase
		sfn, ok := dummySFn.(*mock.SFnClient)
		require.True(tt, ok)
		require.Equal(tt, 2, len(sfn.Input))
	})

	t.Run("Recept single alert via SQS+SNS", func(tt *testing.T) {
		alert := &deepalert.Alert{
			AlertKey: uuid.New().String(),
			RuleID:   "five",
			RuleName: "fifth",
			Detector: "ao",
		}
		raw, err := json.Marshal(alert)
		require.NoError(tt, err)

		snsEntity := &events.SNSEntity{Message: string(raw)}
		body, err := json.Marshal(snsEntity)
		require.NoError(t, err)

		dummySFn, _ := mock.NewSFnClient("")
		dummyRepo := mock.NewRepository("", "")
		args := &handler.Arguments{
			NewRepository: func(string, string) adaptor.Repository { return dummyRepo },
			NewSFn:        func(string) (adaptor.SFnClient, error) { return dummySFn, nil },
			EnvVars: handler.EnvVars{
				InspectorMashine: "arn:aws:states:us-east-1:111122223333:stateMachine:blue",
				ReviewMachine:    "arn:aws:states:us-east-1:111122223333:stateMachine:orange",
			},
			Event: events.SQSEvent{
				Records: []events.SQSMessage{
					{
						Body: string(body),
					},
				},
			},
		}

		resp, err := main.HandleRequest(args)
		require.NoError(tt, err)
		assert.Nil(tt, resp)

		// Check only execution of StepFunctions. More detailed test are in internal/usecase
		sfn, ok := dummySFn.(*mock.SFnClient)
		require.True(tt, ok)
		require.Equal(tt, 2, len(sfn.Input))
	})

}
