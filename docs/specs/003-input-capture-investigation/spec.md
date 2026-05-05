---
id: "003-input-capture-investigation"
title: "Input Capture Bug Fix"
status: "draft"
approved_by: ""
approved_at: ""
---

# Feature: Input Capture Bug Fix

## Descripción funcional

Corrige el bug de captura de input que causa que la serpiente se mueva automáticamente hacia abajo (DirUp en modelo cartesiano) cuando no hay input del usuario. Actualmente `ReadDirectionNonBlocking()` retorna `DirUp` tanto cuando hay timeout (sin input) como para teclas no mapeadas, lo que hace que la serpiente cambie de dirección sin que el usuario haya presionado nada.

El fix introduce `DirNone` como valor que indica "sin dirección válida" y modifica el flujo para que:
1. Sin input (timeout), la serpiente mantenga su dirección actual
2. No exista dirección por defecto cuando no hay input explícito
3. El mapeo input → game direction use la dirección actual como fallback

Status: Approved by Pablo Manuel Rizzo on 2026-05-05

## Contrato

### Firma pública

```go
// input/input.go
type Direction int

const (
    DirUp Direction = iota
    DirDown
    DirLeft
    DirRight
    DirPause
    DirQuit
    DirNone  // NUEVO: indica sin dirección válida
)

// Retorna DirNone cuando no hay input (timeout de 10ms)
func ReadDirectionNonBlocking(screen tcell.Screen) (Direction, error)

// game/game.go
type Direction int

const (
    DirUp Direction = iota
    DirDown
    DirLeft
    DirRight
    DirNone  // NUEVO: indica sin dirección válida
)

// main.go
// convertInputToGameDirection ahora recibe la dirección actual como parámetro
func convertInputToGameDirection(inputDir input.Direction, currentDir game.Direction) game.Direction
```

### Postcondiciones

1. **PC1**: `ReadDirectionNonBlocking()` retorna `DirNone` cuando no hay input disponible (timeout de 10ms).
   - Input: `screen.PollEvent()` timeout después de 10ms sin eventos
   - Output: Retorna `(DirNone, nil)`
   - Error: Si `screen` es `nil`, retorna `(DirNone, nil)` (no `DirUp`)
   - Verificación: Test que mockea `tcell.Screen` con timeout verifica retorno `DirNone`

2. **PC2**: `ReadDirectionNonBlocking()` retorna `DirNone` (no `DirUp`) cuando se recibe una tecla no mapeable.
   - Input: Tecla presionada que no está en el mapeo WASD/flechas (ej: 'x', '1', Enter)
   - Output: Retorna `(DirNone, nil)` y loggea `event:"input_error"`
   - Error: No debe retornar `DirUp` para teclas desconocidas
   - Verificación: Test con teclas aleatorias verifica `DirNone` y log de error

3. **PC3**: `convertInputToGameDirection()` en `main.go` recibe la dirección actual como parámetro y la retorna cuando `inputDir` es `DirNone`.
   - Input: `(input.DirNone, game.DirRight)` — sin input, dirección actual es derecha
   - Output: Retorna `game.DirRight` (mantiene dirección)
   - Error: Si `inputDir` es `DirNone` pero se retorna otra dirección, falla
   - Verificación: Test unitario con tabla de casos verifica fallback correcto

4. **PC4**: Sin input durante el game loop, la serpiente mantiene su dirección actual.
   - Input: Game loop ejecutándose, ninguna tecla presionada por 30 segundos
   - Output: Snake continúa moviéndose en la última dirección válida ingresada
   - Error: Si la serpiente cambia de dirección automáticamente (hacia abajo), falla
   - Verificación: Test E2E simula 200 ticks sin input, verifica dirección constante

5. **PC5**: No hay dirección por defecto cuando no hay input — `DirUp` ya no es valor de fallback.
   - Input: Cualquier situación donde `ReadDirectionNonBlocking()` no reciba input explícito
   - Output: Nunca se retorna `DirUp` como resultado de "sin input"
   - Error: Si el código retorna `DirUp` por ausencia de input, falla
   - Verificación: Code review verifica que `DirUp` solo se retorna para input 'w' o flecha arriba

6. **PC6**: `game.Direction` define `DirNone` como constante válida.
   - Input: Definición de constantes en `src/game/game.go`
   - Output: `const ( DirUp Direction = iota; ...; DirNone )` compila sin error
   - Error: Si falta `DirNone` o tiene valor incorrecto, falla compilación
   - Verificación: `go build` compila correctamente; tests usan `game.DirNone`

7. **PC7**: `input.Direction` define `DirNone` como constante válida.
   - Input: Definición de constantes en `src/input/input.go`
   - Output: `const ( DirUp Direction = iota; ...; DirNone )` compila sin error
   - Error: Si falta `DirNone` en input package, falla compilación
   - Verificación: `go build` compila; `input.DirNone` es exportado y usable

8. **PC8**: `game.Run()` (código muerto en `src/game/game.go:62-97`) usa `DirNone` cuando `ReadDirectionNonBlocking(nil)` retorna sin dirección.
   - Input: Llamada a `input.ReadDirectionNonBlocking(nil)` en línea 72
   - Output: Si retorna `DirNone`, no se llama `g.Update()` (o se llama con dirección anterior si se refactoriza)
   - Error: Si `game.Run()` cambia dirección sin input, falla
   - Verificación: Aunque es código muerto, debe compilarse y no romper tests

## Invariantes verificables

1. **I1**: `DirNone` tiene valor distinto a `DirUp`, `DirDown`, `DirLeft`, `DirRight`, `DirPause`, `DirQuit` (verificable con `iota` secuencial).

2. **I2**: `ReadDirectionNonBlocking()` nunca retorna `DirUp`, `DirDown`, `DirLeft`, `DirRight` como resultado de timeout o tecla no mapeada — solo `DirNone`.

3. **I3**: `convertInputToGameDirection()` preserva la invariante: si `inputDir == input.DirNone`, el output es siempre `currentDir`.

4. **I4**: La dirección de la serpiente solo cambia cuando el usuario presiona una tecla de dirección válida (WASD o flechas) — nunca por ausencia de input.

5. **I5**: No hay magic numbers en el timeout de input — el valor 10ms debe ser constante `InputPollTimeout`.

## Criterios de aceptación

1. Jugar 30 segundos sin tocar el teclado: la serpiente mantiene su dirección actual (no se mueve hacia abajo automáticamente).

2. Ejecutar con `DEBUG=1` y no presionar teclas: los logs NO muestran eventos `input_converted` con dirección alguna — solo posibles `input_error` si se presionan teclas no mapeadas.

3. Presionar teclas no mapeadas (ej: 'x', Enter, Escape): la serpiente continúa en su dirección actual, no cambia hacia arriba ni hacia ninguna dirección por defecto.

4. Test unitario `TestReadDirectionNonBlocking_TimeoutReturnsDirNone` pasa: verifica que timeout retorna `DirNone`.

5. Test unitario `TestReadDirectionNonBlocking_UnknownKeyReturnsDirNone` pasa: verifica que tecla no mapeada retorna `DirNone` con log de error.

6. Test unitario `TestConvertInputToGameDirection_FallbackToCurrent` pasa: verifica que `DirNone` mapea a dirección actual.

7. Test de integración `TestGameLoop_MaintainsDirectionWithoutInput` pasa: simula 200 ticks sin input, verifica que snake no cambió de dirección.

8. `go build` compila sin errores y todos los tests existentes siguen pasando (no regression).

---

Status: Approved by Pablo Manuel Rizzo on 2026-05-05
