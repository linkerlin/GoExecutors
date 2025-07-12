package executors

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/linkerlin/GoExecutors/config"
	"github.com/linkerlin/GoExecutors/logger"
	"github.com/linkerlin/GoExecutors/metrics"
)

// 错误定义
var (
	ErrExecutorShutdown = errors.New("executor has been shutdown")
	ErrTaskRejected     = errors.New("task rejected by executor")
	ErrTaskTimeout      = errors.New("task execution timeout")
	ErrTaskPanic        = errors.New("task execution panic")
)

// Task 任务接口
type Task interface {
	Execute(ctx context.Context) (interface{}, error)
}

// Callable 函数式任务
type Callable func(ctx context.Context) (interface{}, error)

// Execute 实现 Task 接口
func (c Callable) Execute(ctx context.Context) (interface{}, error) {
	return c(ctx)
}

// Result 任务执行结果
type Result struct {
	Value interface{}
	Error error
}

// Future 异步任务的未来结果
type Future struct {
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	result *Result
	once   sync.Once
}

// NewFuture 创建新的 Future
func NewFuture(ctx context.Context) *Future {
	futureCtx, cancel := context.WithCancel(ctx)
	return &Future{
		ctx:    futureCtx,
		cancel: cancel,
		done:   make(chan struct{}),
	}
}

// Get 获取结果，会阻塞直到任务完成
func (f *Future) Get() (interface{}, error) {
	<-f.done
	if f.result == nil {
		return nil, errors.New("future has no result")
	}
	return f.result.Value, f.result.Error
}

// GetWithTimeout 获取结果，带超时
func (f *Future) GetWithTimeout(timeout time.Duration) (interface{}, error) {
	select {
	case <-f.done:
		if f.result == nil {
			return nil, errors.New("future has no result")
		}
		return f.result.Value, f.result.Error
	case <-time.After(timeout):
		return nil, ErrTaskTimeout
	}
}

// IsDone 检查任务是否完成
func (f *Future) IsDone() bool {
	select {
	case <-f.done:
		return true
	default:
		return false
	}
}

// Cancel 取消任务
func (f *Future) Cancel() {
	f.cancel()
}

// complete 完成任务
func (f *Future) complete(result *Result) {
	f.once.Do(func() {
		f.result = result
		close(f.done)
	})
}

// taskWrapper 任务包装器
type taskWrapper struct {
	task   Task
	future *Future
}

// ThreadPoolExecutor 线程池执行器
type ThreadPoolExecutor struct {
	config  *config.Config
	logger  logger.Logger
	metrics *metrics.Metrics

	// 状态管理
	state int32 // 0: running, 1: shutdown, 2: terminated

	// 工作线程管理
	workers     int32
	coreWorkers int32

	// 任务队列
	taskQueue chan *taskWrapper

	// 控制通道
	shutdownCh chan struct{}

	// 等待组
	wg sync.WaitGroup

	// 互斥锁
	mu sync.RWMutex
}

// NewThreadPoolExecutor 创建线程池执行器
func NewThreadPoolExecutor(cfg *config.Config) *ThreadPoolExecutor {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	cfg.Validate()

	var log logger.Logger
	if cfg.EnableLogging {
		log = logger.NewSimpleLogger(cfg.LogLevel)
	} else {
		log = logger.NewNoOpLogger()
	}

	executor := &ThreadPoolExecutor{
		config:     cfg,
		logger:     log,
		metrics:    metrics.NewMetrics(),
		taskQueue:  make(chan *taskWrapper, cfg.QueueSize),
		shutdownCh: make(chan struct{}),
	}

	// 设置指标
	executor.metrics.SetCoreThreads(cfg.CorePoolSize)
	executor.metrics.SetMaxThreads(cfg.MaxPoolSize)
	executor.metrics.SetQueueCapacity(int32(cfg.QueueSize))

	// 启动核心工作线程
	for i := int32(0); i < cfg.CorePoolSize; i++ {
		executor.startWorker(true)
	}

	// 启动监控线程
	if cfg.EnableMetrics {
		go executor.metricsLoop()
	}

	executor.logger.Infof("ThreadPoolExecutor started with config: core=%d, max=%d, queue=%d",
		cfg.CorePoolSize, cfg.MaxPoolSize, cfg.QueueSize)

	return executor
}

// Submit 提交任务
func (e *ThreadPoolExecutor) Submit(task Task) (*Future, error) {
	return e.SubmitWithContext(context.Background(), task)
}

// SubmitWithContext 提交任务带上下文
func (e *ThreadPoolExecutor) SubmitWithContext(ctx context.Context, task Task) (*Future, error) {
	if atomic.LoadInt32(&e.state) != 0 {
		return nil, ErrExecutorShutdown
	}

	future := NewFuture(ctx)
	wrapper := &taskWrapper{
		task:   task,
		future: future,
	}

	// 尝试提交任务
	select {
	case e.taskQueue <- wrapper:
		e.metrics.IncrementTasksSubmitted()
		e.logger.Debugf("Task submitted successfully")

		// 检查是否需要启动新的工作线程
		e.checkAndStartWorker()

		return future, nil
	default:
		// 队列满了，根据拒绝策略处理
		return nil, e.handleRejectedTask(wrapper)
	}
}

// SubmitCallable 提交函数式任务
func (e *ThreadPoolExecutor) SubmitCallable(callable func(ctx context.Context) (interface{}, error)) (*Future, error) {
	return e.Submit(Callable(callable))
}

// checkAndStartWorker 检查并启动工作线程
func (e *ThreadPoolExecutor) checkAndStartWorker() {
	currentWorkers := atomic.LoadInt32(&e.workers)
	queueSize := int32(len(e.taskQueue))

	// 如果队列有积压且工作线程数小于最大值，启动新工作线程
	if queueSize > 0 && currentWorkers < e.config.MaxPoolSize {
		if atomic.CompareAndSwapInt32(&e.workers, currentWorkers, currentWorkers+1) {
			go e.startWorker(false)
		}
	}
}

// startWorker 启动工作线程
func (e *ThreadPoolExecutor) startWorker(isCore bool) {
	if isCore {
		atomic.AddInt32(&e.coreWorkers, 1)
	}
	atomic.AddInt32(&e.workers, 1)

	e.wg.Add(1)
	go e.workerLoop(isCore)
}

// workerLoop 工作线程主循环
func (e *ThreadPoolExecutor) workerLoop(isCore bool) {
	defer e.wg.Done()
	defer func() {
		atomic.AddInt32(&e.workers, -1)
		if isCore {
			atomic.AddInt32(&e.coreWorkers, -1)
		}
	}()

	workerID := fmt.Sprintf("%s-%d", e.config.ThreadNamePrefix, atomic.LoadInt32(&e.workers))
	e.logger.Debugf("Worker %s started", workerID)

	for {
		select {
		case <-e.shutdownCh:
			e.logger.Debugf("Worker %s received shutdown signal", workerID)
			return
		case wrapper := <-e.taskQueue:
			e.executeTask(wrapper, workerID)
		case <-time.After(e.config.KeepAliveTime):
			// 非核心线程空闲超时
			if !isCore || (isCore && e.config.AllowCoreThreadTimeOut) {
				currentWorkers := atomic.LoadInt32(&e.workers)
				minWorkers := e.config.CorePoolSize
				if e.config.AllowCoreThreadTimeOut {
					minWorkers = 0
				}

				if currentWorkers > minWorkers {
					e.logger.Debugf("Worker %s idle timeout, shutting down", workerID)
					return
				}
			}
		}
	}
}

// executeTask 执行任务
func (e *ThreadPoolExecutor) executeTask(wrapper *taskWrapper, workerID string) {
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		e.metrics.RecordExecutionTime(duration)
		e.logger.Debugf("Worker %s completed task in %v", workerID, duration)
	}()

	// 恢复 panic
	defer func() {
		if r := recover(); r != nil {
			e.logger.Errorf("Worker %s panic: %v", workerID, r)
			e.metrics.IncrementTasksPanic()
			wrapper.future.complete(&Result{Error: fmt.Errorf("%w: %v", ErrTaskPanic, r)})
		}
	}()

	// 检查任务是否已取消
	select {
	case <-wrapper.future.ctx.Done():
		e.logger.Debugf("Worker %s task cancelled", workerID)
		wrapper.future.complete(&Result{Error: wrapper.future.ctx.Err()})
		return
	default:
	}

	// 执行任务
	result, err := wrapper.task.Execute(wrapper.future.ctx)

	if err != nil {
		e.metrics.IncrementTasksFailed()
		e.logger.Debugf("Worker %s task failed: %v", workerID, err)
	} else {
		e.metrics.IncrementTasksCompleted()
		e.logger.Debugf("Worker %s task completed successfully", workerID)
	}

	wrapper.future.complete(&Result{Value: result, Error: err})
}

// handleRejectedTask 处理被拒绝的任务
func (e *ThreadPoolExecutor) handleRejectedTask(wrapper *taskWrapper) error {
	e.logger.Warnf("Task rejected due to queue full, policy: %s", e.config.RejectPolicy)

	switch e.config.RejectPolicy {
	case "abort":
		wrapper.future.complete(&Result{Error: ErrTaskRejected})
		return ErrTaskRejected
	case "caller_runs":
		// 在调用者线程中运行任务
		go func() {
			defer func() {
				if r := recover(); r != nil {
					wrapper.future.complete(&Result{Error: fmt.Errorf("%w: %v", ErrTaskPanic, r)})
				}
			}()
			result, err := wrapper.task.Execute(wrapper.future.ctx)
			wrapper.future.complete(&Result{Value: result, Error: err})
		}()
		return nil
	case "discard":
		wrapper.future.complete(&Result{Error: ErrTaskRejected})
		return nil
	default:
		wrapper.future.complete(&Result{Error: ErrTaskRejected})
		return ErrTaskRejected
	}
}

// metricsLoop 指标收集循环
func (e *ThreadPoolExecutor) metricsLoop() {
	ticker := time.NewTicker(e.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.shutdownCh:
			return
		case <-ticker.C:
			e.updateMetrics()
		}
	}
}

// updateMetrics 更新指标
func (e *ThreadPoolExecutor) updateMetrics() {
	e.metrics.SetActiveThreads(atomic.LoadInt32(&e.workers))
	e.metrics.SetQueueSize(int32(len(e.taskQueue)))
}

// GetMetrics 获取指标
func (e *ThreadPoolExecutor) GetMetrics() *metrics.MetricsSnapshot {
	e.updateMetrics()
	return e.metrics.Snapshot()
}

// GetActiveThreadCount 获取活跃线程数
func (e *ThreadPoolExecutor) GetActiveThreadCount() int32 {
	return atomic.LoadInt32(&e.workers)
}

// GetQueueSize 获取队列大小
func (e *ThreadPoolExecutor) GetQueueSize() int {
	return len(e.taskQueue)
}

// Shutdown 优雅关闭
func (e *ThreadPoolExecutor) Shutdown() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if atomic.CompareAndSwapInt32(&e.state, 0, 1) {
		e.logger.Info("Shutting down executor...")
		close(e.shutdownCh)
	}
}

// ShutdownNow 立即关闭
func (e *ThreadPoolExecutor) ShutdownNow() []Task {
	e.mu.Lock()
	defer e.mu.Unlock()

	if atomic.CompareAndSwapInt32(&e.state, 0, 2) {
		e.logger.Info("Shutting down executor immediately...")
		close(e.shutdownCh)

		// 收集未执行的任务
		var unexecutedTasks []Task
		close(e.taskQueue)
		for wrapper := range e.taskQueue {
			wrapper.future.complete(&Result{Error: ErrExecutorShutdown})
			unexecutedTasks = append(unexecutedTasks, wrapper.task)
		}

		return unexecutedTasks
	}

	return nil
}

// AwaitTermination 等待终止
func (e *ThreadPoolExecutor) AwaitTermination(timeout time.Duration) bool {
	done := make(chan struct{})

	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		e.logger.Info("Executor terminated")
		return true
	case <-time.After(timeout):
		e.logger.Warn("Executor termination timeout")
		return false
	}
}

// IsShutdown 检查是否已关闭
func (e *ThreadPoolExecutor) IsShutdown() bool {
	return atomic.LoadInt32(&e.state) != 0
}

// IsTerminated 检查是否已终止
func (e *ThreadPoolExecutor) IsTerminated() bool {
	return atomic.LoadInt32(&e.state) == 2
}

// 为了兼容原有的 API，保留以下类型和函数

// ErrorTimeout 超时错误
type ErrorTimeout string

func (e ErrorTimeout) Error() string { return string(e) }

// Executors 兼容性包装器
type Executors struct {
	executor *ThreadPoolExecutor
}

// NewExecutors 创建兼容性执行器
func NewExecutors() *Executors {
	return &Executors{
		executor: NewThreadPoolExecutor(config.DefaultConfig()),
	}
}

// Submit 提交任务（兼容性方法）
func (e *Executors) Submit(callable func() (interface{}, error)) *Future {
	// 包装原有的 callable 函数
	task := Callable(func(ctx context.Context) (interface{}, error) {
		return callable()
	})

	future, err := e.executor.Submit(task)
	if err != nil {
		// 创建一个失败的 future
		future = NewFuture(context.Background())
		future.complete(&Result{Error: err})
	}

	return future
}

// GetGoNum 获取 goroutine 数量（兼容性方法）
func (e *Executors) GetGoNum() int32 {
	return e.executor.GetActiveThreadCount()
}

// Stop 停止执行器（兼容性方法）
func (e *Executors) Stop() {
	e.executor.Shutdown()
}

// GetResult 获取结果（兼容性方法）
func (f *Future) GetResult(timeout time.Duration) (ret interface{}, timeoutError error, err error, exception interface{}) {
	result, getErr := f.GetWithTimeout(timeout)

	if getErr != nil {
		if errors.Is(getErr, ErrTaskTimeout) {
			return nil, getErr, nil, nil
		}
		if errors.Is(getErr, ErrTaskPanic) {
			return nil, nil, nil, getErr
		}
		return nil, nil, getErr, nil
	}

	return result, nil, nil, nil
}
