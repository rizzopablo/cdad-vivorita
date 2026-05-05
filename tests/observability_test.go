package tests

import (
	"os"
	"testing"
	"vivorita2/src/game"
	"vivorita2/src/observability"
)

func TestPostcondition6_NoLogsWithoutDebug(t *testing.T) {
	os.Unsetenv("DEBUG")

	logsDir := "./logs"
	os.RemoveAll(logsDir)

	err := observability.InitLogging()
	if err != nil {
		t.Fatalf("InitLogging() returned error: %v", err)
	}

	if _, err := os.Stat(logsDir); !os.IsNotExist(err) {
		t.Errorf("Expected ./logs/ to NOT exist without DEBUG=1, but it exists")
	}
}

func TestPostcondition1_InitLogging(t *testing.T) {
	os.Setenv("DEBUG", "1")

	logsDir := "./logs"
	logFile := logsDir + "/vivorita2-debug.log"

	os.RemoveAll(logsDir)

	err := observability.InitLogging()
	if err != nil {
		t.Fatalf("InitLogging() returned error: %v", err)
	}

	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		t.Error("Expected ./logs/ directory to exist after InitLogging() with DEBUG=1")
	}

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Expected ./logs/vivorita2-debug.log to exist after InitLogging() with DEBUG=1")
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected log file to have initial entry, but it's empty")
	}

	os.Unsetenv("DEBUG")
	os.RemoveAll(logsDir)
}

func TestPostcondition2_LogInputEvents(t *testing.T) {
	os.Setenv("DEBUG", "1")

	logsDir := "./logs"
	logFile := logsDir + "/vivorita2-debug.log"

	os.RemoveAll(logsDir)

	err := observability.InitLogging()
	if err != nil {
		t.Fatalf("InitLogging() returned error: %v", err)
	}

	observability.LogEvent("input_raw", map[string]interface{}{
		"char": "d",
	})

	observability.LogEvent("input_converted", map[string]interface{}{
		"direction": "DirRight",
	})

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	if !containsSubstring(logContent, "input_raw") {
		t.Error("Expected log to contain 'input_raw' event")
	}

	if !containsSubstring(logContent, "input_converted") {
		t.Error("Expected log to contain 'input_converted' event")
	}

	if !containsSubstring(logContent, "d") {
		t.Error("Expected log to contain char 'd'")
	}

	if !containsSubstring(logContent, "DirRight") {
		t.Error("Expected log to contain 'DirRight' direction")
	}

	os.Unsetenv("DEBUG")
	os.RemoveAll(logsDir)
}

func TestPostcondition3_LogUpdateEvents(t *testing.T) {
	os.Setenv("DEBUG", "1")

	logsDir := "./logs"
	logFile := logsDir + "/vivorita2-debug.log"

	os.RemoveAll(logsDir)

	err := observability.InitLogging()
	if err != nil {
		t.Fatalf("InitLogging() returned error: %v", err)
	}

	g := game.NewGame()

	g.Update(game.DirRight)

	observability.LogEvent("update", map[string]interface{}{
		"direction":  "DirRight",
		"snake_head": g.Snake().Head(),
		"score":      g.Score(),
		"over":       g.IsOver(),
		"paused":     g.IsPaused(),
	})

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	if !containsSubstring(logContent, "update") {
		t.Error("Expected log to contain 'update' event")
	}

	if !containsSubstring(logContent, "DirRight") {
		t.Error("Expected log to contain direction 'DirRight'")
	}

	if !containsSubstring(logContent, "score") {
		t.Error("Expected log to contain 'score'")
	}

	os.Unsetenv("DEBUG")
	os.RemoveAll(logsDir)
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
