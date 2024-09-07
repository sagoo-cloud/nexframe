package nf

import (
	"net/http"
	"time"
)

// SetCookieMaxAge sets the CookieMaxAge for server.
func (f *APIFramework) SetCookieMaxAge(ttl time.Duration) {
	f.config.CookieMaxAge = ttl
}

// SetCookiePath sets the CookiePath for server.
func (f *APIFramework) SetCookiePath(path string) {
	f.config.CookiePath = path
}

// SetCookieDomain sets the CookieDomain for server.
func (f *APIFramework) SetCookieDomain(domain string) {
	f.config.CookieDomain = domain
}

// GetCookieMaxAge returns the CookieMaxAge of the server.
func (f *APIFramework) GetCookieMaxAge() time.Duration {
	return f.config.CookieMaxAge
}

// GetCookiePath returns the CookiePath of server.
func (f *APIFramework) GetCookiePath() string {
	return f.config.CookiePath
}

// GetCookieDomain returns CookieDomain of server.
func (f *APIFramework) GetCookieDomain() string {
	return f.config.CookieDomain
}

// GetCookieSameSite return CookieSameSite of server.
func (f *APIFramework) GetCookieSameSite() http.SameSite {
	switch f.config.CookieSameSite {
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteDefaultMode
	}
}

func (f *APIFramework) GetCookieSecure() bool {
	return f.config.CookieSecure
}

func (f *APIFramework) GetCookieHttpOnly() bool {
	return f.config.CookieHttpOnly
}
