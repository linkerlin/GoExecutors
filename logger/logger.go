package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger 简单的日志接口
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

// SimpleLogger 简单的日志实现
type SimpleLogger struct {
	level  LogLevel
	logger *log.Logger
	mu     sync.RWMutex
}

// NewSimpleLogger 创建简单日志器
func NewSimpleLogger(level string) *SimpleLogger {
	l := &SimpleLogger{
		level:  parseLogLevel(level),
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
	return l
}

// parseLogLevel 解析日志级别
func parseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// SetLevel 设置日志级别
func (l *SimpleLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Debug 记录调试日志
func (l *SimpleLogger) Debug(args ...interface{}) {
	l.log(DEBUG, args...)
}

// Info 记录信息日志
func (l *SimpleLogger) Info(args ...interface{}) {
	l.log(INFO, args...)
}

// Warn 记录警告日志
func (l *SimpleLogger) Warn(args ...interface{}) {
	l.log(WARN, args...)
}

// Error 记录错误日志
func (l *SimpleLogger) Error(args ...interface{}) {
	l.log(ERROR, args...)
}

// Fatal 记录致命错误日志
func (l *SimpleLogger) Fatal(args ...interface{}) {
	l.log(FATAL, args...)
	os.Exit(1)
}

// Debugf 记录格式化调试日志
func (l *SimpleLogger) Debugf(format string, args ...interface{}) {
	l.logf(DEBUG, format, args...)
}

// Infof 记录格式化信息日志
func (l *SimpleLogger) Infof(format string, args ...interface{}) {
	l.logf(INFO, format, args...)
}

// Warnf 记录格式化警告日志
func (l *SimpleLogger) Warnf(format string, args ...interface{}) {
	l.logf(WARN, format, args...)
}

// Errorf 记录格式化错误日志
func (l *SimpleLogger) Errorf(format string, args ...interface{}) {
	l.logf(ERROR, format, args...)
}

// Fatalf 记录格式化致命错误日志
func (l *SimpleLogger) Fatalf(format string, args ...interface{}) {
	l.logf(FATAL, format, args...)
	os.Exit(1)
}

// log 记录日志
func (l *SimpleLogger) log(level LogLevel, args ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if level >= l.level {
		msg := fmt.Sprint(args...)
		l.logger.Printf("[%s] %s", levelNames[level], msg)
	}
}

// logf 记录格式化日志
func (l *SimpleLogger) logf(level LogLevel, format string, args ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if level >= l.level {
		msg := fmt.Sprintf(format, args...)
		l.logger.Printf("[%s] %s", levelNames[level], msg)
	}
}

// NoOpLogger 空日志实现
type NoOpLogger struct{}

func (n *NoOpLogger) Debug(args ...interface{})                 {}
func (n *NoOpLogger) Info(args ...interface{})                  {}
func (n *NoOpLogger) Warn(args ...interface{})                  {}
func (n *NoOpLogger) Error(args ...interface{})                 {}
func (n *NoOpLogger) Fatal(args ...interface{})                 {}
func (n *NoOpLogger) Debugf(format string, args ...interface{}) {}
func (n *NoOpLogger) Infof(format string, args ...interface{})  {}
func (n *NoOpLogger) Warnf(format string, args ...interface{})  {}
func (n *NoOpLogger) Errorf(format string, args ...interface{}) {}
func (n *NoOpLogger) Fatalf(format string, args ...interface{}) {}

// NewNoOpLogger 创建空日志器
func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

// 全局默认日志器
var defaultLogger Logger = NewNoOpLogger()

// SetDefaultLogger 设置全局默认日志器
func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

// GetDefaultLogger 获取全局默认日志器
func GetDefaultLogger() Logger {
	return defaultLogger
}

// 全局日志函数
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}
