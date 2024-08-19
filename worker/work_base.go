package worker

import (
	"context"
	"github.com/hibiken/asynq"
)

type BaseProcess struct{}

func (q *BaseProcess) GetTopic() string {
	return ""
}

func (q *BaseProcess) GetProcessConfig() *ProcessConfig {
	return nil
}

func (q *BaseProcess) Handle(ctx context.Context, p Payload) (err error) {
	return
}

// HandleAggregate 处理消息
func (q *BaseProcess) HandleAggregate(ctx context.Context, task *asynq.Task) (err error) {
	return
}
