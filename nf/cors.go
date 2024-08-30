package nf

import (
	"github.com/gorilla/mux"
	"github.com/sagoo-cloud/nexframe/utils/gstr"
	"net/http"
	"net/url"
	"strconv"
)

var (
	defaultAllowHeaders    = "Origin,Content-Type,Accept,User-Agent,Cookie,Authorization,X-Auth-Token,X-Requested-With"
	defaultAllowHeadersMap = make(map[string]struct{})
)

func init() {
	array := gstr.SplitAndTrim(defaultAllowHeaders, ",")
	for _, header := range array {
		defaultAllowHeadersMap[header] = struct{}{}
	}
}

const (
	supportedHttpMethods = "GET,PUT,POST,DELETE,PATCH,HEAD,CONNECT,OPTIONS,TRACE"
)

type CORSOptions struct {
	AllowDomain      []string // Used for allowing requests from custom domains
	AllowOrigin      string   // Access-Control-Allow-Origin
	AllowCredentials string   // Access-Control-Allow-Credentials
	ExposeHeaders    string   // Access-Control-Expose-Headers
	MaxAge           int      // Access-Control-Max-Age
	AllowMethods     string   // Access-Control-Allow-Methods
	AllowHeaders     string   // Access-Control-Allow-Headers
}

// CORSDefault 允许跨域请求
func (f *APIFramework) CORSDefault(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware 处理跨域请求
func (f *APIFramework) CORSMiddleware(opts CORSOptions) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 设置允许的来源
			if f.corsAllowedOrigin(r, opts) {
				w.Header().Set("Access-Control-Allow-Origin", opts.AllowOrigin)
			}

			// 设置其他CORS头
			if opts.AllowCredentials != "" {
				w.Header().Set("Access-Control-Allow-Credentials", opts.AllowCredentials)
			}
			if opts.ExposeHeaders != "" {
				w.Header().Set("Access-Control-Expose-Headers", opts.ExposeHeaders)
			}
			if opts.AllowMethods != "" {
				w.Header().Set("Access-Control-Allow-Methods", opts.AllowMethods)
			}
			if opts.AllowHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", opts.AllowHeaders)
			}
			if opts.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(opts.MaxAge))
			}

			// 处理预检请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// 调用下一个中间件
			next.ServeHTTP(w, r)
		})
	}
}

// corsAllowedOrigin CORSAllowed checks whether the current request origin is allowed cross-domain.
func (f *APIFramework) corsAllowedOrigin(r *http.Request, options CORSOptions) bool {
	if options.AllowDomain == nil {
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	for _, v := range options.AllowDomain {
		if gstr.IsSubDomain(parsed.Host, v) {
			return true
		}
	}
	return false
}
