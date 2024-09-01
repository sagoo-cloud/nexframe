package configs

type WebSocketConfig struct {
	WebSocketHost string `json:"webSocketHost"`
	WebSocketPort string `json:"webSocketPort"`
	Path          string
}

func LoadWebSocketConfig() *WebSocketConfig {
	config := &WebSocketConfig{
		Path:          "/ws",
		WebSocketPort: EnvString("servers.websocketPort", "8189"),
		WebSocketHost: EnvString("servers.websocketHost", "127.0.0.1"),
	}
	return config
}
