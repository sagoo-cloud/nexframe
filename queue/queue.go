package queue

import (
	"fmt"
	"sync"
)

const (
	DriverTypeRedis    = "redis"
	DriverTypeAliMns   = "ali_mns"
	DriverTypeAliyunMq = "aliyun_mq"
	DriverTypeRocketMq = "rocket_mq"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Instance)
)

// Instance 是一个函数类型，用于创建 Queue 实例
type Instance func(diName string) Queue

// Register 注册一个队列驱动
func Register(driverType string, driver Instance) {
	driversMu.Lock()
	defer driversMu.Unlock()

	if driver == nil {
		panic("queue: Register driver is nil")
	}
	if _, dup := drivers[driverType]; dup {
		panic("queue: Register called twice for driver " + driverType)
	}
	drivers[driverType] = driver
}

// GetQueue 获取Queue对象
func GetQueue(diName string, driverType string) Queue {
	driversMu.RLock()
	instanceFunc, ok := drivers[driverType]
	driversMu.RUnlock()

	if !ok {
		panic(fmt.Sprintf("queue: unknown driver %s", driverType))
	}

	q := instanceFunc(diName)
	if q == nil {
		panic(fmt.Sprintf("queue: nil queue returned for diName %s", diName))
	}

	return q
}
