package utils

import (
	"log"
	"runtime/debug"
	"sync"
)

// ErrHandler is a function type for custom error handling
type ErrHandler func(interface{}, []byte)

// defaultErrHandler is the default error handler if none is provided
func defaultErrHandler(err interface{}, stack []byte) {
	if stack != nil {
		log.Printf("Panic in goroutine: %v\n%s", err, stack)
	} else {
		log.Printf("Error in goroutine: %v", err)
	}
}

// Go runs the given function in a new goroutine with optional error handling
func Go(fn func() error, errHandler ...ErrHandler) {
	var handler ErrHandler
	if len(errHandler) > 0 && errHandler[0] != nil {
		handler = errHandler[0]
	} else {
		handler = defaultErrHandler
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				handler(r, stack)
			}
		}()

		if err := fn(); err != nil {
			handler(err, nil)
		}
	}()

	// Optionally wait for the goroutine to finish
	// wg.Wait()
}
