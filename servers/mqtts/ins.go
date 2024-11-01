package mqtts

import (
	"context"
	"fmt"
	"github.com/sagoo-cloud/nexframe/configs"
	"github.com/sagoo-cloud/nexframe/net/mqttclient"
	"sync"
	"time"
)

var ins *mqttclient.Client
var once sync.Once

func GetIns() *mqttclient.Client {
	once.Do(func() {
		ins = initMqtt()
	})
	return ins
}
func initMqtt() *mqttclient.Client {
	config := configs.LoadMqttConfig()
	// 创建上下文
	ctx := context.Background()
	// 创建配置
	conf := mqttclient.Config{
		Server:    config.Host,
		Username:  config.UserName,
		Password:  config.PassWord,
		Logger:    nil,
		LogLevel:  mqttclient.IntToLogLevel(config.LogLevel), // 设置日志级别为INFO
		QueueSize: 100,                                       // 设置消息队列大小
	}
	if config.CAFile != "" && config.CertFile != "" {
		conf.CAFile = config.CAFile
		conf.CertFile = config.CertFile
		conf.CertKeyFile = config.CertKeyFile
	}

	// 创建MQTT客户端
	client, err := mqttclient.NewClient(ctx, conf)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// 监控连接状态
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if !client.IsConnected() {
					fmt.Println("Client disconnected, waiting for reconnect...")
				}
			}
		}
	}()

	return client
}
