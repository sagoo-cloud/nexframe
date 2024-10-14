package bloom

import (
	"context"
	"fmt"
	"github.com/sagoo-cloud/nexframe/database/redisdb"
	"strconv"
	"time"
)

// Redis Lua 脚本
const (
	luaSetScript = `
for _, offset in ipairs(ARGV) do
	redis.call('SETBIT', KEYS[1], offset, 1)
end
`
	luaGetScript = `
for _, offset in ipairs(ARGV) do
	if tostring(redis.call('GETBIT', KEYS[1], offset)) == '0' then
			return '0'
		end
	end
return '1'
`
)

// Bloom 结构体定义布隆过滤器
type Bloom struct {
	opts Options
}

// New 创建一个新的布隆过滤器实例
func New(options ...func(*Options)) (*Bloom, error) {
	opts := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(opts)
	}

	// 如果未设置 Redis 客户端，则使用默认的客户端
	if opts.redis == nil {
		opts.redis = redisdb.DB().GetClient()
	}

	return &Bloom{opts: *opts}, nil
}

// Add 将一个或多个字符串添加到布隆过滤器中
func (b *Bloom) Add(ctx context.Context, str ...string) error {
	if len(str) == 0 {
		return nil
	}

	pipe := b.opts.redis.Pipeline()
	for _, item := range str {
		offsets := b.calculateOffsets(item)
		err := pipe.Eval(ctx, luaSetScript, []string{b.opts.key}, offsets).Err()
		if err != nil {
			pipe.Discard()
			return fmt.Errorf("添加项目时出错: %w", err)
		}
	}

	pipe.Expire(ctx, b.opts.key, time.Duration(b.opts.expire)*time.Minute)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("执行管道操作时出错: %w", err)
	}

	return nil
}

// Exist 检查字符串是否可能存在于布隆过滤器中
func (b *Bloom) Exist(ctx context.Context, str string) (bool, error) {
	offsets := b.calculateOffsets(str)
	res, err := b.opts.redis.Eval(ctx, luaGetScript, []string{b.opts.key}, offsets).Result()
	if err != nil {
		return false, fmt.Errorf("检查存在性时出错: %w", err)
	}
	return res == "1", nil
}

// Flush 清空布隆过滤器
func (b *Bloom) Flush(ctx context.Context) error {
	err := b.opts.redis.Del(ctx, b.opts.key).Err()
	if err != nil {
		return fmt.Errorf("清空过滤器时出错: %w", err)
	}
	return nil
}

// calculateOffsets 计算字符串的所有哈希偏移量
func (b *Bloom) calculateOffsets(str string) []string {
	offsets := make([]string, len(b.opts.hash))
	for i, f := range b.opts.hash {
		offset := f(str)
		offsets[i] = strconv.FormatUint(offset, 10)
	}
	return offsets
}

// getDefaultTimeoutCtx 创建一个带有默认超时的上下文
func (b *Bloom) getDefaultTimeoutCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(b.opts.timeout)*time.Second)
}
