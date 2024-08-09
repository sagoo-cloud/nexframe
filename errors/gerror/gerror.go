// Package gerror provides rich functionalities to manipulate errors.
//
// For maintainers, please very note that,
// this package is quite a basic package, which SHOULD NOT import extra packages
// except standard packages and internal packages, to avoid cycle imports.
package gerror

import (
	"github.com/sagoo-cloud/nexframe/errors/gcode"
)

// IIs is the interface for Is feature.
type IIs interface {
	Error() string
	Is(target error) bool
}

// IEqual is the interface for Equal feature.
type IEqual interface {
	Error() string
	Equal(target error) bool
}

// ICode is the interface for Code feature.
type ICode interface {
	Error() string
	Code() gcode.Code
}

// IStack is the interface for Stack feature.
type IStack interface {
	Error() string
	Stack() string
}

// ICause is the interface for Cause feature.
type ICause interface {
	Error() string
	Cause() error
}

// ICurrent is the interface for Current feature.
type ICurrent interface {
	Error() string
	Current() error
}

// IUnwrap is the interface for Unwrap feature.
type IUnwrap interface {
	Error() string
	Unwrap() error
}

const (
	// commaSeparatorSpace is the comma separator with space.
	commaSeparatorSpace = ", "
)
