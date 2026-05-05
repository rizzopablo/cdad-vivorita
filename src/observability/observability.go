package observability

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"vivorita2/src/game"
)

var (
	debugEnabled = false
	logFile      *os.File
)

// InitLogging inicializa el sistema de logging si DEBUG=1
func InitLogging() error {
	if os.Getenv("DEBUG") == "1" {
		debugEnabled = true
		logsDir := "./logs"

		err := os.MkdirAll(logsDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create logs directory: %w", err)
		}

		logPath := filepath.Join(logsDir, "vivorita2-debug.log")
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		logFile = file

		// Write initial entry
		logData := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"event":     "init",
			"data":      map[string]interface{}{},
		}
		jsonData, _ := json.Marshal(logData)
		logFile.WriteString(string(jsonData) + "\n")
	} else {
		debugEnabled = false
	}
	return nil
}

// LogEvent escribe un evento al log en formato JSON
func LogEvent(event string, data map[string]interface{}) {
	if debugEnabled && logFile != nil {
		logData := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"event":     event,
			"data":      data,
		}
		jsonData, _ := json.Marshal(logData)
		logFile.WriteString(string(jsonData) + "\n")
	}
}

// SnapshotBoard guarda el estado del tablero en JSON
func SnapshotBoard(g *game.Game, reason string) error {
	if !debugEnabled {
		return nil
	}

	logsDir := "./logs"
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("board-snapshot-%s.json", timestamp)
	path := filepath.Join(logsDir, filename)

	gameState := map[string]interface{}{
		"snake_segments": g.Snake().Segments(),
		"food_position":  g.Food().Position(),
		"score":          g.Score(),
		"high_score":     g.HighScore(),
		"reason":         reason,
	}

	file, err := os.Create(path)
	if err != nil {
		LogEvent("snapshot_error", map[string]interface{}{
			"error":  err.Error(),
			"reason": reason,
		})
		return fmt.Errorf("failed to create snapshot file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(gameState)
	if err != nil {
		LogEvent("snapshot_error", map[string]interface{}{
			"error":  err.Error(),
			"reason": reason,
		})
		return fmt.Errorf("failed to encode snapshot: %w", err)
	}

	return nil
}
