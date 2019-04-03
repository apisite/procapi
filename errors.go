package pgcall

import (
	"fmt"
)

/*
	type callError interface {
		IsNotFound() bool
		IsBadRequest() bool
		Code() string
		Error() string
		Data() map[string]interface{}
	}
	perr, ok := err.(callError)
	if ok {
		if perr.IsNotFound() {
			status = http.StatusNotFound
		}
	}

*/

// ErrorID
type errorID uint8

const (
	errUnknown = errorID(iota)
	errNotFound
	errArgsMissed
	errBadRequest
	errInternal
)

type callError struct {
	code errorID
	data map[string]interface{}
}

// IsNotFound checks error code
func (ce callError) IsNotFound() bool {
	return ce.code == errNotFound
}

// IsBadRequest checks error code
func (ce callError) IsBadRequest() bool {
	return ce.code == errBadRequest || ce.code == errArgsMissed
}

// Code returns error code
func (ce callError) Code() string {
	// not using stringer cause it has 114Mb distro
	names := [...]string{
		"Unknown",
		"MethodNotFound",
		"RequiredArgsMissed",
		"BadRequest",
		"Internal",
	}
	if ce.code > errInternal {
		return names[errUnknown]
	}
	return names[ce.code]
}

// Message returns error description
func (ce callError) Message() string {
	// not using stringer cause it has 114Mb distro
	names := [...]string{
		"Unknown",
		"Method not found",
		"Required arg(s) missed",
		"BadRequest",
		"Internal",
	}
	if ce.code > errInternal {
		return names[errUnknown]
	}
	return names[ce.code]
}

// Data returns error data map
func (ce callError) Data() map[string]interface{} {
	return ce.data
}

// Error returns error message with data
func (ce callError) Error() string {
	return fmt.Sprintf("%s (%s)", ce.Message(), ce.data)
}

// addContext is an internal method for setting error data
//	err := (&callError{code: NotFound}).addContext("name", method)
func (ce *callError) addContext(name string, value interface{}) *callError {
	if ce.data == nil {
		ce.data = map[string]interface{}{}
	}
	ce.data[name] = value
	return ce
}
