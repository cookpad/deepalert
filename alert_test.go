package deepalert_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	da "github.com/m-mizutani/deepalert"
)

func TestLookupAttribute(t *testing.T) {
	alert := da.Alert{
		Attributes: []da.Attribute{
			{
				Key:  "myaddr",
				Type: da.TypeIPAddr,
				Context: []da.AttrContext{
					da.CtxRemote,
					da.CtxSubject,
				},
			},
		},
	}

	attrs1 := alert.FindAttributes("myaddr")
	assert.Equal(t, 1, len(attrs1))
	assert.True(t, attrs1[0].Match(da.CtxRemote, da.TypeIPAddr))
	assert.False(t, attrs1[0].Match(da.CtxLocal, da.TypeIPAddr))
	assert.False(t, attrs1[0].Match(da.CtxSubject, da.TypeUserName))

	attrs2 := alert.FindAttributes("invalid key")
	assert.Equal(t, 0, len(attrs2))
}
