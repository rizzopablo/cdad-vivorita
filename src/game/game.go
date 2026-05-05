package game

import (
	"os/user"
	"path/filepath"
	"time"

	"vivorita2/src/input"
	"vivorita2/src/score"
)

type Position struct {
	X, Y int
}

type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

type Game struct {
	snake         *Snake
	food          *Food
	score         int
	over          bool
	paused        bool
	highScore     int
	highScorePath string
}

func NewGame() *Game {
	snake := NewSnake()
	highScore, _ := score.LoadHighScore(getDefaultHighScorePath())
	return &Game{
		snake:     snake,
		food:      NewFood(snake),
		score:     0,
		over:      false,
		paused:    false,
		highScore: highScore,
	}
}

func NewGameWithHighScore(path string) *Game {
	snake := NewSnake()
	highScore, _ := score.LoadHighScore(path)
	return &Game{
		snake:         snake,
		food:          NewFood(snake),
		score:         0,
		over:          false,
		paused:        false,
		highScore:     highScore,
		highScorePath: path,
	}
}

func (g *Game) Run() {
	// Start game loop with ticker of 150ms
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	running := true
	for running {
		<-ticker.C

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
				}
			}
		}

		if g.IsOver() {
			// Exit the loop when game is over
			running = false
		}
	}
}

// Convert input.Direction to game.Direction
func convertInputToGameDirection(inputDir input.Direction) Direction {
	switch inputDir {
	case input.DirUp:
		return DirUp
	case input.DirDown:
		return DirDown
	case input.DirLeft:
		return DirLeft
	case input.DirRight:
		return DirRight
	default:
		return DirUp
	}
}

func (g *Game) Update(dir Direction) {
	if g.paused {
		return
	}

	g.snake.Move(dir)

	head := g.snake.Head()

	if head.X < 0 || head.X > 39 || head.Y < 0 || head.Y > 19 {
		g.over = true
		return
	}

	if g.snake.CollidesWithSelf() {
		g.over = true
		return
	}

	if head == g.food.Position() {
		g.snake.Grow()
		g.score++
		g.food.Respawn(g.snake)
	}
}

func (g *Game) IsOver() bool {
	return g.over
}

func (g *Game) Score() int {
	return g.score
}

func (g *Game) Pause() {
	g.paused = true
}

func (g *Game) Resume() {
	g.paused = false
}

func (g *Game) IsPaused() bool {
	return g.paused
}

func (g *Game) IsNewHighScore() bool {
	return g.score > g.highScore
}

func (g *Game) Snake() *Snake {
	return g.snake
}

func (g *Game) Food() *Food {
	return g.food
}

func (g *Game) HighScore() int {
	return g.highScore
}

// getDefaultHighScorePath returns the path to the high score file in the user's home directory
func getDefaultHighScorePath() string {
	user, err := user.Current()
	if err != nil {
		// Fallback to a simple filename if user directory cannot be determined
		return ".vivorita2_highscore.json"
	}
	return filepath.Join(user.HomeDir, ".vivorita2_highscore.json")
}
