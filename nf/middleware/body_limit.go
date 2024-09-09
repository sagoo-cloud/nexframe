package middleware

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/sagoo-cloud/nexframe/utils/bytes"
)

// BodyLimitConfig 定义了BodyLimit中间件的配置
type BodyLimitConfig struct {
	// Limit 指定了请求体的最大允许大小
	// 可以使用如 "4x" 或 "4xB" 的格式，其中x可以是K, M, G, T 或 P
	Limit string
	limit int64
}

// limitedReader 是一个包装了io.ReadCloser的结构体，用于限制读取的字节数
type limitedReader struct {
	config BodyLimitConfig
	reader io.ReadCloser
	read   int64
}

// BodyLimit 返回一个用于限制请求体大小的中间件
func BodyLimit(limit string) mux.MiddlewareFunc {
	config := BodyLimitConfig{Limit: limit}
	limitBytes, err := bytes.Parse(config.Limit)
	if err != nil {
		panic(fmt.Errorf("invalid body-limit=%s", config.Limit))
	}
	config.limit = limitBytes

	pool := &sync.Pool{
		New: func() interface{} {
			return &limitedReader{config: config}
		},
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查Content-Length
			if r.ContentLength > config.limit {
				http.Error(w, "Request Entity Too Large", http.StatusRequestEntityTooLarge)
				return
			}

			// 使用limitedReader包装请求体
			lr := pool.Get().(*limitedReader)
			lr.Reset(r.Body)
			defer pool.Put(lr)
			r.Body = lr

			next.ServeHTTP(w, r)
		})
	}
}

// Read 实现了io.Reader接口
func (r *limitedReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.read += int64(n)
	if r.read > r.config.limit {
		return n, fmt.Errorf("request entity too large")
	}
	return
}

// Close 实现了io.Closer接口
func (r *limitedReader) Close() error {
	return r.reader.Close()
}

// Reset 重置limitedReader以便重用
func (r *limitedReader) Reset(reader io.ReadCloser) {
	r.reader = reader
	r.read = 0
}
