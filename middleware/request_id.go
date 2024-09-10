package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

// RequestIDKey is the key used to store the request ID in the context
const RequestIDKey = "requestID"

// GenerateUUID generates a random UUID
func GenerateUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return ""
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return hex.EncodeToString(uuid)
}

// RequestID is a middleware that adds a unique request id to the request context
func RequestID(logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := GenerateUUID()

			// Add request ID to the context
			ctx := context.WithValue(r.Context(), RequestIDKey, reqID)

			// Create a new logger with the request ID
			reqLogger := logger.With(slog.String(RequestIDKey, reqID))

			// Create a new request with the updated context
			r = r.WithContext(ctx)

			// Call the next handler with the updated request
			next.ServeHTTP(w, r)

			// Log after the request is processed
			reqLogger.Info("Request processed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)
		})
	}
}
