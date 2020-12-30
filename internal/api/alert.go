package api

import (
	"net/http"
	"time"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/m-mizutani/golambda"
)

func postAlert(c *gin.Context) {
	args := getArguments(c)
	now := time.Now().UTC()

	var alert deepalert.Alert
	if err := c.BindJSON(&alert); err != nil {
		resp(c, http.StatusBadRequest, err)
		return
	}

	report, err := usecase.HandleAlert(args, &alert, now)
	if err != nil {
		resp(c, http.StatusInternalServerError, err)
		return
	}

	resp(c, http.StatusOK, report)
}

func getReportByAlertID(c *gin.Context) {
	args := getArguments(c)
	repo, err := args.Repository()
	if err != nil {
		resp(c, http.StatusInternalServerError, err)
		return
	}

	alertID := c.Param(paramAlertID)
	reportID, err := repo.GetReportID(alertID)
	if err != nil {
		resp(c, http.StatusInternalServerError, err)
		return
	}
	if reportID == deepalert.NullReportID {
		resp(c, http.StatusNotFound, golambda.NewError("No such alert").With("alert_id", alertID))
		return
	}

	report, err := repo.GetReport(reportID)
	if err != nil {
		resp(c, http.StatusInternalServerError, err)
		return
	}
	if report == nil {
		resp(c, http.StatusNotFound, golambda.NewError("No such report").With("alert_id", alertID).With("reportID", reportID))
		return
	}

	resp(c, http.StatusOK, report)
}
