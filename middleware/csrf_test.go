package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestCSRFMiddleware(t *testing.T) {
	// 创建一个带有 CSRF 中间件的路由器
	r := mux.NewRouter()
	r.Use(CSRF())

	// 添加一个测试路由
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value("csrf").(string)
		w.Write([]byte(fmt.Sprintf("CSRF Token: %s", token)))
	}).Methods("GET")

	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("POST successful"))
	}).Methods("POST")

	// 测试 GET 请求（应该设置 CSRF cookie 和生成令牌）
	t.Run("GET request should set CSRF cookie", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("处理器返回了错误的状态码: got %v want %v", status, http.StatusOK)
		}

		cookies := rr.Result().Cookies()
		var csrfCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "_csrf" {
				csrfCookie = cookie
				break
			}
		}

		if csrfCookie == nil {
			t.Error("CSRF cookie 未设置")
		}

		if !strings.Contains(rr.Body.String(), "CSRF Token:") {
			t.Error("响应中不包含 CSRF 令牌")
		}
	})

	// 测试没有令牌的 POST 请求（应该被拒绝）
	t.Run("POST without token should be rejected", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/test", nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusForbidden {
			t.Errorf("处理器返回了错误的状态码: got %v want %v", status, http.StatusForbidden)
		}
	})

	// 测试带有有效令牌的 POST 请求
	t.Run("POST with valid token should be accepted", func(t *testing.T) {
		// 首先发送 GET 请求来获取令牌
		getReq, _ := http.NewRequest("GET", "/test", nil)
		getRr := httptest.NewRecorder()
		r.ServeHTTP(getRr, getReq)

		// 从 cookie 中提取令牌
		var csrfToken string
		for _, cookie := range getRr.Result().Cookies() {
			if cookie.Name == "_csrf" {
				csrfToken = cookie.Value
				break
			}
		}

		// 发送带有令牌的 POST 请求
		postReq, _ := http.NewRequest("POST", "/test", nil)
		postReq.Header.Set("X-CSRF-Token", csrfToken)
		for _, cookie := range getRr.Result().Cookies() {
			postReq.AddCookie(cookie)
		}
		postRr := httptest.NewRecorder()

		r.ServeHTTP(postRr, postReq)

		if status := postRr.Code; status != http.StatusOK {
			t.Errorf("处理器返回了错误的状态码: got %v want %v", status, http.StatusOK)
		}

		if body := postRr.Body.String(); body != "POST successful" {
			t.Errorf("未预期的响应体: got %v want %v", body, "POST successful")
		}
	})

	// 测试带有无效令牌的 POST 请求
	t.Run("POST with invalid token should be rejected", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/test", nil)
		req.Header.Set("X-CSRF-Token", "invalid_token")
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusForbidden {
			t.Errorf("处理器返回了错误的状态码: got %v want %v", status, http.StatusForbidden)
		}
	})
}

func TestCSRFWithCustomConfig(t *testing.T) {
	config := CSRFConfig{
		TokenLength: 64,
		CookieName:  "custom_csrf",
		ErrorHandler: func(err error, w http.ResponseWriter, r *http.Request) {
			http.Error(w, "自定义 CSRF 错误", http.StatusUnauthorized)
		},
	}

	r := mux.NewRouter()
	r.Use(CSRFWithConfig(config))

	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}).Methods("POST")

	t.Run("Custom config should be respected", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/test", nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("处理器返回了错误的状态码: got %v want %v", status, http.StatusUnauthorized)
		}

		if body := rr.Body.String(); !strings.Contains(body, "自定义 CSRF 错误") {
			t.Errorf("未预期的错误消息: got %v", body)
		}
	})
}
