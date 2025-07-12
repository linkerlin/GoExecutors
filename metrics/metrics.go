package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics 性能指标
type Metrics struct {
	// 任务相关指标
	TasksSubmitted int64 // 提交的任务数
	TasksCompleted int64 // 完成的任务数
	TasksFailed    int64 // 失败的任务数
	TasksTimeout   int64 // 超时的任务数
	TasksPanic     int64 // 恐慌的任务数

	// 执行时间相关指标
	TotalExecutionTime int64 // 总执行时间(纳秒)
	MinExecutionTime   int64 // 最小执行时间(纳秒)
	MaxExecutionTime   int64 // 最大执行时间(纳秒)

	// 线程池相关指标
	ActiveThreads int32 // 活跃线程数
	CoreThreads   int32 // 核心线程数
	MaxThreads    int32 // 最大线程数
	QueueSize     int32 // 队列大小
	QueueCapacity int32 // 队列容量

	// 时间记录
	StartTime time.Time
	mu        sync.RWMutex
}

// NewMetrics 创建性能指标
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime:        time.Now(),
		MinExecutionTime: int64(^uint64(0) >> 1), // 最大int64值
	}
}

// IncrementTasksSubmitted 增加提交任务数
func (m *Metrics) IncrementTasksSubmitted() {
	atomic.AddInt64(&m.TasksSubmitted, 1)
}

// IncrementTasksCompleted 增加完成任务数
func (m *Metrics) IncrementTasksCompleted() {
	atomic.AddInt64(&m.TasksCompleted, 1)
}

// IncrementTasksFailed 增加失败任务数
func (m *Metrics) IncrementTasksFailed() {
	atomic.AddInt64(&m.TasksFailed, 1)
}

// IncrementTasksTimeout 增加超时任务数
func (m *Metrics) IncrementTasksTimeout() {
	atomic.AddInt64(&m.TasksTimeout, 1)
}

// IncrementTasksPanic 增加恐慌任务数
func (m *Metrics) IncrementTasksPanic() {
	atomic.AddInt64(&m.TasksPanic, 1)
}

// RecordExecutionTime 记录执行时间
func (m *Metrics) RecordExecutionTime(duration time.Duration) {
	nanos := duration.Nanoseconds()
	atomic.AddInt64(&m.TotalExecutionTime, nanos)

	// 更新最小执行时间
	for {
		current := atomic.LoadInt64(&m.MinExecutionTime)
		if nanos >= current {
			break
		}
		if atomic.CompareAndSwapInt64(&m.MinExecutionTime, current, nanos) {
			break
		}
	}

	// 更新最大执行时间
	for {
		current := atomic.LoadInt64(&m.MaxExecutionTime)
		if nanos <= current {
			break
		}
		if atomic.CompareAndSwapInt64(&m.MaxExecutionTime, current, nanos) {
			break
		}
	}
}

// SetActiveThreads 设置活跃线程数
func (m *Metrics) SetActiveThreads(count int32) {
	atomic.StoreInt32(&m.ActiveThreads, count)
}

// SetCoreThreads 设置核心线程数
func (m *Metrics) SetCoreThreads(count int32) {
	atomic.StoreInt32(&m.CoreThreads, count)
}

// SetMaxThreads 设置最大线程数
func (m *Metrics) SetMaxThreads(count int32) {
	atomic.StoreInt32(&m.MaxThreads, count)
}

// SetQueueSize 设置队列大小
func (m *Metrics) SetQueueSize(size int32) {
	atomic.StoreInt32(&m.QueueSize, size)
}

// SetQueueCapacity 设置队列容量
func (m *Metrics) SetQueueCapacity(capacity int32) {
	atomic.StoreInt32(&m.QueueCapacity, capacity)
}

// Snapshot 获取指标快照
func (m *Metrics) Snapshot() *MetricsSnapshot {
	now := time.Now()
	return &MetricsSnapshot{
		TasksSubmitted:     atomic.LoadInt64(&m.TasksSubmitted),
		TasksCompleted:     atomic.LoadInt64(&m.TasksCompleted),
		TasksFailed:        atomic.LoadInt64(&m.TasksFailed),
		TasksTimeout:       atomic.LoadInt64(&m.TasksTimeout),
		TasksPanic:         atomic.LoadInt64(&m.TasksPanic),
		TotalExecutionTime: atomic.LoadInt64(&m.TotalExecutionTime),
		MinExecutionTime:   atomic.LoadInt64(&m.MinExecutionTime),
		MaxExecutionTime:   atomic.LoadInt64(&m.MaxExecutionTime),
		ActiveThreads:      atomic.LoadInt32(&m.ActiveThreads),
		CoreThreads:        atomic.LoadInt32(&m.CoreThreads),
		MaxThreads:         atomic.LoadInt32(&m.MaxThreads),
		QueueSize:          atomic.LoadInt32(&m.QueueSize),
		QueueCapacity:      atomic.LoadInt32(&m.QueueCapacity),
		StartTime:          m.StartTime,
		SnapshotTime:       now,
		Uptime:             now.Sub(m.StartTime),
	}
}

// MetricsSnapshot 指标快照
type MetricsSnapshot struct {
	TasksSubmitted     int64
	TasksCompleted     int64
	TasksFailed        int64
	TasksTimeout       int64
	TasksPanic         int64
	TotalExecutionTime int64
	MinExecutionTime   int64
	MaxExecutionTime   int64
	ActiveThreads      int32
	CoreThreads        int32
	MaxThreads         int32
	QueueSize          int32
	QueueCapacity      int32
	StartTime          time.Time
	SnapshotTime       time.Time
	Uptime             time.Duration
}

// AvgExecutionTime 平均执行时间
func (s *MetricsSnapshot) AvgExecutionTime() time.Duration {
	if s.TasksCompleted == 0 {
		return 0
	}
	return time.Duration(s.TotalExecutionTime / s.TasksCompleted)
}

// TaskThroughput 任务吞吐量（任务/秒）
func (s *MetricsSnapshot) TaskThroughput() float64 {
	seconds := s.Uptime.Seconds()
	if seconds == 0 {
		return 0
	}
	return float64(s.TasksCompleted) / seconds
}

// SuccessRate 成功率
func (s *MetricsSnapshot) SuccessRate() float64 {
	if s.TasksSubmitted == 0 {
		return 0
	}
	return float64(s.TasksCompleted) / float64(s.TasksSubmitted)
}

// QueueUtilization 队列利用率
func (s *MetricsSnapshot) QueueUtilization() float64 {
	if s.QueueCapacity == 0 {
		return 0
	}
	return float64(s.QueueSize) / float64(s.QueueCapacity)
}

// ThreadUtilization 线程利用率
func (s *MetricsSnapshot) ThreadUtilization() float64 {
	if s.MaxThreads == 0 {
		return 0
	}
	return float64(s.ActiveThreads) / float64(s.MaxThreads)
}
