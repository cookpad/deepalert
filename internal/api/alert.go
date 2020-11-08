package api

import (
	"net/http"
	"time"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/errors"
	"github.com/deepalert/deepalert/internal/usecase"
	"github.com/gin-gonic/gin"
)

func postAlert(c *gin.Context) {
	args := getArguments(c)
	now := time.Now().UTC()

	var alert deepalert.Alert
	if err := c.BindJSON(&alert); err != nil {
		resp(c, errors.Wrap(err, "Failed to pase deepalert.Alert").Status(http.StatusBadRequest))
		return
	}

	report, err := usecase.HandleAlert(args, &alert, now)
	if err != nil {
		resp(c, err)
		return
	}

	resp(c, report)
}

func getReportByAlertID(c *gin.Context) {
	args := getArguments(c)
	repo, err := args.Repository()
	if err != nil {
		resp(c, errors.Wrap(err, "Failed to create Repository").
			Status(http.StatusInternalServerError))
		return
	}

	alertID := c.Param(paramAlertID)
	reportID, err := repo.GetReportID(alertID)
	if err != nil {
		resp(c, errors.Wrap(err, "Failed repository access").
			Status(http.StatusInternalServerError))
		return
	}
	if reportID == deepalert.NullReportID {
		resp(c, errors.New("No such alert").
			With("alert_id", alertID).
			Status(http.StatusNotFound))
		return
	}

	report, err := repo.GetReport(reportID)
	if err != nil {
		resp(c, errors.Wrap(err, "Failed GetReport").
			Status(http.StatusInternalServerError))
		return
	}
	if report == nil {
		resp(c, errors.Wrap(err, "Report data is not found, still in transaction").Status(http.StatusNotFound))
		return
	}

	resp(c, report)
}
