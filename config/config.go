package config

import (
	"os"
	"runtime"
	"strconv"
	"time"
)

// Config 执行器配置
type Config struct {
	// 核心线程数
	CorePoolSize int32 `yaml:"core_pool_size"`
	// 最大线程数
	MaxPoolSize int32 `yaml:"max_pool_size"`
	// 线程空闲时间（秒）
	KeepAliveTime time.Duration `yaml:"keep_alive_time"`
	// 任务队列大小
	QueueSize int `yaml:"queue_size"`
	// 是否允许核心线程超时
	AllowCoreThreadTimeOut bool `yaml:"allow_core_thread_timeout"`
	// 拒绝策略
	RejectPolicy string `yaml:"reject_policy"`
	// 线程名称前缀
	ThreadNamePrefix string `yaml:"thread_name_prefix"`
	// 是否开启性能监控
	EnableMetrics bool `yaml:"enable_metrics"`
	// 指标收集间隔
	MetricsInterval time.Duration `yaml:"metrics_interval"`
	// 是否启用日志
	EnableLogging bool `yaml:"enable_logging"`
	// 日志级别
	LogLevel string `yaml:"log_level"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	cpuNum := int32(runtime.NumCPU())
	return &Config{
		CorePoolSize:           cpuNum,
		MaxPoolSize:            cpuNum * 4,
		KeepAliveTime:          60 * time.Second,
		QueueSize:              1000,
		AllowCoreThreadTimeOut: false,
		RejectPolicy:           "abort",
		ThreadNamePrefix:       "goexecutor",
		EnableMetrics:          false,
		MetricsInterval:        10 * time.Second,
		EnableLogging:          false,
		LogLevel:               "info",
	}
}

// LoadFromEnv 从环境变量加载配置
func (c *Config) LoadFromEnv() {
	if val := os.Getenv("GO_EXECUTOR_CORE_POOL_SIZE"); val != "" {
		if size, err := strconv.ParseInt(val, 10, 32); err == nil {
			c.CorePoolSize = int32(size)
		}
	}

	if val := os.Getenv("GO_EXECUTOR_MAX_POOL_SIZE"); val != "" {
		if size, err := strconv.ParseInt(val, 10, 32); err == nil {
			c.MaxPoolSize = int32(size)
		}
	}

	if val := os.Getenv("GO_EXECUTOR_KEEP_ALIVE_TIME"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			c.KeepAliveTime = duration
		}
	}

	if val := os.Getenv("GO_EXECUTOR_QUEUE_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			c.QueueSize = size
		}
	}

	if val := os.Getenv("GO_EXECUTOR_REJECT_POLICY"); val != "" {
		c.RejectPolicy = val
	}

	if val := os.Getenv("GO_EXECUTOR_ENABLE_METRICS"); val != "" {
		c.EnableMetrics = val == "true"
	}

	if val := os.Getenv("GO_EXECUTOR_ENABLE_LOGGING"); val != "" {
		c.EnableLogging = val == "true"
	}

	if val := os.Getenv("GO_EXECUTOR_LOG_LEVEL"); val != "" {
		c.LogLevel = val
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.CorePoolSize <= 0 {
		c.CorePoolSize = 1
	}
	if c.MaxPoolSize < c.CorePoolSize {
		c.MaxPoolSize = c.CorePoolSize
	}
	if c.QueueSize < 0 {
		c.QueueSize = 0
	}
	if c.KeepAliveTime <= 0 {
		c.KeepAliveTime = 60 * time.Second
	}
	if c.ThreadNamePrefix == "" {
		c.ThreadNamePrefix = "goexecutor"
	}
	return nil
}

// 为了兼容性保留旧的函数
func DefaultGoroutinesNum() int32 {
	return DefaultConfig().CorePoolSize
}

func LoadConfig() {
	// 为了兼容性保留，实际上不再使用
}
