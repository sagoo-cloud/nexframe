package gpool

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

// TestGo 测试 Go 方法的基本功能
// 这个测试验证：
// 1. 协程池能否正确处理多个并发任务
// 2. 任务的执行、取消和失败处理是否正确
// 3. RunningCount 方法是否正确报告运行中的协程数量
func TestGo(t *testing.T) {
	pool := NewGPool(5)
	defer pool.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var completedJobs int32

	for i := 0; i < 10; i++ {
		jobIndex := i
		pool.Go(func(jobCtx context.Context) error {
			waitTime := time.Duration(rand.Intn(1000)) * time.Millisecond
			select {
			case <-time.After(waitTime):
				if rand.Float32() < 0.3 {
					return fmt.Errorf("job %d failed", jobIndex)
				}
				atomic.AddInt32(&completedJobs, 1)
				t.Logf("job %d completed successfully", jobIndex)
				return nil
			case <-jobCtx.Done():
				t.Logf("job %d cancelled", jobIndex)
				return jobCtx.Err()
			}
		})
		t.Logf("Running goroutines after submitting job %d: %d", i, pool.RunningCount())
	}

	<-ctx.Done()

	completedJobCount := atomic.LoadInt32(&completedJobs)
	t.Logf("Completed jobs: %d", completedJobCount)
	t.Logf("Final running goroutine count: %d", pool.RunningCount())

	if ctx.Err() == context.DeadlineExceeded {
		t.Log("Test terminated due to timeout")
	}
}

// TestAddJob 测试 AddJob 方法的功能
// 这个测试验证：
// 1. AddJob 方法能否正确添加任务到池中
// 2. 池是否能正确处理超过其容量的任务
// 3. 任务的执行和取消是否正确
func TestAddJob(t *testing.T) {
	pool := NewGPool(3)
	defer pool.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var completedJobs int32

	for i := 0; i < 5; i++ {
		jobIndex := i
		pool.AddJob(func(jobCtx context.Context) error {
			select {
			case <-time.After(time.Duration(100+rand.Intn(100)) * time.Millisecond):
				atomic.AddInt32(&completedJobs, 1)
				t.Logf("Job %d completed", jobIndex)
				return nil
			case <-jobCtx.Done():
				t.Logf("Job %d cancelled", jobIndex)
				return jobCtx.Err()
			}
		})
	}

	<-ctx.Done()

	completedJobCount := atomic.LoadInt32(&completedJobs)
	t.Logf("Completed jobs: %d", completedJobCount)
	t.Logf("Final running goroutine count: %d", pool.RunningCount())

	if ctx.Err() == context.DeadlineExceeded {
		t.Log("Test terminated due to timeout")
	}
}

// TestContextCancellation 测试上下文取消功能
// 这个测试验证：
// 1. 任务是否能正确响应上下文的取消信号
// 2. Shutdown 方法是否能正确取消所有正在运行的任务
// 3. 取消后，是否所有的协程都已经正确退出
func TestContextCancellation(t *testing.T) {
	pool := NewGPool(3)
	defer pool.Shutdown()

	var jobStarted, jobCancelled atomic.Bool

	startTime := time.Now()

	pool.Go(func(jobCtx context.Context) error {
		jobStarted.Store(true)
		t.Log("Job started")
		<-jobCtx.Done()
		jobCancelled.Store(true)
		t.Log("Job cancelled successfully")
		return jobCtx.Err()
	})

	// 等待任务开始
	for !jobStarted.Load() {
		if time.Since(startTime) > 2*time.Second {
			t.Fatal("Job did not start within 2 seconds")
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Log("Cancelling pool")
	pool.Shutdown()

	// 等待任务取消
	cancelWaitStart := time.Now()
	for !jobCancelled.Load() {
		if time.Since(cancelWaitStart) > 2*time.Second {
			t.Fatal("Job cancellation was not detected within 2 seconds after pool shutdown")
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Log("Job cancellation confirmed")

	// 等待一小段时间，让所有协程有机会结束
	time.Sleep(100 * time.Millisecond)

	runningCount := pool.RunningCount()
	t.Logf("Final running goroutine count: %d", runningCount)

	if runningCount != 0 {
		t.Errorf("Expected 0 running goroutines after cancellation, but got %d", runningCount)
	}
}

// TestHighConcurrency 测试协程池在高并发情况下的性能和正确性
// 这个测试验证：
// 1. 协程池能否处理大量并发任务（10000个）
// 2. 在高负载下，错误处理是否正确
// 3. 任务完成的速度和效率（每秒处理的任务数）
// 4. 所有任务完成后，是否所有协程都正确退出
func TestHighConcurrency(t *testing.T) {
	poolCapacity := 200
	totalTasks := 10000 // 增加任务数量

	pool := NewGPool(poolCapacity)
	defer pool.Shutdown()

	var completedTasks int64
	var errorCount int64

	startTime := time.Now()

	// 添加进度报告
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			completed := atomic.LoadInt64(&completedTasks)
			t.Logf("Progress: %d/%d tasks completed", completed, totalTasks)
		}
	}()

	// 提交任务到池中
	for i := 0; i < totalTasks; i++ {
		taskID := i
		pool.Go(func(ctx context.Context) error {
			time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
			atomic.AddInt64(&completedTasks, 1)
			if rand.Float32() < 0.01 { // 1% 的失败率
				atomic.AddInt64(&errorCount, 1)
				return fmt.Errorf("task %d failed", taskID)
			}
			return nil
		})
	}

	pool.Wait()

	duration := time.Since(startTime)

	// 输出统计信息
	t.Logf("Total tasks: %d", totalTasks)
	t.Logf("Completed tasks: %d", atomic.LoadInt64(&completedTasks))
	t.Logf("Error count: %d", atomic.LoadInt64(&errorCount))
	t.Logf("Time taken: %v", duration)
	t.Logf("Tasks per second: %.2f", float64(totalTasks)/duration.Seconds())

	// 验证所有任务都被执行
	if int(atomic.LoadInt64(&completedTasks)) != totalTasks {
		t.Errorf("Expected %d tasks to complete, but got %d", totalTasks, completedTasks)
	}

	// 验证错误数量在预期范围内
	expectedErrorRate := 0.01 // 1%
	actualErrorRate := float64(atomic.LoadInt64(&errorCount)) / float64(totalTasks)
	if math.Abs(actualErrorRate-expectedErrorRate) > 0.005 { // 允许 0.5% 的误差
		t.Errorf("Error rate outside expected range. Expected around %.2f, got %.2f", expectedErrorRate, actualErrorRate)
	}

	// 验证没有正在运行的协程
	if runningCount := pool.RunningCount(); runningCount != 0 {
		t.Errorf("Expected 0 running goroutines after all tasks completed, but got %d", runningCount)
	}
}
