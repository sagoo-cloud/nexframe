package middleware

import (
	"bytes"
	"net/http"
)

// customResponseWriter is a custom response writer that captures the status code and response size
type customResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
	buf    bytes.Buffer
}

func (crw *customResponseWriter) WriteHeader(status int) {
	crw.status = status
	crw.ResponseWriter.WriteHeader(status)
}

func (crw *customResponseWriter) Write(b []byte) (int, error) {
	size, err := crw.ResponseWriter.Write(b)
	crw.size += size
	return size, err
}
