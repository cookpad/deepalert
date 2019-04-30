package deepalert

import "time"

var (
	NewReportCoordinator = newReportCoordinator
)

func TakeReportID(x *reportCoordinator, alertID string, ts time.Time) (ReportID, error) {
	return x.takeReportID(alertID, ts)
}
