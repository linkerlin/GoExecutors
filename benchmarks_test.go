package executors

import (
	"context"
	"testing"
	"time"

	"github.com/linkerlin/GoExecutors/config"
)

// BenchmarkThreadPoolExecutor_Submit 测试任务提交性能
func BenchmarkThreadPoolExecutor_Submit(b *testing.B) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 4
	cfg.MaxPoolSize = 8
	cfg.QueueSize = 10000
	cfg.EnableLogging = false
	cfg.EnableMetrics = false

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	task := Callable(func(ctx context.Context) (interface{}, error) {
		return "done", nil
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			future, err := executor.Submit(task)
			if err != nil {
				b.Error(err)
				continue
			}

			_, err = future.Get()
			if err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkThreadPoolExecutor_SubmitLight 测试轻量级任务提交性能
func BenchmarkThreadPoolExecutor_SubmitLight(b *testing.B) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 4
	cfg.MaxPoolSize = 8
	cfg.QueueSize = 10000
	cfg.EnableLogging = false
	cfg.EnableMetrics = false

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	task := Callable(func(ctx context.Context) (interface{}, error) {
		return 42, nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		future, err := executor.Submit(task)
		if err != nil {
			b.Error(err)
			continue
		}

		_, err = future.Get()
		if err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkThreadPoolExecutor_SubmitHeavy 测试重量级任务提交性能
func BenchmarkThreadPoolExecutor_SubmitHeavy(b *testing.B) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 4
	cfg.MaxPoolSize = 8
	cfg.QueueSize = 1000
	cfg.EnableLogging = false
	cfg.EnableMetrics = false

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	task := Callable(func(ctx context.Context) (interface{}, error) {
		// 模拟一些计算密集型工作
		sum := 0
		for i := 0; i < 1000; i++ {
			sum += i
		}
		return sum, nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		future, err := executor.Submit(task)
		if err != nil {
			b.Error(err)
			continue
		}

		_, err = future.Get()
		if err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkThreadPoolExecutor_Concurrent 测试并发性能
func BenchmarkThreadPoolExecutor_Concurrent(b *testing.B) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 4
	cfg.MaxPoolSize = 8
	cfg.QueueSize = 10000
	cfg.EnableLogging = false
	cfg.EnableMetrics = false

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	task := Callable(func(ctx context.Context) (interface{}, error) {
		time.Sleep(1 * time.Millisecond)
		return "done", nil
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			future, err := executor.Submit(task)
			if err != nil {
				b.Error(err)
				continue
			}

			_, err = future.Get()
			if err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkThreadPoolExecutor_BatchSubmit 测试批量提交性能
func BenchmarkThreadPoolExecutor_BatchSubmit(b *testing.B) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 4
	cfg.MaxPoolSize = 8
	cfg.QueueSize = 10000
	cfg.EnableLogging = false
	cfg.EnableMetrics = false

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	task := Callable(func(ctx context.Context) (interface{}, error) {
		return "done", nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		const batchSize = 100
		futures := make([]*Future, batchSize)

		// 批量提交
		for j := 0; j < batchSize; j++ {
			future, err := executor.Submit(task)
			if err != nil {
				b.Error(err)
				continue
			}
			futures[j] = future
		}

		// 等待所有任务完成
		for _, future := range futures {
			if future != nil {
				_, err := future.Get()
				if err != nil {
					b.Error(err)
				}
			}
		}
	}
}

// BenchmarkExecutors_Compatibility 测试兼容性包装器性能
func BenchmarkExecutors_Compatibility(b *testing.B) {
	executors := NewExecutors()
	defer executors.Stop()

	callable := func() (interface{}, error) {
		return "done", nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		future := executors.Submit(callable)
		ret, timeoutErr, err, exception := future.GetResult(1 * time.Second)

		if timeoutErr != nil || err != nil || exception != nil {
			b.Error("Task execution failed")
		}

		if ret.(string) != "done" {
			b.Error("Unexpected result")
		}
	}
}

// BenchmarkFuture_Get 测试 Future.Get 性能
func BenchmarkFuture_Get(b *testing.B) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 4
	cfg.MaxPoolSize = 8
	cfg.QueueSize = 10000
	cfg.EnableLogging = false
	cfg.EnableMetrics = false

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	// 预先创建 futures
	futures := make([]*Future, b.N)
	for i := 0; i < b.N; i++ {
		task := Callable(func(ctx context.Context) (interface{}, error) {
			return i, nil
		})

		future, err := executor.Submit(task)
		if err != nil {
			b.Fatal(err)
		}
		futures[i] = future
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := futures[i].Get()
		if err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkFuture_GetWithTimeout 测试 Future.GetWithTimeout 性能
func BenchmarkFuture_GetWithTimeout(b *testing.B) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 4
	cfg.MaxPoolSize = 8
	cfg.QueueSize = 10000
	cfg.EnableLogging = false
	cfg.EnableMetrics = false

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	// 预先创建 futures
	futures := make([]*Future, b.N)
	for i := 0; i < b.N; i++ {
		task := Callable(func(ctx context.Context) (interface{}, error) {
			return i, nil
		})

		future, err := executor.Submit(task)
		if err != nil {
			b.Fatal(err)
		}
		futures[i] = future
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := futures[i].GetWithTimeout(1 * time.Second)
		if err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkThreadPoolExecutor_HighContention 测试高竞争情况下的性能
func BenchmarkThreadPoolExecutor_HighContention(b *testing.B) {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 2
	cfg.MaxPoolSize = 4
	cfg.QueueSize = 100 // 小队列增加竞争
	cfg.EnableLogging = false
	cfg.EnableMetrics = false

	executor := NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	task := Callable(func(ctx context.Context) (interface{}, error) {
		// 模拟一些工作
		time.Sleep(1 * time.Millisecond)
		return "done", nil
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			future, err := executor.Submit(task)
			if err != nil {
				// 在高竞争情况下，可能会有任务被拒绝
				continue
			}

			_, err = future.Get()
			if err != nil {
				b.Error(err)
			}
		}
	})
}
