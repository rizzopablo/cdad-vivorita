package main

import (
	"time"

	"github.com/gdamore/tcell/v2"

	"vivorita2/src/game"
	"vivorita2/src/input"
	"vivorita2/src/observability"
	"vivorita2/src/render"
)

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err = screen.Init(); err != nil {
		panic(err)
	}
	defer screen.Fini()

	g := game.NewGameWithHighScore("~/.vivorita2_highscore.json")

	if err := observability.InitLogging(); err != nil {
		panic(err)
	}

	input.LogEvent = observability.LogEvent

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	running := true
	firstInputReceived := false
	for running {
		select {
		case <-ticker.C:
			if !g.IsPaused() && !g.IsOver() {
				render.RenderBoard(screen, g.Snake(), g.Food(), g.Score(), g.HighScore())
				observability.LogEvent("render", map[string]interface{}{
					"source": "ticker",
				})
				screen.Show()
			}
		default:
			if dir, err := input.ReadDirectionNonBlocking(screen); err == nil {
				gameDir := convertInputToGameDirection(dir, g.Snake().Direction())

				switch dir {
				case input.DirQuit:
					running = false
				case input.DirPause:
					if g.IsPaused() {
						g.Resume()
					} else {
						g.Pause()
					}
				case input.DirUp, input.DirDown, input.DirLeft, input.DirRight:
					if !firstInputReceived {
						firstInputReceived = true
					}
					if firstInputReceived && !g.IsOver() {
						g.Update(gameDir)
						observability.LogEvent("update", map[string]interface{}{
							"direction":  directionName(dir),
							"snake_head": g.Snake().Head(),
							"score":      g.Score(),
							"over":       g.IsOver(),
							"paused":     g.IsPaused(),
						})
						if !g.IsPaused() {
							render.RenderBoard(screen, g.Snake(), g.Food(), g.Score(), g.HighScore())
							observability.LogEvent("render", map[string]interface{}{
								"source": "update",
							})
							screen.Show()
						}
					}
				}
			}

			if g.IsOver() {
				if err := observability.SnapshotBoard(g, "game_over"); err != nil {
					observability.LogEvent("snapshot_error", map[string]interface{}{
						"error": err.Error(),
					})
				}
				render.RenderBoard(screen, g.Snake(), g.Food(), g.Score(), g.HighScore())
				observability.LogEvent("render", map[string]interface{}{
					"source": "gameover",
				})
				screen.Show()
				time.Sleep(3 * time.Second)
				running = false
			}
		}
	}
}

func convertInputToGameDirection(inputDir input.Direction, currentDir game.Direction) game.Direction {
	switch inputDir {
	case input.DirUp:
		return game.DirUp
	case input.DirDown:
		return game.DirDown
	case input.DirLeft:
		return game.DirLeft
	case input.DirRight:
		return game.DirRight
	case input.DirNone:
		return currentDir
	default:
		return currentDir
	}
}

func directionName(d input.Direction) string {
	switch d {
	case input.DirUp:
		return "DirUp"
	case input.DirDown:
		return "DirDown"
	case input.DirLeft:
		return "DirLeft"
	case input.DirRight:
		return "DirRight"
	case input.DirPause:
		return "DirPause"
	case input.DirQuit:
		return "DirQuit"
	default:
		return "DirUp"
	}
}
