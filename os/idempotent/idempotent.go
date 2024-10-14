package idempotent

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

// ErrRedisNotEnabled 定义Redis未启用错误
var ErrRedisNotEnabled = errors.New("redis not enabled")

// Idempotent 幂等性检查结构体
type Idempotent struct {
	ops Options
}

// New 创建新的Idempotent实例
func New(options ...func(*Options)) *Idempotent {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return &Idempotent{ops: *ops}
}

// Token 生成幂等性token
func (i *Idempotent) Token(ctx context.Context) (string, error) {
	if i.ops.redis == nil {
		return "", ErrRedisNotEnabled
	}

	token := uuid.NewString()
	key := i.getKey(token)

	err := i.ops.redis.Set(ctx, key, true, time.Duration(i.ops.expire)*time.Minute).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}

// Check 检查幂等性token
func (i *Idempotent) Check(ctx context.Context, token string) (bool, error) {
	if i.ops.redis == nil {
		return false, ErrRedisNotEnabled
	}

	key := i.getKey(token)
	res, err := i.ops.redis.Eval(ctx, luaScript, []string{key}).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	return res == "1", nil
}

// getKey 生成Redis键
func (i *Idempotent) getKey(token string) string {
	return strings.Join([]string{i.ops.prefix, token}, "_")
}

// luaScript Redis Lua脚本：读取 => 删除 => 获取删除标志
const luaScript = `
local current = redis.call('GET', KEYS[1])
if not current then
    return '-1'
end
local del = redis.call('DEL', KEYS[1])
if del == 1 then
     return '1'
end
return '0'
`
