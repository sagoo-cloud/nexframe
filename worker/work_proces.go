package worker

import (
	"context"
	"errors"
	"github.com/hibiken/asynq"
	"github.com/sagoo-cloud/nexframe/g"
	"github.com/sagoo-cloud/nexframe/utils/guid"
	"time"
)

// Scheduled 任务调度器
type Scheduled struct {
	topic string
	w     *Worker
}

const (
	DefaultProcess   = "defaultProcess"   // 默认处理
	AggregateProcess = "aggregateProcess" // 聚合处理
	CronProcess      = "cronProcess"      // 定时任务
)

type ProcessConfig struct {
	Topic       string
	ProcessType string
}

// Process 任务具体处理过程接口，实现该接口即可加入到任务队列中
type Process interface {
	GetTopic() string
	GetProcessConfig() *ProcessConfig                                  // 获取消费主题
	Handle(ctx context.Context, p Payload) (err error)                 // 处理过程的方法
	HandleAggregate(ctx context.Context, task *asynq.Task) (err error) // 处理聚合过程的方法
}

func RegisterProcess(p Process) (s *Scheduled) {
	s = &Scheduled{}
	processConfig := p.GetProcessConfig()
	if processConfig != nil {
		switch processConfig.ProcessType {
		case AggregateProcess:
			s.w = New(
				WithGroup(processConfig.Topic),
				WithUseAggregator(true),
				WithHandlerAggregator(p.HandleAggregate),
			)
		default:
		}
	} else {
		topic := p.GetTopic()
		if topic != "" {
			s.w = New(
				WithGroup(topic),
				WithHandler(p.Handle),
			)
		} else {
			s.w = New(
				WithHandler(p.Handle),
			)
		}
	}
	if s.w == nil {
		g.Log.Debugf(context.Background(), "========== RegisterProcess Worker is nil")
	}
	return
}

// Push 采用消息队列的方式执行任务
func (s *Scheduled) Push(ctx context.Context, topic string, dataId string, data []byte, timeout int) (err error) {
	if dataId == "" {
		dataId = guid.S()
	}
	if s.w == nil {
		return errors.New("worker is nil")
	}
	err = s.w.Once(
		WithRunUuid(dataId),
		WithRunPayload(data), //传递参数
		WithRunGroup(topic),
		WithRunAt(time.Now().Add(time.Duration(1)*time.Second)), //延迟执行
		WithRunTimeout(timeout),                                 //超时时间
	)
	if err != nil {
		g.Log.Debugf(ctx, "Run Queue TaskWorker %s Error: %v", topic, err)
	}
	return
}

// Cron 采用定时任务的方式执行任务
func (s *Scheduled) Cron(ctx context.Context, topic, cronExpr string, data []byte) (err error) {
	s.topic = topic
	err = s.w.Cron(
		WithRunUuid(topic),
		WithRunGroup(topic),
		WithRunExpr(cronExpr),
		WithRunPayload(data), //传递参数
	)
	if err != nil {
		g.Log.Debugf(ctx, "Run Cron TaskWorker %s Error: %v", topic, err)
	}
	return
}
