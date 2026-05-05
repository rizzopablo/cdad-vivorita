package tests

import (
	"os"
	"testing"
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

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
