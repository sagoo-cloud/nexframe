package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"github.com/sagoo-cloud/nexframe/utils/grand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// CSRFConfig 定义 CSRF 中间件的配置。
type CSRFConfig struct {
	// Skipper 定义一个函数来跳过中间件。
	Skipper func(r *http.Request) bool

	// TokenLength 是生成的令牌的长度。
	TokenLength uint8

	// TokenLookup 是一个字符串，用于从请求中提取令牌。
	TokenLookup string

	// ContextKey 是用于将生成的 CSRF 令牌存储到上下文中的键。
	ContextKey string

	// CookieName 是 CSRF cookie 的名称。
	CookieName string

	// CookieDomain 是 CSRF cookie 的域。
	CookieDomain string

	// CookiePath 是 CSRF cookie 的路径。
	CookiePath string

	// CookieMaxAge 是 CSRF cookie 的最大年龄（以秒为单位）。
	CookieMaxAge int

	// CookieSecure 指示 CSRF cookie 是否安全。
	CookieSecure bool

	// CookieHTTPOnly 指示 CSRF cookie 是否为 HTTP only。
	CookieHTTPOnly bool

	// CookieSameSite 指示 CSRF cookie 的 SameSite 模式。
	CookieSameSite http.SameSite

	// ErrorHandler 定义一个用于返回自定义错误的函数。
	ErrorHandler func(err error, w http.ResponseWriter, r *http.Request)
}

// DefaultCSRFConfig 是默认的 CSRF 中间件配置。
var DefaultCSRFConfig = CSRFConfig{
	Skipper:        func(r *http.Request) bool { return false },
	TokenLength:    32,
	TokenLookup:    "header:X-CSRF-Token",
	ContextKey:     "csrf",
	CookieName:     "_csrf",
	CookieMaxAge:   86400,
	CookieSameSite: http.SameSiteDefaultMode,
}

// ErrCSRFInvalid 在 CSRF 检查失败时返回
var ErrCSRFInvalid = errors.New("invalid csrf token")

// CSRF 返回一个跨站请求伪造（CSRF）中间件。
func CSRF() mux.MiddlewareFunc {
	return CSRFWithConfig(DefaultCSRFConfig)
}

// CSRFWithConfig 返回一个带配置的 CSRF 中间件。
func CSRFWithConfig(config CSRFConfig) mux.MiddlewareFunc {
	// 默认值设置
	if config.Skipper == nil {
		config.Skipper = DefaultCSRFConfig.Skipper
	}
	if config.TokenLength == 0 {
		config.TokenLength = DefaultCSRFConfig.TokenLength
	}
	if config.TokenLookup == "" {
		config.TokenLookup = DefaultCSRFConfig.TokenLookup
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultCSRFConfig.ContextKey
	}
	if config.CookieName == "" {
		config.CookieName = DefaultCSRFConfig.CookieName
	}
	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = DefaultCSRFConfig.CookieMaxAge
	}
	if config.CookieSameSite == http.SameSiteNoneMode {
		config.CookieSecure = true
	}

	extractors, err := CreateExtractors(config.TokenLookup)
	if err != nil {
		panic(err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(r) {
				next.ServeHTTP(w, r)
				return
			}

			token := ""
			if c, err := r.Cookie(config.CookieName); err != nil {
				token = grand.S(convert.Int(config.TokenLength))
			} else {
				token = c.Value // 重用令牌
			}

			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			default:
				// 仅对 RFC7231 中未定义为"安全"的请求验证令牌
				var lastExtractorErr error
				var lastTokenErr error
			outer:
				for _, extractor := range extractors {
					clientTokens, err := extractor(r)
					if err != nil {
						lastExtractorErr = err
						continue
					}

					for _, clientToken := range clientTokens {
						if validateCSRFToken(token, clientToken) {
							lastTokenErr = nil
							lastExtractorErr = nil
							break outer
						}
						lastTokenErr = ErrCSRFInvalid
					}
				}
				var finalErr error
				if lastTokenErr != nil {
					finalErr = lastTokenErr
				} else if lastExtractorErr != nil {
					finalErr = lastExtractorErr
				}

				if finalErr != nil {
					if config.ErrorHandler != nil {
						config.ErrorHandler(finalErr, w, r)
					} else {
						http.Error(w, finalErr.Error(), http.StatusForbidden)
					}
					return
				}
			}

			// 设置 CSRF cookie
			cookie := &http.Cookie{
				Name:     config.CookieName,
				Value:    token,
				Path:     config.CookiePath,
				Domain:   config.CookieDomain,
				Expires:  time.Now().Add(time.Duration(config.CookieMaxAge) * time.Second),
				Secure:   config.CookieSecure,
				HttpOnly: config.CookieHTTPOnly,
				SameSite: config.CookieSameSite,
			}
			http.SetCookie(w, cookie)

			// 将令牌存储在请求上下文中
			ctx := context.WithValue(r.Context(), config.ContextKey, token)
			r = r.WithContext(ctx)

			// 保护客户端不缓存响应
			w.Header().Add("Vary", "Cookie")

			next.ServeHTTP(w, r)
		})
	}
}

func validateCSRFToken(token, clientToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(clientToken)) == 1
}

// 这里需要实现 randomString 和 CreateExtractors 函数
// randomString 函数生成指定长度的随机字符串
// CreateExtractors 函数根据 TokenLookup 创建提取器
