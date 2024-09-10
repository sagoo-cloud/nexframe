package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestBodyLimit(t *testing.T) {
	// 创建一个测试用的处理函数
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	})

	// 创建一个带有BodyLimit中间件的路由器
	r := mux.NewRouter()
	r.Use(BodyLimit("10B")) // 设置10字节的限制
	r.Handle("/test", testHandler)

	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "请求体小于限制",
			body:           "12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "请求体等于限制",
			body:           "1234567890",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "请求体大于限制",
			body:           "12345678901",
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个测试请求
			req, err := http.NewRequest("POST", "/test", bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			// 创建一个ResponseRecorder来记录响应
			rr := httptest.NewRecorder()

			// 处理请求
			r.ServeHTTP(rr, req)

			// 检查状态码
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("处理程序返回了错误的状态码：得到 %v 想要 %v",
					status, tt.expectedStatus)
			}

			// 如果期望成功，检查响应体
			if tt.expectedStatus == http.StatusOK {
				if rr.Body.String() != tt.body {
					t.Errorf("处理程序返回了意外的体：得到 %v 想要 %v",
						rr.Body.String(), tt.body)
				}
			}
		})
	}
}
