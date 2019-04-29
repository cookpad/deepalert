package deepalert

import (
	"time"
)

// AttrType shows type of alert attribute.
type AttrType string

const (
	TypeIPAddr     AttrType = "ipaddr"
	TypeDomainName          = "domain"
	TypeUserName            = "username"
)

// AttrContext describes context of the attribute.
type AttrContext string

const (
	CtxRemote  AttrContext = "remote"
	CtxLocal               = "local"
	CtxSubject             = "subject"
	CtxObject              = "object"
)

// Attribute is element of alert
type Attribute struct {
	Type    AttrType      `json:"type"`
	Value   string        `json:"value"`
	Key     string        `json:"key"`
	Context []AttrContext `json:"context"`
}

// Alert is extranted data from KinesisStream
type Alert struct {
	Detector    string `json:"detector"`
	RuleName    string `json:"rule_name"`
	AlertKey    string `json:"alert_key"`
	Description string `json:"description"`

	Timestamp  time.Time   `json:"timestamp"`
	Attributes []Attribute `json:"attributes"`
	Body       interface{} `json:"body,omitempty"`
}

// AddAttribute just appends the attribute to the Alert
func (x *Alert) AddAttribute(attr Attribute) {
	x.Attributes = append(x.Attributes, attr)
}

// AddAttributes appends set of attribute to the Alert
func (x *Alert) AddAttributes(attrs []Attribute) {
	x.Attributes = append(x.Attributes, attrs...)
}

// FindAttributes searches and returns matched attributes
func (x *Alert) FindAttributes(key string) []Attribute {
	var attrs []Attribute
	for _, attr := range x.Attributes {
		if attr.Key == key {
			attrs = append(attrs, attr)
		}
	}

	return attrs
}

// Match checks attribute type and context.
func (x *Attribute) Match(context AttrContext, attrType AttrType) bool {
	if x.Type != attrType {
		return false
	}

	for _, ctx := range x.Context {
		if ctx == context {
			return true
		}
	}

	return false
}
