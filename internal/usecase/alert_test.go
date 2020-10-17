package usecase_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/adaptor"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/mock"
	"github.com/deepalert/deepalert/internal/service"
	"github.com/deepalert/deepalert/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleAlert(t *testing.T) {
	basicSetup := func() (*handler.Arguments, adaptor.SFnClient, adaptor.Repository) {
		dummySFn, _ := mock.NewSFnClient("")
		dummyRepo := mock.NewRepository("", "")
		args := &handler.Arguments{
			NewRepository: func(string, string) adaptor.Repository { return dummyRepo },
			NewSFn:        func(string) (adaptor.SFnClient, error) { return dummySFn, nil },
			EnvVars: handler.EnvVars{
				InspectorMashine: "arn:aws:states:us-east-1:111122223333:stateMachine:blue",
				ReviewMachine:    "arn:aws:states:us-east-1:111122223333:stateMachine:orange",
			},
		}
		return args, dummySFn, dummyRepo
	}

	t.Run("Recept single alert", func(t *testing.T) {
		alert := &deepalert.Alert{
			AlertKey: "5",
			RuleID:   "five",
			RuleName: "fifth",
			Detector: "ao",
		}

		args, dummySFn, dummyRepo := basicSetup()

		report, err := usecase.HandleAlert(args, alert, time.Now())
		require.NoError(t, err)
		assert.NotNil(t, report)
		assert.NotEqual(t, "", report.ID)

		repoSvc := service.NewRepositoryService(dummyRepo, 10)

		t.Run("StepFunctions should be executed", func(t *testing.T) {
			sfn, ok := dummySFn.(*mock.SFnClient)
			require.True(t, ok)
			require.Equal(t, 2, len(sfn.Input))
			assert.Equal(t, "arn:aws:states:us-east-1:111122223333:stateMachine:blue", *sfn.Input[0].StateMachineArn)
			assert.Equal(t, "arn:aws:states:us-east-1:111122223333:stateMachine:orange", *sfn.Input[1].StateMachineArn)

			var report1, report2 deepalert.Report
			require.NoError(t, json.Unmarshal([]byte(*sfn.Input[0].Input), &report1))
			require.Equal(t, 1, len(report1.Alerts))
			assert.Equal(t, alert, report1.Alerts[0])

			require.NoError(t, json.Unmarshal([]byte(*sfn.Input[1].Input), &report2))
			require.Equal(t, 1, len(report2.Alerts))
			assert.Equal(t, alert, report2.Alerts[0])

			assert.Equal(t, report1, report2)
		})

		t.Run("AlertCachce should be stored in repository", func(t *testing.T) {
			alertCache, err := repoSvc.FetchAlertCache(report.ID)
			require.NoError(t, err)
			require.Equal(t, 1, len(alertCache))
			assert.Equal(t, alert, alertCache[0])
		})

		t.Run("Report should be stored in repository", func(t *testing.T) {
			report, err := repoSvc.GetReport(report.ID)
			require.NoError(t, err)
			require.Equal(t, 1, len(report.Alerts))
			assert.Equal(t, alert, report.Alerts[0])
		})
	})

	t.Run("Recept alerts with same AlertID", func(t *testing.T) {
		// AlertID is calculated by AlertKey, RuleID and Detector
		alert1 := &deepalert.Alert{
			AlertKey: "123",
			RuleID:   "blue",
			RuleName: "fifth",
			Detector: "ao",
		}
		alert2 := &deepalert.Alert{
			AlertKey: "123",
			RuleID:   "blue",
			RuleName: "five",
			Detector: "ao",
		}
		args, _, _ := basicSetup()

		report1, err := usecase.HandleAlert(args, alert1, time.Now())
		require.NoError(t, err)
		assert.NotNil(t, report1)
		assert.NotEqual(t, "", report1.ID)

		report2, err := usecase.HandleAlert(args, alert2, time.Now())
		require.NoError(t, err)
		assert.NotNil(t, report2)
		assert.NotEqual(t, "", report2.ID)

		t.Run("ReportIDs should be same", func(t *testing.T) {
			assert.Equal(t, report1.ID, report2.ID)
		})
	})

	t.Run("ReportIDs should be different if AlertID is not same", func(t *testing.T) {
		// AlertID is calculated by AlertKey, RuleID and Detector
		args, _, _ := basicSetup()

		t.Run("Different AlertKey", func(t *testing.T) {
			alert1 := &deepalert.Alert{
				AlertKey: "234",
				RuleID:   "blue",
				RuleName: "fifth",
				Detector: "ao",
			}
			alert2 := &deepalert.Alert{
				AlertKey: "123",
				RuleID:   "blue",
				RuleName: "five",
				Detector: "ao",
			}

			report1, err := usecase.HandleAlert(args, alert1, time.Now())
			require.NoError(t, err)
			report2, err := usecase.HandleAlert(args, alert2, time.Now())
			require.NoError(t, err)
			assert.NotEqual(t, report1.ID, report2.ID)
		})

		t.Run("Different RuleID", func(t *testing.T) {
			alert1 := &deepalert.Alert{
				AlertKey: "123",
				RuleID:   "blue",
				RuleName: "fifth",
				Detector: "ao",
			}
			alert2 := &deepalert.Alert{
				AlertKey: "123",
				RuleID:   "orange",
				RuleName: "five",
				Detector: "ao",
			}

			report1, err := usecase.HandleAlert(args, alert1, time.Now())
			require.NoError(t, err)
			report2, err := usecase.HandleAlert(args, alert2, time.Now())
			require.NoError(t, err)
			assert.NotEqual(t, report1.ID, report2.ID)
		})

		t.Run("Different Detector", func(t *testing.T) {
			alert1 := &deepalert.Alert{
				AlertKey: "123",
				RuleID:   "blue",
				RuleName: "fifth",
				Detector: "ao",
			}
			alert2 := &deepalert.Alert{
				AlertKey: "123",
				RuleID:   "blue",
				RuleName: "five",
				Detector: "tou",
			}

			report1, err := usecase.HandleAlert(args, alert1, time.Now())
			require.NoError(t, err)
			report2, err := usecase.HandleAlert(args, alert2, time.Now())
			require.NoError(t, err)
			assert.NotEqual(t, report1.ID, report2.ID)
		})
	})
}
