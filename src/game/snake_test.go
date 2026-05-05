package game

import (
	"os"
	"testing"
	"vivorita2/src/score"
)

// Helper functions for tests
func osRemove(path string) {
	os.Remove(path)
}

// PC1: Dado un juego nuevo, la serpiente tiene exactamente 3 segmentos en posiciones
// consecutivas centradas, dirección derecha, y no hay colisión.
func TestPostcondition1_InitialSnakeState(t *testing.T) {
	snake := NewSnake()
	if snake == nil {
		t.Fatal("NewSnake() returned nil, expected valid snake")
	}

	// Verificar 3 segmentos
	segments := snake.Segments()
	if segments == nil {
		t.Fatal("Segments() returned nil, expected slice")
	}
	if len(segments) != 3 {
		t.Errorf("Expected 3 segments, got %d", len(segments))
		return
	}

	// Verificar posiciones consecutivas centradas (centro del tablero 40x20: X=20, Y=10)
	// Dirección derecha: serpiente apunta hacia la derecha
	// Cabeza en (21, 10), cuerpo en (20, 10), cola en (19, 10)
	expectedPositions := []Position{
		{X: 21, Y: 10}, // cabeza
		{X: 20, Y: 10}, // medio
		{X: 19, Y: 10}, // cola
	}

	for i, expected := range expectedPositions {
		if i >= len(segments) {
			t.Errorf("Missing segment %d", i)
			continue
		}
		if segments[i] != expected {
			t.Errorf("Segment %d: expected %v, got %v", i, expected, segments[i])
		}
	}

	// Verificar que no hay colisión consigo misma
	if snake.CollidesWithSelf() {
		t.Error("New snake should not collide with itself")
	}
}

// PC2: Tras llamar Move(DirRight) N veces sin comer, la serpiente tiene la misma
// longitud y su cabeza avanzó N posiciones a la derecha.
func TestPostcondition2_MoveWithoutEating(t *testing.T) {
	snake := NewSnake()
	if snake == nil {
		t.Fatal("NewSnake() returned nil")
	}

	initialHead := snake.Head()
	segments := snake.Segments()
	if segments == nil {
		t.Fatal("Segments() returned nil")
	}
	initialLength := len(segments)

	// Mover 5 veces a la derecha
	moves := 5
	for i := 0; i < moves; i++ {
		snake.Move(DirRight)
	}

	// Verificar longitud igual
	segments = snake.Segments()
	if segments == nil {
		t.Fatal("Segments() returned nil after moves")
	}
	finalLength := len(segments)
	if finalLength != initialLength {
		t.Errorf("Length changed: expected %d, got %d", initialLength, finalLength)
	}

	// Verificar que la cabeza avanzó N posiciones a la derecha
	finalHead := snake.Head()
	expectedX := initialHead.X + moves
	if finalHead.X != expectedX || finalHead.Y != initialHead.Y {
		t.Errorf("Head position: expected (%d, %d), got (%d, %d)",
			expectedX, initialHead.Y, finalHead.X, finalHead.Y)
	}
}

// PC3: Tras comer una comida (cabeza coincide con posición de comida),
// la serpiente crece en 1 segmento y el score aumenta en 1.
func TestPostcondition3_EatFoodGrowAndScore(t *testing.T) {
	game := NewGame()
	if game == nil {
		t.Fatal("NewGame() returned nil")
	}

	initialScore := game.Score()

	// Simular que la serpiente come colocando comida justo adelante de la cabeza
	snake := game.Snake()
	if snake == nil {
		t.Fatal("Snake() returned nil")
	}

	segments := snake.Segments()
	if segments == nil {
		t.Fatal("Segments() returned nil")
	}
	initialLength := len(segments)

	// Crear comida en la posición siguiente de la cabeza
	head := snake.Head()
	foodPos := Position{X: head.X + 1, Y: head.Y}
	// We need to modify the game.food.position directly to simulate placing food at a specific position
	// For this test, we'll create a temporary workaround by creating new food at the desired position
	oldFood := game.food
	// Replace with new food temporarily
	game.food = &Food{position: foodPos}

	// Mover hacia la comida
	game.Update(DirRight)

	// Restore the original food if needed
	game.food = oldFood

	// Verify that the snake grew by 1 segment
	segments = snake.Segments()
	if segments == nil {
		t.Fatal("Segments() returned nil after eating")
	}
	finalLength := len(segments)
	expectedLength := initialLength + 1
	if finalLength != expectedLength {
		t.Errorf("Length after eating: expected %d, got %d", expectedLength, finalLength)
	}

	// Verify that the score increased by 1
	finalScore := game.Score()
	expectedScore := initialScore + 1
	if finalScore != expectedScore {
		t.Errorf("Score after eating: expected %d, got %d", expectedScore, finalScore)
	}
}

// PC4: Si la cabeza de la serpiente sale del tablero (X < 0, X > 39, Y < 0, Y > 19),
// IsOver() retorna true.
func TestPostcondition4_WallCollisionGameOver(t *testing.T) {
	// Test colisión con borde derecho (X > 39)
	game := NewGame()
	if game == nil {
		t.Fatal("NewGame() returned nil")
	}

	snake := game.Snake()
	if snake == nil {
		t.Fatal("Snake() returned nil")
	}

	// Mover la serpiente hasta el borde derecho y más allá
	// Cabeza inicial en X=21, necesitamos llegar a X=40 (fuera del tablero)
	for i := 0; i < 100; i++ {
		game.Update(DirRight)
		head := snake.Head()
		if head.X >= 40 {
			break
		}
	}

	if !game.IsOver() {
		t.Error("Game should be over when snake hits right wall")
	}

	// Test colisión con borde izquierdo (X < 0)
	game2 := NewGame()
	if game2 == nil {
		t.Fatal("NewGame() returned nil for game2")
	}

	snake2 := game2.Snake()
	if snake2 == nil {
		t.Fatal("Snake() returned nil for snake2")
	}

	// Primero ir a la izquierda del tablero
	for i := 0; i < 100; i++ {
		game2.Update(DirLeft)
		head := snake2.Head()
		if head.X < 0 {
			break
		}
	}

	if !game2.IsOver() {
		t.Error("Game should be over when snake hits left wall")
	}
}

// PC5: Si la cabeza de la serpiente coincide con cualquier otro segmento del cuerpo,
// IsOver() retorna true.
func TestPostcondition5_SelfCollisionGameOver(t *testing.T) {
	game := NewGame()
	if game == nil {
		t.Fatal("NewGame() returned nil")
	}

	snake := game.Snake()
	if snake == nil {
		t.Fatal("Snake() returned nil")
	}

	// Crear una situación de colisión consigo misma
	// Hacer crecer la serpiente primero para tener más cuerpo
	for i := 0; i < 5; i++ {
		snake.Grow()
	}

	// Mover: arriba, derecha, abajo, izquierda (para formar un bucle)
	game.Update(DirUp)
	game.Update(DirRight)
	game.Update(DirDown)
	game.Update(DirLeft)

	if !game.IsOver() {
		t.Error("Game should be over when snake collides with itself")
	}
}

// PC6: Si el jugador presiona la dirección opuesta a la actual (ej: DirLeft mientras va DirRight),
// la dirección NO cambia (invariante I1).
func TestPostcondition6_NoReverseDirection(t *testing.T) {
	snake := NewSnake()
	if snake == nil {
		t.Fatal("NewSnake() returned nil")
	}

	// La serpiente inicial va hacia la derecha
	// Intentar mover hacia la izquierda (dirección opuesta)
	snake.Move(DirLeft)

	// La cabeza debería haber seguido yendo a la derecha, no a la izquierda
	head := snake.Head()
	expectedX := 22 // If it went left it would be 20, but it should stay going right
	if head.X != expectedX {
		t.Errorf("Direction should not reverse: expected head.X=%d, got %d", expectedX, head.X)
	}

	// Probar con otra dirección: arriba vs abajo
	snake2 := NewSnake()
	if snake2 == nil {
		t.Fatal("NewSnake() returned nil for snake2")
	}

	// Primero subir
	snake2.Move(DirUp)
	headAfterUp := snake2.Head()

	// Luego intentar bajar (opuesto)
	snake2.Move(DirDown)
	headAfterDownAttempt := snake2.Head()

	// La Y debería haber seguido aumentando (subiendo), no disminuyendo
	if headAfterDownAttempt.Y <= headAfterUp.Y {
		t.Error("Direction should not reverse: attempted to go down while going up")
	}
}

// PC7: La comida generada nunca tiene la misma posición que ningún segmento de la serpiente.
func TestPostcondition7_FoodNotOnSnake(t *testing.T) {
	snake := NewSnake()
	if snake == nil {
		t.Fatal("NewSnake() returned nil")
	}

	segments := snake.Segments()
	if segments == nil {
		t.Fatal("Segments() returned nil")
	}

	// Generar comida múltiples veces
	for i := 0; i < 100; i++ {
		food := NewFood(snake)
		if food == nil {
			t.Fatal("NewFood() returned nil")
		}

		foodPos := food.Position()

		// Verificar que la comida no está sobre ningún segmento
		for _, seg := range segments {
			if foodPos == seg {
				t.Errorf("Food spawned on snake segment at %v", foodPos)
			}
		}
	}
}

// PC8: SaveHighScore con un score mayor al actual actualiza el archivo;
// con un score menor o igual, no modifica el archivo.
func TestPostcondition8_SaveHighScoreConditional(t *testing.T) {
	tempFile := "/tmp/test_highscore.json"

	// Limpiar archivo de test si existe
	osRemove(tempFile)

	// Guardar un score inicial
	initialScore := 100
	err := score.SaveHighScore(tempFile, initialScore)
	if err != nil {
		t.Fatalf("Failed to save initial high score: %v", err)
	}

	// Verificar que se guardó
	loaded, _ := score.LoadHighScore(tempFile)
	if loaded != initialScore {
		t.Errorf("Initial score not saved: expected %d, got %d", initialScore, loaded)
	}

	// Intentar guardar un score menor - no debería modificar
	lowerScore := 50
	err = score.SaveHighScore(tempFile, lowerScore)
	if err != nil {
		t.Fatalf("Failed to save lower score: %v", err)
	}

	loaded, _ = score.LoadHighScore(tempFile)
	if loaded != initialScore {
		t.Errorf("Score should not decrease: expected %d, got %d", initialScore, loaded)
	}

	// Intentar guardar score igual - no debería modificar
	err = score.SaveHighScore(tempFile, initialScore)
	if err != nil {
		t.Fatalf("Failed to save equal score: %v", err)
	}

	loaded, _ = score.LoadHighScore(tempFile)
	if loaded != initialScore {
		t.Errorf("Score should not change with equal value: expected %d, got %d", initialScore, loaded)
	}

	// Guardar score mayor - debería actualizar
	higherScore := 200
	err = score.SaveHighScore(tempFile, higherScore)
	if err != nil {
		t.Fatalf("Failed to save higher score: %v", err)
	}

	loaded, _ = score.LoadHighScore(tempFile)
	if loaded != higherScore {
		t.Errorf("Score should update with higher value: expected %d, got %d", higherScore, loaded)
	}
}

// PC9: Tras llamar Pause(), IsPaused() retorna true y Update() no modifica el estado del juego.
// Tras llamar Resume(), IsPaused() retorna false y el juego vuelve a actualizarse normalmente.
func TestPostcondition9_PauseResume(t *testing.T) {
	game := NewGame()
	if game == nil {
		t.Fatal("NewGame() returned nil")
	}

	snake := game.Snake()
	if snake == nil {
		t.Fatal("Snake() returned nil")
	}

	initialHead := snake.Head()

	// Pausar el juego
	game.Pause()

	if !game.IsPaused() {
		t.Error("IsPaused() should return true after Pause()")
	}

	// Intentar actualizar mientras está pausado
	game.Update(DirRight)

	// El estado no debería cambiar
	headAfterPausedUpdate := snake.Head()
	if headAfterPausedUpdate != initialHead {
		t.Error("Update() should not modify state when paused")
	}

	// Reanudar el juego
	game.Resume()

	if game.IsPaused() {
		t.Error("IsPaused() should return false after Resume()")
	}

	// Ahora Update debería funcionar
	game.Update(DirRight)
	headAfterResume := snake.Head()
	if headAfterResume == initialHead {
		t.Error("Update() should modify state after Resume()")
	}
}

// PC10: IsNewHighScore() retorna true si el score actual es estrictamente mayor
// que el high score cargado al iniciar el juego; retorna false en caso contrario.
func TestPostcondition10_IsNewHighScore(t *testing.T) {
	tempFile := "/tmp/test_highscore_pc10.json"
	osRemove(tempFile)

	// Guardar un high score de referencia
	highScore := 100
	score.SaveHighScore(tempFile, highScore)

	// Crear juego con ese high score cargado
	game := NewGameWithHighScore(tempFile)
	if game == nil {
		t.Fatal("NewGameWithHighScore() returned nil")
	}

	// Score inicial debería ser 0, no supera el high score
	if game.IsNewHighScore() {
		t.Error("IsNewHighScore() should be false when score (0) <= highScore (100)")
	}

	// Simular score igual al high score - no es nuevo record
	setGameScore(game, highScore)
	if game.IsNewHighScore() {
		t.Error("IsNewHighScore() should be false when score equals highScore")
	}

	// Simular score menor - no es nuevo record
	setGameScore(game, 50)
	if game.IsNewHighScore() {
		t.Error("IsNewHighScore() should be false when score < highScore")
	}

	// Simular score mayor - sí es nuevo record
	setGameScore(game, 150)
	if !game.IsNewHighScore() {
		t.Error("IsNewHighScore() should be true when score > highScore")
	}
}

// Helper function for tests
func setGameScore(game *Game, score int) {
	game.score = score
}
