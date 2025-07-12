package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/linkerlin/GoExecutors/config"
	"github.com/linkerlin/GoExecutors/executors"
	"github.com/linkerlin/GoExecutors/logger"
)

func main() {
	fmt.Println("=== GoExecutors 演示 ===")

	// 设置日志
	simpleLogger := logger.NewSimpleLogger("info")
	logger.SetDefaultLogger(simpleLogger)

	// 兼容性演示 - 使用原有的 API
	compatibilityDemo()

	// 新 API 演示
	fmt.Println("\n=== 新 API 演示 ===")
	newAPIDemo()
}

// 兼容性演示 - 保持原有的测试逻辑
func compatibilityDemo() {
	config.LoadConfig()
	fmt.Println("Default goroutines number is ", config.DefaultGoroutinesNum())
	es := executors.NewExecutors()
	defer es.Stop()

	// 测试 1: 正常任务
	f := func() (interface{}, error) {
		fmt.Println("这是从一个Callable内部发出的声音。")
		return 1, nil
	}

	var future = es.Submit(f)
	var ret, t, e, ex = future.GetResult(time.Millisecond * 1500)
	switch {
	case ex != nil:
		fmt.Println("异常", ex)
	case t == nil && e == nil:
		fmt.Println("No.1 正常", ret)
	case t != nil:
		fmt.Println("超时！")
	case e != nil:
		fmt.Println("出错", e)
	default:
		fmt.Println("不会到这里", ret)
	}

	// 测试 2: 超时任务
	fTimeout := func() (interface{}, error) {
		time.Sleep(time.Second * 1)
		fmt.Println("这是第二次从Callable内部发出的声音。")
		return 2, errors.New("1s")
	}
	fmt.Println("=================")
	time.Sleep(100 * time.Millisecond)
	future = es.Submit(fTimeout)
	ret2, t, err, ex := future.GetResult(time.Millisecond * 500)
	switch {
	case ex != nil:
		fmt.Println("异常", ex)
	case t == nil && err == nil:
		fmt.Println("执行成功", ret2)
	case err != nil:
		fmt.Println("执行出错", err)
	case t != nil:
		fmt.Println("No.2 超时！", t)
	default:
		fmt.Println("不会到这里", ret2)
	}

	// 测试 3: Panic 任务
	fPanic := func() (interface{}, error) {
		fmt.Println("这是第三次从Callable内部发出的声音。")
		panic(100)
	}
	for i := 0; i < 3; i++ {
		future = es.Submit(fPanic)
	}

	ret3, t, err, ex := future.GetResult(time.Millisecond * 500)
	switch {
	case ex != nil:
		fmt.Printf("No.3 异常 %d\n", es.GetGoNum())
	case err == nil && t == nil:
		fmt.Println("执行失败,没有捕获到错误", ret3)
	case t != nil:
		fmt.Println("执行失败,超时", t)
	case err != nil:
		fmt.Println("执行成功,捕获到", err)
	default:
		fmt.Println("不会到这里", ret3)
	}

	// 测试 4: 错误任务
	f = func() (interface{}, error) {
		fmt.Println("这是从No.4 Callable内部发出的声音。", es.GetGoNum())
		return 1, errors.New("😀")
	}

	future = es.Submit(f)
	ret, t, e, ex = future.GetResult(time.Millisecond * 1500)
	switch {
	case ex != nil:
		fmt.Println("异常", ex)
	case t == nil && e == nil:
		fmt.Println("正常", ret)
	case t != nil:
		fmt.Println("超时！")
	case e != nil:
		fmt.Println("No.4 出错", e)
	default:
		fmt.Println("不会到这里", ret)
	}

	fmt.Println("GoNum:", es.GetGoNum())
	time.Sleep(time.Second * 1)
	fmt.Println("GoNum:", es.GetGoNum())

	// 等待任务完成
	time.Sleep(time.Second * 2)
	fmt.Println("Final GoNum:", es.GetGoNum())
}

// 新 API 演示
func newAPIDemo() {
	// 创建配置
	cfg := config.DefaultConfig()
	cfg.CorePoolSize = 4
	cfg.MaxPoolSize = 8
	cfg.QueueSize = 100
	cfg.EnableLogging = true
	cfg.LogLevel = "info"
	cfg.EnableMetrics = true
	cfg.MetricsInterval = 1 * time.Second

	// 创建执行器
	executor := executors.NewThreadPoolExecutor(cfg)
	defer func() {
		executor.Shutdown()
		executor.AwaitTermination(5 * time.Second)
	}()

	// 提交任务
	task := executors.Callable(func(ctx context.Context) (interface{}, error) {
		fmt.Println("新 API 任务执行中...")
		time.Sleep(100 * time.Millisecond)
		return "新 API 任务完成", nil
	})

	future, err := executor.Submit(task)
	if err != nil {
		fmt.Printf("提交任务失败: %v\n", err)
		return
	}

	// 获取结果
	result, err := future.Get()
	if err != nil {
		fmt.Printf("获取结果失败: %v\n", err)
		return
	}

	fmt.Printf("新 API 任务结果: %v\n", result)

	// 显示性能指标
	time.Sleep(100 * time.Millisecond)
	metrics := executor.GetMetrics()
	fmt.Printf("性能指标: 提交=%d, 完成=%d, 活跃线程=%d\n",
		metrics.TasksSubmitted, metrics.TasksCompleted, metrics.ActiveThreads)
}
