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

type ErrorID uint8

const (
	Unknown = ErrorID(iota)
	NotFound
	ArgsMissed
	BadRequest
	Internal
)

type CallError struct {
	code ErrorID
	data map[string]interface{}
}

func (ce CallError) IsNotFound() bool {
	return ce.code == NotFound
}
func (ce CallError) IsBadRequest() bool {
	return ce.code == BadRequest || ce.code == ArgsMissed
}

//
func (ce CallError) Code() string {
	// not using stringer cause it has 114Mb distro
	names := [...]string{
		"Unknown",
		"MethodNotFound",
		"RequiredArgsMissed",
		"BadRequest",
		"Internal",
	}
	if ce.code > Internal {
		return names[Unknown]
	}
	return names[ce.code]
}

func (ce CallError) Message() string {
	// not using stringer cause it has 114Mb distro
	names := [...]string{
		"Unknown",
		"Method not found",
		"Required arg(s) missed",
		"BadRequest",
		"Internal",
	}
	if ce.code > Internal {
		return names[Unknown]
	}
	return names[ce.code]
}

func (ce CallError) Data() map[string]interface{} {
	return ce.data
}

func (ce CallError) Error() string {
	return fmt.Sprintf("%s (%s)", ce.Message(), ce.data)
}

// addContext is an internal method for setting error data
//	err := (&CallError{code: NotFound}).addContext("name", method)
func (ce *CallError) addContext(name string, value interface{}) *CallError {
	if ce.data == nil {
		ce.data = map[string]interface{}{}
	}
	ce.data[name] = value
	return ce
}
