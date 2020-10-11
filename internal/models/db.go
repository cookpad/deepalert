package models

import (
	"encoding/json"
	"time"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/errors"
)

type RecordBase struct {
	PKey      string    `dynamo:"pk"`
	SKey      string    `dynamo:"sk"`
	ExpiresAt int64     `dynamo:"expires_at"`
	CreatedAt time.Time `dynamo:"created_at,omitempty"`
}

type AlertEntry struct {
	RecordBase
	ReportID deepalert.ReportID `dynamo:"report_id"`
}

type AlertCache struct {
	PKey      string `dynamo:"pk"`
	SKey      string `dynamo:"sk"`
	AlertData []byte `dynamo:"alert_data"`
	ExpiresAt int64  `dynamo:"expires_at"`
}

type InspectorReportRecord struct {
	RecordBase
	Data []byte `dynamo:"data"`
}

type AttributeCache struct {
	RecordBase
	Timestamp   time.Time              `dynamo:"timestamp"`
	AttrKey     string                 `dynamo:"attr_key"`
	AttrType    string                 `dynamo:"attr_type"`
	AttrValue   string                 `dynamo:"attr_value"`
	AttrContext deepalert.AttrContexts `dynamo:"attr_context"`
}

type ReportEntry struct {
	RecordBase
	ID         string `dynamo:"id"`
	Alerts     string `dynamo:"alerts"`
	Attributes string `dynamo:"attributes"`
	Sections   string `dynamo:"sections"`
	Result     string `dynamo:"result"`
	Status     string `dynamo:"status"`
	CreatedAt  int64  `dynamo:"created_at"`
}

func (x *ReportEntry) Import(report *deepalert.Report) error {
	x.ID = string(report.ID)

	raw, err := json.Marshal(report.Alerts)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal report.Alerts").With("report", report)
	}
	x.Alerts = string(raw)

	raw, err = json.Marshal(report.Attributes)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal report.Attributes").With("report", report)
	}
	x.Attributes = string(raw)

	raw, err = json.Marshal(report.Sections)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal report.Contents").With("report", report)
	}
	x.Sections = string(raw)

	raw, err = json.Marshal(report.Result)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal report.Result").With("report", report)
	}
	x.Result = string(raw)

	x.Status = string(report.Status)
	x.CreatedAt = report.CreatedAt.UTC().Unix()

	return nil
}

func (x *ReportEntry) Export() (*deepalert.Report, error) {
	var report deepalert.Report

	report.ID = deepalert.ReportID(x.ID)
	if err := json.Unmarshal([]byte(x.Alerts), &report.Alerts); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal reprot.Alerts").With("entry", *x)
	}

	if err := json.Unmarshal([]byte(x.Alerts), &report.Alerts); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal reprot.Alerts").With("entry", *x)
	}

	if err := json.Unmarshal([]byte(x.Attributes), &report.Attributes); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal reprot.Attributes").With("entry", *x)
	}

	if err := json.Unmarshal([]byte(x.Sections), &report.Sections); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal reprot.Contents").With("entry", *x)
	}

	if err := json.Unmarshal([]byte(x.Result), &report.Result); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal reprot.Result").With("entry", *x)
	}

	report.Status = deepalert.ReportStatus(x.Status)
	report.CreatedAt = time.Unix(x.CreatedAt, 0)

	return &report, nil
}
