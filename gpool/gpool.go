package gpool

import (
	"context"
	"sync"
	"time"
)

// GPool 定义了一个协程池的结构体
type GPool struct {
	sem     chan struct{}      // 带缓冲的信号量，用于控制并发数量
	wg      sync.WaitGroup     // 用于等待所有协程执行完毕
	ctx     context.Context    // 用于取消协程的上下文
	cancel  context.CancelFunc // 触发取消上下文的函数
	errChan chan error         // 带缓冲的错误通道，用于收集协程执行过程中的错误
}

// goroutinePool 是一个 sync.Pool，用于管理协程的复用
var goroutinePool = sync.Pool{
	New: func() interface{} {
		// 每次从 goroutinePool 获取对象时，都会创建一个新的匿名函数
		return func(p *GPool, ctx context.Context, job func(ctx context.Context) error) {
			defer p.Release() // 确保释放信号量
			// 执行任务，如果出现错误，将其发送到 errChan
			if err := job(ctx); err != nil {
				select {
				case p.errChan <- err:
				default:
				}
			}
		}
	},
}

// NewGPool 创建一个新的 GPool 实例，参数 capacity 指定了协程池的容量
func NewGPool(capacity int) *GPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &GPool{
		sem:     make(chan struct{}, capacity),
		ctx:     ctx,
		cancel:  cancel,
		errChan: make(chan error, capacity),
	}
}

// Acquire 获取一个可用的协程，如果获取成功返回 true，否则返回 false
func (p *GPool) Acquire() bool {
	select {
	case p.sem <- struct{}{}: // 尝试获取信号量
		p.wg.Add(1)
		return true
	case <-p.ctx.Done(): // 上下文已取消
		return false
	default: // 信号量已满
		return false
	}
}

// Release 释放一个协程，供其他任务使用
func (p *GPool) Release() {
	select {
	case <-p.sem:
	default:
	}
	p.wg.Done()
}

// Go 提交一个新的任务到协程池中执行，如果当前没有可用的协程，则任务会被加入到待执行队列中
func (p *GPool) Go(job func(ctx context.Context) error) {
	if !p.Acquire() {
		// 如果 Acquire 失败，将任务加入队列中等待执行
		p.AddJob(job)
		return
	}

	// 从 goroutinePool 中获取一个可复用的协程函数
	fn := goroutinePool.Get().(func(p *GPool, ctx context.Context, job func(ctx context.Context) error))
	go func() {
		defer goroutinePool.Put(fn)
		fn(p, p.ctx, job)
	}()
}

// AddJob 将一个新的任务加入到待执行队列中
func (p *GPool) AddJob(job func(ctx context.Context) error) {
	go func() {
		if p.Acquire() {
			defer p.Release()
			// 从 goroutinePool 中获取一个可复用的协程函数
			fn := goroutinePool.Get().(func(p *GPool, ctx context.Context, job func(ctx context.Context) error))
			defer goroutinePool.Put(fn)
			fn(p, p.ctx, job)
		} else {
			// 如果 Acquire 失败，将任务直接执行
			defer p.wg.Done()
			if err := job(p.ctx); err != nil {
				select {
				case p.errChan <- err:
				default:
				}
			}
		}
	}()
}

// Wait 等待所有协程执行完毕，可以设置超时时间
func (p *GPool) Wait(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done: // 所有任务执行完毕
		return true
	case <-time.After(timeout): // 超时
		return false
	}
}

// Shutdown 关闭协程池，等待所有协程执行完毕，并关闭错误通道
func (p *GPool) Shutdown() {
	p.cancel()       // 取消上下文
	p.wg.Wait()      // 等待所有任务完成
	close(p.errChan) // 关闭错误通道
}

// ErrChan 获取错误通道，以便获取协程执行过程中出现的错误
func (p *GPool) ErrChan() <-chan error {
	return p.errChan
}
