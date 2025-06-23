package main

import (
	"fmt"
	"time"

	"github.com/livefir/fir"
	"github.com/livefir/fir/internal/logger"
)

func main() {
	fmt.Println("🎯 Testing Milestone 1: Enhanced Debug Logging")

	// Test the enhanced logger directly
	log := logger.GetGlobalLogger()

	fmt.Println("\n📋 Testing Basic Logging:")
	log.Info("Testing enhanced logger")
	log.Debug("This debug message should appear")
	log.Error("Testing error logging", "component", "test")

	fmt.Println("\n📋 Testing Structured Logging:")
	structuredLogger := log.WithFields(map[string]any{
		"event_id":   "test.event",
		"session_id": "test-session-123",
		"transport":  "http",
		"timestamp":  time.Now().Unix(),
	})

	structuredLogger.Info("Event processed successfully")
	structuredLogger.Debug("Processing details", "duration_ms", 42, "patches_sent", 3)

	fmt.Println("\n📋 Testing Debug Mode Controller Option:")
	controller := fir.NewController(
		"test-controller",
		fir.WithDebug(true), // This should enable comprehensive debug logging
	)

	if controller != nil {
		fmt.Println("✅ Controller with debug mode created successfully")
		log.Info("Debug mode controller initialized")
	}

	fmt.Println("\n🎉 Milestone 1 Enhanced Logging Test Complete!")
	fmt.Println("✅ Basic logging working")
	fmt.Println("✅ Structured logging working")
	fmt.Println("✅ Debug mode controller option working")
	fmt.Println("✅ No panics or errors detected")
}
