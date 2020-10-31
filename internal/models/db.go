package models

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/errors"
)

const (
	// DynamoPKeyName is common name of PartitionKey (HashKey) of DynamoDB
	DynamoPKeyName = "pk"
	// DynamoSKeyName = "sk"
)

type RecordBase struct {
	PKey      string `dynamo:"pk"`
	SKey      string `dynamo:"sk"`
	ExpiresAt int64  `dynamo:"expires_at"`
	CreatedAt int64  `dynamo:"created_at,omitempty"`
}

type AlertEntry struct {
	RecordBase
	ReportID deepalert.ReportID `dynamo:"report_id"`
}

type AlertCache struct {
	RecordBase
	AlertData []byte `dynamo:"alert_data"`
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
	ID     string `dynamo:"id"`
	Result string `dynamo:"result"`
	Status string `dynamo:"status"`
}

// ErrRecordIsNotReport means DynamoDB record is not event of add/modify report.
var ErrRecordIsNotReport = errors.New("Reocrd is not report")

// ImportDynamoRecord copies values from record data in DynamoDB stream
func (x *ReportEntry) ImportDynamoRecord(record *events.DynamoDBEventRecord) error {
	if record.Change.NewImage == nil {
		return ErrRecordIsNotReport
	}

	getString := func(key string) string {
		value, ok := record.Change.NewImage[key]
		if !ok {
			return ""
		}
		return value.String()
	}

	x.ID = getString("id")
	x.Result = getString("result")
	x.Status = getString("status")

	createdAtValue, ok := record.Change.NewImage["created_at"]
	if !ok {
		return errors.New("created_at is not available in DynamDB event").With("record", record)
	}

	v, err := strconv.ParseInt(createdAtValue.Number(), 10, 64)
	if err != nil {
		return errors.Wrap(err, "Failed to parse createdAt of DynamoRecord").
			With("record", record)
	}
	x.CreatedAt = v

	return nil
}

// Import copies values from deepalert.Report to own
func (x *ReportEntry) Import(report *deepalert.Report) error {
	x.ID = string(report.ID)

	raw, err := json.Marshal(report.Result)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal report.Result").With("report", report)
	}
	x.Result = string(raw)

	x.Status = string(report.Status)
	x.CreatedAt = report.CreatedAt.UTC().Unix()

	return nil
}

// Export creates a new deepalert.Report from own values
func (x *ReportEntry) Export() (*deepalert.Report, error) {
	var report deepalert.Report

	report.ID = deepalert.ReportID(x.ID)
	if err := json.Unmarshal([]byte(x.Result), &report.Result); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal reprot.Result").With("entry", *x)
	}

	report.Status = deepalert.ReportStatus(x.Status)
	report.CreatedAt = time.Unix(x.CreatedAt, 0)

	return &report, nil
}
