package main

import (
	"github.com/sagoo-cloud/nexframe/os/zlog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

func main() {
	// 创建一个新的 LogHTTPWriter 实例
	httpWriter := zlog.NewLogHTTPWriter("http://localhost:8089/logs", true)

	// 创建一个多写入器，同时写入到控制台和 HTTP
	multiWriter := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stdout}, httpWriter)

	// 配置 zerolog
	logger := zerolog.New(multiWriter).With().Timestamp().Logger()

	// 创建一个新的 mux 路由器
	r := mux.NewRouter()

	// 定义一个中间件来记录请求
	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Msg("Incoming request")
			next.ServeHTTP(w, r)
		})
	}

	// 使用日志中间件
	r.Use(loggingMiddleware)

	// 定义路由
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/api", APIHandler) //在浏览器端访问 http://localhost:8080/api

	// 启动服务器
	http.ListenAndServe(":8080", r)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to the home page!"))
}

func APIHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is the API endpoint"))
}
