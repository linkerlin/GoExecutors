package config

import (
	"os"
	"runtime"
	"testing"
	"time"
)

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	expectedCorePoolSize := int32(runtime.NumCPU())
	if cfg.CorePoolSize != expectedCorePoolSize {
		t.Errorf("Expected CorePoolSize %d, got %d", expectedCorePoolSize, cfg.CorePoolSize)
	}

	expectedMaxPoolSize := expectedCorePoolSize * 4
	if cfg.MaxPoolSize != expectedMaxPoolSize {
		t.Errorf("Expected MaxPoolSize %d, got %d", expectedMaxPoolSize, cfg.MaxPoolSize)
	}

	if cfg.QueueSize != 1000 {
		t.Errorf("Expected QueueSize 1000, got %d", cfg.QueueSize)
	}

	if cfg.KeepAliveTime != 60*time.Second {
		t.Errorf("Expected KeepAliveTime 60s, got %v", cfg.KeepAliveTime)
	}

	if cfg.RejectPolicy != "abort" {
		t.Errorf("Expected RejectPolicy 'abort', got %s", cfg.RejectPolicy)
	}

	if cfg.ThreadNamePrefix != "goexecutor" {
		t.Errorf("Expected ThreadNamePrefix 'goexecutor', got %s", cfg.ThreadNamePrefix)
	}
}

// TestConfig_LoadFromEnv 测试从环境变量加载配置
func TestConfig_LoadFromEnv(t *testing.T) {
	// 设置环境变量
	os.Setenv("GO_EXECUTOR_CORE_POOL_SIZE", "10")
	os.Setenv("GO_EXECUTOR_MAX_POOL_SIZE", "20")
	os.Setenv("GO_EXECUTOR_QUEUE_SIZE", "500")
	os.Setenv("GO_EXECUTOR_KEEP_ALIVE_TIME", "30s")
	os.Setenv("GO_EXECUTOR_REJECT_POLICY", "discard")
	os.Setenv("GO_EXECUTOR_ENABLE_METRICS", "true")
	os.Setenv("GO_EXECUTOR_ENABLE_LOGGING", "true")
	os.Setenv("GO_EXECUTOR_LOG_LEVEL", "debug")

	defer func() {
		// 清理环境变量
		os.Unsetenv("GO_EXECUTOR_CORE_POOL_SIZE")
		os.Unsetenv("GO_EXECUTOR_MAX_POOL_SIZE")
		os.Unsetenv("GO_EXECUTOR_QUEUE_SIZE")
		os.Unsetenv("GO_EXECUTOR_KEEP_ALIVE_TIME")
		os.Unsetenv("GO_EXECUTOR_REJECT_POLICY")
		os.Unsetenv("GO_EXECUTOR_ENABLE_METRICS")
		os.Unsetenv("GO_EXECUTOR_ENABLE_LOGGING")
		os.Unsetenv("GO_EXECUTOR_LOG_LEVEL")
	}()

	cfg := DefaultConfig()
	cfg.LoadFromEnv()

	if cfg.CorePoolSize != 10 {
		t.Errorf("Expected CorePoolSize 10, got %d", cfg.CorePoolSize)
	}

	if cfg.MaxPoolSize != 20 {
		t.Errorf("Expected MaxPoolSize 20, got %d", cfg.MaxPoolSize)
	}

	if cfg.QueueSize != 500 {
		t.Errorf("Expected QueueSize 500, got %d", cfg.QueueSize)
	}

	if cfg.KeepAliveTime != 30*time.Second {
		t.Errorf("Expected KeepAliveTime 30s, got %v", cfg.KeepAliveTime)
	}

	if cfg.RejectPolicy != "discard" {
		t.Errorf("Expected RejectPolicy 'discard', got %s", cfg.RejectPolicy)
	}

	if !cfg.EnableMetrics {
		t.Error("Expected EnableMetrics true, got false")
	}

	if !cfg.EnableLogging {
		t.Error("Expected EnableLogging true, got false")
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel 'debug', got %s", cfg.LogLevel)
	}
}

// TestConfig_Validate 测试配置验证
func TestConfig_Validate(t *testing.T) {
	cfg := &Config{
		CorePoolSize:     -1,
		MaxPoolSize:      5,
		QueueSize:        -10,
		KeepAliveTime:    -1 * time.Second,
		ThreadNamePrefix: "",
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 检查修正后的值
	if cfg.CorePoolSize != 1 {
		t.Errorf("Expected CorePoolSize 1, got %d", cfg.CorePoolSize)
	}

	if cfg.MaxPoolSize != 5 {
		t.Errorf("Expected MaxPoolSize 5, got %d", cfg.MaxPoolSize)
	}

	if cfg.QueueSize != 0 {
		t.Errorf("Expected QueueSize 0, got %d", cfg.QueueSize)
	}

	if cfg.KeepAliveTime != 60*time.Second {
		t.Errorf("Expected KeepAliveTime 60s, got %v", cfg.KeepAliveTime)
	}

	if cfg.ThreadNamePrefix != "goexecutor" {
		t.Errorf("Expected ThreadNamePrefix 'goexecutor', got %s", cfg.ThreadNamePrefix)
	}
}

// TestConfig_ValidateMaxPoolSize 测试最大池大小验证
func TestConfig_ValidateMaxPoolSize(t *testing.T) {
	cfg := &Config{
		CorePoolSize: 10,
		MaxPoolSize:  5, // 小于核心池大小
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// MaxPoolSize 应该被调整为等于 CorePoolSize
	if cfg.MaxPoolSize != cfg.CorePoolSize {
		t.Errorf("Expected MaxPoolSize %d, got %d", cfg.CorePoolSize, cfg.MaxPoolSize)
	}
}

// TestDefaultGoroutinesNum 测试兼容性函数
func TestDefaultGoroutinesNum(t *testing.T) {
	expected := DefaultConfig().CorePoolSize
	actual := DefaultGoroutinesNum()

	if actual != expected {
		t.Errorf("Expected %d, got %d", expected, actual)
	}
}

// TestLoadConfig 测试兼容性函数
func TestLoadConfig(t *testing.T) {
	// 这个函数应该不会 panic
	LoadConfig()
}
