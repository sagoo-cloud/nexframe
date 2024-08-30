package httputil

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
)

// TestNewGetRequest 测试创建 GET 请求
func TestNewGetRequest(t *testing.T) {
	urlStr := "http://baidu.com"
	params := map[string]interface{}{"key": "value"}
	req, err := NewGetRequest(urlStr, params)
	if err != nil {
		t.Fatalf("NewGetRequest failed: %v", err)
	}
	if req.Method != http.MethodGet {
		t.Errorf("Expected GET method, got %s", req.Method)
	}
	if req.URL.String() != "http://baidu.com?key=value" {
		t.Errorf("Unexpected URL: %s", req.URL.String())
	}
}

// TestNewFormPostRequest 测试创建表单 POST 请求
func TestNewFormPostRequest(t *testing.T) {
	testURL := "http://example.com"
	params := map[string]interface{}{
		"key1": "value1",
		"key2": []string{"value2", "value3"},
		"key3": 123,
	}
	req, err := NewFormPostRequest(testURL, params)
	if err != nil {
		t.Fatalf("NewFormPostRequest failed: %v", err)
	}

	// 检查请求方法
	if req.Method != http.MethodPost {
		t.Errorf("Expected POST method, got %s", req.Method)
	}

	// 检查 Content-Type
	if req.Header.Get("Content-Type") != ContentTypeForm {
		t.Errorf("Unexpected Content-Type: %s", req.Header.Get("Content-Type"))
	}

	// 检查 URL
	if req.URL.String() != testURL {
		t.Errorf("Unexpected URL: got %s, want %s", req.URL.String(), testURL)
	}

	// 读取请求体
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("Failed to read request body: %v", err)
	}

	// 解析和检查表单数据
	formData, err := url.ParseQuery(string(body))
	if err != nil {
		t.Fatalf("Failed to parse form data: %v", err)
	}

	expectedData := url.Values{
		"key1": []string{"value1"},
		"key2": []string{"value2", "value3"},
		"key3": []string{"123"},
	}

	for key, expectedValues := range expectedData {
		if !reflect.DeepEqual(formData[key], expectedValues) {
			t.Errorf("Unexpected value for %s: got %v, want %v", key, formData[key], expectedValues)
		}
	}
}

// TestNewJSONPostRequest 测试创建 JSON POST 请求
func TestNewJSONPostRequest(t *testing.T) {
	urlStr := "http://example.com"
	params := map[string]interface{}{"key": "value"}
	req, err := NewJSONPostRequest(urlStr, params)
	if err != nil {
		t.Fatalf("NewJSONPostRequest failed: %v", err)
	}
	if req.Method != http.MethodPost {
		t.Errorf("Expected POST method, got %s", req.Method)
	}
	if req.Header.Get("Content-Type") != ContentTypeJSON {
		t.Errorf("Unexpected Content-Type: %s", req.Header.Get("Content-Type"))
	}
}

// TestGet 测试 GET 请求
func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	resp, err := Get(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}
}

// TestPost 测试 POST 请求
func TestPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != ContentTypeForm {
			t.Errorf("Unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	resp, err := Post(ctx, server.URL, map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("Post request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}
}

// TestPostJSON 测试 JSON POST 请求
func TestPostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != ContentTypeJSON {
			t.Errorf("Unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}
		var data map[string]interface{}
		json.NewDecoder(r.Body).Decode(&data)
		if data["key"] != "value" {
			t.Errorf("Unexpected body content")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	resp, err := PostJSON(ctx, server.URL, map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("PostJSON request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}
}

// TestSetTraceIDInHeader 测试设置 Trace ID
func TestSetTraceIDInHeader(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	ctx := context.WithValue(context.Background(), "traceID", "test-trace-id")
	setTraceIDInHeader(ctx, req)
	if req.Header.Get("X-TRACE-ID") != "test-trace-id" {
		t.Errorf("Expected X-TRACE-ID header to be set")
	}
}

// TestGetTimeout 测试获取超时时间
func TestGetTimeout(t *testing.T) {
	timeout := getTimeout(map[string]interface{}{"timeout": 10})
	if timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", timeout)
	}

	timeout = getTimeout()
	if timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", timeout)
	}
}
