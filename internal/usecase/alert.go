package usecase

import (
	"net/http"
	"time"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/errors"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

// HandleAlert creates a report from alert and invoke delay machines
func HandleAlert(args *handler.Arguments, alert *deepalert.Alert, now time.Time) (*deepalert.Report, error) {
	logger.With("alert", alert).Info("Taking report")

	if err := alert.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid alert format").
			Status(http.StatusBadRequest)
	}

	sfnSvc := args.SFnService()
	repo, err := args.Repository()
	if err != nil {
		return nil, err
	}

	report, err := repo.TakeReport(*alert, now)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to take reportID for alert").With("alert", alert)
	}
	if report == nil {
		return nil, errors.Wrap(err, "No report in cache").
			With("alert", alert)

	}

	logger.
		With("ReportID", report.ID).
		With("Status", report.Status).
		With("Error", err).
		With("AlertID", alert.AlertID()).
		Info("ReportID has been retrieved")

	report.Alerts = []*deepalert.Alert{alert}

	if err := repo.SaveAlertCache(report.ID, *alert, now); err != nil {
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

	if err := repo.PutReport(report); err != nil {
		return nil, errors.Wrap(err, "Fail PutReport")

	}

	return report, nil
}
