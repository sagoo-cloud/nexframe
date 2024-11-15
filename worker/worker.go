package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-module/carbon/v2"
	"github.com/gorhill/cronexpr"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/sagoo-cloud/nexframe/g"
	"github.com/sagoo-cloud/nexframe/os/nx"
	"github.com/sagoo-cloud/nexframe/utils/guid"
	"io"
	"net/http"
	"strings"
	"time"
)

type Worker struct {
	ops       Options
	redis     redis.UniversalClient
	redisOpt  asynq.RedisConnOpt
	lock      *nx.Nx
	client    *asynq.Client
	inspector *asynq.Inspector
	Error     error
}

type periodTask struct {
	Expr      string `json:"expr"`
	Group     string `json:"group"`
	Uid       string `json:"uid"`
	Payload   []byte `json:"payload"`
	Next      int64  `json:"next"`
	Processed int64  `json:"processed"`
	MaxRetry  int    `json:"maxRetry"`
	Timeout   int    `json:"timeout"`
}

// 将周期任务转化为JSON字符串
func (p *periodTask) String() (str string) {
	bs, _ := json.Marshal(p)
	str = string(bs)
	return
}

// FromString 从JSON字符串解析周期任务
func (p *periodTask) FromString(str string) {
	err := json.Unmarshal([]byte(str), p)
	if err != nil {
		return
	}
	return
}

type periodTaskHandler struct {
	tk Worker
}

type Payload struct {
	Group   string `json:"group"`
	Uid     string `json:"uid"`
	Payload []byte `json:"payload"`
}

func (p Payload) String() (str string) {
	bs, _ := json.Marshal(p)
	str = string(bs)
	return
}

// ProcessTask 处理任务
func (p periodTaskHandler) ProcessTask(ctx context.Context, t *asynq.Task) (err error) {
	uid := guid.S()
	group := strings.TrimSuffix(strings.TrimSuffix(t.Type(), ".once"), ".cron")
	payload := Payload{
		Group:   group,
		Uid:     t.ResultWriter().TaskID(),
		Payload: t.Payload(),
	}
	defer func() {
		if err != nil {
			g.Log.Debugf(ctx, "run task failed. uuid: %s task: %s Error:%s", uid, payload, err)
		}
	}()
	if p.tk.ops.handler != nil {
		err = p.tk.ops.handler(ctx, payload)
	} else if p.tk.ops.handlerAggregator != nil {
		err = p.tk.ops.handlerAggregator(ctx, t)
	} else if p.tk.ops.handlerNeedWorker != nil {
		err = p.tk.ops.handlerNeedWorker(p.tk, ctx, payload)
	} else if p.tk.ops.callback != "" {
		err = p.httpCallback(ctx, payload)
	} else {
		g.Log.Debugf(ctx, "no task handler. uuid: %s task: %s Error:%s", uid, payload, err)
	}
	// 保存处理次数
	p.tk.processed(ctx, payload.Uid)
	return
}

// HTTP回调函数
func (p periodTaskHandler) httpCallback(ctx context.Context, payload Payload) (err error) {
	client := &http.Client{}
	body := payload.String()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tk.ops.callback, bytes.NewReader([]byte(body)))
	if err != nil {
		g.Log.Errorf(ctx, "创建HTTP请求失败: %v", err)
		return err
	}
	r.Header.Add("Content-Type", "application/json")

	// 确保响应体正确关闭，防止资源泄漏
	res, err := client.Do(r)
	if err != nil {
		g.Log.Errorf(ctx, "执行HTTP请求失败: %v", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			g.Log.Errorf(ctx, "body close err: %v", err)
		}
	}(res.Body)

	responseBytes, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		err = errors.New("HTTP回调状态码异常：" + res.Status)
		g.Log.Debugf(ctx, "HTTP回调状态码非200: %v 响应内容: %s", err, string(responseBytes))
		return err
	}
	return nil
}

// New 创建一个新的任务处理器
func New(options ...func(*Options)) *Worker {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	// 修改：确保 ops 不为 nil
	if ops == nil {
		return &Worker{Error: fmt.Errorf("options cannot be nil")}
	}
	if ops.redisUri == "" {
		return &Worker{Error: fmt.Errorf("redis URI is empty: %w", ErrRedisNil)}
	}

	rs, err := asynq.ParseRedisURI(ops.redisUri)
	if err != nil {
		return &Worker{Error: fmt.Errorf("invalid redis URI: %w", err)}
	}
	addsList := strings.Split(ops.redisUri, ",")

	switch ops.redisLinkMode {
	case "cluster":
		rs = asynq.RedisClusterClientOpt{
			Addrs: addsList,
		}
	case "sentinel":
		rs = &asynq.RedisFailoverClientOpt{
			MasterName:    "sagoo-master",
			SentinelAddrs: addsList,
		}
	}

	// 创建 Redis 客户端
	redisClient, ok := rs.MakeRedisClient().(redis.UniversalClient)
	// 修改：确保 redisClient 类型转换成功
	if !ok {
		return &Worker{Error: fmt.Errorf("failed to create redis client: %w", ErrRedisInvalid)}
	}

	client := asynq.NewClient(rs)
	inspector := asynq.NewInspector(rs)
	// 创建一个简单的锁实例，减少锁的范围，仅在需要时使用
	nxLock, err := nx.New(nx.WithRedis(redisClient), nx.WithExpire(10), nx.WithKey("lock:"+ops.redisPeriodKey))
	if err != nil {
		return &Worker{Error: fmt.Errorf("failed to create lock: %w", err)}
	}

	worker := &Worker{
		ops:       *ops,
		redis:     redisClient,
		redisOpt:  rs,
		lock:      nxLock,
		client:    client,
		inspector: inspector,
	}

	if ops.handlerAggregator != nil {
		srv := asynq.NewServer(
			rs,
			asynq.Config{
				Queues:           map[string]int{ops.group: 10},
				GroupAggregator:  asynq.GroupAggregatorFunc(aggregate),
				GroupGracePeriod: time.Duration(ops.groupGracePeriod) * time.Second, // 多久聚合一次
				GroupMaxDelay:    time.Duration(ops.groupMaxDelay) * time.Second,    // 最晚多久聚合一次
				GroupMaxSize:     ops.groupMaxSize,                                  // 每多少个任务聚合一次
				LogLevel:         4,
			},
		)

		go func() {
			mux := asynq.NewServeMux()
			var h periodTaskHandler
			h.tk = *worker
			mux.Handle(ops.group, h)

			if err := srv.Run(mux); err != nil {
				g.Log.Debugf(context.Background(), "running task handler failed: %v", err)
			}
		}()

	} else {

		// 启动任务处理服务器
		srv := asynq.NewServer(rs, asynq.Config{
			Concurrency: 100,                           //服务器同时处理的最大任务数
			Queues:      map[string]int{ops.group: 10}, //定义了每个队列的最大并发任务数
			LogLevel:    4,
		})

		go func() {
			var h periodTaskHandler
			h.tk = *worker
			if err := srv.Run(h); err != nil {
				g.Log.Debugf(context.Background(), "running task handler failed: %v", err)
			}
		}()
	}

	// 定期扫描和清理归档任务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.schedulePeriodicTasks(ctx)

	return worker
}

// 传入上下文，以便在需要时取消定时任务
func (wk *Worker) schedulePeriodicTasks(ctx context.Context) {
	// 确保在函数退出前取消定时器
	scanTicker := time.NewTicker(10 * time.Second)
	clearTicker := time.NewTicker(time.Duration(wk.ops.clearArchived) * time.Second)
	defer scanTicker.Stop()
	defer clearTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-scanTicker.C:
			wk.scan()
		case <-clearTicker.C:
			wk.clearArchived()
		}
	}
}

func (wk *Worker) Once(options ...func(*RunOptions)) (err error) {
	if wk == nil {
		return
	}
	ops := getRunOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.uid == "" {
		err = errors.Unwrap(ErrUuidNil)
		return
	}
	taskOpts := []asynq.Option{
		asynq.Queue(wk.ops.group),
		asynq.MaxRetry(wk.ops.maxRetry),
		asynq.Timeout(time.Duration(ops.timeout) * time.Second),
	}

	if wk.ops.useAggregator {

		data := Payload{
			Group:   ops.group,
			Uid:     ops.uid,
			Payload: ops.payload,
		}
		//
		sendData, err2 := json.Marshal(data)
		if err2 != nil {
			return
		}
		taskOpts = append(taskOpts, asynq.Group(ops.group))
		task := asynq.NewTask(wk.ops.group, sendData)
		_, err := wk.client.Enqueue(task, taskOpts...)
		if err != nil {
			g.Log.Debugf(context.Background(), "无法加入任务队列：%v", err)
		}

	} else {
		if ops.maxRetry > 0 {
			taskOpts = append(taskOpts, asynq.MaxRetry(ops.maxRetry))
		}
		if ops.retention > 0 {
			taskOpts = append(taskOpts, asynq.Retention(time.Duration(ops.retention)*time.Second))
		} else {
			taskOpts = append(taskOpts, asynq.Retention(time.Duration(wk.ops.retention)*time.Second))
		}
		if ops.in != nil {
			taskOpts = append(taskOpts, asynq.ProcessIn(*ops.in))
		} else if ops.at != nil {
			taskOpts = append(taskOpts, asynq.ProcessAt(*ops.at))
		} else if ops.now {
			taskOpts = append(taskOpts, asynq.ProcessIn(time.Millisecond))
		}
		t := asynq.NewTask(strings.Join([]string{ops.group, "once"}, "."), ops.payload, asynq.TaskID(ops.uid))
		_, err = wk.client.Enqueue(t, taskOpts...)
		if ops.replace && errors.Is(err, asynq.ErrTaskIDConflict) {
			ctx := wk.getDefaultTimeoutCtx()
			if ops.ctx != nil {
				ctx = ops.ctx
			}
			err = wk.Remove(ctx, ops.uid)
			if err != nil {
				return
			}
			_, err = wk.client.Enqueue(t, taskOpts...)
		}
	}

	return
}

// Cron 设置周期性任务
func (wk *Worker) Cron(options ...func(*RunOptions)) (err error) {
	ops := getRunOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.uid == "" {
		err = errors.Unwrap(ErrUuidNil)
		return
	}
	var next int64
	next, err = getNext(ops.expr, 0)
	if err != nil {
		err = errors.Unwrap(ErrExprInvalid)
		return
	}
	t := periodTask{
		Expr:     ops.expr,
		Group:    strings.Join([]string{ops.group, "cron"}, "."),
		Uid:      ops.uid,
		Payload:  ops.payload,
		Next:     next,
		MaxRetry: ops.maxRetry,
		Timeout:  ops.timeout,
	}
	ctx := wk.getDefaultTimeoutCtx()
	res, err := wk.redis.HGet(ctx, wk.ops.redisPeriodKey, ops.uid).Result()
	if err == nil {
		var oldT periodTask
		err = json.Unmarshal([]byte(res), &oldT)
		if err != nil {
			return err
		}
		if oldT.Expr != t.Expr {
			err = wk.Remove(ctx, t.Uid)
			if err != nil {
				return err
			}
		}
	}
	_, err = wk.redis.HSet(ctx, wk.ops.redisPeriodKey, ops.uid, t.String()).Result()
	if err != nil {
		err = errors.Unwrap(ErrSaveCron)
		return
	}
	return
}

// Remove 移除任务
func (wk *Worker) Remove(ctx context.Context, uid string) (err error) {
	err = wk.lock.Lock(ctx)
	if err != nil {
		return
	}
	defer func(lock *nx.Nx, ctx context.Context) {
		err := lock.Unlock(ctx)
		if err != nil {
			return
		}
	}(wk.lock, ctx)
	wk.redis.HDel(ctx, wk.ops.redisPeriodKey, uid)

	err = wk.inspector.DeleteTask(wk.ops.group, uid)
	return
}

func (wk *Worker) processed(ctx context.Context, uid string) {
	err := wk.lock.Lock(ctx)
	if err != nil {
		return
	}
	defer func(lock *nx.Nx, ctx context.Context) {
		err := lock.Unlock(ctx)
		if err != nil {
			return
		}
	}(wk.lock, ctx)
	t, e := wk.redis.HGet(ctx, wk.ops.redisPeriodKey, uid).Result()
	if e == nil || !errors.Is(e, redis.Nil) {
		var item periodTask
		item.FromString(t)
		item.Processed++
		wk.redis.HSet(ctx, wk.ops.redisPeriodKey, uid, item.String())
	}
	return
}

// scan 扫描并处理任务队列
func (wk *Worker) scan() {
	ctx := wk.getDefaultTimeoutCtx()
	if err := wk.lock.Lock(ctx); err != nil {
		return
	}
	defer func() {
		err := wk.lock.Unlock(ctx)
		if err != nil {
			g.Log.Errorf(ctx, "unlock failed: %v", err)
		}
	}()
	m, _ := wk.redis.HGetAll(ctx, wk.ops.redisPeriodKey).Result()
	p := wk.redis.Pipeline()
	ops := wk.ops
	if ops.group == "task" {
		for _, v := range m {
			var item periodTask
			item.FromString(v)
			next, _ := getNext(item.Expr, item.Next)
			t := asynq.NewTask(item.Group, item.Payload, asynq.TaskID(item.Uid))
			taskOpts := []asynq.Option{
				asynq.Queue(ops.group),
				asynq.MaxRetry(ops.maxRetry),
				asynq.Timeout(time.Duration(item.Timeout) * time.Second),
			}
			if item.MaxRetry > 0 {
				taskOpts = append(taskOpts, asynq.MaxRetry(item.MaxRetry))
			}
			diff := next - item.Next
			if diff > 10 {
				retention := diff / 3
				if diff > 600 {
					retention = 600
				}
				taskOpts = append(taskOpts, asynq.Retention(time.Duration(retention)*time.Second))
			}
			taskOpts = append(taskOpts, asynq.ProcessAt(time.Unix(item.Next, 0)))
			if _, err := wk.client.Enqueue(t, taskOpts...); err == nil {
				item.Next = next
				p.HSet(ctx, wk.ops.redisPeriodKey, item.Uid, item.String())
			}
		}
		if _, err := p.Exec(ctx); err != nil {
			return
		}
	}

	return
}

// clearArchived 清除已归档的任务
func (wk *Worker) clearArchived() {
	list, err := wk.inspector.ListArchivedTasks(wk.ops.group, asynq.Page(1), asynq.PageSize(100))
	if err != nil {
		return
	}
	ctx := wk.getDefaultTimeoutCtx()
	for _, item := range list {
		last := carbon.CreateFromStdTime(item.LastFailedAt)
		if !last.IsZero() && item.Retried < item.MaxRetry {
			continue
		}
		uid := item.ID
		var flag bool
		if strings.HasSuffix(item.Type, ".cron") {
			t, e := wk.redis.HGet(ctx, wk.ops.redisPeriodKey, uid).Result()
			if e == nil || !errors.Is(e, redis.Nil) {
				var task periodTask
				task.FromString(t)
				next, _ := getNext(task.Expr, task.Next)
				diff := next - task.Next
				if diff <= 60 {
					if carbon.Now().Gt(last.AddMinutes(5)) {
						flag = true
					}
				} else if diff <= 600 {
					if carbon.Now().Gt(last.AddMinutes(30)) {
						flag = true
					}
				} else if diff <= 3600 {
					if carbon.Now().Gt(last.AddHours(2)) {
						flag = true
					}
				} else {
					if carbon.Now().Gt(last.AddHours(5)) {
						flag = true
					}
				}
			}
		} else {
			if carbon.Now().Gt(last.AddMinutes(5)) {
				flag = true
			}
		}
		if flag {
			err := wk.inspector.DeleteTask(wk.ops.group, uid)
			if err != nil {
				return
			}
		}
	}
}

// getDefaultTimeoutCtx 获取带有默认超时的上下文
func (wk *Worker) getDefaultTimeoutCtx() context.Context {
	c, cancel := context.WithTimeout(context.Background(), time.Duration(wk.ops.timeout)*time.Second)
	defer cancel()
	return c
}

// getNext 计算下一次执行时间
func getNext(expr string, timestamp int64) (next int64, err error) {
	var e *cronexpr.Expression
	e, err = cronexpr.Parse(expr)
	if err != nil {
		return
	}
	t := carbon.Now().StdTime()
	if timestamp > 0 {
		t = carbon.CreateFromTimestamp(timestamp).StdTime()
	}
	next = e.Next(t).Unix()
	return
}

// 从组中聚合任务
func aggregate(group string, tasks []*asynq.Task) *asynq.Task {
	//g.Log.Debugf(context.Background(), "从组 %v 聚合了 %d 个任务", group, len(tasks))
	var payloads []Payload
	for _, payload := range tasks {
		var p Payload
		if err := json.Unmarshal(payload.Payload(), &p); err != nil {
			continue
		}
		payloads = append(payloads, p)
	}
	sendData, err := json.Marshal(payloads)
	if err != nil {
		return nil
	}
	return asynq.NewTask(group, sendData)
}
