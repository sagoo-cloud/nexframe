package zlog

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog"
)

type LogHTTPWriter struct {
	Writer io.Writer

	serverURL     string
	warnOnHttpErr bool
	errorLogger   zerolog.Logger
}

func NewLogHTTPWriter(serverURL string, warnOnHttpErr bool) *LogHTTPWriter {
	return &LogHTTPWriter{
		serverURL:     serverURL,
		warnOnHttpErr: warnOnHttpErr,
		errorLogger:   zerolog.New(os.Stderr).With().Caller().Timestamp().Logger(),
	}
}

func (w *LogHTTPWriter) Write(p []byte) (n int, err error) {
	req, err := http.NewRequest("POST", w.serverURL, bytes.NewBuffer(p))
	if err != nil {
		if w.warnOnHttpErr {
			w.errorLogger.Warn().Err(err).Msg("Failed to create HTTP request for logging")
		}
		return len(p), err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if w.warnOnHttpErr {
			w.errorLogger.Warn().Err(err).Msg("Failed to send log via HTTP")
		}
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	return len(p), nil
}
