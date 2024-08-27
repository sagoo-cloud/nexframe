package database

import (
	"errors"
	"log"
	"runtime"
	"sync"
	"time"
)

// Aggregator 聚合器结构体
type Aggregator struct {
	option          AggregatorOption
	wg              sync.WaitGroup
	quit            chan struct{}
	eventQueue      chan interface{}
	batchProcessor  BatchProcessFunc
	pool            *sync.Pool
	lingerTimer     *time.Timer
	lastProcessTime time.Time
}

// AggregatorOption 聚合器选项
type AggregatorOption struct {
	BatchSize         int
	Workers           int
	ChannelBufferSize int
	LingerTime        time.Duration
	ErrorHandler      ErrorHandlerFunc
	Logger            *log.Logger
}

// BatchProcessFunc 批处理函数类型
type BatchProcessFunc func([]interface{}) error

// SetAggregatorOptionFunc 聚合器选项设置函数类型
type SetAggregatorOptionFunc func(option *AggregatorOption)

// ErrorHandlerFunc 错误处理函数类型
type ErrorHandlerFunc func(err error, items []interface{}, batchProcessFunc BatchProcessFunc, aggregator *Aggregator)

// NewAggregator 创建新的聚合器实例
func NewAggregator(batchProcessor BatchProcessFunc, optionFuncs ...SetAggregatorOptionFunc) (*Aggregator, error) {
	option := AggregatorOption{
		BatchSize:  8,
		Workers:    runtime.NumCPU(),
		LingerTime: 1 * time.Minute,
	}

	for _, optionFunc := range optionFuncs {
		optionFunc(&option)
	}

	if option.ChannelBufferSize < option.Workers {
		option.ChannelBufferSize = option.Workers
	}

	pool := &sync.Pool{
		New: func() interface{} {
			return make([]interface{}, 0, option.BatchSize)
		},
	}

	lingerTimer := time.NewTimer(option.LingerTime)

	aggregator := &Aggregator{
		eventQueue:      make(chan interface{}, option.ChannelBufferSize),
		option:          option,
		quit:            make(chan struct{}),
		batchProcessor:  batchProcessor,
		pool:            pool,
		lingerTimer:     lingerTimer,
		lastProcessTime: time.Now(),
	}

	return aggregator, nil
}

// TryEnqueue 尝试入队一个项目，非阻塞
func (agt *Aggregator) TryEnqueue(item interface{}) bool {
	select {
	case agt.eventQueue <- item:
		return true
	default:
		if agt.option.Logger != nil {
			agt.option.Logger.Println("Aggregator: 事件队列已满，尝试重新安排")
		}
		runtime.Gosched() // 让出CPU时间片
		select {
		case agt.eventQueue <- item:
			return true
		default:
			if agt.option.Logger != nil {
				agt.option.Logger.Printf("Aggregator: 事件队列仍然已满，并且跳过了 %+v \n", item)
			}
			return false
		}
	}
}

// Enqueue 入队一个项目，会阻塞直到有空间
func (agt *Aggregator) Enqueue(item interface{}) error {
	select {
	case agt.eventQueue <- item:
		return nil
	case <-agt.quit:
		return errors.New("aggregator is stopping")
	}
}

// EnqueueWithRetry 入队一个项目，会重试直到成功或者达到最大重试次数
func (agt *Aggregator) EnqueueWithRetry(item interface{}, maxRetries int, backoff time.Duration) bool {
	for i := 0; i < maxRetries; i++ {
		if err := agt.Enqueue(item); err == nil {
			return true // 入队成功
		}
		time.Sleep(backoff) // 等待一段时间后重试
		backoff *= 2        // 指数退避
	}
	return false // 最终尝试失败
}

// Start 启动聚合器
func (agt *Aggregator) Start() {
	agt.wg.Add(agt.option.Workers)
	for i := 0; i < agt.option.Workers; i++ {
		go agt.work()
	}
}

// Stop 停止聚合器
func (agt *Aggregator) Stop() {
	close(agt.quit)
	agt.wg.Wait()
}

// SafeStop 安全停止聚合器，确保所有项目都被处理
func (agt *Aggregator) SafeStop() {
	close(agt.quit)
	agt.wg.Wait() // 等待所有工作协程退出

	// 等待最后一个计时器完成
	agt.lingerTimer.Stop()
	<-agt.lingerTimer.C

	// 处理剩余的事件
	agt.processBatch(agt.getBatchFromQueue())
}

func (agt *Aggregator) work() {
	defer agt.wg.Done()

	batch := agt.pool.Get().([]interface{})
	defer agt.pool.Put(batch[:0])

	for {
		select {
		case item := <-agt.eventQueue:
			batch = append(batch, item)
			if len(batch) == cap(batch) || time.Since(agt.lastProcessTime) >= agt.option.LingerTime {
				agt.processBatch(batch)
				batch = batch[:0]                // 清空切片
				agt.lastProcessTime = time.Now() // 更新最近处理时间
			}
		case <-agt.quit:
			if len(batch) > 0 {
				agt.processBatch(batch)
			}
			return // 退出工作协程
		}
	}
}

func (agt *Aggregator) processBatch(items []interface{}) {
	defer agt.wg.Add(1)
	defer agt.wg.Done()
	if err := agt.batchProcessor(items); err != nil {
		if agt.option.Logger != nil {
			agt.option.Logger.Println("Aggregator: 处理批次时发生错误")
		}
		if agt.option.ErrorHandler != nil {
			agt.option.ErrorHandler(err, items, agt.batchProcessor, agt)
		}
		// 记录错误信息，例如将错误事件保存到其他地方
	} else if agt.option.Logger != nil {
		agt.option.Logger.Printf("Aggregator: 成功处理了%d个项目。\n", len(items))
	}
}

// 示例: 设置聚合器选项的函数
func WithBatchSize(batchSize int) SetAggregatorOptionFunc {
	return func(option *AggregatorOption) {
		option.BatchSize = batchSize
	}
}

func WithWorkers(workers int) SetAggregatorOptionFunc {
	return func(option *AggregatorOption) {
		option.Workers = workers
	}
}

func WithChannelBufferSize(size int) SetAggregatorOptionFunc {
	return func(option *AggregatorOption) {
		option.ChannelBufferSize = size
	}
}

func WithLingerTime(duration time.Duration) SetAggregatorOptionFunc {
	return func(option *AggregatorOption) {
		option.LingerTime = duration
	}
}

func WithLogger(logger *log.Logger) SetAggregatorOptionFunc {
	return func(option *AggregatorOption) {
		option.Logger = logger
	}
}

func WithErrorHandler(handler ErrorHandlerFunc) SetAggregatorOptionFunc {
	return func(option *AggregatorOption) {
		option.ErrorHandler = handler
	}
}

// getBatchFromQueue 从事件队列中获取一个批处理
func (agt *Aggregator) getBatchFromQueue() []interface{} {
	batch := make([]interface{}, 0, agt.option.BatchSize)
	for {
		select {
		case item := <-agt.eventQueue:
			batch = append(batch, item)
		case <-time.After(agt.option.LingerTime):
			return batch
		}
	}
}
