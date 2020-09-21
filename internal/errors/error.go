package errors

import "fmt"

// Error is error interface for deepalert to handle related variables
type Error struct {
	Wrapped error                  `json:"wrapped"`
	Message string                 `json:"message"`
	Context map[string]interface{} `json:"context"`
}

// New creates a new error with message
func New(msg string) *Error {
	err := &Error{
		Message: msg,
		Context: make(map[string]interface{}),
	}
	handleSentryError(err)
	return err
}

// Newf creates a new error with Sprintf message
func Newf(msgfmt string, args ...interface{}) *Error {
	return New(fmt.Sprintf(msgfmt, args...))
}

// Wrap creates a new error with existing error
func Wrap(err error, msg string) *Error {
	e := New(msg)
	e.Wrapped = err
	return e
}

// Wrapf creates a new error with existing error and format message
func Wrapf(err error, msgfmt string, args ...interface{}) *Error {
	e := Newf(msgfmt, args...)
	e.Wrapped = err
	return e
}

// Error returns error message for error interface
func (x *Error) Error() string {
	if x.Wrapped != nil {
		return fmt.Sprintf("%s %+v", x.Message, x.Wrapped)
	}
	return x.Message
}

// With adds key and value related to the error event
func (x *Error) With(key string, value interface{}) *Error {
	x.Context[key] = value
	return x
}
