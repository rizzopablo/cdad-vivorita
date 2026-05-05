package render

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"vivorita2/src/game"
)

func RenderBoard(screen tcell.Screen, snake *game.Snake, food *game.Food, score, highScore int) {
	// Clear screen
	screen.Clear()

	// Create board representation
	board := make([][]rune, 20)
	for i := range board {
		board[i] = make([]rune, 40)
		for j := range board[i] {
			board[i][j] = ' '
		}
	}

	// Draw snake
	segments := snake.Segments()
	for i, seg := range segments {
		if seg.X >= 0 && seg.X < 40 && seg.Y >= 0 && seg.Y < 20 {
			if i == 0 {
				board[seg.Y][seg.X] = '█' // Head
			} else {
				board[seg.Y][seg.X] = '█' // Body
			}
		}
	}

	// Draw food
	foodPos := food.Position()
	if foodPos.X >= 0 && foodPos.X < 40 && foodPos.Y >= 0 && foodPos.Y < 20 {
		board[foodPos.Y][foodPos.X] = '●'
	}

	// Draw borders
	for x := 0; x < 40; x++ {
		board[0][x] = '─'  // Top border
		board[19][x] = '─' // Bottom border
	}
	for y := 0; y < 20; y++ {
		board[y][0] = '│'  // Left border
		board[y][39] = '│' // Right border
	}
	// Corners
	board[0][0] = '┌'   // Top-left
	board[0][39] = '┐'  // Top-right
	board[19][0] = '└'  // Bottom-left
	board[19][39] = '┘' // Bottom-right

	// Draw board to screen
	style := tcell.StyleDefault
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			screen.SetContent(x, y, board[y][x], nil, style)
		}
	}

	// Print score below the board
	scoreStr := fmt.Sprintf("Score: %d | High Score: %d", score, highScore)
	for i, ch := range scoreStr {
		if i < 80 { // Limit to prevent overflow
			screen.SetContent(i, 21, ch, nil, style)
		}
	}

	screen.Show()
}
