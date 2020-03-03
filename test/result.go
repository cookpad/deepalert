package test

import (
	"fmt"
	"math"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/m-mizutani/deepalert"
	"github.com/pkg/errors"
)

// Repository is accessor to DynamoDB
type Repository struct {
	table dynamo.Table
}

type result interface {
	setKeys(pk, sk string)
}

type baseResult struct {
	PKey string `dynamo:"pk"`
	SKey string `dynamo:"sk"`
	TTL  int64  `dynamo:"ttl"`
}

func (x *baseResult) setKeys(pk, sk string) {
	x.PKey = pk
	x.SKey = sk
	x.TTL = time.Now().UTC().Add(time.Hour * 12).Unix()
}

// NewRepository is constructor of repository to access DynamoDB
func NewRepository(region, tableName string) *Repository {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})

	return &Repository{
		table: db.Table(tableName),
	}
}

func toDynamoKeys(alertKey, testTable string) (string, string) {
	return fmt.Sprintf("alert:%s", alertKey), testTable
}

// InspectorResult is record of Inspector test
type InspectorResult struct {
	baseResult
	Key   string `dynamo:"key"`
	Value string `dynamo:"value"`
}

// PutInspectorResult puts InspectorResult to DynamoDB
func (x *Repository) PutInspectorResult(id deepalert.ReportID, key, value string) error {
	values := InspectorResult{
		Key:   key,
		Value: value,
	}
	return x.put("inspector:"+string(id), key, &values)
}

// GetInspectorResult looks up InspectorResult from DynamoDB
func (x *Repository) GetInspectorResult(id deepalert.ReportID, key string) (*InspectorResult, error) {
	var values InspectorResult
	if err := x.get("inspector:"+string(id), key, &values); err != nil {
		return nil, err
	}
	return &values, nil
}

// EmitterResult is record of Emitter test
type EmitterResult struct {
	baseResult
	Timestamp time.Time `dynamo:"timestamp"`
}

// PutEmitterResult puts EmitterResult to DynamoDB
func (x *Repository) PutEmitterResult(reportID deepalert.ReportID) error {
	values := EmitterResult{
		Timestamp: time.Now().UTC(),
	}
	return x.put("emitter:"+string(reportID), "@", &values)
}

// GetEmitterResult looks up EmitterResult from DynamoDB
func (x *Repository) GetEmitterResult(reportID deepalert.ReportID) (*EmitterResult, error) {
	var values EmitterResult
	if err := x.get("emitter:"+string(reportID), "@", &values); err != nil {
		return nil, err
	}
	return &values, nil
}

// Internal methods
func (x *Repository) put(pk, sk string, res result) error {
	res.setKeys(pk, sk)

	if err := x.table.Put(res).Run(); err != nil {
		return errors.Wrapf(err, "Fail to put result: %v", res)
	}

	return nil
}

func (x *Repository) get(pk, sk string, res result) error {
	const maxRetry = 30
	start := time.Now()

	for i := 0; i < maxRetry; i++ {
		if err := x.table.Get("pk", pk).Range("sk", dynamo.Equal, sk).One(res); err != nil {
			if err != dynamo.ErrNotFound {
				return errors.Wrapf(err, "Fail to get result: %v + %v", pk, sk)
			}

			sleep := math.Pow(1.1, float64(i))
			if sleep > 20 {
				sleep = 20
			}
			time.Sleep(time.Second * time.Duration(sleep))
		} else {
			return nil
		}
	}

	end := time.Now()

	return fmt.Errorf("Timeout to get value (waited %f sec): %v + %v", end.Sub(start).Seconds(), pk, sk)
}
