package nf

import (
	"time"
)

// SetSessionMaxAge sets the SessionMaxAge for server.
func (f *APIFramework) SetSessionMaxAge(ttl time.Duration) {
	f.config.SessionMaxAge = ttl
}

// SetSessionIdName sets the SessionIdName for server.
func (f *APIFramework) SetSessionIdName(name string) {
	f.config.SessionIdName = name
}

// SetSessionCookieOutput sets the SetSessionCookieOutput for server.
func (f *APIFramework) SetSessionCookieOutput(enabled bool) {
	f.config.SessionCookieOutput = enabled
}

// SetSessionCookieMaxAge sets the SessionCookieMaxAge for server.
func (f *APIFramework) SetSessionCookieMaxAge(maxAge time.Duration) {
	f.config.SessionCookieMaxAge = maxAge
}

// GetSessionMaxAge returns the SessionMaxAge of server.
func (f *APIFramework) GetSessionMaxAge() time.Duration {
	return f.config.SessionMaxAge
}

// GetSessionIdName returns the SessionIdName of server.
func (f *APIFramework) GetSessionIdName() string {
	return f.config.SessionIdName
}

// GetSessionCookieMaxAge returns the SessionCookieMaxAge of server.
func (f *APIFramework) GetSessionCookieMaxAge() time.Duration {
	return f.config.SessionCookieMaxAge
}
