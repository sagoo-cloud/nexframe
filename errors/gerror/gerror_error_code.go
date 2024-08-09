package gerror

import (
	"github.com/sagoo-cloud/nexframe/errors/gcode"
)

// Code returns the error code.
// It returns CodeNil if it has no error code.
func (err *Error) Code() gcode.Code {
	if err == nil {
		return gcode.CodeNil
	}
	if err.code == gcode.CodeNil {
		return Code(err.Unwrap())
	}
	return err.code
}

// SetCode updates the internal code with given code.
func (err *Error) SetCode(code gcode.Code) {
	if err == nil {
		return
	}
	err.code = code
}
