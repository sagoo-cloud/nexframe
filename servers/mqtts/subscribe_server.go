package mqtts

import (
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sagoo-cloud/nexframe/configs"
	"github.com/sagoo-cloud/nexframe/net/mqttclient"
	"github.com/sagoo-cloud/nexframe/servers/commons"
	"log/slog"
)

type Server struct {
	topics       map[string]*commons.CommHandler
	Logger       *slog.Logger
	Parallel     bool //并行处理
	SubscribeQos byte
}

func NewServer() *Server {
	config := configs.LoadMqttConfig()
	ss := &Server{
		topics:       make(map[string]*commons.CommHandler),
		Parallel:     config.Parallel,
		SubscribeQos: config.SubscribeQos,
	}
	return ss
}
func (s *Server) Register(name string, handler *commons.CommHandler) {
	s.topics[name] = handler

}
func (s *Server) Serve() error {
	if GetIns() != nil {
		errChans := make(map[string]chan error)
		s.work(errChans)
		for _, errChan := range errChans {
			if errChan != nil {
				s.Logger.Info("errChan:", <-errChan)
			}
		}
	} else {
		s.Logger.Info("MQTT链接失败")
	}
	return nil
}

func (s *Server) work(errChans map[string]chan error) {
	s.Logger.Info("MQTT Subscribe Server Start")
	for topic, handler := range s.topics {
		errChans[topic] = make(chan error)
		go s.worker(topic, handler, errChans[topic])
	}

}
func (s *Server) worker(t string, h *commons.CommHandler, e chan error) {
	s.Logger.Info("Subscribe topic:%s", t)
	// 创建消息处理器
	handler := mqttclient.Handler{
		Topic: t,
		Qos:   s.SubscribeQos,
		Handle: func(
			client mqtt.Client, message mqtt.Message) {
			if s.Parallel {
				go s.process(h, message)
			} else {
				s.process(h, message)
			}
		},
	}
	// 注册处理器
	if err := GetIns().RegisterHandler(handler); err != nil {
		panic(err)
	}
}
func (s *Server) process(h *commons.CommHandler, Message mqtt.Message) {
	s.Logger.Info("subscribe topic:", Message.Topic())
	resp, err := h.Handle(context.Background(), Message.Payload())
	if err != nil {
		s.Logger.Info(err.Error())

	} else {
		s.Logger.Info("resp:", resp)
	}
}

func (s *Server) Close() {
	if GetIns() != nil {
		for topic := range s.topics {
			GetIns().Close()
			s.Logger.Info("Unsubscribe topic:%s", topic)
		}
		GetIns().GetClient().Disconnect(uint(250))
	}
}
