package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger 日志接口
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

// DefaultLogger 默认日志实现
type DefaultLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	warnLogger  *log.Logger
	debugLogger *log.Logger
	logFile     *os.File
}

// NewLogger 创建新的日志记录器
func NewLogger() Logger {
	logger := &DefaultLogger{}
	
	// 确保日志目录存在
	logDir := "./data/logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("创建日志目录失败: %v", err)
	}

	// 创建日志文件
	logFile := filepath.Join(logDir, fmt.Sprintf("crawler_%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("打开日志文件失败: %v", err)
		// 如果无法创建文件，则只输出到控制台
		logger.infoLogger = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
		logger.errorLogger = log.New(os.Stderr, "[ERROR] ", log.LstdFlags)
		logger.warnLogger = log.New(os.Stdout, "[WARN] ", log.LstdFlags)
		logger.debugLogger = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags)
	} else {
		logger.logFile = file
		// 创建多个输出目标（控制台和文件）
		logger.infoLogger = log.New(NewMultiWriter(os.Stdout, file), "[INFO] ", log.LstdFlags)
		logger.errorLogger = log.New(NewMultiWriter(os.Stderr, file), "[ERROR] ", log.LstdFlags)
		logger.warnLogger = log.New(NewMultiWriter(os.Stdout, file), "[WARN] ", log.LstdFlags)
		logger.debugLogger = log.New(NewMultiWriter(os.Stdout, file), "[DEBUG] ", log.LstdFlags)
	}

	return logger
}

// Info 记录信息日志
func (l *DefaultLogger) Info(msg string, keysAndValues ...interface{}) {
	formatted := l.formatMessage(msg, keysAndValues...)
	l.infoLogger.Print(formatted)
}

// Error 记录错误日志
func (l *DefaultLogger) Error(msg string, keysAndValues ...interface{}) {
	formatted := l.formatMessage(msg, keysAndValues...)
	l.errorLogger.Print(formatted)
}

// Warn 记录警告日志
func (l *DefaultLogger) Warn(msg string, keysAndValues ...interface{}) {
	formatted := l.formatMessage(msg, keysAndValues...)
	l.warnLogger.Print(formatted)
}

// Debug 记录调试日志
func (l *DefaultLogger) Debug(msg string, keysAndValues ...interface{}) {
	formatted := l.formatMessage(msg, keysAndValues...)
	l.debugLogger.Print(formatted)
}

// formatMessage 格式化消息
func (l *DefaultLogger) formatMessage(msg string, keysAndValues ...interface{}) string {
	if len(keysAndValues) == 0 {
		return msg
	}

	formatted := msg
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := keysAndValues[i]
			value := keysAndValues[i+1]
			formatted += fmt.Sprintf(" %v=%v", key, value)
		}
	}
	return formatted
}

// Close 关闭日志文件
func (l *DefaultLogger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// MultiWriter 多重写入器
type MultiWriter struct {
	writers []interface{ Write([]byte) (int, error) }
}

// NewMultiWriter 创建多重写入器
func NewMultiWriter(writers ...interface{ Write([]byte) (int, error) }) *MultiWriter {
	return &MultiWriter{writers: writers}
}

// Write 写入到所有写入器
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}
	return len(p), nil
}

// FileLogger 文件日志记录器
type FileLogger struct {
	*DefaultLogger
	filename string
}

// NewFileLogger 创建文件日志记录器
func NewFileLogger(filename string) (Logger, error) {
	// 确保日志目录存在
	logDir := filepath.Dir(filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 打开日志文件
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}

	logger := &FileLogger{
		DefaultLogger: &DefaultLogger{
			infoLogger:  log.New(NewMultiWriter(os.Stdout, file), "[INFO] ", log.LstdFlags),
			errorLogger: log.New(NewMultiWriter(os.Stderr, file), "[ERROR] ", log.LstdFlags),
			warnLogger:  log.New(NewMultiWriter(os.Stdout, file), "[WARN] ", log.LstdFlags),
			debugLogger: log.New(NewMultiWriter(os.Stdout, file), "[DEBUG] ", log.LstdFlags),
			logFile:     file,
		},
		filename: filename,
	}

	return logger, nil
}

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// LevelLogger 带级别的日志记录器
type LevelLogger struct {
	*DefaultLogger
	level LogLevel
}

// NewLevelLogger 创建带级别的日志记录器
func NewLevelLogger(level LogLevel) Logger {
	baseLogger := NewLogger().(*DefaultLogger)
	return &LevelLogger{
		DefaultLogger: baseLogger,
		level:         level,
	}
}

// Debug 记录调试日志（带级别检查）
func (ll *LevelLogger) Debug(msg string, keysAndValues ...interface{}) {
	if ll.level <= LogLevelDebug {
		ll.DefaultLogger.Debug(msg, keysAndValues...)
	}
}

// Info 记录信息日志（带级别检查）
func (ll *LevelLogger) Info(msg string, keysAndValues ...interface{}) {
	if ll.level <= LogLevelInfo {
		ll.DefaultLogger.Info(msg, keysAndValues...)
	}
}

// Warn 记录警告日志（带级别检查）
func (ll *LevelLogger) Warn(msg string, keysAndValues ...interface{}) {
	if ll.level <= LogLevelWarn {
		ll.DefaultLogger.Warn(msg, keysAndValues...)
	}
}

// Error 记录错误日志（带级别检查）
func (ll *LevelLogger) Error(msg string, keysAndValues ...interface{}) {
	if ll.level <= LogLevelError {
		ll.DefaultLogger.Error(msg, keysAndValues...)
	}
}

// ParseLogLevel 解析日志级别
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
} 