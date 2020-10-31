package api

import (
	"net/http"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/errors"
	"github.com/gin-gonic/gin"
)

func getReport(c *gin.Context) {
	args := getArguments(c)
	reportID := deepalert.ReportID(c.Param(paramReportID))

	repo, err := args.Repository()
	if err != nil {
		resp(c, errors.Wrap(err, "Failed to create Repository").
			Status(http.StatusInternalServerError))
		return
	}

	report, err := repo.GetReport(reportID)
	if err != nil {
		resp(c, errors.Wrap(err, "Failed GetReport").
			Status(http.StatusInternalServerError))
		return
	}

	resp(c, report)
}

func getReportAlerts(c *gin.Context) {
	args := getArguments(c)
	reportID := deepalert.ReportID(c.Param(paramReportID))

	repo, err := args.Repository()
	if err != nil {
		resp(c, errors.Wrap(err, "Failed to set up Repository").
			Status(http.StatusInternalServerError))
		return
	}

	alerts, err := repo.FetchAlertCache(reportID)
	if err != nil {
		resp(c, errors.Wrap(err, "Failed to fetch alerts").
			Status(http.StatusInternalServerError))
		return
	}
	if alerts == nil {
		resp(c, errors.New("alerts not found").
			Status(http.StatusNotFound))
		return
	}

	resp(c, alerts)
}

func getReportSections(c *gin.Context) {
	args := getArguments(c)
	reportID := deepalert.ReportID(c.Param(paramReportID))

	repo, err := args.Repository()
	if err != nil {
		resp(c, errors.Wrap(err, "Failed to set up Repository").
			Status(http.StatusInternalServerError))
		return
	}

	sections, err := repo.FetchInspectReport(reportID)
	if err != nil {
		resp(c, errors.Wrap(err, "Failed to fetch sections").
			Status(http.StatusInternalServerError))
		return
	}
	if sections == nil {
		resp(c, errors.New("sections not found").
			Status(http.StatusNotFound))
		return
	}

	resp(c, sections)
}

func getReportAttributes(c *gin.Context) {
	args := getArguments(c)
	reportID := deepalert.ReportID(c.Param(paramReportID))

	repo, err := args.Repository()
	if err != nil {
		resp(c, errors.Wrap(err, "Failed to set up Repository").
			Status(http.StatusInternalServerError))
		return
	}

	attributes, err := repo.FetchAttributeCache(reportID)
	if err != nil {
		resp(c, errors.Wrap(err, "Failed to fetch attributes").
			Status(http.StatusInternalServerError))
		return
	}
	if attributes == nil {
		resp(c, errors.New("attributes not found").
			Status(http.StatusNotFound))
		return
	}

	resp(c, attributes)
}
