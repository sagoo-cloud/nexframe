// Package gcode provides universal error code definition and common error codes implementation.
package gcode

import (
	"fmt"
	"sync"
)

// Code is universal error code interface definition.
type Code interface {
	// Code returns the integer number of current error code.
	Code() int

	// Message returns the brief message for current error code.
	Message() string

	// Detail returns the detailed information of current error code,
	// which is mainly designed as an extension field for error code.
	Detail() interface{}
}

// localCode is the internal implementation of Code interface.
type localCode struct {
	code    int
	message string
	detail  interface{}
}

func (c localCode) Code() int           { return c.code }
func (c localCode) Message() string     { return c.message }
func (c localCode) Detail() interface{} { return c.detail }
func (c localCode) String() string {
	if c.detail != nil {
		return fmt.Sprintf(`%d:%s %v`, c.code, c.message, c.detail)
	}
	if c.message != "" {
		return fmt.Sprintf(`%d:%s`, c.code, c.message)
	}
	return fmt.Sprintf(`%d`, c.code)
}

// Use a sync.Pool to reduce GC pressure for frequently used error codes
var codePool = sync.Pool{
	New: func() interface{} {
		return new(localCode)
	},
}

// Common error code definition.
// There are reserved internal error codes by framework: code < 1000.
var (
	CodeNil                       = newLocalCode(-1, "", nil)
	CodeOK                        = newLocalCode(0, "OK", nil)
	CodeInternalError             = newLocalCode(50, "Internal Error", nil)
	CodeValidationFailed          = newLocalCode(51, "Validation Failed", nil)
	CodeDbOperationError          = newLocalCode(52, "Database Operation Error", nil)
	CodeInvalidParameter          = newLocalCode(53, "Invalid Parameter", nil)
	CodeMissingParameter          = newLocalCode(54, "Missing Parameter", nil)
	CodeInvalidOperation          = newLocalCode(55, "Invalid Operation", nil)
	CodeInvalidConfiguration      = newLocalCode(56, "Invalid Configuration", nil)
	CodeMissingConfiguration      = newLocalCode(57, "Missing Configuration", nil)
	CodeNotImplemented            = newLocalCode(58, "Not Implemented", nil)
	CodeNotSupported              = newLocalCode(59, "Not Supported", nil)
	CodeOperationFailed           = newLocalCode(60, "Operation Failed", nil)
	CodeNotAuthorized             = newLocalCode(61, "Not Authorized", nil)
	CodeSecurityReason            = newLocalCode(62, "Security Reason", nil)
	CodeServerBusy                = newLocalCode(63, "Server Is Busy", nil)
	CodeUnknown                   = newLocalCode(64, "Unknown Error", nil)
	CodeNotFound                  = newLocalCode(65, "Not Found", nil)
	CodeInvalidRequest            = newLocalCode(66, "Invalid Request", nil)
	CodeNecessaryPackageNotImport = newLocalCode(67, "Necessary Package Not Import", nil)
	CodeInternalPanic             = newLocalCode(68, "Internal Panic", nil)
	CodeBusinessValidationFailed  = newLocalCode(300, "Business Validation Failed", nil)
)

// newLocalCode creates a new localCode instance.
func newLocalCode(code int, message string, detail interface{}) Code {
	return localCode{
		code:    code,
		message: message,
		detail:  detail,
	}
}

// New creates and returns an error code.
// Note that it returns an interface object of Code.
func New(code int, message string, detail interface{}) Code {
	c := codePool.Get().(*localCode)
	c.code = code
	c.message = message
	c.detail = detail
	return c
}

// WithCode creates and returns a new error code based on given Code.
// The code and message is from given `code`, but the detail is from given `detail`.
func WithCode(code Code, detail interface{}) Code {
	c := codePool.Get().(*localCode)
	c.code = code.Code()
	c.message = code.Message()
	c.detail = detail
	return c
}

// Release puts the Code back to the pool.
// It should be called when the Code is no longer needed.
func Release(c Code) {
	if lc, ok := c.(*localCode); ok {
		codePool.Put(lc)
	}
}
