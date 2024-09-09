package nf

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe/contracts"
	"net/http"
)

// ErrorResponse 定义了统一的错误响应结构
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewErrorResponse 创建一个新的ErrorResponse
func NewErrorResponse(code string, message string) ErrorResponse {
	return ErrorResponse{
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	}
}

// ErrorHandlingMiddleware 是一个自定义的错误处理中间件
func (f *APIFramework) ErrorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 创建一个自定义的ResponseWriter来捕获状态码
		crw := &customResponseWriter{ResponseWriter: w, framework: f}

		// 使用recover来捕获可能的panic
		defer func() {
			if err := recover(); err != nil {
				f.handleError(crw, fmt.Errorf("%v", err), http.StatusInternalServerError)
			}
		}()

		// 调用下一个处理器
		next.ServeHTTP(crw, r)
	})
}

// handleError 处理错误并发送JSON格式的错误响应
func (f *APIFramework) handleError(w http.ResponseWriter, err error, status int) {
	var errorCode int
	var errorMessage string

	switch status {
	case http.StatusBadRequest:
		errorCode = 400
		errorMessage = "Invalid request"
	case http.StatusUnauthorized:
		errorCode = 401
		errorMessage = "Authentication required"
	case http.StatusForbidden:
		errorCode = 403
		errorMessage = "Access denied"
	case http.StatusNotFound:
		errorCode = 404
		errorMessage = "Resource not found"
	case http.StatusMethodNotAllowed:
		errorCode = 405
		errorMessage = "Method not allowed"
	default:
		errorCode = 500
		errorMessage = "An unexpected error occurred"
	}
	if err != nil {
		errorMessage = err.Error()
	}

	contracts.JsonExit(w, errorCode, errorMessage)
}

// customResponseWriter 是一个自定义的ResponseWriter，用于捕获状态码和错误
type customResponseWriter struct {
	http.ResponseWriter
	status    int
	framework *APIFramework
}

func (crw *customResponseWriter) WriteHeader(status int) {
	crw.status = status
	if status >= 400 {
		// 如果是错误状态码，调用handleError
		crw.framework.handleError(crw.ResponseWriter, nil, status)
	} else {
		crw.ResponseWriter.WriteHeader(status)
	}
}

func (crw *customResponseWriter) Write(b []byte) (int, error) {
	if crw.status == 0 {
		crw.status = 200
	}
	return crw.ResponseWriter.Write(b)
}

// UseErrorHandlingMiddleware 在APIFramework结构体中添加一个方法来应用这个中间件
func (f *APIFramework) UseErrorHandlingMiddleware() {
	f.WithMiddleware(f.ErrorHandlingMiddleware)
}

// CustomError 是一个辅助函数，用于在控制器中返回自定义错误
func CustomError(w http.ResponseWriter, err error, status int) {
	if cw, ok := w.(*customResponseWriter); ok {
		cw.framework.handleError(w, err, status)
	} else {
		http.Error(w, err.Error(), status)
	}
}
