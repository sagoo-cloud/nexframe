package idempotent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// 创建 Redis 客户端
func newRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
}

// TestNew 测试 New 函数
func TestNew(t *testing.T) {
	client := newRedisClient()
	defer client.Close()

	tests := []struct {
		name           string
		options        []func(*Options)
		expectedPrefix string
		expectedExpire int
		expectedRedis  bool
	}{
		{
			name:           "默认选项",
			options:        []func(*Options){},
			expectedPrefix: "idempotent",
			expectedExpire: 60,
			expectedRedis:  false,
		},
		{
			name: "自定义选项",
			options: []func(*Options){
				WithRedis(client),
				WithPrefix("custom"),
				WithExpire(30),
			},
			expectedPrefix: "custom",
			expectedExpire: 30,
			expectedRedis:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := New(tt.options...)
			if i.ops.prefix != tt.expectedPrefix {
				t.Errorf("prefix = %v, want %v", i.ops.prefix, tt.expectedPrefix)
			}
			if i.ops.expire != tt.expectedExpire {
				t.Errorf("expire = %v, want %v", i.ops.expire, tt.expectedExpire)
			}
			if (i.ops.redis != nil) != tt.expectedRedis {
				t.Errorf("redis presence = %v, want %v", i.ops.redis != nil, tt.expectedRedis)
			}
		})
	}
}

// TestToken 测试 Token 方法
func TestToken(t *testing.T) {
	client := newRedisClient()
	defer client.Close()

	i := New(WithRedis(client))

	ctx := context.Background()
	token, err := i.Token(ctx)
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if token == "" {
		t.Error("Token() returned empty string")
	}

	// 验证 token 是否已存储在 Redis 中
	key := i.getKey(token)
	exists, err := client.Exists(ctx, key).Result()
	if err != nil {
		t.Fatalf("Failed to check key existence: %v", err)
	}
	if exists != 1 {
		t.Error("Token not stored in Redis")
	}

	// 测试 Redis 未启用的情况
	iNoRedis := New()
	_, err = iNoRedis.Token(ctx)
	if !errors.Is(err, ErrRedisNotEnabled) {
		t.Errorf("Expected ErrRedisNotEnabled, got %v", err)
	}
}

// TestCheck 测试 Check 方法
func TestCheck(t *testing.T) {
	client := newRedisClient()
	defer client.Close()

	i := New(WithRedis(client), WithExpire(1))

	ctx := context.Background()
	token, _ := i.Token(ctx)

	// 首次检查应该通过
	pass, err := i.Check(ctx, token)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !pass {
		t.Error("First check should pass")
	}

	// 第二次检查应该失败
	pass, err = i.Check(ctx, token)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if pass {
		t.Error("Second check should fail")
	}

	// 测试无效的 token
	pass, err = i.Check(ctx, "invalid_token")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if pass {
		t.Error("Check with invalid token should fail")
	}

	// 测试过期情况
	token, _ = i.Token(ctx)
	time.Sleep(time.Minute + time.Second) // 等待超过过期时间
	pass, err = i.Check(ctx, token)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if pass {
		t.Error("Check with expired token should fail")
	}

	// 测试 Redis 未启用的情况
	iNoRedis := New()
	_, err = iNoRedis.Check(ctx, token)
	if !errors.Is(err, ErrRedisNotEnabled) {
		t.Errorf("Expected ErrRedisNotEnabled, got %v", err)
	}
}
