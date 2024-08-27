package timers

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

// TestServer 测试 Server 的主要功能
func TestServer(t *testing.T) {
	// 创建一个测试用的日志记录器
	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)

	// 创建一个新的 Server 实例
	server := NewServer(logger)

	// 测试注册服务
	t.Run("Register Service", func(t *testing.T) {
		err := server.Register("test1", time.Second, testHandler, nil)
		if err != nil {
			t.Errorf("注册服务失败: %v", err)
		}

		// 测试重复注册
		err = server.Register("test1", time.Second, testHandler, nil)
		if err == nil {
			t.Error("预期重复注册会失败，但是成功了")
		}

		// 测试无效参数
		err = server.Register("", time.Second, testHandler, nil)
		if err == nil {
			t.Error("预期使用空名称注册会失败，但是成功了")
		}
	})

	// 测试服务执行
	t.Run("Run Services", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		callCount := 0
		mutex := &sync.Mutex{}

		err := server.Register("test2", 100*time.Millisecond, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			mutex.Lock()
			callCount++
			mutex.Unlock()
			if callCount >= 3 {
				wg.Done()
			}
			return "ok", nil
		}, nil)

		if err != nil {
			t.Fatalf("注册服务失败: %v", err)
		}

		err = server.Run()
		if err != nil {
			t.Fatalf("启动服务失败: %v", err)
		}

		// 等待服务执行几次
		wg.Wait()

		mutex.Lock()
		if callCount < 3 {
			t.Errorf("服务没有按预期执行，执行次数: %d", callCount)
		}
		mutex.Unlock()
	})

	// 测试关闭服务
	t.Run("Close Server", func(t *testing.T) {
		err := server.Close()
		if err != nil {
			t.Errorf("关闭服务器失败: %v", err)
		}

		// 给一些时间让 goroutines 完全退出
		time.Sleep(200 * time.Millisecond)

		// 检查是否所有 goroutines 都已退出
		server.wg.Wait()
	})
}

// testHandler 是一个用于测试的处理函数
func testHandler(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return "test completed", nil
}

// TestConcurrency 测试并发注册和执行
func TestConcurrency(t *testing.T) {
	logger := log.New(os.Stdout, "CONCURRENCY TEST: ", log.Ldate|log.Ltime|log.Lshortfile)
	server := NewServer(logger)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := server.Register(fmt.Sprintf("service-%d", id), time.Millisecond*10, testHandler, nil)
			if err != nil {
				t.Errorf("并发注册服务失败: %v", err)
			}
		}(i)
	}

	wg.Wait()

	err := server.Run()
	if err != nil {
		t.Fatalf("启动服务失败: %v", err)
	}

	// 让服务运行一段时间
	time.Sleep(100 * time.Millisecond)

	err = server.Close()
	if err != nil {
		t.Errorf("关闭服务器失败: %v", err)
	}
}
