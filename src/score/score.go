package score

import (
	"encoding/json"
	"os"
)

func LoadHighScore(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, nil
	}

	var score int
	if err := json.Unmarshal(data, &score); err != nil {
		return 0, nil
	}

	return score, nil
}

func SaveHighScore(path string, score int) error {
	current, _ := LoadHighScore(path)

	if score <= current {
		return nil
	}

	data, err := json.Marshal(score)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
