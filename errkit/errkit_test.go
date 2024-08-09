package errkit

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("test error")
	if err.Message != "test error" {
		t.Errorf("Expected message 'test error', got '%s'", err.Message)
	}
	if len(err.Code) != 4 {
		t.Errorf("Expected code length 4, got %d", len(err.Code))
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, "wrapped error")
	if !strings.Contains(wrappedErr.Message, "original error") {
		t.Errorf("Wrapped error should contain original error message")
	}
	if !strings.Contains(wrappedErr.Message, "wrapped error") {
		t.Errorf("Wrapped error should contain wrap message")
	}
}

func TestIs(t *testing.T) {
	err1 := New("test error")
	err2 := New("test error")
	if !errors.Is(err1, err2) {
		t.Errorf("Errors with same message should be considered equal")
	}
}

func TestAs(t *testing.T) {
	originalErr := New("test error")
	var err *Error
	if !errors.As(originalErr, &err) {
		t.Errorf("As should succeed for *Error type")
	}
}

func TestAnnotate(t *testing.T) {
	err := New("test error")
	annotatedErr := Annotate(err, "key", "value")
	if e, ok := annotatedErr.(*Error); !ok {
		t.Errorf("Annotate should return *Error")
	} else {
		if v, ok := e.Context["key"]; !ok || v != "value" {
			t.Errorf("Annotation not set correctly")
		}
	}
}

func TestGetAnnotation(t *testing.T) {
	err := New("test error")
	Annotate(err, "key", "value")
	if v, ok := GetAnnotation(err, "key"); !ok || v != "value" {
		t.Errorf("GetAnnotation failed to retrieve annotation")
	}
}

func TestErrorGroup(t *testing.T) {
	eg := NewErrorGroup()
	eg.Add(New("error 1"))
	eg.Add(New("error 2"))
	if eg.Err() == nil {
		t.Errorf("ErrorGroup should return error when it contains errors")
	}
	if !strings.Contains(eg.Err().Error(), "2 errors occurred") {
		t.Errorf("ErrorGroup error message incorrect")
	}
}

func TestRecover(t *testing.T) {
	var err error
	func() {
		defer Recover(&err)
		panic("test panic")
	}()
	if err == nil {
		t.Errorf("Recover should catch panic")
	}
	if !strings.Contains(err.Error(), "test panic") {
		t.Errorf("Recovered error should contain panic message, got: %v", err)
	}
}

func TestAssert(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Assert should panic on false condition")
		}
	}()
	Assert(false, "assertion failed")
}

func TestFormatErrorResponse(t *testing.T) {
	err := New("test error")
	resp := FormatErrorResponse(err)
	if _, ok := resp["code"]; !ok {
		t.Errorf("FormatErrorResponse should include 'code'")
	}
	if _, ok := resp["message"]; !ok {
		t.Errorf("FormatErrorResponse should include 'message'")
	}
}

func TestLogError(t *testing.T) {
	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := New("test error")
	LogError(err)

	// Reset stdout
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check if output contains expected information
	if !strings.Contains(output, "Error Code:") {
		t.Errorf("LogError output should contain 'Error Code:'")
	}
	if !strings.Contains(output, "test error") {
		t.Errorf("LogError output should contain the error message")
	}
	if !strings.Contains(output, "Stack Trace:") {
		t.Errorf("LogError output should contain 'Stack Trace:'")
	}
}
