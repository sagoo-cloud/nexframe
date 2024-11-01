package configs

import "time"

type MqttConfig struct {
	Host         string `json:"host"`         // MQTT broker地址
	UserName     string `json:"userName"`     // 用户名
	PassWord     string `json:"passWord"`     // 密码
	ClientID     string `json:"clientID"`     // 客户端ID
	Parallel     bool   `json:"parallel"`     // 并行处理
	SubscribeQos uint8  `json:"subscribeQos"` // 订阅Qos
	PublishQos   uint8  `json:"publishQos"`   // 发布Qos

	CleanSession         bool          `json:"cleanSession"`         // 清理会话标志
	MaxReconnectInterval time.Duration `json:"maxReconnectInterval"` // 重连间隔
	CAFile               string        `json:"cAFile"`               // CA证书文件
	CertFile             string        `json:"certFile"`             // 客户端证书
	CertKeyFile          string        `json:"certKeyFile"`          // 客户端密钥
	LogLevel             int           `json:"logLevel"`             // 日志级别
	QueueSize            int           `json:"queueSize"`            // 消息队列大小
}

func LoadMqttConfig() *MqttConfig {
	config := &MqttConfig{
		Host:                 EnvString(MqttHost, "tcp://127.0.0.1:1883"),
		UserName:             EnvString(MattUsername, ""),
		PassWord:             EnvString(MqttPassword, ""),
		ClientID:             EnvString(MqttClientID, "211"),
		Parallel:             EnvBool(MqttParallel, false),
		SubscribeQos:         uint8(EnvInt(MqttSubscribeQos, 0)),
		PublishQos:           uint8(EnvInt(MqttPublishQos, 0)),
		CleanSession:         EnvBool(MqttCleanSession, true),
		MaxReconnectInterval: time.Duration(EnvInt(MqttMaxReconnectInterval, 60)) * time.Second,
		CAFile:               EnvString(MqttCAFile, ""),
		CertFile:             EnvString(MqttCertFile, ""),
		CertKeyFile:          EnvString(MqttCertKeyFile, ""),
		LogLevel:             EnvInt(MqttLogLevel, 0),
		QueueSize:            EnvInt(MqttQueueSize, 100),
	}
	return config
}
