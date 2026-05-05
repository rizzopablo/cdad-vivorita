package main

import (
	"time"

	"github.com/gdamore/tcell/v2"

	"vivorita2/src/game"
	"vivorita2/src/input"
	"vivorita2/src/render"
)

func main() {
	// Initialize screen
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err = screen.Init(); err != nil {
		panic(err)
	}
	defer screen.Fini()

	g := game.NewGameWithHighScore("~/.vivorita2_highscore.json")

	// Start game loop with ticker
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	running := true
	for running {
		select {
		case <-ticker.C:
			if !g.IsPaused() && !g.IsOver() {
				render.RenderBoard(screen, g.Snake(), g.Food(), g.Score(), g.HighScore())
				screen.Show()
			}
		default:
			// Non-blocking read of input
			if dir, err := input.ReadDirectionNonBlocking(); err == nil {
				// Convert input.Direction to game.Direction
				gameDir := convertInputToGameDirection(dir)

				switch dir {
				case input.DirQuit:
					running = false
				case input.DirPause:
					if g.IsPaused() {
						g.Resume()
					} else {
						g.Pause()
					}
				default:
					if !g.IsOver() {
						g.Update(gameDir)
						if !g.IsPaused() {
							render.RenderBoard(screen, g.Snake(), g.Food(), g.Score(), g.HighScore())
							screen.Show()
						}
					}
				}
			}

			if g.IsOver() {
				// Show game over screen briefly before exiting
				render.RenderBoard(screen, g.Snake(), g.Food(), g.Score(), g.HighScore())
				screen.Show()
				time.Sleep(3 * time.Second)
				running = false
			}
		}
	}
}

// Convert input.Direction to game.Direction
func convertInputToGameDirection(inputDir input.Direction) game.Direction {
	switch inputDir {
	case input.DirUp:
		return game.DirUp
	case input.DirDown:
		return game.DirDown
	case input.DirLeft:
		return game.DirLeft
	case input.DirRight:
		return game.DirRight
	default:
		return game.DirUp
	}
}
