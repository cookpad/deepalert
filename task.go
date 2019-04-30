package deepalert

// Task is invoke argument of inspectors
type Task struct {
	ReportID  ReportID  `json:"report_id"`
	Attribute Attribute `json:"attribute"`
}
