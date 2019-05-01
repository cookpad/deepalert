package deepalert

import "time"

var (
	NewDataStoreService = newDataStoreService
	NewReportID         = newReportID
)

func TakeReportID(x *dataStoreService, alertID string, ts time.Time) (ReportID, error) {
	return x.takeReportID(alertID, ts)
}

func SaveAlertCache(x *dataStoreService, reportID ReportID, alert Alert) error {
	return x.saveAlertCache(reportID, alert)
}

func FetchAlertCache(x *dataStoreService, reportID ReportID) ([]Alert, error) {
	return x.fetchAlertCache(reportID)
}

func SaveReportContent(x *dataStoreService, content ReportContent) error {
	return x.saveReportContent(content)
}

func FetchReportContent(x *dataStoreService, reportID ReportID) ([]ReportContent, error) {
	return x.fetchReportContent(reportID)
}
