package configs

import "time"

type QueueConfig struct {
	Prefix      string
	Listen      []string
	Interval    time.Duration
	Concurrency int
}

func LoadQueueConfig() *QueueConfig {
	interval := EnvInt(QueueInterval, 1)
	config := &QueueConfig{
		Prefix:      EnvString(QueuePrefix, "sagoo"),
		Listen:      EnvStringSlice(QueueListen),
		Interval:    time.Duration(interval) * time.Second,
		Concurrency: EnvInt(QueueConcurrency, 1),
	}
	return config
}
