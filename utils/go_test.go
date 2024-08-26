package utils

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGo(t *testing.T) {
	t.Run("Normal Execution", func(t *testing.T) {
		var executed bool
		Go(func() error {
			executed = true
			return nil
		})
		time.Sleep(10 * time.Millisecond) // Give some time for goroutine to execute
		if !executed {
			t.Error("Function was not executed")
		}
	})

	t.Run("Error Handling", func(t *testing.T) {
		expectedErr := errors.New("test error")
		var capturedErr error
		var wg sync.WaitGroup
		wg.Add(1)

		Go(
			func() error {
				return expectedErr
			},
			func(err interface{}, _ []byte) {
				capturedErr = err.(error)
				wg.Done()
			},
		)

		wg.Wait()
		if capturedErr != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, capturedErr)
		}
	})

	t.Run("Panic Recovery", func(t *testing.T) {
		panicMsg := "test panic"
		var capturedPanic interface{}
		var capturedStack []byte
		var wg sync.WaitGroup
		wg.Add(1)

		Go(
			func() error {
				panic(panicMsg)
			},
			func(err interface{}, stack []byte) {
				capturedPanic = err
				capturedStack = stack
				wg.Done()
			},
		)

		wg.Wait()
		if capturedPanic != panicMsg {
			t.Errorf("Expected panic message %v, got %v", panicMsg, capturedPanic)
		}
		if !strings.Contains(string(capturedStack), "runtime/panic.go") {
			t.Error("Stack trace does not contain expected content")
		}
	})

	t.Run("Default Error Handler", func(t *testing.T) {
		expectedErr := errors.New("test error")
		Go(func() error {
			return expectedErr
		})
		// Note: This test doesn't actually verify the log output.
		// In a real scenario, you might want to use a custom logger or
		// capture stdout to verify the log message.
	})

	t.Run("Concurrent Execution", func(t *testing.T) {
		const count = 100
		var wg sync.WaitGroup
		wg.Add(count)

		for i := 0; i < count; i++ {
			Go(func() error {
				defer wg.Done()
				time.Sleep(10 * time.Millisecond)
				return nil
			})
		}

		wg.Wait() // If this doesn't deadlock, the test passes
	})
}
