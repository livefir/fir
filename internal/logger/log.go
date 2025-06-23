package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Config holds the logger configuration
type Config struct {
	Level       LogLevel
	Format      string // "json" or "text"
	EnableDebug bool
	Output      io.Writer
}

// Logger holds the configured logger instance
type Logger struct {
	slogger   *slog.Logger
	config    Config
	debugMode bool
	mu        sync.RWMutex
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// DefaultConfig returns the default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:       LevelInfo,
		Format:      "json",
		EnableDebug: false,
		Output:      os.Stdout,
	}
}

// NewLogger creates a new logger with the given configuration
func NewLogger(config Config) *Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: replaceAttr,
		Level:       slogLevelFromConfig(config.Level),
	}

	if config.Format == "text" {
		handler = slog.NewTextHandler(config.Output, opts)
	} else {
		handler = slog.NewJSONHandler(config.Output, opts)
	}

	return &Logger{
		slogger:   slog.New(handler),
		config:    config,
		debugMode: config.EnableDebug,
	}
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	defaultLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	once.Do(func() {
		if defaultLogger == nil {
			defaultLogger = NewLogger(DefaultConfig())
		}
	})
	return defaultLogger
}

// EnableDebugMode enables debug mode for enhanced logging
func (l *Logger) EnableDebugMode() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugMode = true
}

// DisableDebugMode disables debug mode
func (l *Logger) DisableDebugMode() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugMode = false
}

// IsDebugEnabled returns whether debug mode is enabled
func (l *Logger) IsDebugEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.debugMode
}

// WithFields creates a logger with additional fields
func (l *Logger) WithFields(fields map[string]any) *Logger {
	var attrs []any
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	return &Logger{
		slogger:   l.slogger.With(attrs...),
		config:    l.config,
		debugMode: l.debugMode,
	}
}

// Info logs an info level message
func (l *Logger) Info(msg string, fields ...any) {
	l.logRecord(slog.LevelInfo, msg, fields...)
}

// Debug logs a debug level message (only if debug mode is enabled)
func (l *Logger) Debug(msg string, fields ...any) {
	if l.IsDebugEnabled() {
		l.logRecord(slog.LevelDebug, msg, fields...)
	}
}

// Warn logs a warning level message
func (l *Logger) Warn(msg string, fields ...any) {
	l.logRecord(slog.LevelWarn, msg, fields...)
}

// Error logs an error level message
func (l *Logger) Error(msg string, fields ...any) {
	l.logRecord(slog.LevelError, msg, fields...)
}

// Infof logs an info level message with formatting
func (l *Logger) Infof(format string, args ...any) {
	l.logRecord(slog.LevelInfo, fmt.Sprintf(format, args...))
}

// Debugf logs a debug level message with formatting (only if debug mode is enabled)
func (l *Logger) Debugf(format string, args ...any) {
	if l.IsDebugEnabled() {
		l.logRecord(slog.LevelDebug, fmt.Sprintf(format, args...))
	}
}

// Warnf logs a warning level message with formatting
func (l *Logger) Warnf(format string, args ...any) {
	l.logRecord(slog.LevelWarn, fmt.Sprintf(format, args...))
}

// Errorf logs an error level message with formatting
func (l *Logger) Errorf(format string, args ...any) {
	l.logRecord(slog.LevelError, fmt.Sprintf(format, args...))
}

// logRecord handles the actual logging with proper caller information
func (l *Logger) logRecord(level slog.Level, msg string, fields ...any) {
	if !l.slogger.Enabled(context.Background(), level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip [Callers, logRecord, Info/Debug/etc]
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])

	// Add any additional fields
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			r.Add(fields[i].(string), fields[i+1])
		}
	}

	_ = l.slogger.Handler().Handle(context.Background(), r)
}

// replaceAttr customizes log record attributes
func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	// Remove the directory from the source's filename.
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		source.Function = ""
		source.File = filepath.Base(source.File)
	}

	// format time RFC3339 in utc
	if a.Key == slog.TimeKey {
		a.Value = slog.StringValue(a.Value.Time().UTC().Format(time.RFC3339))
	}

	return a
}

// slogLevelFromConfig converts our LogLevel to slog.Level
func slogLevelFromConfig(level LogLevel) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Legacy functions for backward compatibility
var jsonLogger = slog.New(slog.NewJSONHandler(
	os.Stdout,
	&slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: replaceAttr,
	}))

func logRecord(level slog.Level, msg string) {
	if !jsonLogger.Enabled(context.Background(), level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip [Callers, logRecord, Infof/Debugf]
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	_ = jsonLogger.Handler().Handle(context.Background(), r)
}

func LegacyLogger() *slog.Logger {
	return jsonLogger
}

func Infof(format string, args ...any) {
	logRecord(slog.LevelInfo, fmt.Sprintf(format, args...))
}

func Debugf(format string, args ...any) {
	logRecord(slog.LevelDebug, fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...any) {
	logRecord(slog.LevelError, fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...any) {
	logRecord(slog.LevelWarn, fmt.Sprintf(format, args...))
}
