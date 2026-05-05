# Spec: 001-snake-game

## Descripción
Clon clásico del juego Snake (viborita) ejecutable en terminal. El jugador controla una serpiente que se mueve por un tablero de 40x20 celdas, comiendo comida para crecer y sumar puntos. El juego termina cuando la serpiente choca contra los bordes o contra su propio cuerpo. El high score se persiste entre sesiones en un archivo local JSON.

## Contrato

### Estructura del proyecto
```
src/
├── main.go          ← punto de entrada
├── game/
│   ├── game.go      ← game loop y estado
│   ├── snake.go     ← lógica de la serpiente
│   └── food.go      ← lógica de la comida
├── render/
│   └── render.go    ← dibujo en terminal
├── input/
│   └── input.go     ← lectura de teclado
└── score/
    └── score.go     ← persistencia de high score
```

### Game Loop
- `NewGame() *Game` — crea una partida nueva con estado inicial.
- `(g *Game) Run()` — loop principal: lee input, actualiza estado, renderiza.
- `(g *Game) Update(dir Direction)` — avanza un tick en la dirección dada.
- `(g *Game) IsOver() bool` — retorna true si hay colisión.
- `(g *Game) Score() int` — puntaje actual.
- `(g *Game) Pause()` — pone el juego en pausa.
- `(g *Game) Resume()` — reanuda el juego desde pausa.
- `(g *Game) IsPaused() bool` — retorna true si el juego está en pausa.
- `(g *Game) IsNewHighScore() bool` — retorna true si el score actual supera el high score cargado.

### Snake
- `NewSnake() *Snake` — serpiente inicial (3 segmentos, centro del tablero, dirección derecha).
- `(s *Snake) Move(dir Direction)` — mueve la serpiente un paso.
- `(s *Snake) Grow()` — agrega un segmento al final.
- `(s *Snake) Head() Position` — posición de la cabeza.
- `(s *Snake) CollidesWithSelf() bool` — detecta colisión con el cuerpo.
- `(s *Snake) Segments() []Position` — retorna todas las posiciones.

### Food
- `NewFood(snake *Snake) *Food` — genera comida en posición aleatoria no ocupada por la serpiente.
- `(f *Food) Position() Position` — posición actual.
- `(f *Food) Respawn(snake *Snake)` — reposiciona la comida.

### Render
- `RenderBoard(w io.Writer, snake *Snake, food *Food, score, highScore int)` — dibuja el tablero completo.

### Input
- `ReadDirection() (Direction, error)` — lee una tecla y retorna la dirección.

### Score
- `LoadHighScore(path string) (int, error)` — carga el high score del archivo. Si el archivo no existe, retorna `(0, nil)` — no es un error, simplemente no hay high score previo.
- `SaveHighScore(path string, score int) error` — guarda el high score solo si `score` es mayor al actual. Si `score` es menor o igual, no modifica el archivo.

### Tipos
```go
type Position struct { X, Y int }
type Direction int
const (
    DirUp Direction = iota
    DirDown
    DirLeft
    DirRight
)
```

## Invariantes

1. **I1**: La serpiente nunca puede invertir dirección 180° (si va a la derecha, no puede ir a la izquierda inmediatamente). **Esta validación ocurre en `Move`, no en `ReadDirection`.**
2. **I2**: La comida nunca aparece sobre un segmento de la serpiente.
3. **I3**: El tablero es de tamaño fijo 40x20 (coordenadas válidas: X ∈ [0,39], Y ∈ [0,19]).
4. **I4**: El high score nunca decrece — solo se actualiza si el score actual lo supera.
5. **I5**: La velocidad del juego es constante (tick fijo de 150ms). **Esta invariante se valida en tests de integración/E2E, no en unit tests.**

## Criterios de Aceptación

1. El juego se ejecuta con `go run .` desde la raíz del proyecto.
2. El tablero se renderiza en terminal usando caracteres Unicode (`█` para serpiente, `●` para comida, `─│┌┐└┘┤` para bordes).
3. Controles: flechas y WASD mapeados a las 4 direcciones.
4. Tecla `P` pausa/reanuda. Tecla `Q` sale.
5. Al morir: muestra "Game Over", puntaje final, y si es nuevo high score lo indica.
6. High score se persiste en `~/.vivorita2_highscore.json` y se lee al iniciar.
7. La serpiente inicial tiene 3 segmentos, empieza en el centro, se mueve a la derecha.
8. Cada comida suma 1 punto y crece la serpiente en 1 segmento.
9. Colisión con borde o cuerpo propio = game over.

## Postcondiciones

1. **PC1**: Dado un juego nuevo, la serpiente tiene exactamente 3 segmentos en posiciones consecutivas centradas, dirección derecha, y no hay colisión.
2. **PC2**: Tras llamar `Move(DirRight)` N veces sin comer, la serpiente tiene la misma longitud y su cabeza avanzó N posiciones a la derecha.
3. **PC3**: Tras comer una comida (cabeza coincide con posición de comida), la serpiente crece en 1 segmento y el score aumenta en 1.
4. **PC4**: Si la cabeza de la serpiente sale del tablero (X < 0, X > 39, Y < 0, Y > 19), `IsOver()` retorna true.
5. **PC5**: Si la cabeza de la serpiente coincide con cualquier otro segmento del cuerpo, `IsOver()` retorna true.
6. **PC6**: Si el jugador presiona la dirección opuesta a la actual (ej: DirLeft mientras va DirRight), la dirección NO cambia (invariante I1).
7. **PC7**: La comida generada nunca tiene la misma posición que ningún segmento de la serpiente.
8. **PC8**: `SaveHighScore` con un score mayor al actual actualiza el archivo; con un score menor o igual, no modifica el archivo.
9. **PC9**: Tras llamar `Pause()`, `IsPaused()` retorna true y `Update()` no modifica el estado del juego. Tras llamar `Resume()`, `IsPaused()` retorna false y el juego vuelve a actualizarse normalmente.
10. **PC10**: `IsNewHighScore()` retorna true si el score actual es estrictamente mayor que el high score cargado al iniciar el juego; retorna false en caso contrario.

Status: Approved by Pablo Manuel Rizzo on 2026-05-04
