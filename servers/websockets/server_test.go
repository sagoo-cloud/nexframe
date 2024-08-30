package websockets

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/sagoo-cloud/nexframe/contracts"
	"github.com/sagoo-cloud/nexframe/servers/commons"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type mockHandler struct{}

func (m *mockHandler) ServeHandle(ctx context.Context, request interface{}) (interface{}, error) {
	return map[string]string{"message": "Hello, WebSocket!"}, nil
}

func TestWebSocketServer(t *testing.T) {
	logger := slog.Default()
	server := NewServer(WithLogger(logger), WithMaxConnections(10))

	server.Register("test", &commons.CommHandler{Handler: &mockHandler{}})

	testServer := httptest.NewServer(http.HandlerFunc(server.wsHandler))
	defer testServer.Close()

	url := "ws" + strings.TrimPrefix(testServer.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not connect to WebSocket server: %v", err)
	}
	defer ws.Close()

	testPayload := contracts.Payload{
		Route:  "test",
		Params: map[string]interface{}{"key": "value"},
	}
	err = ws.WriteJSON(testPayload)
	if err != nil {
		t.Fatalf("Could not send test message: %v", err)
	}

	_, message, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read response: %v", err)
	}

	t.Logf("Raw response: %s", string(message))

	var response map[string]interface{}
	err = json.Unmarshal(message, &response)
	if err != nil {
		t.Fatalf("Could not parse response: %v", err)
	}

	t.Logf("Parsed response: %+v", response)

	expectedMessage := "Hello, WebSocket!"
	if msg, ok := response["message"].(string); !ok || msg != expectedMessage {
		t.Errorf("Expected message '%s', got '%v'", expectedMessage, response["message"])
	}
}

func TestMaxConnections(t *testing.T) {
	maxConns := 2
	server := NewServer(WithMaxConnections(maxConns))

	testServer := httptest.NewServer(http.HandlerFunc(server.wsHandler))
	defer testServer.Close()

	url := "ws" + strings.TrimPrefix(testServer.URL, "http")

	// Connect maxConns times
	for i := 0; i < maxConns; i++ {
		_, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatalf("Connection %d failed: %v", i+1, err)
		}
	}

	// Try to connect one more time, which should fail
	_, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err == nil {
		t.Fatalf("Expected connection to fail, but it succeeded")
	}
	if resp == nil {
		t.Fatalf("Expected response to be non-nil")
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("Expected status 503, got %d", resp.StatusCode)
	}
}

func TestServerClose(t *testing.T) {
	server := NewServer()
	testServer := httptest.NewServer(http.HandlerFunc(server.wsHandler))
	defer testServer.Close()

	url := "ws" + strings.TrimPrefix(testServer.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Could not connect to WebSocket server: %v", err)
	}
	defer ws.Close()

	// Close the server
	err = server.Close()
	if err != nil {
		t.Fatalf("Failed to close server: %v", err)
	}

	// Function to check if the connection is closed
	isConnectionClosed := func() bool {
		err := ws.WriteMessage(websocket.PingMessage, []byte{})
		if err != nil {
			return true
		}
		_, _, err = ws.ReadMessage()
		return err != nil
	}

	// Wait for the connection to close with a timeout
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if isConnectionClosed() {
				return // Test passes
			}
		case <-timeout:
			t.Fatalf("Connection did not close within the expected timeframe")
		}
	}
}
