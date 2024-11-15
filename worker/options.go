package worker

import (
	"context"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/sagoo-cloud/nexframe/g"
	"reflect"
	"time"
)

type Options struct {
	group             string                                                    //任务处理器的组名
	redisUri          string                                                    //redis连接地址
	redisPeriodKey    string                                                    //redis周期任务key
	retention         int                                                       //redis周期任务key过期时间
	maxRetry          int                                                       //任务最大重试次数
	handler           func(ctx context.Context, p Payload) error                //任务的处理函数
	handlerNeedWorker func(worker Worker, ctx context.Context, p Payload) error //需要Worker参数的任务处理函数
	handlerAggregator func(ctx context.Context, task *asynq.Task) error         //聚合任务的处理函数
	callback          string                                                    //任务完成后的回调地址
	clearArchived     int                                                       //清除已归档任务的时间间隔
	timeout           int                                                       //任务处理器的超时时间

	redisLinkMode  string //redis连接模式
	concurrencyNum int    //任务处理器的并发数

	//聚合参数
	useAggregator    bool //是否使用聚合器
	groupMaxDelay    int  //最晚多少秒聚合一次
	groupGracePeriod int  //每多少秒聚合一次
	groupMaxSize     int  // 每多少个任务聚合一次
}

// WithRedisLinkMode 设置redis连接模式
func WithRedisLinkMode(s string) func(*Options) {
	return func(options *Options) {
		if s == "" {
			s = "default"
		}
		getOptionsOrSetDefault(options).redisLinkMode = s
	}
}

// WithGroup 设置任务处理器的组名
func WithGroup(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).group = s
	}
}

// WithRedisUri 设置redis连接地址，默认值redis://127.0.0.1:6379/0
func WithRedisUri(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).redisUri = s
	}
}

// WithRedisPeriodKey 设置redis周期任务key
func WithRedisPeriodKey(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).redisPeriodKey = s
	}
}

// WithRetention 成功任务存储时间，默认 60 秒，如果提供此选项，任务将在成功处理后作为已完成任务存储
func WithRetention(second int) func(*Options) {
	return func(options *Options) {
		if second > 0 {
			getOptionsOrSetDefault(options).retention = second
		}
	}
}

// WithMaxRetry 任务出错时的最大重试次数，默认为 3
func WithMaxRetry(count int) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).maxRetry = count
	}
}

// WithHandler 设置任务的回调处理器
func WithHandler(fun func(ctx context.Context, p Payload) error) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).handler = fun
		}
	}
}

// WithHandlerNeedWorker 设置需要Worker参数的任务处理函数
func WithHandlerNeedWorker(fun func(worker Worker, ctx context.Context, p Payload) error) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).handlerNeedWorker = fun
		}
	}
}

// WithCallback 设置任务完成后的回调地址
func WithCallback(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).callback = s
	}
}

// WithClearArchived 清除已存档任务的间隔，默认为 300 秒
func WithClearArchived(second int) func(*Options) {
	return func(options *Options) {
		if second > 0 {
			getOptionsOrSetDefault(options).clearArchived = second
		}
	}
}

// WithTimeout 任务超时时间，默认为 10 秒
func WithTimeout(second int) func(*Options) {
	return func(options *Options) {
		if second > 0 {
			getOptionsOrSetDefault(options).timeout = second
		}
	}
}

// WithConcurrencyNum 设置任务处理器的并发数
func WithConcurrencyNum(s int) func(*Options) {
	return func(options *Options) {
		if s > 0 {
			getOptionsOrSetDefault(options).concurrencyNum = s
		}
	}
}

// WithHandleAggregator 设置聚合任务的回调处理器
func WithHandlerAggregator(fun func(ctx context.Context, task *asynq.Task) error) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).handlerAggregator = fun
		}
	}
}

// WithUseAggregator 是否使用聚合器
func WithUseAggregator(s bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).useAggregator = s
	}
}

// WithGroupMaxDelay 最晚多少秒聚合一次
func WithGroupMaxDelay(second int) func(*Options) {
	return func(options *Options) {
		if second > 0 {
			getOptionsOrSetDefault(options).groupMaxDelay = second
		}
	}
}

// WithGroupGracePeriod 每多少秒聚合一次
func WithGroupGracePeriod(second int) func(*Options) {
	return func(options *Options) {
		if second > 0 {
			getOptionsOrSetDefault(options).groupGracePeriod = second
		}
	}
}

// WithGroupMaxSize 多少个任务聚合一次
func WithGroupMaxSize(second int) func(*Options) {
	return func(options *Options) {
		if second > 0 {
			getOptionsOrSetDefault(options).groupMaxSize = second
		}
	}
}

// WithOptions 设置任务处理器的配置
func getOptionsOrSetDefault(options *Options) *Options {
	addr := g.Cfg.GetString("redis.default.address")
	db := g.Cfg.GetString("redis.default.db")
	user := g.Cfg.GetString("redis.default.user", "default")
	pass := g.Cfg.GetString("redis.default.pass")
	redisUri := "redis://127.0.0.1/0"
	if pass != "" {
		redisUri = fmt.Sprintf("redis://%s:%s@%s/%s", user, pass, addr, db)
	} else {
		redisUri = fmt.Sprintf("redis://%s/%s", addr, db)

	}

	retention := g.Cfg.GetInt("task.retention", 60)
	maxRetry := g.Cfg.GetInt("task.maxRetry", 2)
	clearArchived := g.Cfg.GetInt("task.clearArchived", 300)
	timeout := g.Cfg.GetInt("task.timeout", 30)
	concurrencyNum := g.Cfg.GetInt("task.concurrencyNum", 100)

	groupMaxSize := g.Cfg.GetInt("task.groupMaxSize", 100)
	groupMaxDelay := g.Cfg.GetInt("task.groupMaxDelay", 3000)
	groupGracePeriod := g.Cfg.GetInt("task.groupGracePeriod", 300)
	redisLinkMode := g.Cfg.GetString("redis.default.mode", "single")

	if options == nil {
		return &Options{
			group:          "task",
			redisUri:       redisUri,
			redisPeriodKey: "period",
			retention:      retention,
			maxRetry:       maxRetry,
			clearArchived:  clearArchived,
			timeout:        timeout,
			concurrencyNum: concurrencyNum,
			redisLinkMode:  redisLinkMode,
			//聚合参数
			groupMaxDelay:    groupMaxDelay,
			groupGracePeriod: groupGracePeriod,
			groupMaxSize:     groupMaxSize,
		}
	}
	return options
}

type RunOptions struct {
	uid       string
	group     string
	payload   []byte
	expr      string          // only period task
	in        *time.Duration  // only once task
	at        *time.Time      // only once task
	now       bool            // only once task
	retention int             // only once task
	replace   bool            // only once task
	ctx       context.Context // only once task
	maxRetry  int
	timeout   int
}

// WithRunUuid 任务唯一id
func WithRunUuid(s string) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).uid = s
	}
}

// WithRunGroup 组前缀，默认组
func WithRunGroup(s string) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).group = s
	}
}

// WithRunPayload 任务负载，任务回调会使用
func WithRunPayload(s []byte) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).payload = s
	}
}

// WithRunExpr Cron表达式, 最小单位1分钟, 参见gorhill/cronexpr
func WithRunExpr(s string) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).expr = s
	}
}

// WithRunIn 任务延迟执行，在xxx秒内运行
func WithRunIn(in time.Duration) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).in = &in
	}
}

// WithRunAt 运行任务的时间
func WithRunAt(at time.Time) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).at = &at
	}
}

// WithRunNow 立即运行任务
func WithRunNow(flag bool) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).now = flag
	}
}

// WithRunRetention 任务过期时间，默认60秒
func WithRunRetention(second int) func(*RunOptions) {
	return func(options *RunOptions) {
		if second > 0 {
			getRunOptionsOrSetDefault(options).retention = second
		}
	}
}

// WithRunReplace 当uid重复时，删除旧的并创建新的
func WithRunReplace(flag bool) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).replace = flag
	}
}

// WithRunCtx 任务上下文
func WithRunCtx(ctx context.Context) func(*RunOptions) {
	return func(options *RunOptions) {
		if !interfaceIsNil(ctx) {
			getRunOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

// WithRunMaxRetry 最大重试次数, 任务回调发生error会重试，默认3次
func WithRunMaxRetry(count int) func(*RunOptions) {
	return func(options *RunOptions) {
		getRunOptionsOrSetDefault(options).maxRetry = count
	}
}

// WithRunTimeout 任务超时，默认60秒
func WithRunTimeout(second int) func(*RunOptions) {
	return func(options *RunOptions) {
		if second > 0 {
			getRunOptionsOrSetDefault(options).timeout = second
		}
	}
}

// 获取运行选项或设置默认值
func getRunOptionsOrSetDefault(options *RunOptions) *RunOptions {
	if options == nil {
		return &RunOptions{
			group:   "group",
			timeout: 60,
		}
	}
	return options
}

func interfaceIsNil(i interface{}) bool {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return i == nil
}
