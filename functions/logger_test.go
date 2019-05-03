package functions_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	f "github.com/m-mizutani/deepalert/functions"
)

func TestLoggerHook(t *testing.T) {
	cfg := loadTestConfig()
	testID1 := uuid.New().String()

	reportID := f.NewReportID()
	now := time.Now().UTC()

	old := f.GetLogOutput()
	f.SetLogOutput(new(bytes.Buffer))
	defer func() { f.SetLogOutput(old) }()

	f.SetLogDestination(cfg.LogGroup, cfg.LogStream, cfg.Region)
	f.SetLoggerContext(nil, "Test", reportID)
	f.Logger.WithField("test_id", testID1).Info("Test")

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
	}))
	cwlogs := cloudwatchlogs.New(ssn)

	detected := func() bool {
		for i := 0; i < 10; i++ {
			resp, err := cwlogs.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
				LogGroupName:  aws.String(cfg.LogGroup),
				LogStreamName: aws.String(cfg.LogStream),
				StartTime:     aws.Int64(now.Unix() * 1000),
			})

			require.NoError(t, err)
			for _, event := range resp.Events {
				if strings.Index(event.GoString(), testID1) >= 0 {
					return true
				}
			}
			time.Sleep(time.Second * 2)
		}
		return false
	}()

	assert.True(t, detected)
}
