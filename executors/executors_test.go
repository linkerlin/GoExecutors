package executors

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/linkerlin/GoExecutors/config"
)

// TestThreadPoolExecutor_Basic 基本功能测试
func TestThreadPoolExecutor_Basic(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 2
	cfg.MaxPoolSize = 4
	cfg.QueueSize = 10

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	// 测试简单任务提交
	task := Callable(func(ctx context.Context) (interface{}, error) {
		return "hello", nil
	})

	future, err := executor.Submit(task)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	result, err := future.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result.(string) != "hello" {
		t.Errorf("Expected 'hello', got %v", result)
	}
}

// TestThreadPoolExecutor_Timeout 超时测试
func TestThreadPoolExecutor_Timeout(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 1
	cfg.MaxPoolSize = 1
	cfg.QueueSize = 1

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	// 提交长时间运行的任务
	task := Callable(func(ctx context.Context) (interface{}, error) {
		time.Sleep(200 * time.Millisecond)
		return "done", nil
	})

	future, err := executor.Submit(task)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// 短超时时间
	_, err = future.GetWithTimeout(50 * time.Millisecond)
	if err != ErrTaskTimeout {
		t.Errorf("Expected timeout error, got %v", err)
	}
}

// TestThreadPoolExecutor_Panic 恐慌恢复测试
func TestThreadPoolExecutor_Panic(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 1
	cfg.MaxPoolSize = 1
	cfg.QueueSize = 1

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	// 提交会恐慌的任务
	task := Callable(func(ctx context.Context) (interface{}, error) {
		panic("test panic")
	})

	future, err := executor.Submit(task)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	_, err = future.Get()
	if err == nil {
		t.Error("Expected panic error")
	}
}

// TestThreadPoolExecutor_Cancel 取消测试
func TestThreadPoolExecutor_Cancel(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 1
	cfg.MaxPoolSize = 1
	cfg.QueueSize = 1

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	// 提交长时间运行的任务
	task := Callable(func(ctx context.Context) (interface{}, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(200 * time.Millisecond):
			return "done", nil
		}
	})

	future, err := executor.Submit(task)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// 取消任务
	future.Cancel()

	_, err = future.Get()
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

// TestThreadPoolExecutor_Concurrent 并发测试
func TestThreadPoolExecutor_Concurrent(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 4
	cfg.MaxPoolSize = 8
	cfg.QueueSize = 100

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	const numTasks = 50
	var wg sync.WaitGroup
	var completed int32

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()

			task := Callable(func(ctx context.Context) (interface{}, error) {
				time.Sleep(10 * time.Millisecond)
				return taskID, nil
			})

			future, err := executor.Submit(task)
			if err != nil {
				t.Errorf("Submit failed: %v", err)
				return
			}

			result, err := future.Get()
			if err != nil {
				t.Errorf("Get failed: %v", err)
				return
			}

			if result.(int) != taskID {
				t.Errorf("Expected %d, got %v", taskID, result)
				return
			}

			atomic.AddInt32(&completed, 1)
		}(i)
	}

	wg.Wait()

	if atomic.LoadInt32(&completed) != numTasks {
		t.Errorf("Expected %d completed tasks, got %d", numTasks, atomic.LoadInt32(&completed))
	}
}

// TestThreadPoolExecutor_QueueFull 队列满测试
func TestThreadPoolExecutor_QueueFull(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 1
	cfg.MaxPoolSize = 1
	cfg.QueueSize = 1
	cfg.RejectPolicy = "abort"
	cfg.EnableLogging = false

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	// 提交阻塞任务占用工作线程
	blockTask := Callable(func(ctx context.Context) (interface{}, error) {
		time.Sleep(200 * time.Millisecond)
		return "blocked", nil
	})

	future1, err := executor.Submit(blockTask)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// 等待一小段时间确保任务开始执行
	time.Sleep(10 * time.Millisecond)

	// 提交任务填满队列
	normalTask := Callable(func(ctx context.Context) (interface{}, error) {
		return "normal", nil
	})

	future2, err := executor.Submit(normalTask)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// 再提交任务应该被拒绝
	_, err = executor.Submit(normalTask)
	if err == nil {
		t.Error("Expected task to be rejected, but it was accepted")
	}

	// 等待任务完成
	result1, err := future1.Get()
	if err != nil {
		t.Errorf("Future1 get failed: %v", err)
	}
	if result1 != nil && result1.(string) != "blocked" {
		t.Errorf("Expected 'blocked', got %v", result1)
	}

	result2, err := future2.Get()
	if err != nil {
		t.Errorf("Future2 get failed: %v", err)
	}
	if result2 != nil && result2.(string) != "normal" {
		t.Errorf("Expected 'normal', got %v", result2)
	}
}

// TestThreadPoolExecutor_Shutdown 关闭测试
func TestThreadPoolExecutor_Shutdown(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 2
	cfg.MaxPoolSize = 2
	cfg.QueueSize = 10

	executor := NewThreadPoolExecutor(cfg)

	// 提交一些任务
	for i := 0; i < 5; i++ {
		task := Callable(func(ctx context.Context) (interface{}, error) {
			time.Sleep(50 * time.Millisecond)
			return "done", nil
		})

		_, err := executor.Submit(task)
		if err != nil {
			t.Errorf("Submit failed: %v", err)
		}
	}

	// 关闭执行器
	executor.Shutdown()

	// 等待终止
	if !executor.AwaitTermination(1 * time.Second) {
		t.Error("Executor did not terminate in time")
	}

	// 尝试提交新任务应该失败
	task := Callable(func(ctx context.Context) (interface{}, error) {
		return "new", nil
	})

	_, err := executor.Submit(task)
	if err != ErrExecutorShutdown {
		t.Errorf("Expected ErrExecutorShutdown, got %v", err)
	}
}

// TestThreadPoolExecutor_Metrics 指标测试
func TestThreadPoolExecutor_Metrics(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 2
	cfg.MaxPoolSize = 4
	cfg.QueueSize = 10
	cfg.EnableMetrics = true

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	// 提交一些任务
	for i := 0; i < 5; i++ {
		task := Callable(func(ctx context.Context) (interface{}, error) {
			time.Sleep(10 * time.Millisecond)
			return "done", nil
		})

		future, err := executor.Submit(task)
		if err != nil {
			t.Errorf("Submit failed: %v", err)
		}

		_, err = future.Get()
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
	}

	// 检查指标
	metrics := executor.GetMetrics()

	if metrics.TasksSubmitted != 5 {
		t.Errorf("Expected 5 submitted tasks, got %d", metrics.TasksSubmitted)
	}

	if metrics.TasksCompleted != 5 {
		t.Errorf("Expected 5 completed tasks, got %d", metrics.TasksCompleted)
	}

	if metrics.CoreThreads != 2 {
		t.Errorf("Expected 2 core threads, got %d", metrics.CoreThreads)
	}

	if metrics.MaxThreads != 4 {
		t.Errorf("Expected 4 max threads, got %d", metrics.MaxThreads)
	}
}

// TestExecutors_Compatibility 兼容性测试
func TestExecutors_Compatibility(t *testing.T) {
	executors := NewExecutors()
	defer executors.Stop()

	// 测试原有的 Submit 方法
	callable := func() (interface{}, error) {
		return "hello", nil
	}

	future := executors.Submit(callable)

	// 测试原有的 GetResult 方法
	ret, timeoutErr, err, exception := future.GetResult(1 * time.Second)

	if timeoutErr != nil {
		t.Errorf("Unexpected timeout error: %v", timeoutErr)
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if exception != nil {
		t.Errorf("Unexpected exception: %v", exception)
	}

	if ret.(string) != "hello" {
		t.Errorf("Expected 'hello', got %v", ret)
	}

	// 测试 GetGoNum 方法
	goNum := executors.GetGoNum()
	if goNum < 0 {
		t.Errorf("Invalid goroutine number: %d", goNum)
	}
}
