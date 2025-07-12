package main

import (
	"context"
	"fmt"
	"time"

	"github.com/linkerlin/GoExecutors/config"
	"github.com/linkerlin/GoExecutors/executors"
	"github.com/linkerlin/GoExecutors/logger"
)

func main() {
	// 使用新的 API 演示
	fmt.Println("=== 使用新的 ThreadPoolExecutor API ===")
	newAPIDemo()

	fmt.Println("\n=== 使用兼容的 Executors API ===")
	compatibilityDemo()

	fmt.Println("\n=== 性能监控演示 ===")
	metricsDemo()

	fmt.Println("\n=== 错误处理演示 ===")
	errorHandlingDemo()
}

// 新 API 演示
func newAPIDemo() {
	// 创建自定义配置
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 4
	cfg.MaxPoolSize = 8
	cfg.QueueSize = 100
	cfg.EnableLogging = true
	cfg.LogLevel = "info"
	cfg.EnableMetrics = true

	// 创建执行器
	executor := executors.NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	// 提交简单任务
	task1 := executors.Callable(func(ctx context.Context) (interface{}, error) {
		fmt.Println("执行任务 1")
		time.Sleep(100 * time.Millisecond)
		return "任务 1 完成", nil
	})

	future1, err := executor.Submit(task1)
	if err != nil {
		fmt.Printf("提交任务失败: %v\n", err)
		return
	}

	// 获取结果
	result1, err := future1.Get()
	if err != nil {
		fmt.Printf("获取结果失败: %v\n", err)
		return
	}

	fmt.Printf("任务 1 结果: %v\n", result1)

	// 提交带超时的任务
	task2 := executors.Callable(func(ctx context.Context) (interface{}, error) {
		fmt.Println("执行任务 2（长时间运行）")
		time.Sleep(200 * time.Millisecond)
		return "任务 2 完成", nil
	})

	future2, err := executor.Submit(task2)
	if err != nil {
		fmt.Printf("提交任务失败: %v\n", err)
		return
	}

	// 带超时获取结果
	result2, err := future2.GetWithTimeout(50 * time.Millisecond)
	if err != nil {
		fmt.Printf("任务 2 超时: %v\n", err)

		// 等待任务实际完成
		result2, err = future2.Get()
		if err != nil {
			fmt.Printf("获取结果失败: %v\n", err)
		} else {
			fmt.Printf("任务 2 最终结果: %v\n", result2)
		}
	} else {
		fmt.Printf("任务 2 结果: %v\n", result2)
	}
}

// 兼容性演示
func compatibilityDemo() {
	// 使用兼容的 API
	es := executors.NewExecutors()
	defer es.Stop()

	// 提交任务
	callable := func() (interface{}, error) {
		fmt.Println("兼容性任务执行中...")
		time.Sleep(50 * time.Millisecond)
		return "兼容性任务完成", nil
	}

	future := es.Submit(callable)

	// 使用原有的 GetResult 方法
	ret, timeoutErr, err, exception := future.GetResult(1 * time.Second)

	switch {
	case exception != nil:
		fmt.Printf("任务异常: %v\n", exception)
	case timeoutErr != nil:
		fmt.Printf("任务超时: %v\n", timeoutErr)
	case err != nil:
		fmt.Printf("任务错误: %v\n", err)
	default:
		fmt.Printf("任务结果: %v\n", ret)
	}

	fmt.Printf("当前活跃 Goroutine 数: %d\n", es.GetGoNum())
}

// 性能监控演示
func metricsDemo() {
	// 启用性能监控
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 2
	cfg.MaxPoolSize = 4
	cfg.QueueSize = 10
	cfg.EnableMetrics = true
	cfg.EnableLogging = true
	cfg.LogLevel = "info"

	executor := executors.NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	// 提交多个任务
	for i := 0; i < 10; i++ {
		taskID := i
		task := executors.Callable(func(ctx context.Context) (interface{}, error) {
			fmt.Printf("执行任务 %d\n", taskID)
			time.Sleep(time.Duration(taskID*10) * time.Millisecond)
			return fmt.Sprintf("任务 %d 完成", taskID), nil
		})

		future, err := executor.Submit(task)
		if err != nil {
			fmt.Printf("提交任务 %d 失败: %v\n", taskID, err)
			continue
		}

		// 异步获取结果
		go func(id int, f *executors.Future) {
			result, err := f.Get()
			if err != nil {
				fmt.Printf("任务 %d 执行失败: %v\n", id, err)
			} else {
				fmt.Printf("任务 %d 结果: %v\n", id, result)
			}
		}(taskID, future)
	}

	// 等待一段时间让任务执行
	time.Sleep(200 * time.Millisecond)

	// 获取性能指标
	metrics := executor.GetMetrics()
	fmt.Printf("性能指标:\n")
	fmt.Printf("  提交任务数: %d\n", metrics.TasksSubmitted)
	fmt.Printf("  完成任务数: %d\n", metrics.TasksCompleted)
	fmt.Printf("  失败任务数: %d\n", metrics.TasksFailed)
	fmt.Printf("  活跃线程数: %d\n", metrics.ActiveThreads)
	fmt.Printf("  队列大小: %d\n", metrics.QueueSize)
	fmt.Printf("  平均执行时间: %v\n", metrics.AvgExecutionTime())
	fmt.Printf("  任务吞吐量: %.2f 任务/秒\n", metrics.TaskThroughput())
	fmt.Printf("  成功率: %.2f%%\n", metrics.SuccessRate()*100)
	fmt.Printf("  线程利用率: %.2f%%\n", metrics.ThreadUtilization()*100)

	// 等待剩余任务完成
	time.Sleep(500 * time.Millisecond)
}

// 错误处理演示
func errorHandlingDemo() {
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 2
	cfg.MaxPoolSize = 2
	cfg.QueueSize = 2
	cfg.EnableLogging = true
	cfg.LogLevel = "info"
	cfg.RejectPolicy = "abort"

	executor := executors.NewThreadPoolExecutor(cfg)
	defer executor.Shutdown()

	// 1. 正常任务
	normalTask := executors.Callable(func(ctx context.Context) (interface{}, error) {
		fmt.Println("正常任务执行")
		return "正常完成", nil
	})

	future1, err := executor.Submit(normalTask)
	if err != nil {
		fmt.Printf("提交正常任务失败: %v\n", err)
	} else {
		result, err := future1.Get()
		if err != nil {
			fmt.Printf("正常任务执行失败: %v\n", err)
		} else {
			fmt.Printf("正常任务结果: %v\n", result)
		}
	}

	// 2. 返回错误的任务
	errorTask := executors.Callable(func(ctx context.Context) (interface{}, error) {
		fmt.Println("错误任务执行")
		return nil, fmt.Errorf("这是一个预期的错误")
	})

	future2, err := executor.Submit(errorTask)
	if err != nil {
		fmt.Printf("提交错误任务失败: %v\n", err)
	} else {
		result, err := future2.Get()
		if err != nil {
			fmt.Printf("错误任务执行失败: %v\n", err)
		} else {
			fmt.Printf("错误任务结果: %v\n", result)
		}
	}

	// 3. 会发生 panic 的任务
	panicTask := executors.Callable(func(ctx context.Context) (interface{}, error) {
		fmt.Println("恐慌任务执行")
		panic("这是一个预期的恐慌")
	})

	future3, err := executor.Submit(panicTask)
	if err != nil {
		fmt.Printf("提交恐慌任务失败: %v\n", err)
	} else {
		result, err := future3.Get()
		if err != nil {
			fmt.Printf("恐慌任务执行失败: %v\n", err)
		} else {
			fmt.Printf("恐慌任务结果: %v\n", result)
		}
	}

	// 4. 可以取消的任务
	cancelableTask := executors.Callable(func(ctx context.Context) (interface{}, error) {
		fmt.Println("可取消任务开始执行")
		select {
		case <-ctx.Done():
			fmt.Println("可取消任务被取消")
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
			fmt.Println("可取消任务执行完成")
			return "可取消任务完成", nil
		}
	})

	future4, err := executor.Submit(cancelableTask)
	if err != nil {
		fmt.Printf("提交可取消任务失败: %v\n", err)
	} else {
		// 等待一小段时间后取消
		time.Sleep(100 * time.Millisecond)
		future4.Cancel()

		result, err := future4.Get()
		if err != nil {
			fmt.Printf("可取消任务执行失败: %v\n", err)
		} else {
			fmt.Printf("可取消任务结果: %v\n", result)
		}
	}

	// 5. 提交大量任务测试拒绝策略
	fmt.Println("测试任务拒绝策略...")

	// 先提交阻塞任务占用所有线程
	blockingTask := executors.Callable(func(ctx context.Context) (interface{}, error) {
		time.Sleep(200 * time.Millisecond)
		return "阻塞任务完成", nil
	})

	for i := 0; i < 2; i++ {
		_, err := executor.Submit(blockingTask)
		if err != nil {
			fmt.Printf("提交阻塞任务 %d 失败: %v\n", i, err)
		}
	}

	// 填满队列
	quickTask := executors.Callable(func(ctx context.Context) (interface{}, error) {
		return "快速任务完成", nil
	})

	for i := 0; i < 2; i++ {
		_, err := executor.Submit(quickTask)
		if err != nil {
			fmt.Printf("提交快速任务 %d 失败: %v\n", i, err)
		}
	}

	// 这个任务应该被拒绝
	_, err = executor.Submit(quickTask)
	if err != nil {
		fmt.Printf("任务被拒绝: %v\n", err)
	}

	// 等待任务完成
	time.Sleep(300 * time.Millisecond)
}

// 配置演示函数
func configDemo() {
	// 1. 使用默认配置
	fmt.Println("=== 默认配置 ===")
	defaultCfg := config.DefaultConfig()
	fmt.Printf("核心线程数: %d\n", defaultCfg.CorePoolSize)
	fmt.Printf("最大线程数: %d\n", defaultCfg.MaxPoolSize)
	fmt.Printf("队列大小: %d\n", defaultCfg.QueueSize)
	fmt.Printf("保持活跃时间: %v\n", defaultCfg.KeepAliveTime)
	fmt.Printf("拒绝策略: %s\n", defaultCfg.RejectPolicy)

	// 2. 从环境变量加载配置
	fmt.Println("\n=== 从环境变量加载配置 ===")
	envCfg := config.DefaultConfig()
	envCfg.LoadFromEnv()
	fmt.Printf("核心线程数: %d\n", envCfg.CorePoolSize)
	fmt.Printf("最大线程数: %d\n", envCfg.MaxPoolSize)

	// 3. 自定义配置
	fmt.Println("\n=== 自定义配置 ===")
	customCfg := &config.Config{
		CorePoolSize:    2,
		MaxPoolSize:     4,
		QueueSize:       50,
		KeepAliveTime:   30 * time.Second,
		RejectPolicy:    "discard",
		EnableLogging:   true,
		LogLevel:        "debug",
		EnableMetrics:   true,
		MetricsInterval: 5 * time.Second,
	}

	err := customCfg.Validate()
	if err != nil {
		fmt.Printf("配置验证失败: %v\n", err)
	} else {
		fmt.Println("自定义配置验证通过")
	}
}

// 日志演示函数
func loggerDemo() {
	// 1. 使用全局日志器
	fmt.Println("=== 全局日志器演示 ===")

	// 设置日志器
	simpleLogger := logger.NewSimpleLogger("debug")
	logger.SetDefaultLogger(simpleLogger)

	// 使用全局日志函数
	logger.Debug("这是一个调试消息")
	logger.Info("这是一个信息消息")
	logger.Warn("这是一个警告消息")
	logger.Error("这是一个错误消息")

	// 格式化日志
	logger.Infof("用户 %s 的年龄是 %d", "Alice", 30)

	// 2. 使用自定义日志器
	fmt.Println("\n=== 自定义日志器演示 ===")
	customLogger := logger.NewSimpleLogger("warn")

	customLogger.Debug("这条调试消息不会显示")
	customLogger.Info("这条信息消息不会显示")
	customLogger.Warn("这条警告消息会显示")
	customLogger.Error("这条错误消息会显示")

	// 3. 使用空日志器
	fmt.Println("\n=== 空日志器演示 ===")
	noOpLogger := logger.NewNoOpLogger()
	noOpLogger.Info("这条消息不会显示")
	noOpLogger.Error("这条错误消息也不会显示")
}
