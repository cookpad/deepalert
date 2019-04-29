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

func TestAttributeHash(t *testing.T) {
	a1 := da.Attribute{
		Key:  "myaddr",
		Type: da.TypeIPAddr,
		Context: []da.AttrContext{
			da.CtxRemote,
			da.CtxSubject,
		},
	}

	a2 := da.Attribute{
		Key:  "hoge",
		Type: da.TypeIPAddr,
		Context: []da.AttrContext{
			da.CtxRemote,
			da.CtxSubject,
		},
	}

	a3 := da.Attribute{
		Key:  "myaddr",
		Type: da.TypeIPAddr,
		Context: []da.AttrContext{
			// Reversed
			da.CtxSubject,
			da.CtxRemote,
		},
	}
	assert.NotEqual(t, a1.Hash(), "")
	assert.NotEqual(t, a1.Hash(), a2.Hash())
	assert.Equal(t, a1.Hash(), a3.Hash())
}
