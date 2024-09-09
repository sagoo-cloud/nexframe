package session

import (
	"fmt"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type (
	// Config defines the config for Session middleware.
	Config struct {
		// Skipper defines a function to skip middleware.
		Skipper func(r *http.Request) bool

		// Session store.
		// Required.
		Store sessions.Store
	}
)

const (
	key = "_session_store"
)

var (
	// DefaultConfig is the default Session middleware config.
	DefaultConfig = Config{
		Skipper: defaultSkipper,
	}
)

// defaultSkipper returns false which processes the middleware.
func defaultSkipper(r *http.Request) bool {
	return false
}

// Get returns a named session.
func Get(name string, r *http.Request) (*sessions.Session, error) {
	s := context.Get(r, key)
	if s == nil {
		return nil, fmt.Errorf("%q session store not found", key)
	}
	store := s.(sessions.Store)
	return store.Get(r, name)
}

// Middleware returns a Session middleware.
func Middleware(store sessions.Store) mux.MiddlewareFunc {
	c := DefaultConfig
	c.Store = store
	return MiddlewareWithConfig(c)
}

// MiddlewareWithConfig returns a Sessions middleware with config.
func MiddlewareWithConfig(config Config) mux.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultConfig.Skipper
	}
	if config.Store == nil {
		panic("gorilla/mux: session middleware requires store")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(r) {
				next.ServeHTTP(w, r)
				return
			}
			defer context.Clear(r)
			context.Set(r, key, config.Store)
			next.ServeHTTP(w, r)
		})
	}
}
