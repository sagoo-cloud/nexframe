package configs

const (
	MqttHost         = "mqtt.host"
	MattUsername     = "mqtt.username"
	MqttPassword     = "mqtt.password"
	MqttClientID     = "mqtt.client_id"
	MqttParallel     = "mqtt.parallel"
	MqttSubscribeQos = "mqtt.subscribe_qos"
	MqttPublishQos   = "mqtt.publish_qos"
)

type MqttConfig struct {
	Host         string `json:"host"`
	UserName     string `json:"user_name"`
	PassWord     string `json:"pass_word"`
	ClientID     string `json:"client_id"`
	Parallel     bool   `json:"parallel"`
	SubscribeQos uint8  `json:"subscribe_qos"`
	PublishQos   uint8  `json:"publish_qos"`
}

func LoadMqttConfig() *MqttConfig {
	config := &MqttConfig{
		Host:         EnvString(MqttHost, "tcp://127.0.0.1:1883"),
		UserName:     EnvString(MattUsername, ""),
		PassWord:     EnvString(MqttPassword, ""),
		ClientID:     EnvString(MqttClientID, "211"),
		Parallel:     EnvBool(MqttParallel, false),
		SubscribeQos: uint8(EnvInt(MqttSubscribeQos, 2)),
		PublishQos:   uint8(EnvInt(MqttPublishQos, 2)),
	}
	return config
}
