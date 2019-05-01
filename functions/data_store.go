package functions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"

	"github.com/m-mizutani/deepalert"
)

type DataStoreService struct {
	tableName  string
	region     string
	table      dynamo.Table
	timeToLive time.Duration
}

func NewDataStoreService(tableName, region string) *DataStoreService {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})
	x := DataStoreService{
		tableName:  tableName,
		region:     region,
		table:      db.Table(tableName),
		timeToLive: time.Hour * 3,
	}

	return &x
}

type recordBase struct {
	PKey      string    `dynamo:"pk"`
	SKey      string    `dynamo:"sk"`
	ExpiresAt time.Time `dynamo:"expires_at"`
}

// -----------------------------------------------------------
// Control alertEntry to manage AlertID to ReportID mapping
//
type alertEntry struct {
	recordBase
	ReportID deepalert.ReportID `dynamo:"report_id"`
}

func isConditionalCheckErr(err error) bool {
	if ae, ok := err.(awserr.RequestFailure); ok {
		return ae.Code() == "ConditionalCheckFailedException"
	}
	return false
}

func NewReportID() deepalert.ReportID {
	return deepalert.ReportID(uuid.New().String())
}

func (x *DataStoreService) TakeReportID(alert deepalert.Alert) (deepalert.ReportID, error) {
	fixedKey := "Fixed"
	nullID := deepalert.ReportID("")
	alertID := alert.AlertID()
	ts := alert.Timestamp

	cache := alertEntry{
		recordBase: recordBase{
			PKey:      "alertmap/" + alertID,
			SKey:      fixedKey,
			ExpiresAt: ts.Add(time.Hour * 3),
		},
		ReportID: NewReportID(),
	}

	if err := x.table.Put(cache).If("(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR expires_at < ?", ts).Run(); err != nil {
		if isConditionalCheckErr(err) {
			var existedEntry alertEntry
			if err := x.table.Get("pk", cache.PKey).Range("sk", dynamo.Equal, cache.SKey).One(&existedEntry); err != nil {
				return nullID, errors.Wrapf(err, "Fail to get cached reportID, AlertID=%s", alertID)
			}

			return existedEntry.ReportID, nil
		}

		return nullID, errors.Wrapf(err, "Fail to get cached reportID, AlertID=%s", alertID)
	}

	return cache.ReportID, nil
}

// -----------------------------------------------------------
// Control alertCache to manage published alert data
//
type alertCache struct {
	PKey      string    `dynamo:"pk"`
	SKey      string    `dynamo:"sk"`
	AlertData []byte    `dynamo:"alert_data"`
	ExpiresAt time.Time `dynamo:"expires_at"`
}

func toAlertCacheKey(reportID deepalert.ReportID) (string, string) {
	return fmt.Sprintf("alert/%s", reportID), "cache/" + uuid.New().String()
}

func (x *DataStoreService) SaveAlertCache(reportID deepalert.ReportID, alert deepalert.Alert) error {
	raw, err := json.Marshal(alert)
	if err != nil {
		return errors.Wrapf(err, "Fail to marshal alert: %v", alert)
	}

	pk, sk := toAlertCacheKey(reportID)
	cache := alertCache{
		PKey:      pk,
		SKey:      sk,
		AlertData: raw,
		ExpiresAt: alert.Timestamp.Add(x.timeToLive),
	}

	if err := x.table.Put(cache).Run(); err != nil {
		return errors.Wrap(err, "")
	}

	return nil
}

func (x *DataStoreService) FetchAlertCache(reportID deepalert.ReportID) ([]deepalert.Alert, error) {
	pk, _ := toAlertCacheKey(reportID)
	var caches []alertCache
	var alerts []deepalert.Alert

	if err := x.table.Get("pk", pk).All(&caches); err != nil {
		return nil, errors.Wrapf(err, "Fail to retrieve alertCache: %s", reportID)
	}

	for _, cache := range caches {
		var alert deepalert.Alert
		if err := json.Unmarshal(cache.AlertData, &alert); err != nil {
			return nil, errors.Wrapf(err, "Fail to unmarshal alert: %s", string(cache.AlertData))
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// -----------------------------------------------------------
// Control reportRecord to manage report contents by inspector
//
type reportContentRecord struct {
	recordBase
	Data []byte `dynamo:"data"`
}

func toReportContentRecord(reportID deepalert.ReportID, content *deepalert.ReportContent) (string, string) {
	pk := fmt.Sprintf("content/%s", reportID)
	sk := ""
	if content != nil {
		sk = fmt.Sprintf("%s/%s", content.Attribute.Hash(), content.Author)
	}
	return pk, sk
}

func (x *DataStoreService) SaveReportContent(content deepalert.ReportContent) error {
	raw, err := json.Marshal(content)
	if err != nil {
		return errors.Wrapf(err, "Fail to marshal ReportContent: %v", content)
	}

	pk, sk := toReportContentRecord(content.ReportID, &content)
	record := reportContentRecord{
		recordBase: recordBase{
			PKey:      pk,
			SKey:      sk,
			ExpiresAt: time.Now().UTC().Add(time.Hour * 24),
		},
		Data: raw,
	}

	if err := x.table.Put(record).Run(); err != nil {
		return errors.Wrap(err, "Fail to put report record")
	}

	return nil
}

func (x *DataStoreService) FetchReportContent(reportID deepalert.ReportID) ([]deepalert.ReportContent, error) {
	var records []reportContentRecord
	pk, _ := toReportContentRecord(reportID, nil)

	if err := x.table.Get("pk", pk).All(&records); err != nil {
		return nil, errors.Wrap(err, "Fail to fetch report records")
	}

	var contents []deepalert.ReportContent
	for _, record := range records {
		var content deepalert.ReportContent
		if err := json.Unmarshal(record.Data, &content); err != nil {
			return nil, errors.Wrapf(err, "Fail to unmarshal report content: %v %s", record, string(record.Data))
		}

		contents = append(contents, content)
	}

	return contents, nil
}
