package observability

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"vivorita2/src/game"
)

const (
	logsDir     = "./logs"
	logFilename = "vivorita2-debug.log"
)

var (
	debugEnabled = false
	logFile      *os.File
)

// InitLogging initializes the debug logging system when DEBUG=1.
// Creates the logs directory and opens the log file for appending.
func InitLogging() error {
	if os.Getenv("DEBUG") == "1" {
		debugEnabled = true

		if err := os.MkdirAll(logsDir, 0755); err != nil {
			return fmt.Errorf("failed to create logs directory: %w", err)
		}

		file, err := os.OpenFile(filepath.Join(logsDir, logFilename), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		logFile = file

		writeJSONLog(logFile, "init", nil)
	} else {
		debugEnabled = false
	}
	return nil
}

// LogEvent writes a structured event to the debug log in JSON format.
// Each entry includes a RFC3339 timestamp, the event name, and optional data.
func LogEvent(event string, data map[string]interface{}) {
	if debugEnabled && logFile != nil {
		writeJSONLog(logFile, event, data)
	}
}

// SnapshotBoard saves the current game board state as a JSON file in the logs directory.
// The filename includes a timestamp for uniqueness. Returns an error if the file
// cannot be created or the state cannot be serialized.
func SnapshotBoard(g *game.Game, reason string) error {
	if !debugEnabled {
		return nil
	}

	timestamp := time.Now().Format("20060102-150405")
	path := filepath.Join(logsDir, fmt.Sprintf("board-snapshot-%s.json", timestamp))

	gameState := map[string]interface{}{
		"snake_segments": g.Snake().Segments(),
		"food_position":  g.Food().Position(),
		"score":          g.Score(),
		"high_score":     g.HighScore(),
		"reason":         reason,
	}

	file, err := os.Create(path)
	if err != nil {
		logSnapshotError(err, reason)
		return fmt.Errorf("failed to create snapshot file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(gameState); err != nil {
		logSnapshotError(err, reason)
		return fmt.Errorf("failed to encode snapshot: %w", err)
	}

	return nil
}

// writeJSONLog marshals the event data to JSON and writes it to the log file.
func writeJSONLog(file *os.File, event string, data map[string]interface{}) {
	entry := struct {
		Timestamp string                 `json:"timestamp"`
		Event     string                 `json:"event"`
		Data      map[string]interface{} `json:"data"`
	}{
		Timestamp: time.Now().Format(time.RFC3339),
		Event:     event,
		Data:      data,
	}
	if jsonData, err := json.Marshal(entry); err == nil {
		file.Write(append(jsonData, '\n'))
	}
}

// logSnapshotError logs a snapshot-related error with the given reason.
func logSnapshotError(err error, reason string) {
	LogEvent("snapshot_error", map[string]interface{}{
		"error":  err.Error(),
		"reason": reason,
	})
}
