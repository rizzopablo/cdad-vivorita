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

func TestPostcondition4_LogRenderEvents(t *testing.T) {
	os.Setenv("DEBUG", "1")

	logsDir := "./logs"
	logFile := logsDir + "/vivorita2-debug.log"

	os.RemoveAll(logsDir)

	err := observability.InitLogging()
	if err != nil {
		t.Fatalf("InitLogging() returned error: %v", err)
	}

	observability.LogEvent("render", map[string]interface{}{
		"timestamp": "2026-05-05T12:00:00Z",
	})

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	if !containsSubstring(logContent, "render") {
		t.Error("Expected log to contain 'render' event")
	}

	if !containsSubstring(logContent, "timestamp") {
		t.Error("Expected log to contain 'timestamp'")
	}

	os.Unsetenv("DEBUG")
	os.RemoveAll(logsDir)
}

func TestPostcondition5_SnapshotOnGameOver(t *testing.T) {
	os.Setenv("DEBUG", "1")

	logsDir := "./logs"
	os.RemoveAll(logsDir)

	err := observability.InitLogging()
	if err != nil {
		t.Fatalf("InitLogging() returned error: %v", err)
	}

	g := game.NewGame()

	for i := 0; i < 100; i++ {
		g.Update(game.DirRight)
		if g.IsOver() {
			break
		}
	}

	err = observability.SnapshotBoard(g, "game_over_wall_collision")
	if err != nil {
		t.Fatalf("SnapshotBoard() returned error: %v", err)
	}

	files, err := os.ReadDir(logsDir)
	if err != nil {
		t.Fatalf("Failed to read logs directory: %v", err)
	}

	snapshotFound := false
	for _, f := range files {
		if containsSubstring(f.Name(), "board-snapshot") {
			snapshotFound = true

			snapshotFile := logsDir + "/" + f.Name()
			content, err := os.ReadFile(snapshotFile)
			if err != nil {
				t.Fatalf("Failed to read snapshot file: %v", err)
			}

			snapshotContent := string(content)

			if !containsSubstring(snapshotContent, "snake_segments") {
				t.Error("Expected snapshot to contain 'snake_segments'")
			}

			if !containsSubstring(snapshotContent, "food_position") {
				t.Error("Expected snapshot to contain 'food_position'")
			}

			if !containsSubstring(snapshotContent, "score") {
				t.Error("Expected snapshot to contain 'score'")
			}

			if !containsSubstring(snapshotContent, "high_score") {
				t.Error("Expected snapshot to contain 'high_score'")
			}

			if !containsSubstring(snapshotContent, "reason") {
				t.Error("Expected snapshot to contain 'reason'")
			}

			break
		}
	}

	if !snapshotFound {
		t.Error("Expected board snapshot file to be created in ./logs/")
	}

	os.Unsetenv("DEBUG")
	os.RemoveAll(logsDir)
}

func TestPostcondition7_IntegrationFullFlow(t *testing.T) {
	os.Setenv("DEBUG", "1")

	logsDir := "./logs"
	logFile := logsDir + "/vivorita2-debug.log"

	os.RemoveAll(logsDir)

	err := observability.InitLogging()
	if err != nil {
		t.Fatalf("InitLogging() returned error: %v", err)
	}

	g := game.NewGame()
	initialHead := g.Snake().Head()

	observability.LogEvent("input_converted", map[string]interface{}{
		"direction": "DirRight",
	})

	g.Update(game.DirRight)

	observability.LogEvent("update", map[string]interface{}{
		"direction":  "DirRight",
		"snake_head": g.Snake().Head(),
		"score":      g.Score(),
	})

	observability.LogEvent("render", map[string]interface{}{
		"timestamp": "2026-05-05T12:00:00Z",
	})

	finalHead := g.Snake().Head()
	if finalHead.X == initialHead.X {
		t.Error("Expected snake head to move after Update(DirRight)")
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	if !containsSubstring(logContent, "input_converted") {
		t.Error("Expected log to contain 'input_converted' event")
	}

	if !containsSubstring(logContent, "update") {
		t.Error("Expected log to contain 'update' event")
	}

	if !containsSubstring(logContent, "render") {
		t.Error("Expected log to contain 'render' event")
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
