package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	config := Config{
		Level:       LevelDebug,
		Format:      "json",
		EnableDebug: true,
		Output:      os.Stdout,
	}

	logger := NewLogger(config)
	if logger == nil {
		t.Fatal("Expected logger to be created, got nil")
	}

	if !logger.IsDebugEnabled() {
		t.Error("Expected debug mode to be enabled")
	}
}

func TestLoggerDebugMode(t *testing.T) {
	logger := NewLogger(DefaultConfig())

	// Initially debug should be disabled
	if logger.IsDebugEnabled() {
		t.Error("Expected debug mode to be disabled by default")
	}

	// Enable debug mode
	logger.EnableDebugMode()
	if !logger.IsDebugEnabled() {
		t.Error("Expected debug mode to be enabled after EnableDebugMode()")
	}

	// Disable debug mode
	logger.DisableDebugMode()
	if logger.IsDebugEnabled() {
		t.Error("Expected debug mode to be disabled after DisableDebugMode()")
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:       LevelDebug,
		Format:      "json",
		EnableDebug: true,
		Output:      &buf,
	}

	logger := NewLogger(config)
	loggerWithFields := logger.WithFields(map[string]any{
		"user_id": "123",
		"action":  "test",
	})

	loggerWithFields.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "user_id") {
		t.Error("Expected log output to contain user_id field")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Expected log output to contain the message")
	}
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:       LevelDebug,
		Format:      "json",
		EnableDebug: true,
		Output:      &buf,
	}

	logger := NewLogger(config)

	tests := []struct {
		name  string
		logFn func(string, ...any)
		msg   string
		level string
	}{
		{"Info", logger.Info, "info message", "INFO"},
		{"Debug", logger.Debug, "debug message", "DEBUG"},
		{"Warn", logger.Warn, "warn message", "WARN"},
		{"Error", logger.Error, "error message", "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFn(tt.msg)

			output := buf.String()
			if !strings.Contains(output, tt.msg) {
				t.Errorf("Expected log output to contain message: %s", tt.msg)
			}
			if !strings.Contains(output, tt.level) {
				t.Errorf("Expected log output to contain level: %s", tt.level)
			}
		})
	}
}

func TestLoggerFormattedMethods(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:       LevelDebug,
		Format:      "json",
		EnableDebug: true,
		Output:      &buf,
	}

	logger := NewLogger(config)

	tests := []struct {
		name   string
		logFn  func(string, ...any)
		format string
		args   []any
		expect string
	}{
		{"Infof", logger.Infof, "user %s logged in", []any{"john"}, "user john logged in"},
		{"Debugf", logger.Debugf, "debug value: %d", []any{42}, "debug value: 42"},
		{"Warnf", logger.Warnf, "warning: %s", []any{"test"}, "warning: test"},
		{"Errorf", logger.Errorf, "error: %v", []any{"failed"}, "error: failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFn(tt.format, tt.args...)

			output := buf.String()
			if !strings.Contains(output, tt.expect) {
				t.Errorf("Expected log output to contain: %s, got: %s", tt.expect, output)
			}
		})
	}
}

func TestLoggerDebugModeFiltering(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:       LevelDebug,
		Format:      "json",
		EnableDebug: false, // Debug disabled
		Output:      &buf,
	}

	logger := NewLogger(config)

	// Debug message should not be logged when debug mode is disabled
	logger.Debug("debug message")
	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("Expected debug message to be filtered out when debug mode is disabled")
	}

	// Enable debug mode
	logger.EnableDebugMode()
	buf.Reset()

	// Debug message should now be logged
	logger.Debug("debug message enabled")
	output = buf.String()
	if !strings.Contains(output, "debug message enabled") {
		t.Error("Expected debug message to be logged when debug mode is enabled")
	}
}

func TestLoggerJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:       LevelInfo,
		Format:      "json",
		EnableDebug: false,
		Output:      &buf,
	}

	logger := NewLogger(config)
	logger.Info("test message", "key", "value")

	output := buf.String()

	// Verify it's valid JSON
	var logEntry map[string]any
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}

	// Check required fields
	if logEntry["msg"] != "test message" {
		t.Errorf("Expected msg field to be 'test message', got: %v", logEntry["msg"])
	}

	if logEntry["key"] != "value" {
		t.Errorf("Expected key field to be 'value', got: %v", logEntry["key"])
	}
}

func TestGlobalLogger(t *testing.T) {
	// Test that GetGlobalLogger returns a logger
	logger := GetGlobalLogger()
	if logger == nil {
		t.Fatal("Expected global logger to be available")
	}

	// Test setting a custom global logger
	customConfig := Config{
		Level:       LevelWarn,
		Format:      "json",
		EnableDebug: true,
		Output:      os.Stdout,
	}
	customLogger := NewLogger(customConfig)
	SetGlobalLogger(customLogger)

	retrievedLogger := GetGlobalLogger()
	if retrievedLogger != customLogger {
		t.Error("Expected retrieved logger to be the same as the set custom logger")
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that legacy functions still work - they should not panic
	Infof("test info: %s", "value")
	Debugf("test debug: %d", 42)
	Errorf("test error: %v", "error")
	Warnf("test warn: %s", "warning")

	// Test legacy Logger function
	legacyLogger := LegacyLogger()
	if legacyLogger == nil {
		t.Error("Expected legacy logger to be available")
	}
}

func TestSlogLevelConversion(t *testing.T) {
	tests := []struct {
		input    LogLevel
		expected slog.Level
	}{
		{LevelDebug, slog.LevelDebug},
		{LevelInfo, slog.LevelInfo},
		{LevelWarn, slog.LevelWarn},
		{LevelError, slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.input)), func(t *testing.T) {
			result := slogLevelFromConfig(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != LevelInfo {
		t.Errorf("Expected default level to be LevelInfo, got %v", config.Level)
	}

	if config.Format != "json" {
		t.Errorf("Expected default format to be 'json', got %s", config.Format)
	}

	if config.EnableDebug {
		t.Error("Expected default EnableDebug to be false")
	}

	if config.Output != os.Stdout {
		t.Error("Expected default output to be os.Stdout")
	}
}
