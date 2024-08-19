package nx

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Nx struct {
	ops    Options
	mu     sync.Mutex
	owners map[string]bool
}

// New 创建一个新的 nx 锁实例
func New(options ...func(*Options)) (*Nx, error) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if err := ops.Validate(); err != nil {
		return nil, err
	}
	return &Nx{
		ops:    *ops,
		owners: make(map[string]bool),
	}, nil
}

// Lock 尝试获取锁，如果获取失败，则在 interval 秒内自动重试 retry 次以获得锁定，如果失败，则返回错误
func (nx *Nx) Lock(ctx context.Context) error {
	var attempts int
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ok, err := nx.tryLock(ctx)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			if attempts >= nx.ops.retry {
				return errors.New("lock timeout")
			}
			time.Sleep(nx.ops.interval)
			attempts++
		}
	}
}

// tryLock 尝试获取锁。如果获取成功，则返回 true，否则返回 false
func (nx *Nx) tryLock(ctx context.Context) (bool, error) {
	script := `
		if redis.call("get", KEYS[1]) == false then
			return redis.call("setex", KEYS[1], ARGV[1], "locked")
		else
			return 0
		end
	`
	result, err := nx.ops.redis.Eval(ctx, script, []string{nx.ops.key}, nx.ops.expire).Result()
	if err != nil {
		return false, err
	}
	return result == "OK", nil
}

// Unlock 释放锁
func (nx *Nx) Unlock(ctx context.Context) error {
	nx.mu.Lock()
	defer nx.mu.Unlock()

	if !nx.owners[nx.ops.key] {
		return nil
	}
	delete(nx.owners, nx.ops.key)

	return nx.ops.redis.Del(ctx, nx.ops.key).Err()
}
