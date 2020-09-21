package api

import (
	"github.com/deepalert/deepalert"
	"github.com/gin-gonic/gin"
)

func getReportAlerts(c *gin.Context) {
	args := getArguments(c)
	reportID := deepalert.ReportID(c.Param(paramReportID))

	repo, err := args.Repository()
	if err != nil {
		resp(c, wrapSystemError(err, "Failed to set up Repository"))
		return
	}

	alerts, err := repo.FetchAlertCache(reportID)
	if err != nil {
		resp(c, wrapSystemError(err, "Failed to fetch alerts"))
		return
	}
	if alerts == nil {
		resp(c, newUserError("alerts not found").SetStatusCode(404))
		return
	}

	resp(c, alerts)
	return
}

func getReportSections(c *gin.Context) {
	args := getArguments(c)
	reportID := deepalert.ReportID(c.Param(paramReportID))

	repo, err := args.Repository()
	if err != nil {
		resp(c, wrapSystemError(err, "Failed to set up Repository"))
		return
	}

	sections, err := repo.FetchReportSection(reportID)
	if err != nil {
		resp(c, wrapSystemError(err, "Failed to fetch sections"))
		return
	}
	if sections == nil {
		resp(c, newUserError("sections not found").SetStatusCode(404))
		return
	}

	resp(c, sections)
	return
}

func getReportAttributes(c *gin.Context) {
	args := getArguments(c)
	reportID := deepalert.ReportID(c.Param(paramReportID))

	repo, err := args.Repository()
	if err != nil {
		resp(c, wrapSystemError(err, "Failed to set up Repository"))
		return
	}

	attributes, err := repo.FetchAttributeCache(reportID)
	if err != nil {
		resp(c, wrapSystemError(err, "Failed to fetch attributes"))
		return
	}
	if attributes == nil {
		resp(c, newUserError("attributes not found").SetStatusCode(404))
		return
	}

	resp(c, attributes)
	return
}
