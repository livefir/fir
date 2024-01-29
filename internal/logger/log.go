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

var jsonLogger = slog.New(slog.NewJSONHandler(
	os.Stdout,
	&slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: replace,
	}))

func replace(groups []string, a slog.Attr) slog.Attr {
	// Remove the directory from the source's filename.
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		source.Function = ""
		source.File = filepath.Base(source.File)
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
