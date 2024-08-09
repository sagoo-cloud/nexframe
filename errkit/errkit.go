package errkit

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Pre defined error types
var (
	ErrNotFound   = errors.New("not found")
	ErrPermission = errors.New("permission denied")
	ErrTimeout    = errors.New("operation timed out")
)

// Error represents a custom error type that includes an error code and stack trace
type Error struct {
	Err        error
	Code       string
	Message    string
	Internal   string
	StackTrace string
	Context    map[string]interface{}
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

// Is implements error comparison for the errors.Is function
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return errors.Is(e.Err, target)
	}
	return e.Code == t.Code
}

// As implements type assertion for the errors.As function
func (e *Error) As(target interface{}) bool {
	t, ok := target.(**Error)
	if !ok {
		return errors.As(e.Err, target)
	}
	*t = e
	return true
}

// New creates a new Error
func New(message string) *Error {
	return &Error{
		Err:        errors.New(message),
		Message:    message,
		Code:       generateErrorCode(message),
		StackTrace: getStackTrace(),
		Context:    make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}
	wrappedErr := fmt.Errorf("%s: %w", message, err)
	return &Error{
		Err:        wrappedErr,
		Message:    wrappedErr.Error(),
		Code:       generateErrorCode(wrappedErr.Error()),
		StackTrace: getStackTrace(),
		Context:    make(map[string]interface{}),
	}
}

// Annotate adds context to an error without changing its type
func Annotate(err error, key string, value interface{}) error {
	if err == nil {
		return nil
	}
	e, ok := err.(*Error)
	if !ok {
		e = Wrap(err, err.Error())
	}
	e.Context[key] = value
	return e
}

// GetAnnotation retrieves context from an error
func GetAnnotation(err error, key string) (interface{}, bool) {
	e, ok := err.(*Error)
	if !ok {
		return nil, false
	}
	value, exists := e.Context[key]
	return value, exists
}

// generateErrorCode creates a 4-character error code from a string
func generateErrorCode(s string) string {
	h := md5.Sum([]byte(s))
	u := binary.BigEndian.Uint32(h[:4])
	code := fmt.Sprintf("%04s", strings.ToUpper(fmt.Sprintf("%x", u&0xFFFFF)))
	return strings.NewReplacer("O", "0", "I", "1").Replace(code)[:4]
}

// Handle is a helper function for uniform error handling
func Handle(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return &Error{
		Err:        err,
		Message:    "An unexpected error occurred",
		Code:       generateErrorCode(err.Error()),
		Internal:   err.Error(),
		StackTrace: getStackTrace(),
		Context:    make(map[string]interface{}),
	}
}

// Assert provides a simple assertion mechanism
func Assert(condition bool, message string) {
	if !condition {
		panic(New(message))
	}
}

// Recover is a helper function to recover from panics
func Recover(err *error) {
	if r := recover(); r != nil {
		switch x := r.(type) {
		case string:
			*err = New(x)
		case error:
			*err = Wrap(x, "panic recovered")
		default:
			*err = New(fmt.Sprintf("panic recovered: %v", r))
		}
	}
}

// TryCatch simulates try-catch behavior
func TryCatch(try func() error, catch func(error)) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = Handle(fmt.Errorf("%v", r))
		}
		if err != nil {
			catch(err)
		}
	}()
	err = try()
}

// getStackTrace returns the stack trace as a string
func getStackTrace() string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, 2*len(buf))
	}
}

// FormatErrorResponse formats the error for service-level response
func FormatErrorResponse(err error) map[string]interface{} {
	e := Handle(err)
	return map[string]interface{}{
		"code":    e.Code,
		"message": fmt.Sprintf("An error occurred. Error code: %s. If you need assistance, please contact support.", e.Code),
	}
}

// LogError logs the full error details for backend tracking
func LogError(err error) {
	e := Handle(err)
	fmt.Printf("Error Code: %s\nMessage: %s\nInternal: %s\nStack Trace:\n%s\nContext: %v\n",
		e.Code, e.Message, e.Internal, e.StackTrace, e.Context)
}

// ErrorGroup represents a group of errors
type ErrorGroup struct {
	errors []error
}

// NewErrorGroup creates a new ErrorGroup
func NewErrorGroup() *ErrorGroup {
	return &ErrorGroup{errors: []error{}}
}

// Add adds an error to the ErrorGroup
func (eg *ErrorGroup) Add(err error) {
	if err != nil {
		eg.errors = append(eg.errors, err)
	}
}

// Err returns an error representing the ErrorGroup
func (eg *ErrorGroup) Err() error {
	if len(eg.errors) == 0 {
		return nil
	}
	return fmt.Errorf("%d errors occurred: %v", len(eg.errors), eg.errors)
}
