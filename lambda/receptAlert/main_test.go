package main_test

import (
	"testing"

	"github.com/deepalert/deepalert/internal/adaptor"
	"github.com/deepalert/deepalert/internal/handler"
	"github.com/deepalert/deepalert/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/deepalert/deepalert/lambda/receptAlert"
)

func TestReceptAlert(t *testing.T) {
	t.Run("Recept single alert", func(tt *testing.T) {
		dummySFn := mock.NewSFnClient("")
		dummyRepo := mock.NewRepository("", "")
		args := &handler.Arguments{
			NewRepository: func(string, string) adaptor.Repository { return dummyRepo },
			NewSFn:        func(string) (adaptor.SFnClient, error) { return dummySFn, nil },
		}

		resp, err := main.HandleRequest(args)
		require.NoError(tt, err)
		assert.Nil(tt, resp)
	})
}
