package models

import (
	"time"

	"github.com/deepalert/deepalert"
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

type ReportSectionRecord struct {
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
