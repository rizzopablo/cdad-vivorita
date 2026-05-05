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

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
