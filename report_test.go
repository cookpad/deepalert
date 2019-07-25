package deepalert_test

import (
	"testing"
	"time"

	"github.com/m-mizutani/deepalert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportExtraction(t *testing.T) {
	report := deepalert.Report{
		Sections: []deepalert.ReportSection{
			{
				Author: "Familiar1",
				Attribute: deepalert.Attribute{
					Type:    deepalert.TypeIPAddr,
					Key:     "source",
					Value:   "192.168.0.1",
					Context: []deepalert.AttrContext{deepalert.CtxRemote},
				},
				Type: deepalert.ContentHost,
				Content: deepalert.ReportHost{
					RelatedDomains: []deepalert.EntityDomain{
						{
							Name:      "example.com",
							Timestamp: time.Now(),
							Source:    "tester",
						},
					},
				},
			},
			{
				Author: "Familiar2",
				Attribute: deepalert.Attribute{
					Type:    deepalert.TypeIPAddr,
					Key:     "source",
					Value:   "192.168.0.1",
					Context: []deepalert.AttrContext{deepalert.CtxRemote},
				},
				Type: deepalert.ContentHost,
				Content: deepalert.ReportHost{
					RelatedDomains: []deepalert.EntityDomain{
						{
							Name:      "example.org",
							Timestamp: time.Now(),
							Source:    "tester",
						},
					},
				},
			},
		},
	}

	reportMap, err := report.ExtractContents()
	require.NoError(t, err)
	require.NotNil(t, reportMap)

	// pp.Println(reportMap)
	hv := report.Sections[0].Attribute.Hash()
	assert.Equal(t, 1, len(reportMap.Attributes))
	assert.Equal(t, 1, len(reportMap.Hosts))
	hosts, ok := reportMap.Hosts[hv]
	require.True(t, ok)
	assert.Equal(t, 2, len(hosts))
	assert.Equal(t, "example.com", hosts[0].RelatedDomains[0].Name)

	assert.Equal(t, 0, len(reportMap.Users))
	assert.Equal(t, 0, len(reportMap.Binaries))
}
