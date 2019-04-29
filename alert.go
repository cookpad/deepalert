package deepalert

import (
	"time"
)

// Attribute is element of alert
type Attribute struct {
	Type    string   `json:"type"`
	Value   string   `json:"value"`
	Key     string   `json:"key"`
	Context []string `json:"context"`
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
func (x *Attribute) Match(context, attrType string) bool {
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
