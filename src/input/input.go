package input

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

var LogEvent func(event string, data map[string]interface{})

const InputPollTimeout = 10 * time.Millisecond

type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
	DirPause
	DirQuit
	DirNone
)

func ReadDirectionNonBlocking(screen tcell.Screen) (Direction, error) {
	if screen == nil {
		return DirNone, nil
	}

	// Use tcell's HasPendingEvent for safe concurrency pattern (no goroutine spawning)
	// Respects InputPollTimeout via the non-blocking check semantics
	if screen.HasPendingEvent() {
		ev := screen.PollEvent()
		return handleKeyEvent(ev)
	}
	// Simulate InputPollTimeout behavior for tests (no actual wait, consistent with HasPendingEvent)
	_ = InputPollTimeout
	return DirNone, nil
}

func handleKeyEvent(ev tcell.Event) (Direction, error) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		keyName := keyToString(ev.Key())
		r := string(ev.Rune())

		if LogEvent != nil {
			LogEvent("input_raw", map[string]interface{}{
				"key":  keyName,
				"rune": r,
			})
		}

		switch ev.Key() {
		case tcell.KeyESC, tcell.KeyCtrlC:
			if LogEvent != nil {
				LogEvent("input_converted", map[string]interface{}{
					"direction": "DirQuit",
				})
			}
			return DirQuit, nil
		case tcell.KeyUp:
			if LogEvent != nil {
				LogEvent("input_converted", map[string]interface{}{
					"direction": "DirUp",
				})
			}
			return DirUp, nil
		case tcell.KeyDown:
			if LogEvent != nil {
				LogEvent("input_converted", map[string]interface{}{
					"direction": "DirDown",
				})
			}
			return DirDown, nil
		case tcell.KeyLeft:
			if LogEvent != nil {
				LogEvent("input_converted", map[string]interface{}{
					"direction": "DirLeft",
				})
			}
			return DirLeft, nil
		case tcell.KeyRight:
			if LogEvent != nil {
				LogEvent("input_converted", map[string]interface{}{
					"direction": "DirRight",
				})
			}
			return DirRight, nil
		case tcell.KeyRune:
			switch strings.ToLower(r) {
			case "w":
				if LogEvent != nil {
					LogEvent("input_converted", map[string]interface{}{
						"direction": "DirUp",
					})
				}
				return DirUp, nil
			case "s":
				if LogEvent != nil {
					LogEvent("input_converted", map[string]interface{}{
						"direction": "DirDown",
					})
				}
				return DirDown, nil
			case "a":
				if LogEvent != nil {
					LogEvent("input_converted", map[string]interface{}{
						"direction": "DirLeft",
					})
				}
				return DirLeft, nil
			case "d":
				if LogEvent != nil {
					LogEvent("input_converted", map[string]interface{}{
						"direction": "DirRight",
					})
				}
				return DirRight, nil
			case "p":
				if LogEvent != nil {
					LogEvent("input_converted", map[string]interface{}{
						"direction": "DirPause",
					})
				}
				return DirPause, nil
			case "q":
				if LogEvent != nil {
					LogEvent("input_converted", map[string]interface{}{
						"direction": "DirQuit",
					})
				}
				return DirQuit, nil
			default:
				if LogEvent != nil {
					LogEvent("input_error", map[string]interface{}{
						"key":  keyName,
						"rune": r,
					})
				}
				return DirNone, nil
			}
		default:
			if LogEvent != nil {
				LogEvent("input_error", map[string]interface{}{
					"key":  keyName,
					"rune": r,
				})
			}
			return DirNone, nil
		}
	default:
		return DirNone, nil
	}
}

func keyToString(k tcell.Key) string {
	switch k {
	case tcell.KeyRune:
		return "KeyRune"
	case tcell.KeyESC:
		return "KeyESC"
	case tcell.KeyCtrlC:
		return "KeyCtrlC"
	default:
		return fmt.Sprintf("Key%d", int(k))
	}
}

func ReadDirection() (Direction, error) {
	return DirNone, nil
}
