package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// TestMiddleware 测试中间件的基本功能
func TestMiddleware(t *testing.T) {
	// 创建一个测试用的存储
	store := sessions.NewCookieStore([]byte("test-secret"))

	// 创建一个使用中间件的路由器
	r := mux.NewRouter()
	r.Use(Middleware(store))

	// 添加一个测试路由
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		// 尝试获取会话
		session, err := Get("test-session", r)
		if err != nil {
			t.Errorf("获取会话失败: %v", err)
			return
		}
		// 设置一个会话值
		session.Values["test-key"] = "test-value"
		session.Save(r, w)
		w.WriteHeader(http.StatusOK)
	})

	// 创建一个测试请求
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(rr, req)

	// 检查响应状态码
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("处理器返回了错误的状态码: 得到 %v 想要 %v", status, http.StatusOK)
	}

	// 检查设置的 cookie
	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Error("没有设置 cookie")
	}
}

// TestMiddlewareWithConfig 测试使用自定义配置的中间件
func TestMiddlewareWithConfig(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test-secret"))

	config := Config{
		Skipper: func(r *http.Request) bool {
			return r.URL.Path == "/skip"
		},
		Store: store,
	}

	r := mux.NewRouter()
	r.Use(MiddlewareWithConfig(config))

	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		_, err := Get("test-session", r)
		if err != nil {
			t.Errorf("获取会话失败: %v", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	r.HandleFunc("/skip", func(w http.ResponseWriter, r *http.Request) {
		_, err := Get("test-session", r)
		if err == nil {
			t.Error("跳过的路由不应该有会话")
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// 测试正常路由
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("处理器返回了错误的状态码: 得到 %v 想要 %v", status, http.StatusOK)
	}

	// 测试跳过的路由
	req = httptest.NewRequest("GET", "/skip", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("处理器返回了错误的状态码: 得到 %v 想要 %v", status, http.StatusOK)
	}
}
