# GoExecutors

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/linkerlin/GoExecutors)](https://goreportcard.com/report/github.com/linkerlin/GoExecutors)

一个高性能、工业级的 Go 语言并发执行器库，灵感来自 Java 的 ExecutorService。提供了线程池管理、任务调度、Future 模式等功能。

## 🚀 特性

- **高性能线程池**：基于 goroutine 池的高效任务执行
- **灵活配置**：支持自定义线程池大小、队列容量、超时策略等
- **Future 模式**：支持异步任务执行和结果获取
- **错误处理**：完善的错误处理和 panic 恢复机制
- **性能监控**：内置性能指标收集和监控
- **优雅关闭**：支持优雅关闭和强制关闭
- **多种拒绝策略**：支持 abort、caller-runs、discard 等拒绝策略
- **日志系统**：内置日志系统，支持多种日志级别
- **上下文支持**：完整的 context.Context 支持，便于取消和超时控制
- **兼容性**：保持与旧版本 API 的兼容性

## 📦 安装

```bash
go get github.com/linkerlin/GoExecutors
```

## 🎯 快速开始

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/linkerlin/GoExecutors/config"
    "github.com/linkerlin/GoExecutors/executors"
)

func main() {
    // 创建配置
    cfg := config.DefaultConfig()
    cfg.CorePoolSize = 4
    cfg.MaxPoolSize = 8
    cfg.QueueSize = 100
    
    // 创建执行器
    executor := executors.NewThreadPoolExecutor(cfg)
    defer executor.Shutdown()
    
    // 提交任务
    task := executors.Callable(func(ctx context.Context) (interface{}, error) {
        fmt.Println("Hello, GoExecutors!")
        return "任务完成", nil
    })
    
    future, err := executor.Submit(task)
    if err != nil {
        panic(err)
    }
    
    // 获取结果
    result, err := future.Get()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("结果: %v\n", result)
}
```

### 兼容性用法

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/linkerlin/GoExecutors/executors"
)

func main() {
    // 使用兼容的 API
    es := executors.NewExecutors()
    defer es.Stop()
    
    // 提交任务
    callable := func() (interface{}, error) {
        return "Hello, World!", nil
    }
    
    future := es.Submit(callable)
    
    // 获取结果
    ret, timeoutErr, err, exception := future.GetResult(1 * time.Second)
    if err != nil {
        fmt.Printf("错误: %v\n", err)
    } else if timeoutErr != nil {
        fmt.Printf("超时: %v\n", timeoutErr)
    } else if exception != nil {
        fmt.Printf("异常: %v\n", exception)
    } else {
        fmt.Printf("结果: %v\n", ret)
    }
}
```

## 📚 详细文档

### 配置选项

```go
cfg := &config.Config{
    CorePoolSize:           4,                    // 核心线程数
    MaxPoolSize:            8,                    // 最大线程数
    QueueSize:              100,                  // 队列大小
    KeepAliveTime:          60 * time.Second,     // 线程空闲时间
    AllowCoreThreadTimeOut: false,                // 是否允许核心线程超时
    RejectPolicy:           "abort",              // 拒绝策略
    ThreadNamePrefix:       "goexecutor",         // 线程名称前缀
    EnableMetrics:          true,                 // 启用性能监控
    MetricsInterval:        10 * time.Second,     // 指标收集间隔
    EnableLogging:          true,                 // 启用日志
    LogLevel:               "info",               // 日志级别
}
```

### 环境变量配置

```bash
# 设置环境变量
export GO_EXECUTOR_CORE_POOL_SIZE=8
export GO_EXECUTOR_MAX_POOL_SIZE=16
export GO_EXECUTOR_QUEUE_SIZE=200
export GO_EXECUTOR_KEEP_ALIVE_TIME=30s
export GO_EXECUTOR_REJECT_POLICY=discard
export GO_EXECUTOR_ENABLE_METRICS=true
export GO_EXECUTOR_ENABLE_LOGGING=true
export GO_EXECUTOR_LOG_LEVEL=debug
```

### 任务类型

#### 1. Callable 函数

```go
task := executors.Callable(func(ctx context.Context) (interface{}, error) {
    // 执行任务逻辑
    return "结果", nil
})
```

#### 2. 自定义 Task

```go
type MyTask struct {
    Data string
}

func (t *MyTask) Execute(ctx context.Context) (interface{}, error) {
    // 执行任务逻辑
    return t.Data + " 处理完成", nil
}

// 使用
task := &MyTask{Data: "测试数据"}
future, err := executor.Submit(task)
```

### Future 操作

```go
// 提交任务
future, err := executor.Submit(task)

// 阻塞获取结果
result, err := future.Get()

// 带超时获取结果
result, err := future.GetWithTimeout(5 * time.Second)

// 检查是否完成
if future.IsDone() {
    fmt.Println("任务已完成")
}

// 取消任务
future.Cancel()
```

### 性能监控

```go
// 启用性能监控
cfg.EnableMetrics = true

// 获取性能指标
metrics := executor.GetMetrics()

fmt.Printf("提交任务数: %d\n", metrics.TasksSubmitted)
fmt.Printf("完成任务数: %d\n", metrics.TasksCompleted)
fmt.Printf("失败任务数: %d\n", metrics.TasksFailed)
fmt.Printf("活跃线程数: %d\n", metrics.ActiveThreads)
fmt.Printf("平均执行时间: %v\n", metrics.AvgExecutionTime())
fmt.Printf("任务吞吐量: %.2f 任务/秒\n", metrics.TaskThroughput())
fmt.Printf("成功率: %.2f%%\n", metrics.SuccessRate()*100)
```

### 错误处理

```go
// 1. 正常错误
task := executors.Callable(func(ctx context.Context) (interface{}, error) {
    return nil, errors.New("业务错误")
})

// 2. Panic 恢复
task := executors.Callable(func(ctx context.Context) (interface{}, error) {
    panic("发生恐慌") // 会被自动恢复
})

// 3. 上下文取消
task := executors.Callable(func(ctx context.Context) (interface{}, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // 执行任务
        return "完成", nil
    }
})
```

### 拒绝策略

| 策略 | 描述 |
|------|------|
| `abort` | 抛出异常（默认） |
| `caller_runs` | 在调用者线程中运行 |
| `discard` | 直接丢弃任务 |

### 优雅关闭

```go
// 启动优雅关闭
executor.Shutdown()

// 等待任务完成（带超时）
if executor.AwaitTermination(30 * time.Second) {
    fmt.Println("执行器已优雅关闭")
} else {
    fmt.Println("关闭超时，强制关闭")
    executor.ShutdownNow()
}
```

## 🔧 高级用法

### 批量任务处理

```go
// 批量提交任务
tasks := []executors.Task{
    executors.Callable(func(ctx context.Context) (interface{}, error) {
        return "任务1", nil
    }),
    executors.Callable(func(ctx context.Context) (interface{}, error) {
        return "任务2", nil
    }),
    // ... 更多任务
}

futures := make([]*executors.Future, len(tasks))
for i, task := range tasks {
    future, err := executor.Submit(task)
    if err != nil {
        fmt.Printf("提交任务 %d 失败: %v\n", i, err)
        continue
    }
    futures[i] = future
}

// 等待所有任务完成
for i, future := range futures {
    if future == nil {
        continue
    }
    
    result, err := future.Get()
    if err != nil {
        fmt.Printf("任务 %d 失败: %v\n", i, err)
    } else {
        fmt.Printf("任务 %d 结果: %v\n", i, result)
    }
}
```

### 自定义日志

```go
import "github.com/linkerlin/GoExecutors/logger"

// 创建自定义日志器
customLogger := logger.NewSimpleLogger("debug")

// 设置为全局日志器
logger.SetDefaultLogger(customLogger)

// 或者在配置中启用
cfg.EnableLogging = true
cfg.LogLevel = "debug"
```

## 🧪 测试

```bash
# 运行所有测试
go test -v ./...

# 运行基准测试
go test -v -bench=. ./...

# 运行覆盖率测试
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📊 性能基准

在 MacBook Pro (M1, 16GB) 上的基准测试结果：

```
BenchmarkThreadPoolExecutor_Submit-8           1000000    1203 ns/op
BenchmarkThreadPoolExecutor_SubmitLight-8      2000000     856 ns/op
BenchmarkThreadPoolExecutor_Concurrent-8       500000     2456 ns/op
BenchmarkFuture_Get-8                          5000000     234 ns/op
BenchmarkFuture_GetWithTimeout-8               3000000     456 ns/op
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📜 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](LICENSE) 文件。

## 🔗 相关链接

- [API 文档](https://pkg.go.dev/github.com/linkerlin/GoExecutors)
- [示例代码](examples/)
- [性能测试](benchmarks/)
- [更新日志](CHANGELOG.md)

## 🙏 致谢

- 感谢 Java 的 ExecutorService 提供的设计灵感
- 感谢 Go 社区的优秀工具和库

---

如果这个项目对你有帮助，请给个 ⭐️ 支持一下！
