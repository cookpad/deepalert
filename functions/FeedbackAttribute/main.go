package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/m-mizutani/deepalert"
	f "github.com/m-mizutani/deepalert/functions"
)

type lambdaArguments struct {
	Event            events.SNSEvent
	TaskNotification string
	CacheTable       string
	Region           string
}

func mainHandler(args lambdaArguments) error {
	svc := f.NewDataStoreService(args.CacheTable, args.Region)

	for _, msg := range f.SNStoMessages(args.Event) {
		var reportedAttr deepalert.ReportAttribute
		if err := json.Unmarshal(msg, &reportedAttr); err != nil {
			return errors.Wrapf(err, "Fail to unmarshal ReportAttribute from AttributeaNotification: %s", string(msg))
		}

		f.Logger.WithField("reportedAttr", reportedAttr).Info("unmarshaled reported attribute")

		for _, attr := range reportedAttr.Attributes {
			sendable, err := svc.PutAttributeCache(reportedAttr.ReportID, attr)
			if err != nil {
				return errors.Wrapf(err, "Fail to manage attribute cache: %v", attr)
			}

			f.Logger.WithFields(logrus.Fields{"sendable": sendable, "attr": attr}).Info("attribute")
			if !sendable {
				continue
			}

			task := deepalert.Task{
				ReportID:  reportedAttr.ReportID,
				Attribute: attr,
			}

			if err := f.PublishSNS(args.TaskNotification, args.Region, &task); err != nil {
				return errors.Wrapf(err, "Fail to publihsh task notification: %v", task)
			}

		}
	}

	return nil
}

func handleRequest(ctx context.Context, event events.SNSEvent) error {
	f.SetLoggerContext(ctx, "FeedbackAttribute", deepalert.ReportID(""))
	f.Logger.WithField("event", event).Info("Start")

	args := lambdaArguments{
		Event:            event,
		TaskNotification: os.Getenv("TASK_NOTIFICATION"),
		CacheTable:       os.Getenv("CACHE_TABLE"),
		Region:           os.Getenv("AWS_REGION"),
	}

	if err := mainHandler(args); err != nil {
		f.Logger.WithError(err).Error("Fail")
		return err
	}

	return nil
}

func main() {
	lambda.Start(handleRequest)
}
