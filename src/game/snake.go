package game

type Snake struct {
	segments  []Position
	direction Direction
}

func NewSnake() *Snake {
	return &Snake{
		segments: []Position{
			{X: 21, Y: 10},
			{X: 20, Y: 10},
			{X: 19, Y: 10},
		},
		direction: DirRight,
	}
}

func (s *Snake) Move(dir Direction) {
	opposite := map[Direction]Direction{
		DirUp:    DirDown,
		DirDown:  DirUp,
		DirLeft:  DirRight,
		DirRight: DirLeft,
	}

	if opposite[dir] == s.direction {
		dir = s.direction
	} else {
		s.direction = dir
	}

	head := s.Head()
	var newHead Position

	switch dir {
	case DirUp:
		newHead = Position{X: head.X, Y: head.Y + 1}
	case DirDown:
		newHead = Position{X: head.X, Y: head.Y - 1}
	case DirLeft:
		newHead = Position{X: head.X - 1, Y: head.Y}
	case DirRight:
		newHead = Position{X: head.X + 1, Y: head.Y}
	}

	s.segments = append([]Position{newHead}, s.segments[:len(s.segments)-1]...)
}

func (s *Snake) Grow() {
	s.segments = append(s.segments, s.segments[len(s.segments)-1])
}

func (s *Snake) Head() Position {
	return s.segments[0]
}

func (s *Snake) CollidesWithSelf() bool {
	head := s.Head()
	for i := 1; i < len(s.segments); i++ {
		if s.segments[i] == head {
			return true
		}
	}
	return false
}

func (s *Snake) Segments() []Position {
	return s.segments
}

func (s *Snake) Direction() Direction {
	return s.direction
}
