package deepalert_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	da "github.com/cookpad/deepalert"
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

func TestValidate(t *testing.T) {
	base := da.Alert{
		Detector: "det",
		RuleID:   "rid",
	}

	t.Run("Valid minimal alert passes", func(t *testing.T) {
		a := base
		require.NoError(t, a.Validate())
	})

	t.Run("Missing Detector is rejected", func(t *testing.T) {
		a := base
		a.Detector = ""
		err := a.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, da.ErrInvalidAlert)
	})

	t.Run("Missing RuleID is rejected", func(t *testing.T) {
		a := base
		a.RuleID = ""
		err := a.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, da.ErrInvalidAlert)
	})

	for _, tc := range []struct {
		name  string
		field func(a *da.Alert, v string)
		limit int
	}{
		{"Detector", func(a *da.Alert, v string) { a.Detector = v }, 1024},
		{"RuleID", func(a *da.Alert, v string) { a.RuleID = v }, 1024},
		{"RuleName", func(a *da.Alert, v string) { a.RuleName = v }, 1024},
		{"AlertKey", func(a *da.Alert, v string) { a.AlertKey = v }, 1024},
		{"Description", func(a *da.Alert, v string) { a.Description = v }, 4096},
	} {
		tc := tc
		t.Run(tc.name+" at limit passes", func(t *testing.T) {
			a := base
			tc.field(&a, strings.Repeat("x", tc.limit))
			require.NoError(t, a.Validate())
		})
		t.Run(tc.name+" over limit is rejected", func(t *testing.T) {
			a := base
			tc.field(&a, strings.Repeat("x", tc.limit+1))
			err := a.Validate()
			require.Error(t, err)
			assert.ErrorIs(t, err, da.ErrInvalidAlert)
		})
	}

	t.Run("Attribute.Key at limit passes", func(t *testing.T) {
		a := base
		a.Attributes = []da.Attribute{{Key: strings.Repeat("k", 1024), Value: "v"}}
		require.NoError(t, a.Validate())
	})

	t.Run("Attribute.Key over limit is rejected", func(t *testing.T) {
		a := base
		a.Attributes = []da.Attribute{{Key: strings.Repeat("k", 1025), Value: "v"}}
		err := a.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, da.ErrInvalidAlert)
	})

	t.Run("Attribute.Value at limit passes", func(t *testing.T) {
		a := base
		a.Attributes = []da.Attribute{{Key: "k", Value: strings.Repeat("v", 1024)}}
		require.NoError(t, a.Validate())
	})

	t.Run("Attribute.Value over limit is rejected", func(t *testing.T) {
		a := base
		a.Attributes = []da.Attribute{{Key: "k", Value: strings.Repeat("v", 1025)}}
		err := a.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, da.ErrInvalidAlert)
	})

	t.Run("Attributes at count limit passes", func(t *testing.T) {
		a := base
		for i := 0; i < 100; i++ {
			a.Attributes = append(a.Attributes, da.Attribute{Key: "k", Value: "v"})
		}
		require.NoError(t, a.Validate())
	})

	t.Run("Attributes over count limit is rejected", func(t *testing.T) {
		a := base
		for i := 0; i < 101; i++ {
			a.Attributes = append(a.Attributes, da.Attribute{Key: "k", Value: "v"})
		}
		err := a.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, da.ErrInvalidAlert)
	})

	t.Run("Body at size limit passes", func(t *testing.T) {
		a := base
		// JSON string of exactly 65536 bytes: `"<65534 x's>"` = 65536 chars
		a.Body = strings.Repeat("x", 65534)
		require.NoError(t, a.Validate())
	})

	t.Run("Body over size limit is rejected", func(t *testing.T) {
		a := base
		a.Body = strings.Repeat("x", 65535) // marshals to 65537 bytes as a JSON string
		err := a.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, da.ErrInvalidAlert)
	})
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
