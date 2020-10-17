package errors_test

import (
	"testing"

	"github.com/deepalert/deepalert/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Run("Simple error", func(tt *testing.T) {
		err := errors.New("something wrong").With("color", "blue")
		assert.Equal(tt, "something wrong", err.Error())
	})

	t.Run("Wrap error", func(tt *testing.T) {
		err1 := errors.New("bomb!!").With("code", "cast")
		err2 := errors.Wrap(err1, "something wrong").With("color", "blue")
		assert.Equal(tt, "something wrong: bomb!!", err2.Error())
	})
}
