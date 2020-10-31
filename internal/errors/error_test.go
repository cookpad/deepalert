package errors_test

import (
	"testing"

	"github.com/deepalert/deepalert/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	t.Run("Simple error", func(t *testing.T) {
		err := errors.New("something wrong").With("color", "blue")
		assert.Equal(t, "something wrong", err.Error())
	})

	t.Run("Wrap error", func(t *testing.T) {
		err1 := errors.New("bomb!!").With("code", "cast")
		err2 := errors.Wrap(err1, "something wrong").With("color", "blue")
		assert.Equal(t, "something wrong: bomb!!", err2.Error())
	})

	t.Run("inherit value", func(t *testing.T) {
		err1 := errors.New("bomb!!").With("code", "cast")
		err2 := errors.Wrap(err1, "something wrong").With("color", "blue")
		v, ok := err2.Values["code"]
		require.True(t, ok)
		assert.Equal(t, "cast", v)
	})
}
