package game

import (
	"math/rand"
	"time"
)

type Food struct {
	position Position
}

// Global random generator initialized once
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func NewFood(snake *Snake) *Food {
	// Create a map of occupied positions
	segments := snake.Segments()
	occupied := make(map[Position]bool)
	for _, seg := range segments {
		occupied[seg] = true
	}

	// Collect all available positions
	var available []Position
	for x := 0; x < 40; x++ {
		for y := 0; y < 20; y++ {
			pos := Position{X: x, Y: y}
			if !occupied[pos] {
				available = append(available, pos)
			}
		}
	}

	// Select a random position from available ones
	if len(available) > 0 {
		randomIndex := rng.Intn(len(available))
		return &Food{position: available[randomIndex]}
	}

	// Fallback: if somehow all positions are occupied (shouldn't happen), return a default
	return &Food{position: Position{X: 0, Y: 0}}
}

func (f *Food) Position() Position {
	return f.position
}

func (f *Food) Respawn(snake *Snake) {
	// Create a map of occupied positions
	segments := snake.Segments()
	occupied := make(map[Position]bool)
	for _, seg := range segments {
		occupied[seg] = true
	}

	// Collect all available positions
	var available []Position
	for x := 0; x < 40; x++ {
		for y := 0; y < 20; y++ {
			pos := Position{X: x, Y: y}
			if !occupied[pos] {
				available = append(available, pos)
			}
		}
	}

	// Select a random position from available ones
	if len(available) > 0 {
		randomIndex := rng.Intn(len(available))
		f.position = available[randomIndex]
	}
}
