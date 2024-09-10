package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// SecureConfig 定义 Secure 中间件的配置。
type SecureConfig struct {
	// Skipper 定义一个函数来跳过中间件。
	Skipper func(r *http.Request) bool

	// XSSProtection 提供对跨站脚本攻击(XSS)的保护。
	XSSProtection string

	// ContentTypeNosniff 提供对覆盖 Content-Type 头的保护。
	ContentTypeNosniff string

	// XFrameOptions 用于指示浏览器是否应该被允许在 frame 中渲染页面。
	XFrameOptions string

	// HSTSMaxAge 设置 Strict-Transport-Security 头。
	HSTSMaxAge int

	// HSTSExcludeSubdomains 在 Strict Transport Security 头中不包含 subdomains 标签。
	HSTSExcludeSubdomains bool

	// ContentSecurityPolicy 设置 Content-Security-Policy 头。
	ContentSecurityPolicy string

	// CSPReportOnly 使用 Content-Security-Policy-Report-Only 头。
	CSPReportOnly bool

	// HSTSPreloadEnabled 在 Strict Transport Security 头中添加 preload 标签。
	HSTSPreloadEnabled bool

	// ReferrerPolicy 设置 Referrer-Policy 头。
	ReferrerPolicy string
}

// DefaultSecureConfig 是 Secure 中间件的默认配置。
var DefaultSecureConfig = SecureConfig{
	Skipper:            func(r *http.Request) bool { return false },
	XSSProtection:      "1; mode=block",
	ContentTypeNosniff: "nosniff",
	XFrameOptions:      "SAMEORIGIN",
	HSTSPreloadEnabled: false,
}

// Secure 返回一个 Secure 中间件。
func Secure() mux.MiddlewareFunc {
	return SecureWithConfig(DefaultSecureConfig)
}

// SecureWithConfig 返回一个带配置的 Secure 中间件。
func SecureWithConfig(config SecureConfig) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper != nil && config.Skipper(r) {
				next.ServeHTTP(w, r)
				return
			}

			if config.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", config.XSSProtection)
			}
			if config.ContentTypeNosniff != "" {
				w.Header().Set("X-Content-Type-Options", config.ContentTypeNosniff)
			}
			if config.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.XFrameOptions)
			}

			isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
			if isSecure && config.HSTSMaxAge != 0 {
				subdomains := ""
				if !config.HSTSExcludeSubdomains {
					subdomains = "; includeSubdomains"
				}
				if config.HSTSPreloadEnabled {
					subdomains = fmt.Sprintf("%s; preload", subdomains)
				}
				w.Header().Set("Strict-Transport-Security", fmt.Sprintf("max-age=%d%s", config.HSTSMaxAge, subdomains))
			}
			if config.ContentSecurityPolicy != "" {
				if config.CSPReportOnly {
					w.Header().Set("Content-Security-Policy-Report-Only", config.ContentSecurityPolicy)
				} else {
					w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
				}
			}
			if config.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}
