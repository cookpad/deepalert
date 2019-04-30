package deepalert

import (
	"encoding/json"
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

type alertEntry struct {
	AlertID   string    `dynamo:"pk"`
	SortKey   string    `dynamo:"sk"`
	ReportID  ReportID  `dynamo:"report_id"`
	ExpiresAt time.Time `dynamo:"expires_at"`
}

func isConditionalCheckErr(err error) bool {
	if ae, ok := err.(awserr.RequestFailure); ok {
		return ae.Code() == "ConditionalCheckFailedException"
	}
	return false
}

func (x *reportCoordinator) takeReportID(alertID string, ts time.Time) (ReportID, error) {
	fixedKey := "Fixed"
	cache := alertEntry{
		AlertID:   alertID,
		SortKey:   fixedKey,
		ReportID:  ReportID(uuid.New().String()),
		ExpiresAt: ts.Add(time.Hour * 3),
	}

	if err := x.table.Put(cache).If("(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR expires_at < ?", ts).Run(); err != nil {
		if isConditionalCheckErr(err) {
			var existedEntry alertEntry
			if err := x.table.Get("pk", alertID).Range("sk", dynamo.Equal, fixedKey).One(&existedEntry); err != nil {
				return ReportID(""), errors.Wrapf(err, "Fail to get cached reportID, AlertID=%s", alertID)
			}

			return existedEntry.ReportID, nil
		}

		return ReportID(""), errors.Wrapf(err, "Fail to get cached reportID, AlertID=%s", alertID)
	}

	return cache.ReportID, nil
}

type alertCache struct {
	ReportID  ReportID  `dynamo:"pk"`
	Timestamp time.Time `dynamo:"sk"`
	AlertData []byte    `dynamo:"alert_data"`
	ExpiresAt time.Time `dynamo:"expires_at"`
}

func (x *reportCoordinator) saveReportCache(reportID ReportID, alert Alert) error {
	raw, err := json.Marshal(alert)
	if err != nil {
		return errors.Wrapf(err, "Fail to marshal alert: %v", alert)
	}

	cache := alertCache{
		ReportID:  reportID,
		Timestamp: alert.Timestamp,
		AlertData: raw,
		ExpiresAt: alert.Timestamp.Add(x.timeToLive),
	}

	if err := x.table.Put(cache).Run(); err != nil {
		return errors.Wrap(err, "")
	}

	return nil
}

func (x *reportCoordinator) fetchReportCache(reportID ReportID) ([]Alert, error) {
	return nil, nil
}

func (x *reportCoordinator) saveReport(record *ReportRecord) error {
	record.TimeToLive = time.Now().UTC().Add(time.Hour * 24)
	if err := x.table.Put(record).Run(); err != nil {
		return errors.Wrap(err, "Fail to put report record")
	}

	return nil
}

func (x *reportCoordinator) fetchReportRecords(reportID ReportID) ([]ReportRecord, error) {
	var records []ReportRecord
	pk := reportIDtoRecordKey(reportID)
	if err := x.table.Get("pk", pk).All(&records); err != nil {
		return nil, errors.Wrap(err, "Fail to fetch report records")
	}

	return records, nil
}
