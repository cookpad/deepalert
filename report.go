package deepalert

import (
	"time"
)

// ReportID is a unique ID of a report. Multiple alerts can be aggregated to
// one report by same Detector, RuleName and AlertKey.
type ReportID string

// NullReportID means not available ID
const NullReportID ReportID = ""

// ReportStatus shows "new" or "published". "new" means that the report have
// not been reviewed by Reviewer and inspection may be still ongoing.
// "publihsed" means that the report is already submitted to ReportNotification
// as a reviwed report.
type ReportStatus string

// ReportSeverity has three statuses: "safe", "unclassified", "urgent".
// - "safe": Reviewer determined the alert has no or minimal risk.
//           E.g. Win32 malware is detected in a host, but the host's OS is MacOS.
// - "unclassified": Reviewer has no suitable policy or can not determine risk.
// - "urgent": The alert has a big impact and a security operator must
//             respond it immediately.
type ReportSeverity string

// ReportContentType shows "user", "host" or "binary". It helps to parse
// Content field in ReportContnet.
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
	// ContentUser means Content field is ReportUser.
	ContentUser ReportContentType = "user"
	// ContentHost means Content field is ReportHost.
	ContentHost = "host"
	// ContentBinary means Content field is ReportBinary.
	ContentBinary = "binary"
)

// ReportResult shows output of Reviewer invoked to evaluate risk of the alert.
type ReportResult struct {
	Severity ReportSeverity `json:"severity"`
	Reason   string         `json:"reason"`
}

const (
	// SevSafe : Reviewer determined the alert has no or minimal risk.
	// E.g. Win32 malware is detected in a host, but the host's OS is MacOS.
	SevSafe ReportSeverity = "safe"
	// SevUnclassified : Reviewer has no suitable policy or can not determine risk.
	SevUnclassified = "unclassified"
	// SevUrgent : The alert has a big impact and a security operator must respond it immediately.
	SevUrgent = "urgent"
)

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

// EntityActivity shows history of user/host activity such as accessing a cloud service.
type EntityActivity struct {
	ServiceName string    `json:"service_name"`
	RemoteAddr  string    `json:"remote_addr"`
	Principal   string    `json:"principal"`
	Action      string    `json:"action"`
	Target      string    `json:"target"`
	LastSeen    time.Time `json:"last_seen"`
}

// EntityMalware shows set of malware scan result by AntiVirus software.
type EntityMalware struct {
	SHA256    string              `json:"sha256"`
	Timestamp time.Time           `json:"timestamp"`
	Scans     []EntityMalwareScan `json:"scans"`
	Relation  string              `json:"relation"`
}

// EntityMalwareScan shows a result of malware scan.
type EntityMalwareScan struct {
	Vendor   string `json:"vendor"`
	Name     string `json:"name"`
	Positive bool   `json:"positive"`
	Source   string `json:"source"`
}

// EntityDomain shows a related domain to the host.
type EntityDomain struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// EntityURL shows a related URL to the host.
type EntityURL struct {
	URL       string    `json:"url"`
	Reference string    `json:"reference"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// EntitySoftware shows installed software to the host.
type EntitySoftware struct {
	Name     string    `json:"name"`
	Location string    `json:"location"`
	LastSeen time.Time `json:"last_seen"`
}
