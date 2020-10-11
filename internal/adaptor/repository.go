package adaptor

import (
	"time"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/models"
)

// RepositoryFactory is interface Repository constructor
type RepositoryFactory func(region, tableName string) Repository

// Repository is interface of AWS SDK SQS
type Repository interface {
	PutAlertEntry(entry *models.AlertEntry, ts time.Time) error
	GetAlertEntry(pk, sk string) (*models.AlertEntry, error)
	PutAlertCache(cache *models.AlertCache) error
	GetAlertCaches(pk string) ([]*models.AlertCache, error)
	PutInspectorReport(record *models.InspectorReportRecord) error
	GetInspectorReports(pk string) ([]*models.InspectorReportRecord, error)
	PutAttributeCache(attr *models.AttributeCache, ts time.Time) error
	GetAttributeCaches(pk string) ([]*models.AttributeCache, error)
	PutReport(pk string, report *deepalert.Report) error
	GetReport(pk string) (*deepalert.Report, error)

	IsConditionalCheckErr(err error) bool
}

// NewRepository creates actual AWS SFn SDK client
func NewRepository(region, tableName string) Repository {
	return nil
}
