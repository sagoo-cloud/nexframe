package mqtts

import (
	"encoding/json"
	"errors"
	"github.com/sagoo-cloud/nexframe/configs"
)

func Publish(topic string, payload interface{}) error {
	if GetIns() == nil {
		return errors.New("MQTT链接失败")
	}
	param, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	config := configs.LoadMqttConfig()
	token := GetIns().Publish(topic, config.PublishQos, false, param)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
