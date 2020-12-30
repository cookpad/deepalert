package repository

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/adaptor"
	"github.com/deepalert/deepalert/internal/models"
	"github.com/guregu/dynamo"
	"github.com/m-mizutani/golambda"
)

type DynamoDBRepositry struct {
	tableName  string
	region     string
	table      dynamo.Table
	timeToLive time.Duration
}

// NewDynamoDB is constructor of DynamoDBRepositry
func NewDynamoDB(region, tableName string) (adaptor.Repository, error) {
	ssn, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, golambda.WrapError(err, "Failed session.NewSession for DynamoDB").With("region", region)
	}
	db := dynamo.New(ssn)
	x := &DynamoDBRepositry{
		tableName:  tableName,
		region:     region,
		table:      db.Table(tableName),
		timeToLive: time.Hour * 3,
	}

	return x, nil
}

func (x *DynamoDBRepositry) PutAlertEntry(entry *models.AlertEntry, ts time.Time) error {
	cond := "(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR expires_at < ?"
	if err := x.table.Put(entry).If(cond, ts.UTC().Unix()).Run(); err != nil {
		return err
	}

	return nil
}

func (x *DynamoDBRepositry) GetAlertEntry(pk, sk string) (*models.AlertEntry, error) {
	var output models.AlertEntry
	if err := x.table.Get("pk", pk).Range("sk", dynamo.Equal, sk).One(&output); err != nil {
		if err == dynamo.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &output, nil
}

func (x *DynamoDBRepositry) PutAlertCache(cache *models.AlertCache) error {
	if err := x.table.Put(cache).Run(); err != nil {
		return golambda.WrapError(err, "Failed PutAlertCache").With("cache", cache)
	}

	return nil
}

func (x *DynamoDBRepositry) GetAlertCaches(pk string) ([]*models.AlertCache, error) {
	var caches []*models.AlertCache

	if err := x.table.Get("pk", pk).All(&caches); err != nil {
		if err == dynamo.ErrNotFound {
			return nil, nil
		}

		return nil, golambda.WrapError(err, "Failed GetAlertCaches").With("pk", pk)
	}

	return caches, nil
}

func (x *DynamoDBRepositry) PutInspectorReport(record *models.InspectorReportRecord) error {
	if err := x.table.Put(record).Run(); err != nil {
		return golambda.WrapError(err, "Failed PutInspectorReport").With("record", record)
	}

	return nil
}

func (x *DynamoDBRepositry) GetInspectorReports(pk string) ([]*models.InspectorReportRecord, error) {
	var records []*models.InspectorReportRecord

	if err := x.table.Get("pk", pk).All(&records); err != nil {
		return nil, golambda.WrapError(err, "Failed GetInspectorReports").With("pk", pk)
	}

	return records, nil
}

func (x *DynamoDBRepositry) PutAttributeCache(attr *models.AttributeCache, ts time.Time) error {
	if err := x.table.Put(attr).If("(attribute_not_exists(pk) AND attribute_not_exists(sk)) OR expires_at < ?", ts.UTC().Unix()).Run(); err != nil {
		return err
	}

	return nil
}

func (x *DynamoDBRepositry) GetAttributeCaches(pk string) ([]*models.AttributeCache, error) {
	var attrs []*models.AttributeCache

	if err := x.table.Get("pk", pk).All(&attrs); err != nil {
		return nil, golambda.WrapError(err, "Failed GetAttributeCaches").With("pk", pk)
	}

	return attrs, nil
}

func (x *DynamoDBRepositry) PutReport(pk string, report *deepalert.Report) error {
	var entry models.ReportEntry
	if err := entry.Import(report); err != nil {
		return err
	}
	entry.PKey = pk
	entry.SKey = "-"

	if err := x.table.Put(&entry).Run(); err != nil {
		return err
	}
	return nil
}

func (x *DynamoDBRepositry) GetReport(pk string) (*deepalert.Report, error) {
	var entry models.ReportEntry
	if err := x.table.Get("pk", pk).Range("sk", dynamo.Equal, "-").One(&entry); err != nil {
		if err == dynamo.ErrNotFound {
			return nil, nil
		}
		return nil, golambda.WrapError(err, "Failed to get report").With("pk", pk)
	}

	report, err := entry.Export()
	if err != nil {
		return nil, err
	}
	return report, nil
}

// Error handling

func (x *DynamoDBRepositry) IsConditionalCheckErr(err error) bool {
	if ae, ok := err.(awserr.RequestFailure); ok {
		return ae.Code() == "ConditionalCheckFailedException"
	}
	return false
}
