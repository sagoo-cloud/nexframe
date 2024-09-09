package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestSecureMiddleware(t *testing.T) {
	// 创建一个带有自定义配置的 Secure 中间件
	config := SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		HSTSPreloadEnabled:    true,
	}

	// 创建一个简单的处理函数
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// 创建一个带有 Secure 中间件的路由器
	router := mux.NewRouter()
	router.Use(SecureWithConfig(config))
	router.HandleFunc("/", handler)

	// 创建一个测试服务器
	server := httptest.NewServer(router)
	defer server.Close()

	// 发送一个 HTTPS 请求（通过设置 X-Forwarded-Proto 头来模拟）
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Forwarded-Proto", "https")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// 检查响应头
	tests := []struct {
		header string
		expect string
	}{
		{"X-XSS-Protection", "1; mode=block"},
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"Strict-Transport-Security", "max-age=31536000; includeSubdomains; preload"},
		{"Content-Security-Policy", "default-src 'self'"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}

	for _, tt := range tests {
		if got := resp.Header.Get(tt.header); got != tt.expect {
			t.Errorf("Expected header %s to be %s, but got %s", tt.header, tt.expect, got)
		}
	}
}

func TestSecureMiddlewareWithEmptyConfig(t *testing.T) {
	// 创建一个空的 Secure 中间件配置
	config := SecureConfig{}

	// 创建一个简单的处理函数
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// 创建一个带有 Secure 中间件的路由器
	router := mux.NewRouter()
	router.Use(SecureWithConfig(config))
	router.HandleFunc("/", handler)

	// 创建一个测试服务器
	server := httptest.NewServer(router)
	defer server.Close()

	// 发送请求
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// 检查是否没有设置任何安全头部
	securityHeaders := []string{
		"X-XSS-Protection",
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Strict-Transport-Security",
		"Content-Security-Policy",
		"Referrer-Policy",
	}

	for _, header := range securityHeaders {
		if got := resp.Header.Get(header); got != "" {
			t.Errorf("Expected header %s to be empty, but got %s", header, got)
		}
	}
}
func TestSecureMiddlewareSkipper(t *testing.T) {
	// 创建一个带有 skipper 的 Secure 中间件
	config := SecureConfig{
		XSSProtection: "1; mode=block",
		Skipper: func(r *http.Request) bool {
			return r.URL.Path == "/skip"
		},
	}

	// 创建一个简单的处理函数
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// 创建一个带有 Secure 中间件的路由器
	router := mux.NewRouter()
	router.Use(SecureWithConfig(config))
	router.HandleFunc("/", handler)
	router.HandleFunc("/skip", handler)

	// 创建一个测试服务器
	server := httptest.NewServer(router)
	defer server.Close()

	// 测试正常路径
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("X-XSS-Protection"); got != "1; mode=block" {
		t.Errorf("Expected X-XSS-Protection header to be set, but got %s", got)
	}

	// 测试跳过路径
	resp, err = http.Get(server.URL + "/skip")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("X-XSS-Protection"); got != "" {
		t.Errorf("Expected X-XSS-Protection header to be empty, but got %s", got)
	}
}
