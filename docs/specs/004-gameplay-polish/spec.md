---
id: "004-gameplay-polish"
title: "Gameplay Polish"
status: "approved"
approved_by: "Pablo Manuel Rizzo"
approved_at: "2026-05-05"
---

# Feature: Gameplay Polish

## Descripción funcional

Mejora la experiencia de juego ajustando tres aspectos clave:

1. **Snake estática al inicio**: La serpiente no se mueve hasta que el jugador presiona su primer input de dirección. El tablero, la serpiente inicial y la comida son visibles desde el inicio, pero sin movimiento automático.

2. **Velocidad del game loop**: Ajusta el tick duration del game loop a 200ms (incremento del 33% respecto a los 150ms actuales) para dar más tiempo de reacción al jugador.

3. **Constante InputPollTimeout**: Extrae el magic number de 10ms del timeout en `screen.PollEvent()` a una constante privada en `src/input/input.go`.

Estos cambios mejoran la accesibilidad y la claridad de intención del código sin cambiar la mecánica del juego.

## Contrato

### Firma pública

```go
// src/main.go
// Cambios en el game loop para manejar "primer input recibido"

// src/input/input.go
const InputPollTimeout = 10 * time.Millisecond

func ReadDirectionNonBlocking(screen tcell.Screen) (Direction, error)
// Utiliza InputPollTimeout en lugar de magic number
```

### Postcondiciones

1. **PC1**: El game loop usa un ticker con duración de 200ms en lugar de 150ms.
   - Input: Línea 32 en `src/main.go`: `time.NewTicker(150 * time.Millisecond)`
   - Output: Se cambia a `time.NewTicker(200 * time.Millisecond)`
   - Error: Si el ticker usa otra duración, falla
   - Verificación: Verificar código en `src/main.go` línea 32; medir intervalos reales con logs de render (con `DEBUG=1`, dos eventos `render` consecutivos están separados por ~200ms)

2. **PC2**: `src/input/input.go` define la constante `InputPollTimeout` con valor 10ms.
   - Input: Definición en `src/input/input.go` antes de la función `ReadDirectionNonBlocking()`
   - Output: `const InputPollTimeout = 10 * time.Millisecond`
   - Error: Si no existe la constante o tiene otro valor, falla
   - Verificación: `go build` compila; la constante está accesible en el archivo

3. **PC3**: `ReadDirectionNonBlocking()` usa `InputPollTimeout` en lugar del magic number 10ms.
   - Input: `time.After(10 * time.Millisecond)` en línea 37 de `src/input/input.go`
   - Output: Se reemplaza por `time.After(InputPollTimeout)`
   - Error: Si aún usa el número literal, falla
   - Verificación: Búsqueda de "10 * time.Millisecond" en `src/input/input.go` no encuentra resultados (excepto en comentarios); `go build` compila sin cambios de comportamiento

4. **PC4**: El game loop mantiene un flag interno "firstInputReceived" que se inicializa en false.
   - Input: Línea ~24-33 en `src/main.go` donde se crea el game y el ticker
   - Output: Existe un booleano `firstInputReceived := false` declarado antes del game loop
   - Error: Si no existe el flag o está mal inicializado, falla
   - Verificación: Code review del game loop detecta la variable

5. **PC5**: El flag "firstInputReceived" se asigna a true cuando se recibe el primer input que NO sea DirNone, DirPause ni DirQuit.
   - Input: Primer input que mapea a una dirección válida (DirUp, DirDown, DirLeft, DirRight)
   - Output: `firstInputReceived` cambia a `true` y permanece así para el resto de la ejecución
   - Error: Si el flag se asigna incorrectamente o se reinicia, falla
   - Verificación: Test unitario/E2E que simula inputs y verifica el estado interno del flag

6. **PC6**: Mientras `firstInputReceived` es false, `g.Update()` NO se invoca en el game loop.
   - Input: Game loop ejecutándose, ningún input aún presionado
   - Output: La serpiente no cambia de posición, permanece en su posición inicial
   - Error: Si la serpiente se mueve antes del primer input, falla
   - Verificación: Test E2E que ejecuta 100 ticks del game loop sin input verifica que `snake.Head()` permanece en la posición inicial

7. **PC7**: Cuando se recibe el primer input de dirección (PC5), `g.Update()` se invoca con la dirección correspondiente y el movimiento comienza.
   - Input: Tecla presionada (ej: 'd' para DirRight) cuando `firstInputReceived == false`
   - Output: `g.Update()` se ejecuta; en el siguiente tick, la serpiente estará en la posición esperada en la nueva dirección
   - Error: Si se invocan varias llamadas a `Update()` por un solo input, o si el input se ignora, falla
   - Verificación: Test E2E simula input "d" en tick 0, verifica `Update()` se llamó exactamente una vez, y snake.Head() cambió en la dirección esperada

8. **PC8**: El tablero inicial (serpiente, comida, bordes) es visible desde el inicio sin necesidad de input.
   - Input: Inicio del juego, antes de cualquier input
   - Output: `render.RenderBoard()` se ejecuta al menos una vez antes del primer input, mostrando el estado inicial
   - Error: Si el tablero no se renderiza hasta que hay input, falla
   - Verificación: Test E2E verifica que hay un evento `render` loggeado con `source:"ticker"` en los primeros 200ms con `DEBUG=1`

9. **PC9**: Después del primer input de dirección, el game loop se comporta exactamente como antes: respeta el ticker de 200ms y procesa inputs en cada iteración del `default` del select.
   - Input: Game loop tras recibir primer input de dirección
   - Output: Comportamiento idéntico al ciclo actual: render cada 200ms (ticker), procesar input en el `default` del select, actualizar serpiente cuando hay input válido
   - Error: Si hay cambios adicionales en la lógica del game loop después de `firstInputReceived == true`, falla
   - Verificación: Comparación de comportamiento: tests de snake movement, pause, quit con y sin flag "firstInputReceived" son idénticos

10. **PC10**: Los eventos de logging NO incluyen mensajes de "waiting for input" o similar durante la espera del primer input.
    - Input: Game loop ejecutándose, fase pre-primer-input, con `DEBUG=1`
    - Output: No hay entradas de log tipo `event:"waiting_for_input"` ni mensajes en `./logs/vivorita2-debug.log` indicando espera
    - Error: Si hay logs que documentan la espera, falla
    - Verificación: Con `DEBUG=1`, durante 2 segundos sin presionar teclas, el log NO contiene mensajes de "waiting"

## Invariantes verificables

### I1: Ticker y InputPollTimeout son constantes (no cambian dinámicamente)

El valor del ticker es siempre 200ms durante toda la ejecución del juego, independientemente de input o estado del juego.

El valor de `InputPollTimeout` es siempre 10ms en `src/input/input.go`.

**Verificación**: Búsqueda de asignaciones dinámicas a `ticker` (aparte de su creación) y a `InputPollTimeout` no encuentra modificaciones.

### I2: El flag "firstInputReceived" es monótono (once true, siempre true)

Una vez que `firstInputReceived` se asigna a true (cuando se recibe el primer input de dirección válida), nunca vuelve a false.

**Verificación**: Code review verifica que no existe `firstInputReceived = false` después de su inicialización; test E2E simula 50 inputs consecutivos y verifica que el flag sigue siendo true.

### I3: El estado inicial del snake es visual pero sin movimiento

Antes de `firstInputReceived == true`, la serpiente visible en pantalla nunca se mueve de su posición inicial, incluso si el game loop está ejecutándose.

**Verificación**: Test E2E mide `snake.Head()` en los ticks 0-5 (primeros 1000ms) sin input, verifica que todas las medidas dan la misma posición.

### I4: El comportamiento post-primer-input es idéntico al original

Una vez que `firstInputReceived == true`, el comportamiento del game loop es exactamente el mismo al que tenía antes de esta feature (sin cambios en lógica de movimiento, colisión, pause, etc.).

**Verificación**: Tests parametrizados comparan traces de ejecución (secuencias de posiciones de snake, scores) entre versión anterior y actual post-primer-input.

## Criterios de aceptación

1. **Ticker es 200ms**: Medición empírica con logs. Con `DEBUG=1`, dos eventos `render` separados en tiempo por ~200ms (margen ±20ms tolerado).

2. **InputPollTimeout es constante**: Búsqueda en código fuente de la constante en `src/input/input.go`; su uso en `time.After()` en la misma función.

3. **Constante sin magic numbers**: No existe `10 * time.Millisecond` literal en `src/input/input.go` fuera de comentarios.

4. **Snake estática al inicio**: Test E2E que ejecuta 10 ticks sin input verifica que `snake.Head()` es igual en todos los ticks (posición inicial).

5. **Primer input dispara movimiento**: Test E2E que simula input "d" en tick 5 verifica que snake.Head().X incrementa en el tick 6 (primera oportunidad post-input).

6. **Tablero visible sin input**: Con `DEBUG=1`, el log contiene evento `render` dentro de los primeros 200ms (antes de esperar input).

7. **Sin logs silenciosos de "waiting"**: Con `DEBUG=1`, búsqueda en `./logs/vivorita2-debug.log` de palabras "waiting", "poll", "input_timeout" no encuentra resultados relacionados con espera del primer input.

8. **Suite tests verde**: `go test ./... -v` retorna 0 tests fallidos.

9. **Compilación correcta**: `go build` compila sin errores de tipo ni imports.

10. **Comportamiento post-primer-input sin cambios**: Tests existentes (PC4 de 003-input-capture-investigation, tests de observability) siguen pasando sin modificación.

---

**Status**: Draft — Pendiente aprobación humana
