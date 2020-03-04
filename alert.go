package deepalert

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

// AttrType shows type of alert attribute.
type AttrType string

const (
	TypeIPAddr        AttrType = "ipaddr"
	TypeDomainName             = "domain"
	TypeUserName               = "username"
	TypeFileHashValue          = "filehashvalue"
	TypeJSON                   = "json"
	TypeURL                    = "url"
)

// AttrContext describes context of the attribute.
type AttrContext string

const (
	CtxRemote  AttrContext = "remote"
	CtxLocal               = "local"
	CtxSubject             = "subject"
	CtxObject              = "object"
	CtxClient              = "client"
	CtxServer              = "server"
	CtxFile                = "file"

	// Meta contexts
	CtxAdditionalInfo = "additional"
)

// Attribute is element of alert
type Attribute struct {
	Type AttrType `json:"type"`

	// Key should be unique in alert.Attributes, but not must.
	Key string `json:"key"`

	// Value is main value of attribute.
	Value string `json:"value"`

	// Context explains background of the attribute value.
	Context []AttrContext `json:"context"`

	// Timestamp indicates observed time of the attribute.
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

// Alert is extranted data from KinesisStream
type Alert struct {
	Detector    string `json:"detector"`
	RuleName    string `json:"rule_name"`
	RuleID      string `json:"rule_id"`
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

// AlertID calculate ID of the alert from Detector, RuleID and AlertKey.
func (x *Alert) AlertID() string {
	key := strings.Join([]string{
		base64.StdEncoding.EncodeToString([]byte(x.Detector)),
		base64.StdEncoding.EncodeToString([]byte(x.RuleID)),
		base64.StdEncoding.EncodeToString([]byte(x.AlertKey)),
	}, ":")

	hasher := sha256.New()
	hasher.Write([]byte(key))
	return fmt.Sprintf("alert:%x", hasher.Sum(nil))
}

// Match checks attribute type and context.
func (x *Attribute) Match(context AttrContext, attrType AttrType) bool {
	if x.Type != attrType {
		return false
	}
	if x.Context == nil {
		return false
	}

	for _, ctx := range x.Context {
		if ctx == context {
			return true
		}
	}

	return false
}

// Hash provides an unique value for the Attribute.
// Hash value must be same if it has same Type, Key, Value and Context.
func (x Attribute) Hash() string {
	sort.Slice(x.Context, func(i, j int) bool {
		return x.Context[i] < x.Context[j]
	})

	raw, err := json.Marshal(x)
	if err != nil {
		// Must marshal
		log.Fatal("Fail to unmarshal attribute", x, err)
	}
	hasher := sha256.New()
	hasher.Write(raw)
	sha := fmt.Sprintf("%x", hasher.Sum(nil))

	return sha
}
