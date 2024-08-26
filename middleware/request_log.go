package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	timeISO8601 = "2006-01-02T15:04:05.000Z0700"
)

var sensitiveHeaders = map[string]bool{
	"authorization": true,
	"cookie":        true,
	"token":         true,
	"session":       true,
}

// RequestLog is a middleware that logs request details
func RequestLog(logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Capture the request body
			var body []byte
			if r.Body != nil {
				body, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(body))
			}

			// Capture headers (excluding sensitive ones)
			headers := make(map[string]string)
			for k, v := range r.Header {
				if !sensitiveHeaders[strings.ToLower(k)] {
					headers[k] = strings.Join(v, ", ")
				}
			}

			// Create a custom response writer to capture the status code and size
			crw := &customResponseWriter{ResponseWriter: w}

			// Process request
			next.ServeHTTP(crw, r)

			// Calculate request duration
			duration := time.Since(start)

			// Prepare log entry
			logEntry := []slog.Attr{
				slog.Int("status", crw.status),
				slog.String("method", r.Method),
				slog.String("path", r.URL.String()),
				slog.Any("headers", headers),
				slog.Int("size", crw.size),
				slog.String("clientIP", getClientIP(r)),
				slog.String("user-agent", r.UserAgent()),
				slog.String("start", start.Format(timeISO8601)),
				slog.Duration("latency", duration),
			}

			// Add request ID if available
			if reqID, ok := r.Context().Value(RequestIDKey).(string); ok {
				logEntry = append(logEntry, slog.String(RequestIDKey, reqID))
			}

			// Add body if not empty
			if len(body) > 0 {
				logEntry = append(logEntry, slog.String("body", string(body)))
			}

			// Log the request details
			logger.LogAttrs(r.Context(), slog.LevelInfo, "Request processed", logEntry...)
		})
	}
}

// getClientIP tries to get the real client IP
func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}
	}
	return ip
}
