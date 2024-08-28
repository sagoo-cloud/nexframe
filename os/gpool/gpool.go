package gpool

import (
	"context"
	"sync"
	"sync/atomic"
)

type GPool struct {
	workers   chan struct{}
	workQueue chan func()
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	running   int64
}

// NewGPool 创建并初始化具有指定容量的新GPool。
func NewGPool(capacity int) *GPool {
	ctx, cancel := context.WithCancel(context.Background())
	p := &GPool{
		workers:   make(chan struct{}, capacity),
		workQueue: make(chan func(), capacity*2), // Double capacity for work queue
		ctx:       ctx,
		cancel:    cancel,
	}

	// Start worker goroutines
	for i := 0; i < capacity; i++ {
		go p.worker()
	}

	return p
}

func (p *GPool) worker() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case job := <-p.workQueue:
			atomic.AddInt64(&p.running, 1)
			job()
			atomic.AddInt64(&p.running, -1)
			p.wg.Done()
		}
	}
}

// Go 将任务提交到池中执行。
func (p *GPool) Go(job func(ctx context.Context) error) {
	p.wg.Add(1)
	select {
	case p.workQueue <- func() {
		err := job(p.ctx)
		if err != nil {
			println("Error executing job:", err.Error())
		}
	}:
		//作业已成功添加到队列
	default:
		//如果队列已满，请在新的goroutine中执行作业
		go func() {
			defer p.wg.Done()
			err := job(p.ctx)
			if err != nil {
				println("Error executing job:", err.Error())
			}
		}()
	}
}

// AddJob 将新作业添加到池中。如果池已满，它将阻塞，直到有工作人员可用。
func (p *GPool) AddJob(job func(ctx context.Context) error) {
	p.wg.Add(1)
	p.workQueue <- func() {
		err := job(p.ctx)
		if err != nil {
			println("Error executing job:", err.Error())
		}
	}
}

// Wait 直到所有提交的任务都完成。
func (p *GPool) Wait() {
	p.wg.Wait()
}

// Shutdown 取消所有挂起的任务，并等待正在运行的任务完成。
func (p *GPool) Shutdown() {
	p.cancel()
	p.Wait()
}

// RunningCount 返回当前正在运行的任务数。
func (p *GPool) RunningCount() int64 {
	return atomic.LoadInt64(&p.running)
}
