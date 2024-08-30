package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// Recovery 中间件用于捕获 Panic 并返回 500 错误
func Recovery(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// 记录 Panic 信息
				fmt.Printf("panic: %v\n", err)
				debug.PrintStack()

				// 返回 500 错误
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
