package input

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
	DirPause
	DirQuit
)

// ReadDirectionNonBlocking reads a direction key press without blocking
func ReadDirectionNonBlocking() (Direction, error) {
	if runtime.GOOS == "windows" {
		return readDirectionWindows()
	}
	return readDirectionUnix()
}

// On Unix systems, we can use a non-blocking approach
func readDirectionUnix() (Direction, error) {
	// Try to read a character with a short timeout
	cmd := exec.Command("bash", "-c", "read -n 1 -t 0.01 char && echo -n $char || exit 1")
	output, err := cmd.Output()

	if err != nil {
		return DirUp, nil // Not really an error, just no input
	}

	if len(output) == 0 {
		return DirUp, nil // No input
	}

	char := string(output[0])

	// Map inputs to directions
	switch strings.ToLower(char) {
	case "w":
		return DirUp, nil
	case "s":
		return DirDown, nil
	case "a":
		return DirLeft, nil
	case "d":
		return DirRight, nil
	case "q":
		return DirQuit, nil
	case "p":
		return DirPause, nil
	default:
		return DirUp, nil
	}
}

// On Windows, we'll implement a simpler version
func readDirectionWindows() (Direction, error) {
	// For now, just return a default
	return DirUp, nil
}

// ReadDirection is the original blocking version
func ReadDirection() (Direction, error) {
	// For a full implementation, we'd need a more sophisticated approach
	// For now we'll just simulate reading from stdin with a timeout
	done := make(chan Direction, 1)
	go func() {
		var char rune
		_, err := os.Stdin.Read([]byte{byte(char)})
		if err != nil {
			return
		}

		switch strings.ToLower(string(char)) {
		case "w":
			done <- DirUp
		case "s":
			done <- DirDown
		case "a":
			done <- DirLeft
		case "d":
			done <- DirRight
		case "q":
			done <- DirQuit
		case "p":
			done <- DirPause
		default:
			done <- DirUp
		}
	}()

	select {
	case dir := <-done:
		return dir, nil
	case <-time.After(10 * time.Millisecond): // Short timeout to mimic non-blocking
		return DirUp, nil // No input
	}
}
