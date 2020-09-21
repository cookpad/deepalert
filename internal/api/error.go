package api

import (
	"fmt"

	"github.com/pkg/errors"
)

type apiError interface {
	Error() string
	StatusCode() int
	Message() string

	SetStatusCode(code int) apiError
}

type baseError struct {
	err        error
	message    string
	statusCode int
}

func (x *baseError) StatusCode() int { return x.statusCode }
func (x *baseError) Message() string {
	var errmsg string
	if x.err != nil {
		errmsg = ": " + x.err.Error()
	}
	return x.message + errmsg
}

type userError struct{ baseError }
type systemError struct{ baseError }

func (x *userError) Error() string   { return "UserError: " + x.Message() }
func (x *systemError) Error() string { return "SystemError" }

func (x *userError) SetStatusCode(code int) apiError   { x.statusCode = code; return x }
func (x *systemError) SetStatusCode(code int) apiError { x.statusCode = code; return x }

func wrapUserError(err error, msg string, args ...interface{}) apiError {
	return &userError{
		baseError: baseError{
			message:    fmt.Sprintf(msg, args...),
			err:        errors.Wrapf(err, msg, args...),
			statusCode: 400,
		},
	}
}

func newUserError(msg string, args ...interface{}) apiError {
	return &userError{
		baseError: baseError{
			message:    fmt.Sprintf(msg, args...),
			statusCode: 400,
		},
	}
}

func wrapSystemError(err error, msg string, args ...interface{}) apiError {
	return &systemError{
		baseError: baseError{
			err:        errors.Wrapf(err, msg, args...),
			statusCode: 500,
		},
	}
}

func newSystemError(msg string, args ...interface{}) apiError {
	return &systemError{
		baseError: baseError{
			message:    fmt.Sprintf(msg, args...),
			statusCode: 500,
		},
	}
}
