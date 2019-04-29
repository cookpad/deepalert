package deepalert

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"

	"github.com/m-mizutani/deepalert/functions"
)

type ReportID string
type ReportStatus string
type ReportSeverity string

// Report is a container to deliver contents and inspection results of the alert.
type Report struct {
	ID       ReportID       `json:"id"`
	Alert    Alert          `json:"alert"`
	Entities []ReportEntity `json:"entities"`
	Result   ReportResult   `json:"result"`
	Status   ReportStatus   `json:"status"`
}

// ReportEntity contians results of inspectors about user and host.
type ReportEntity interface {
	GetAttr() Attribute
}

// ReportEntityBase is base structure of report entity.
type ReportEntityBase struct {
	Author    string    `json:"author"`
	Attribute Attribute `json:"attribute"`
}

// GetAttr of ReportEntity returns a target attribute.
func (x *ReportEntityBase) GetAttr() Attribute {
	return x.Attribute
}

// Submit of ReportRecord sends record to SNS
func (x *ReportEntityBase) Submit(topicArn, region string) error {
	if err := functions.PublishSNS(topicArn, region, x); err != nil {
		return errors.Wrapf(err, "Fail to publish ReportEntity: %v", *x)
	}

	return nil
}

// ReportResult shows output of Reviewer invoked to evaluate risk of the alert.
type ReportResult struct {
	Severity ReportSeverity `json:"severity"`
	Reason   string         `json:"reason"`
}

// IsNew returns status of the report
func (x *Report) IsNew() bool { return x.Status == StatusNew }

// IsPublished returns status of the report
func (x *Report) IsPublished() bool { return x.Status == StatusPublished }

const (
	StatusNew       ReportStatus = "new"
	StatusPublished              = "published"
)

// -----------------------------------------------
// Entities

// ReportUser describes a user indicator on remote services.
type ReportUser struct {
	ReportEntityBase
	Activities []EntityActivity `json:"activities"`
}

// ReportBinary describes a binary file indicator including executable format.
type ReportBinary struct {
	ReportEntityBase
	RelatedMalware []EntityMalware  `json:"related_malware,omitempty"`
	Software       []string         `json:"software,omitempty"`
	OS             []string         `json:"os,omitempty"`
	Activities     []EntityActivity `json:"activities,omitempty"`
}

// ReportHost describes a host indicator binding IP address, domain name
type ReportHost struct {
	ReportEntityBase

	// Network related entities
	IPAddr         []string         `json:"ipaddr,omitempty"`
	Country        []string         `json:"country,omitempty"`
	ASOwner        []string         `json:"as_owner,omitempty"`
	RelatedDomains []EntityDomain   `json:"related_domains,omitempty"`
	RelatedURLs    []EntityURL      `json:"related_urls,omitempty"`
	RelatedMalware []EntityMalware  `json:"related_malware,omitempty"`
	Activities     []EntityActivity `json:"activities,omitempty"`

	// Internal environment
	UserName []string         `json:"username,omitempty"`
	Owner    []string         `json:"owner,omitempty"`
	OS       []string         `json:"os,omitempty"`
	MACAddr  []string         `json:"macaddr,omitempty"`
	HostName []string         `json:"hostname,omitempty"`
	Software []EntitySoftware `json:"software,omitempty"`
}

// -----------------------------------------------
// Entity Objects

type EntityActivity struct {
	ServiceName string    `json:"service_name"`
	RemoteAddr  string    `json:"remote_addr"`
	Principal   string    `json:"principal"`
	Action      string    `json:"action"`
	Target      string    `json:"target"`
	LastSeen    time.Time `json:"last_seen"`
}

type EntityMalware struct {
	SHA256    string              `json:"sha256"`
	Timestamp time.Time           `json:"timestamp"`
	Scans     []EntityMalwareScan `json:"scans"`
	Relation  string              `json:"relation"`
}

type EntityMalwareScan struct {
	Vendor   string `json:"vendor"`
	Name     string `json:"name"`
	Positive bool   `json:"positive"`
	Source   string `json:"source"`
}

type EntityDomain struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

type EntityURL struct {
	URL       string    `json:"url"`
	Reference string    `json:"reference"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

type EntitySoftware struct {
	Name     string    `json:"name"`
	Location string    `json:"location"`
	LastSeen time.Time `json:"last_seen"`
}

// ---------------------------------------------
// Component

// ReportRecord is a format to store a report to DynamoDB
type ReportRecord struct {
	ReportID   ReportID  `dynamo:"report_id"`
	AttrID     string    `dynamo:"attr_id"`
	Data       []byte    `dynamo:"data"`
	TimeToLive time.Time `dynamo:"ttl"`
}

// NewReportRecord is a constructor of ReportRecord
func NewReportRecord(id ReportID, entity ReportEntity) *ReportRecord {
	raw, err := json.Marshal(entity)
	if err != nil {
		// Must success
		log.Fatal("Fail to unmarshal ReportEntity.", err)
	}

	rec := ReportRecord{
		ReportID: id,
		AttrID:   entity.GetAttr().Hash(),
		Data:     raw,
	}

	return &rec
}

func (x *ReportRecord) Save(tableName, region string) error {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})
	table := db.Table(tableName)

	x.TimeToLive = time.Now().UTC().Add(time.Hour * 24)
	if err := table.Put(x).Run(); err != nil {
		return errors.Wrap(err, "Fail to put report record")
	}

	return nil
}

func reportIDtoRecordKey(reportID ReportID) string {
	return fmt.Sprintf("record/%s", reportID)
}

func FetchReportRecords(tableName, region string, reportID ReportID) ([]ReportRecord, error) {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})
	table := db.Table(tableName)

	var records []ReportRecord
	pk := reportIDtoRecordKey(reportID)
	if err := table.Get("pk", pk).All(&records); err != nil {
		return nil, errors.Wrap(err, "Fail to fetch report records")
	}

	return records, nil
}

/*
// NewReportComponent is a constructor of ReportComponent
func NewReportComponent(reportID ReportID) *ReportComponent {
	data := ReportComponent{
		ReportID: reportID,
		DataID:   uuid.NewV4().String(),
	}

	return &data
}

// SetPage sets page data with serialization.
func (x *ReportComponent) SetPage(page ReportPage) {
	data, err := json.Marshal(&page)
	if err != nil {
		log.Println("Fail to marshal report page:", page)
	}

	x.Data = data
}

// Page returns deserialized page structure
func (x *ReportComponent) Page() *ReportPage {
	if len(x.Data) == 0 {
		return nil
	}

	var page ReportPage
	err := json.Unmarshal(x.Data, &page)
	if err != nil {
		log.Println("Invalid report page data foramt", string(x.Data))
		return nil
	}

	return &page
}

func (x *ReportComponent) Submit(tableName, region string) error {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})
	table := db.Table(tableName)

	x.TimeToLive = time.Now().UTC().Add(time.Second * 864000)

	log.WithFields(log.Fields{
		"component": x,
		"tableName": tableName,
	}).Info("Put component")
	err := table.Put(x).Run()
	if err != nil {
		return errors.Wrap(err, "Fail to put report data")
	}

	return nil
}

func FetchReportPages(tableName, region string, reportID ReportID) ([]*ReportPage, error) {
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})
	table := db.Table(tableName)

	dataList := []ReportComponent{}
	err := table.Get("report_id", reportID).All(&dataList)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to fetch report data")
	}

	pages := []*ReportPage{}
	for _, data := range dataList {
		pages = append(pages, data.Page())
	}
	return pages, nil
}

func NewReport(reportID ReportID, alert Alert) Report {
	report := Report{
		ID:      reportID,
		Alert:   alert,
		Content: newReportContent(),
	}

	return report
}

func NewReportID() ReportID {
	return ReportID(uuid.NewV4().String())
}
*/
