package main_test

import (
	"context"
	"testing"

	"github.com/deepalert/deepalert"
	main "github.com/deepalert/deepalert/examples/inspector"
	"github.com/deepalert/deepalert/inspector"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInspectorExample(t *testing.T) {
	attrURL := "https://sqs.ap-northeast-1.amazonaws.com/123456789xxx/attribute-queue"
	contentURL := "https://sqs.ap-northeast-1.amazonaws.com/123456789xxx/content-queue"

	args := inspector.Arguments{
		Handler:         main.Handler,
		Author:          "blue",
		AttrQueueURL:    attrURL,
		ContentQueueURL: contentURL,
	}

	t.Run("With IPaddr attribute", func(tt *testing.T) {
		mock, newSQS := inspector.NewSQSMock()
		args.NewSQS = newSQS

		task := deepalert.Task{
			ReportID: deepalert.ReportID(uuid.New().String()),
			Attribute: deepalert.Attribute{
				Type:  deepalert.TypeIPAddr,
				Key:   "dst",
				Value: "192.10.0.1",
			},
		}

		err := inspector.HandleTask(context.Background(), args, task)
		require.NoError(tt, err)
		sections, err := mock.GetSections(contentURL)
		require.NoError(tt, err)
		require.Equal(tt, 1, len(sections))
	})

	t.Run("With not IPaddr attribute", func(tt *testing.T) {
		mock, newSQS := inspector.NewSQSMock()
		args.NewSQS = newSQS

		task := deepalert.Task{
			ReportID: deepalert.ReportID(uuid.New().String()),
			Attribute: deepalert.Attribute{
				Type:  deepalert.TypeUserName,
				Key:   "login-name",
				Value: "mizutani",
			},
		}

		err := inspector.HandleTask(context.Background(), args, task)
		require.NoError(tt, err)
		sections, err := mock.GetSections(contentURL)
		require.NoError(tt, err)
		require.Equal(tt, 0, len(sections))
	})
}
