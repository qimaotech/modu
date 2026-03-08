package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel 日志级别
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Logger 日志记录器
type Logger struct {
	infoLogger  *lumberjack.Logger
	warnLogger  *lumberjack.Logger
	errorLogger *lumberjack.Logger
	debugLogger *lumberjack.Logger
}

// New 创建日志记录器
func New() *Logger {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}

	logDir := filepath.Join(homeDir, ".modu", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create log directory: %v\n", err)
	}

	return &Logger{
		infoLogger: &lumberjack.Logger{
			Filename:   filepath.Join(logDir, "info.log"),
			MaxSize:    10, // MB
			MaxBackups: 5,
			MaxAge:     30, // days
			Compress:   true,
		},
		warnLogger: &lumberjack.Logger{
			Filename:   filepath.Join(logDir, "warn.log"),
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		},
		errorLogger: &lumberjack.Logger{
			Filename:   filepath.Join(logDir, "error.log"),
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		},
		debugLogger: &lumberjack.Logger{
			Filename:   filepath.Join(logDir, "debug.log"),
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		},
	}
}

func (l *Logger) formatMessage(level, format string, args ...interface{}) string {
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("[%s] [%s] %s\n", timeStr, level, msg)
}

// Debug 记录调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.debugLogger != nil {
		if _, err := l.debugLogger.Write([]byte(l.formatMessage("DEBUG", format, args...))); err != nil {
			fmt.Fprintf(os.Stderr, "log error: %v\n", err)
		}
	}
}

// Info 记录信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	if l.infoLogger != nil {
		if _, err := l.infoLogger.Write([]byte(l.formatMessage("INFO", format, args...))); err != nil {
			fmt.Fprintf(os.Stderr, "log error: %v\n", err)
		}
	}
}

// Warn 记录警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.warnLogger != nil {
		if _, err := l.warnLogger.Write([]byte(l.formatMessage("WARN", format, args...))); err != nil {
			fmt.Fprintf(os.Stderr, "log error: %v\n", err)
		}
	}
}

// Error 记录错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	if l.errorLogger != nil {
		if _, err := l.errorLogger.Write([]byte(l.formatMessage("ERROR", format, args...))); err != nil {
			fmt.Fprintf(os.Stderr, "log error: %v\n", err)
		}
	}
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	var errs []error
	if l.infoLogger != nil {
		if err := l.infoLogger.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if l.warnLogger != nil {
		if err := l.warnLogger.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if l.errorLogger != nil {
		if err := l.errorLogger.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if l.debugLogger != nil {
		if err := l.debugLogger.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Global 全局日志记录器
var global = New()

// Debug 记录调试日志（全局）
func Debug(format string, args ...interface{}) {
	global.Debug(format, args...)
}

// Info 记录信息日志（全局）
func Info(format string, args ...interface{}) {
	global.Info(format, args...)
}

// Warn 记录警告日志（全局）
func Warn(format string, args ...interface{}) {
	global.Warn(format, args...)
}

// Error 记录错误日志（全局）
func Error(format string, args ...interface{}) {
	global.Error(format, args...)
}

// Close 关闭全局日志记录器
func Close() error {
	return global.Close()
}
