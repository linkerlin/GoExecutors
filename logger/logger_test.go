package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

// TestSimpleLogger_Basic 基本日志测试
func TestSimpleLogger_Basic(t *testing.T) {
	var buf bytes.Buffer
	logger := &SimpleLogger{
		level:  INFO,
		logger: log.New(&buf, "", 0),
	}

	// 测试不同级别的日志
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()

	// DEBUG 级别的日志不应该输出
	if strings.Contains(output, "debug message") {
		t.Error("DEBUG message should not be logged at INFO level")
	}

	// INFO 级别及以上的日志应该输出
	if !strings.Contains(output, "info message") {
		t.Error("INFO message should be logged")
	}

	if !strings.Contains(output, "warn message") {
		t.Error("WARN message should be logged")
	}

	if !strings.Contains(output, "error message") {
		t.Error("ERROR message should be logged")
	}
}

// TestSimpleLogger_Levels 测试不同日志级别
func TestSimpleLogger_Levels(t *testing.T) {
	tests := []struct {
		level    LogLevel
		message  string
		method   func(*SimpleLogger)
		expected bool
	}{
		{DEBUG, "debug", func(l *SimpleLogger) { l.Debug("test") }, true},
		{DEBUG, "info", func(l *SimpleLogger) { l.Info("test") }, true},
		{DEBUG, "warn", func(l *SimpleLogger) { l.Warn("test") }, true},
		{DEBUG, "error", func(l *SimpleLogger) { l.Error("test") }, true},
		{INFO, "debug", func(l *SimpleLogger) { l.Debug("test") }, false},
		{INFO, "info", func(l *SimpleLogger) { l.Info("test") }, true},
		{INFO, "warn", func(l *SimpleLogger) { l.Warn("test") }, true},
		{INFO, "error", func(l *SimpleLogger) { l.Error("test") }, true},
		{WARN, "debug", func(l *SimpleLogger) { l.Debug("test") }, false},
		{WARN, "info", func(l *SimpleLogger) { l.Info("test") }, false},
		{WARN, "warn", func(l *SimpleLogger) { l.Warn("test") }, true},
		{WARN, "error", func(l *SimpleLogger) { l.Error("test") }, true},
		{ERROR, "debug", func(l *SimpleLogger) { l.Debug("test") }, false},
		{ERROR, "info", func(l *SimpleLogger) { l.Info("test") }, false},
		{ERROR, "warn", func(l *SimpleLogger) { l.Warn("test") }, false},
		{ERROR, "error", func(l *SimpleLogger) { l.Error("test") }, true},
	}

	for _, tt := range tests {
		var buf bytes.Buffer
		logger := &SimpleLogger{
			level:  tt.level,
			logger: log.New(&buf, "", 0),
		}

		tt.method(logger)
		output := buf.String()

		if tt.expected && !strings.Contains(output, "test") {
			t.Errorf("Expected '%s' to be logged at level %s", tt.message, levelNames[tt.level])
		}

		if !tt.expected && strings.Contains(output, "test") {
			t.Errorf("Did not expect '%s' to be logged at level %s", tt.message, levelNames[tt.level])
		}
	}
}

// TestSimpleLogger_Formatted 测试格式化日志
func TestSimpleLogger_Formatted(t *testing.T) {
	var buf bytes.Buffer
	logger := &SimpleLogger{
		level:  INFO,
		logger: log.New(&buf, "", 0),
	}

	logger.Infof("Hello %s, you are %d years old", "Alice", 30)

	output := buf.String()
	if !strings.Contains(output, "Hello Alice, you are 30 years old") {
		t.Errorf("Expected formatted message, got: %s", output)
	}
}

// TestParseLogLevel 测试日志级别解析
func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"DEBUG", DEBUG},
		{"debug", DEBUG},
		{"INFO", INFO},
		{"info", INFO},
		{"WARN", WARN},
		{"warn", WARN},
		{"ERROR", ERROR},
		{"error", ERROR},
		{"FATAL", FATAL},
		{"fatal", FATAL},
		{"invalid", INFO}, // 默认值
		{"", INFO},        // 默认值
	}

	for _, tt := range tests {
		result := parseLogLevel(tt.input)
		if result != tt.expected {
			t.Errorf("parseLogLevel(%s) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

// TestNewSimpleLogger 测试简单日志器创建
func TestNewSimpleLogger(t *testing.T) {
	logger := NewSimpleLogger("debug")

	if logger.level != DEBUG {
		t.Errorf("Expected level DEBUG, got %v", logger.level)
	}

	if logger.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

// TestSimpleLogger_SetLevel 测试设置日志级别
func TestSimpleLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := &SimpleLogger{
		level:  INFO,
		logger: log.New(&buf, "", 0),
	}

	// 初始级别为 INFO，DEBUG 不应该输出
	logger.Debug("debug message")
	if strings.Contains(buf.String(), "debug message") {
		t.Error("DEBUG message should not be logged at INFO level")
	}

	// 设置级别为 DEBUG
	logger.SetLevel(DEBUG)
	buf.Reset()

	logger.Debug("debug message after level change")
	if !strings.Contains(buf.String(), "debug message after level change") {
		t.Error("DEBUG message should be logged at DEBUG level")
	}
}

// TestNoOpLogger 测试空日志器
func TestNoOpLogger(t *testing.T) {
	logger := NewNoOpLogger()

	// 所有方法都应该不做任何事情，不应该 panic
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
	logger.Debugf("test %s", "formatted")
	logger.Infof("test %s", "formatted")
	logger.Warnf("test %s", "formatted")
	logger.Errorf("test %s", "formatted")

	// 如果没有 panic，测试通过
}

// TestDefaultLogger 测试默认日志器
func TestDefaultLogger(t *testing.T) {
	// 保存原始的默认日志器
	originalLogger := GetDefaultLogger()

	// 创建一个新的日志器
	var buf bytes.Buffer
	newLogger := &SimpleLogger{
		level:  INFO,
		logger: log.New(&buf, "", 0),
	}

	// 设置新的默认日志器
	SetDefaultLogger(newLogger)

	// 使用全局函数
	Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("Expected message to be logged by default logger")
	}

	// 恢复原始的默认日志器
	SetDefaultLogger(originalLogger)
}

// TestDefaultLogger_Formatted 测试默认日志器格式化
func TestDefaultLogger_Formatted(t *testing.T) {
	// 保存原始的默认日志器
	originalLogger := GetDefaultLogger()

	// 创建一个新的日志器
	var buf bytes.Buffer
	newLogger := &SimpleLogger{
		level:  INFO,
		logger: log.New(&buf, "", 0),
	}

	// 设置新的默认日志器
	SetDefaultLogger(newLogger)

	// 使用全局格式化函数
	Infof("Hello %s, you are %d years old", "Bob", 25)

	output := buf.String()
	if !strings.Contains(output, "Hello Bob, you are 25 years old") {
		t.Errorf("Expected formatted message, got: %s", output)
	}

	// 恢复原始的默认日志器
	SetDefaultLogger(originalLogger)
}
