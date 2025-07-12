package metrics

import (
	"testing"
	"time"
)

// TestMetrics_Basic 基本指标测试
func TestMetrics_Basic(t *testing.T) {
	m := NewMetrics()

	// 测试初始值
	if m.TasksSubmitted != 0 {
		t.Errorf("Expected TasksSubmitted 0, got %d", m.TasksSubmitted)
	}

	if m.TasksCompleted != 0 {
		t.Errorf("Expected TasksCompleted 0, got %d", m.TasksCompleted)
	}

	if m.MinExecutionTime != int64(^uint64(0)>>1) {
		t.Errorf("Expected MinExecutionTime max int64, got %d", m.MinExecutionTime)
	}

	if m.MaxExecutionTime != 0 {
		t.Errorf("Expected MaxExecutionTime 0, got %d", m.MaxExecutionTime)
	}
}

// TestMetrics_Increment 测试计数器增加
func TestMetrics_Increment(t *testing.T) {
	m := NewMetrics()

	// 测试提交任务数
	m.IncrementTasksSubmitted()
	if m.TasksSubmitted != 1 {
		t.Errorf("Expected TasksSubmitted 1, got %d", m.TasksSubmitted)
	}

	// 测试完成任务数
	m.IncrementTasksCompleted()
	if m.TasksCompleted != 1 {
		t.Errorf("Expected TasksCompleted 1, got %d", m.TasksCompleted)
	}

	// 测试失败任务数
	m.IncrementTasksFailed()
	if m.TasksFailed != 1 {
		t.Errorf("Expected TasksFailed 1, got %d", m.TasksFailed)
	}

	// 测试超时任务数
	m.IncrementTasksTimeout()
	if m.TasksTimeout != 1 {
		t.Errorf("Expected TasksTimeout 1, got %d", m.TasksTimeout)
	}

	// 测试恐慌任务数
	m.IncrementTasksPanic()
	if m.TasksPanic != 1 {
		t.Errorf("Expected TasksPanic 1, got %d", m.TasksPanic)
	}
}

// TestMetrics_RecordExecutionTime 测试执行时间记录
func TestMetrics_RecordExecutionTime(t *testing.T) {
	m := NewMetrics()

	// 记录第一个执行时间
	m.RecordExecutionTime(100 * time.Millisecond)

	expectedTotal := (100 * time.Millisecond).Nanoseconds()
	if m.TotalExecutionTime != expectedTotal {
		t.Errorf("Expected TotalExecutionTime %d, got %d", expectedTotal, m.TotalExecutionTime)
	}

	if m.MinExecutionTime != expectedTotal {
		t.Errorf("Expected MinExecutionTime %d, got %d", expectedTotal, m.MinExecutionTime)
	}

	if m.MaxExecutionTime != expectedTotal {
		t.Errorf("Expected MaxExecutionTime %d, got %d", expectedTotal, m.MaxExecutionTime)
	}

	// 记录第二个执行时间（更短）
	m.RecordExecutionTime(50 * time.Millisecond)

	expectedMin := (50 * time.Millisecond).Nanoseconds()
	expectedMax := (100 * time.Millisecond).Nanoseconds()
	expectedTotal = expectedMin + expectedMax

	if m.TotalExecutionTime != expectedTotal {
		t.Errorf("Expected TotalExecutionTime %d, got %d", expectedTotal, m.TotalExecutionTime)
	}

	if m.MinExecutionTime != expectedMin {
		t.Errorf("Expected MinExecutionTime %d, got %d", expectedMin, m.MinExecutionTime)
	}

	if m.MaxExecutionTime != expectedMax {
		t.Errorf("Expected MaxExecutionTime %d, got %d", expectedMax, m.MaxExecutionTime)
	}

	// 记录第三个执行时间（更长）
	m.RecordExecutionTime(200 * time.Millisecond)

	expectedMax = (200 * time.Millisecond).Nanoseconds()
	expectedTotal = expectedMin + (100 * time.Millisecond).Nanoseconds() + expectedMax

	if m.TotalExecutionTime != expectedTotal {
		t.Errorf("Expected TotalExecutionTime %d, got %d", expectedTotal, m.TotalExecutionTime)
	}

	if m.MinExecutionTime != expectedMin {
		t.Errorf("Expected MinExecutionTime %d, got %d", expectedMin, m.MinExecutionTime)
	}

	if m.MaxExecutionTime != expectedMax {
		t.Errorf("Expected MaxExecutionTime %d, got %d", expectedMax, m.MaxExecutionTime)
	}
}

// TestMetrics_ThreadCounts 测试线程数设置
func TestMetrics_ThreadCounts(t *testing.T) {
	m := NewMetrics()

	m.SetActiveThreads(5)
	if m.ActiveThreads != 5 {
		t.Errorf("Expected ActiveThreads 5, got %d", m.ActiveThreads)
	}

	m.SetCoreThreads(3)
	if m.CoreThreads != 3 {
		t.Errorf("Expected CoreThreads 3, got %d", m.CoreThreads)
	}

	m.SetMaxThreads(10)
	if m.MaxThreads != 10 {
		t.Errorf("Expected MaxThreads 10, got %d", m.MaxThreads)
	}

	m.SetQueueSize(15)
	if m.QueueSize != 15 {
		t.Errorf("Expected QueueSize 15, got %d", m.QueueSize)
	}

	m.SetQueueCapacity(100)
	if m.QueueCapacity != 100 {
		t.Errorf("Expected QueueCapacity 100, got %d", m.QueueCapacity)
	}
}

// TestMetrics_Snapshot 测试快照
func TestMetrics_Snapshot(t *testing.T) {
	m := NewMetrics()

	// 设置一些值
	m.IncrementTasksSubmitted()
	m.IncrementTasksCompleted()
	m.IncrementTasksFailed()
	m.RecordExecutionTime(100 * time.Millisecond)
	m.SetActiveThreads(5)
	m.SetCoreThreads(3)
	m.SetMaxThreads(10)
	m.SetQueueSize(15)
	m.SetQueueCapacity(100)

	// 获取快照
	snapshot := m.Snapshot()

	// 验证快照值
	if snapshot.TasksSubmitted != 1 {
		t.Errorf("Expected TasksSubmitted 1, got %d", snapshot.TasksSubmitted)
	}

	if snapshot.TasksCompleted != 1 {
		t.Errorf("Expected TasksCompleted 1, got %d", snapshot.TasksCompleted)
	}

	if snapshot.TasksFailed != 1 {
		t.Errorf("Expected TasksFailed 1, got %d", snapshot.TasksFailed)
	}

	if snapshot.ActiveThreads != 5 {
		t.Errorf("Expected ActiveThreads 5, got %d", snapshot.ActiveThreads)
	}

	if snapshot.CoreThreads != 3 {
		t.Errorf("Expected CoreThreads 3, got %d", snapshot.CoreThreads)
	}

	if snapshot.MaxThreads != 10 {
		t.Errorf("Expected MaxThreads 10, got %d", snapshot.MaxThreads)
	}

	if snapshot.QueueSize != 15 {
		t.Errorf("Expected QueueSize 15, got %d", snapshot.QueueSize)
	}

	if snapshot.QueueCapacity != 100 {
		t.Errorf("Expected QueueCapacity 100, got %d", snapshot.QueueCapacity)
	}

	// 验证时间
	if snapshot.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}

	if snapshot.SnapshotTime.IsZero() {
		t.Error("Expected SnapshotTime to be set")
	}

	if snapshot.Uptime <= 0 {
		t.Error("Expected Uptime to be positive")
	}
}

// TestMetricsSnapshot_Calculations 测试快照计算
func TestMetricsSnapshot_Calculations(t *testing.T) {
	m := NewMetrics()

	// 设置一些值
	m.IncrementTasksSubmitted()
	m.IncrementTasksSubmitted()
	m.IncrementTasksCompleted()
	m.RecordExecutionTime(100 * time.Millisecond)
	m.SetQueueSize(15)
	m.SetQueueCapacity(100)
	m.SetActiveThreads(5)
	m.SetMaxThreads(10)

	// 等待一小段时间以确保有正的运行时间
	time.Sleep(10 * time.Millisecond)

	snapshot := m.Snapshot()

	// 测试平均执行时间
	expectedAvg := 100 * time.Millisecond
	if snapshot.AvgExecutionTime() != expectedAvg {
		t.Errorf("Expected AvgExecutionTime %v, got %v", expectedAvg, snapshot.AvgExecutionTime())
	}

	// 测试任务吞吐量
	throughput := snapshot.TaskThroughput()
	if throughput <= 0 {
		t.Error("Expected positive throughput")
	}

	// 测试成功率
	expectedSuccessRate := 0.5 // 1 完成 / 2 提交
	if snapshot.SuccessRate() != expectedSuccessRate {
		t.Errorf("Expected SuccessRate %f, got %f", expectedSuccessRate, snapshot.SuccessRate())
	}

	// 测试队列利用率
	expectedQueueUtilization := 0.15 // 15 / 100
	if snapshot.QueueUtilization() != expectedQueueUtilization {
		t.Errorf("Expected QueueUtilization %f, got %f", expectedQueueUtilization, snapshot.QueueUtilization())
	}

	// 测试线程利用率
	expectedThreadUtilization := 0.5 // 5 / 10
	if snapshot.ThreadUtilization() != expectedThreadUtilization {
		t.Errorf("Expected ThreadUtilization %f, got %f", expectedThreadUtilization, snapshot.ThreadUtilization())
	}
}

// TestMetricsSnapshot_EdgeCases 测试快照边界情况
func TestMetricsSnapshot_EdgeCases(t *testing.T) {
	m := NewMetrics()
	snapshot := m.Snapshot()

	// 测试除零情况
	if snapshot.AvgExecutionTime() != 0 {
		t.Error("Expected AvgExecutionTime 0 when no tasks completed")
	}

	if snapshot.TaskThroughput() != 0 {
		t.Error("Expected TaskThroughput 0 when uptime is 0")
	}

	if snapshot.SuccessRate() != 0 {
		t.Error("Expected SuccessRate 0 when no tasks submitted")
	}

	if snapshot.QueueUtilization() != 0 {
		t.Error("Expected QueueUtilization 0 when queue capacity is 0")
	}

	if snapshot.ThreadUtilization() != 0 {
		t.Error("Expected ThreadUtilization 0 when max threads is 0")
	}
}
