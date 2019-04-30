package deepalert

import (
	"time"

	"github.com/pkg/errors"

	"github.com/m-mizutani/deepalert/functions"
)

type ReportID string
type ReportStatus string
type ReportSeverity string
type ReportContentType string

// Report is a container to deliver contents and inspection results of the alert.
type Report struct {
	ID       ReportID        `json:"id"`
	Alert    Alert           `json:"alert"`
	Contents []ReportContent `json:"entities"`
	Result   ReportResult    `json:"result"`
	Status   ReportStatus    `json:"status"`
}

// ReportContent is base structure of report entity.
type ReportContent struct {
	ReportID  ReportID          `json:"report_id"`
	Author    string            `json:"author"`
	Attribute Attribute         `json:"attribute"`
	Type      ReportContentType `json:"type"`
	Content   interface{}       `json:"content"`
}

const (
	ContentUser   ReportContentType = "user"
	ContentHost                     = "host"
	ContentBinary                   = "binary"
)

// SubmitReportContent sends record to SNS
func SubmitReportContent(content ReportContent, topicArn, region string) error {
	if err := functions.PublishSNS(topicArn, region, content); err != nil {
		return errors.Wrapf(err, "Fail to publish ReportEntity: %v", content)
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
	Activities []EntityActivity `json:"activities"`
}

// ReportBinary describes a binary file indicator including executable format.
type ReportBinary struct {
	RelatedMalware []EntityMalware  `json:"related_malware,omitempty"`
	Software       []string         `json:"software,omitempty"`
	OS             []string         `json:"os,omitempty"`
	Activities     []EntityActivity `json:"activities,omitempty"`
}

// ReportHost describes a host indicator binding IP address, domain name
type ReportHost struct {
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
