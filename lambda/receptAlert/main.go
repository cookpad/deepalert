package main

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/m-mizutani/deepalert"
	"github.com/m-mizutani/deepalert/internal/errors"
	"github.com/m-mizutani/deepalert/internal/handler"
	"github.com/m-mizutani/deepalert/internal/logging"
)

var logger = logging.Logger

func main() {
	handler.StartLambda(HandleRequest)
}

// HandleRequest is main logic of ReceptAlert
func HandleRequest(args *handler.Arguments) (handler.Response, error) {
	messages, err := args.DecapSNSEvent()
	if err != nil {
		return nil, err
	}

	repo, err := args.Repository()
	if err != nil {
		return nil, err
	}
	snsSvc := args.SNSService()
	sfnSvc := args.SFnService()
	now := time.Now()

	for _, msg := range messages {
		var alert deepalert.Alert
		if err := json.Unmarshal(msg, &alert); err != nil {
			return nil, errors.Wrap(err, "Fail to unmarshal alert").With("alert", string(msg))
		}

		logger.WithField("alert", alert).Info("Taking report")
		report, err := repo.TakeReport(alert, now)
		if err != nil {
			return nil, errors.Wrapf(err, "Fail to take reportID for alert").With("alert", alert)
		}
		if report == nil {
			return nil, errors.Wrapf(err, "No report in cache").With("alert", alert)
		}

		logger.WithFields(logrus.Fields{
			"ReportID": report.ID,
			"Status":   report.Status,
			"Error":    err,
			"AlertID":  alert.AlertID(),
		}).Info("ReportID has been retrieved")

		report.Alerts = []deepalert.Alert{alert}

		if err := repo.SaveAlertCache(report.ID, alert, now); err != nil {
			return nil, errors.Wrap(err, "Fail to save alert cache")
		}

		if err := sfnSvc.Exec(args.InspectorMashine, &report); err != nil {
			return nil, errors.Wrap(err, "Fail to execute InspectorDelayMachine")
		}

		if report.IsNew() {
			if err := sfnSvc.Exec(args.ReviewMachine, &report); err != nil {
				return nil, errors.Wrap(err, "Fail to execute ReviewerDelayMachine")
			}
		}

		if err := snsSvc.Publish(args.ReportTopic, &report); err != nil {
			return nil, errors.Wrap(err, "Fail to publish report")
		}
	}

	return nil, nil
}
