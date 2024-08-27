package mqtts

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/sagoo-cloud/nexframe/configs"
	"sync"
)

var ins mqtt.Client
var once sync.Once

func GetIns() mqtt.Client {
	once.Do(func() {
		ins = init_mc()
	})
	return ins
}
func init_mc() mqtt.Client {
	config := configs.LoadMqttConfig()
	opts := mqtt.NewClientOptions().AddBroker(config.Host)
	opts.SetUsername(config.UserName)
	opts.SetPassword(config.PassWord)
	opts.SetClientID(config.ClientID)
	mc := mqtt.NewClient(opts)
	if token := mc.Connect(); token.Wait() && token.Error() != nil {
		return nil
	}
	return mc
}
