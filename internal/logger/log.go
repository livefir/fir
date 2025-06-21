package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// getLogLevel returns the log level from environment variable or defaults to INFO
// If development mode is enabled, it will return DEBUG level unless explicitly overridden
func getLogLevel() slog.Level {
	switch os.Getenv("FIR_LOG_LEVEL") {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		// Check if development mode is enabled globally
		if isDevelopmentMode() {
			return slog.LevelDebug
		}
		return slog.LevelInfo
	}
}

var developmentMode bool

// SetDevelopmentMode enables or disables development mode logging
func SetDevelopmentMode(enabled bool) {
	developmentMode = enabled
	// Recreate the logger with the new log level
	jsonLogger = slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			AddSource:   true,
			Level:       getLogLevel(),
			ReplaceAttr: replace,
		}))
}

// isDevelopmentMode returns true if development mode is enabled
func isDevelopmentMode() bool {
	return developmentMode
}

var jsonLogger = slog.New(slog.NewJSONHandler(
	os.Stdout,
	&slog.HandlerOptions{
		AddSource:   true,
		Level:       getLogLevel(),
		ReplaceAttr: replace,
	}))

func replace(groups []string, a slog.Attr) slog.Attr {
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

func logRecord(level slog.Level, msg string) {
	if !jsonLogger.Enabled(context.Background(), level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip [Callers, logRecord, Infof/Debugf]
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	_ = jsonLogger.Handler().Handle(context.Background(), r)
}

func Logger() *slog.Logger {
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

// Context-aware logging functions for better tracing
func WithContext(ctx context.Context) *slog.Logger {
	return jsonLogger.With()
}

func WithSession(sessionID string) *slog.Logger {
	return jsonLogger.With("session_id", sessionID)
}

func WithRoute(routeID string) *slog.Logger {
	return jsonLogger.With("route_id", routeID)
}

func WithRequest(sessionID, routeID, eventID string) *slog.Logger {
	return jsonLogger.With(
		"session_id", sessionID,
		"route_id", routeID,
		"event_id", eventID,
	)
}

func WithTemplate(templateName string) *slog.Logger {
	return jsonLogger.With("template", templateName)
}

func WithAction(actionName string) *slog.Logger {
	return jsonLogger.With("action", actionName)
}
