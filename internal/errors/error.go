package errors

import (
	"fmt"
)

// Error is error interface for deepalert to handle related variables
type Error struct {
	base       error
	Message    string
	Values     map[string]interface{} `json:"values"`
	StatusCode int                    `json:"status_code"`
}

// New creates a new error with message
func New(msg string) *Error {
	err := &Error{
		base:    fmt.Errorf(msg),
		Message: msg,
		Values:  make(map[string]interface{}),
	}
	handleSentryError(err)
	return err
}

// Wrap creates a new error with existing error
func Wrap(err error, msg string) *Error {
	e := New(msg)
	if cause, ok := err.(*Error); ok {
		e.base = fmt.Errorf("%s: %w", msg, cause.base)
	} else {
		e.base = fmt.Errorf("%s: %w", msg, err)
	}

	if cause, ok := err.(*Error); ok {
		for k, v := range cause.Values {
			e.Values[k] = v
		}
		e.StatusCode = cause.StatusCode
	}
	return e
}

// Error returns error message for error interface
func (x *Error) Error() string {
	return x.base.Error()
}

// With adds key and value related to the error event
func (x *Error) With(key string, value interface{}) *Error {
	x.Values[key] = value
	return x
}

// Status sets HTTP status code
func (x *Error) Status(code int) *Error {
	x.StatusCode = code
	return x
}
