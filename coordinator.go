package deepalert

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
)

type reportCoordinator struct {
	tableName  string
	region     string
	table      dynamo.Table
	timeToLive time.Duration
}

func newReportCoordinator(tableName, region string) *reportCoordinator {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})
	x := reportCoordinator{
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
	ReportID ReportID `dynamo:"report_id"`
}

func isConditionalCheckErr(err error) bool {
	if ae, ok := err.(awserr.RequestFailure); ok {
		return ae.Code() == "ConditionalCheckFailedException"
	}
	return false
}

func newReportID() ReportID {
	return ReportID(uuid.New().String())
}

func (x *reportCoordinator) takeReportID(alertID string, ts time.Time) (ReportID, error) {
	fixedKey := "Fixed"
	cache := alertEntry{
		recordBase: recordBase{
			PKey:      "alert/" + alertID,
			SKey:      fixedKey,
			ExpiresAt: ts.Add(time.Hour * 3),
		},
		ReportID: newReportID(),
	}

	if err := x.table.Put(cache).If("(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR expires_at < ?", ts).Run(); err != nil {
		if isConditionalCheckErr(err) {
			var existedEntry alertEntry
			if err := x.table.Get("pk", cache.PKey).Range("sk", dynamo.Equal, cache.SKey).One(&existedEntry); err != nil {
				return ReportID(""), errors.Wrapf(err, "Fail to get cached reportID, AlertID=%s", alertID)
			}

			return existedEntry.ReportID, nil
		}

		return ReportID(""), errors.Wrapf(err, "Fail to get cached reportID, AlertID=%s", alertID)
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

func toAlertCacheKey(reportID ReportID) (string, string) {
	return fmt.Sprintf("report/%s", reportID), "cache/" + uuid.New().String()
}

func (x *reportCoordinator) saveAlertCache(reportID ReportID, alert Alert) error {
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

func (x *reportCoordinator) fetchAlertCache(reportID ReportID) ([]Alert, error) {
	pk, _ := toAlertCacheKey(reportID)
	var caches []alertCache
	var alerts []Alert

	if err := x.table.Get("pk", pk).All(&caches); err != nil {
		return nil, errors.Wrapf(err, "Fail to retrieve alertCache: %s", reportID)
	}

	for _, cache := range caches {
		var alert Alert
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

func toReportContentRecord(reportID ReportID, content *ReportContent) (string, string) {
	pk := fmt.Sprintf("content/%s", reportID)
	sk := ""
	if content != nil {
		sk = fmt.Sprintf("%s/%s", content.Attribute.Hash(), content.Author)
	}
	return pk, sk
}

func (x *reportCoordinator) saveReportContent(reportID ReportID, content ReportContent) error {
	pk, sk := toReportContentRecord(reportID, &content)
	record := reportContentRecord{
		recordBase: recordBase{
			PKey:      pk,
			SKey:      sk,
			ExpiresAt: time.Now().UTC().Add(time.Hour * 24),
		},
	}

	if err := x.table.Put(record).Run(); err != nil {
		return errors.Wrap(err, "Fail to put report record")
	}

	return nil
}

func (x *reportCoordinator) fetchReportRecords(reportID ReportID) ([]ReportContent, error) {
	var records []reportContentRecord
	pk, _ := toReportContentRecord(reportID, nil)

	if err := x.table.Get("pk", pk).All(&records); err != nil {
		return nil, errors.Wrap(err, "Fail to fetch report records")
	}

	var contents []ReportContent
	for _, record := range records {
		var content ReportContent
		if err := json.Unmarshal(record.Data, &content); err != nil {
			return nil, errors.Wrapf(err, "Fail to unmarshal report content: %v %s", record, string(record.Data))
		}

		contents = append(contents, content)
	}

	return contents, nil
}
