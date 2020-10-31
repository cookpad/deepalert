package workflow

import (
	"encoding/json"
	"math"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/errors"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
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
	Data string `dynamo:"data"`
}

func (x *baseResult) setKeys(pk, sk string) {
	x.PKey = pk
	x.SKey = sk
	x.TTL = time.Now().UTC().Add(time.Hour * 12).Unix()
}

// NewRepository is constructor of repository to access DynamoDB
func NewRepository(region, tableName string) (*Repository, error) {
	ssn, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, errors.Wrap(err, "Failed session.NewSession")
	}

	db := dynamo.New(ssn, &aws.Config{Region: aws.String(region)})

	return &Repository{
		table: db.Table(tableName),
	}, nil
}

// EmitterResult is record of Emitter test
type EmitterResult struct {
	baseResult
	Timestamp time.Time `dynamo:"timestamp"`
}

// PutEmitterResult puts EmitterResult to DynamoDB
func (x *Repository) PutEmitterResult(report *deepalert.Report) error {
	raw, err := json.Marshal(report)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal report").With("report", report)
	}

	value := EmitterResult{
		Timestamp: time.Now().UTC(),
		baseResult: baseResult{
			PKey: "emitter/" + string(report.ID),
			SKey: uuid.New().String(),
			Data: string(raw),
		},
	}

	if err := x.table.Put(value).Run(); err != nil {
		return errors.Wrap(err, "Fail to put result").With("value", value)
	}

	return nil
}

// GetEmitterResult looks up EmitterResult from DynamoDB
func (x *Repository) GetEmitterResult(reportID deepalert.ReportID) ([]*EmitterResult, error) {
	var values []*EmitterResult

	const maxRetry = 30
	start := time.Now()
	pk := "emitter/" + string(reportID)

	for i := 0; i < maxRetry; i++ {
		if err := x.table.Get("pk", pk).All(&values); err != nil {
			if err != dynamo.ErrNotFound {
				return nil, errors.Wrap(err, "Fail to get result").With("pk", pk)
			}

			sleep := math.Pow(1.1, float64(i))
			if sleep > 20 {
				sleep = 20
			}
			time.Sleep(time.Second * time.Duration(sleep))
		} else {
			return values, nil
		}
	}

	end := time.Now()
	sec := end.Sub(start).Seconds()
	return nil, errors.New("Timeout to get value").With("waited", sec).With("pk", pk)
}
