package api

import (
	"net/http"

	"github.com/deepalert/deepalert"
	"github.com/gin-gonic/gin"
	"github.com/m-mizutani/golambda"
)

func getReport(c *gin.Context) {
	args := getArguments(c)
	reportID := deepalert.ReportID(c.Param(paramReportID))

	repo, err := args.Repository()
	if err != nil {
		resp(c, http.StatusInternalServerError, err)
		return
	}

	report, err := repo.GetReport(reportID)
	if err != nil {
		resp(c, http.StatusInternalServerError, golambda.WrapError(err, "Failed GetReport"))
		return
	}

	resp(c, http.StatusOK, report)
}

func getReportAlerts(c *gin.Context) {
	args := getArguments(c)
	reportID := deepalert.ReportID(c.Param(paramReportID))

	repo, err := args.Repository()
	if err != nil {
		resp(c, http.StatusInternalServerError, err)
		return
	}

	alerts, err := repo.FetchAlertCache(reportID)
	if err != nil {
		resp(c, http.StatusInternalServerError, err)
		return
	}
	if alerts == nil {
		resp(c, http.StatusNotFound, golambda.NewError("alerts not found"))
		return
	}

	resp(c, http.StatusOK, alerts)
}

func getSections(c *gin.Context) {
	args := getArguments(c)
	reportID := deepalert.ReportID(c.Param(paramReportID))

	repo, err := args.Repository()
	if err != nil {
		resp(c, http.StatusInternalServerError, err)
		return
	}

	sections, err := repo.FetchSection(reportID)
	if err != nil {
		resp(c, http.StatusInternalServerError, err)
		return
	}
	if sections == nil {
		resp(c, http.StatusNotFound, golambda.NewError("sections not found"))
		return
	}

	resp(c, http.StatusOK, sections)
}

func getReportAttributes(c *gin.Context) {
	args := getArguments(c)
	reportID := deepalert.ReportID(c.Param(paramReportID))

	repo, err := args.Repository()
	if err != nil {
		resp(c, http.StatusInternalServerError, err)
		return
	}

	attributes, err := repo.FetchAttributeCache(reportID)
	if err != nil {
		resp(c, http.StatusInternalServerError, err)
		return
	}
	if attributes == nil {
		resp(c, http.StatusNotFound, golambda.NewError("attributes not found"))
		return
	}

	resp(c, http.StatusOK, attributes)
}
