package handler_test

import (
	"testing"

	"github.com/deepalert/deepalert/internal/handler"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
)

func TestDecodeEvents(t *testing.T) {
	args := handler.Arguments{
		Event: map[string][]map[string]string{
			"Records": {
				{
					"body": `{"a":1}`,
				},
				{
					"body": `{"a":765}`,
				},
			},
		},
	}

	var ev2 struct {
		A int `json:"a"`
	}

	records, err := args.DecapSQSEvent()
	require.NoError(t, err)
	assert.Equal(t, 2, len(records))
	assert.Equal(t, `{"a":1}`, string(records[0]))
	require.NoError(t, records[1].Bind(&ev2))
	assert.Equal(t, 765, ev2.A)
}
