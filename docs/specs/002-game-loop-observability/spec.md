---
id: "002-game-loop-observability"
title: "Game Loop Observability"
status: "approved"
approved_by: "Pablo Manuel Rizzo"
approved_at: "2026-05-05"
extended_at: "2026-05-05"
extended_by: "Pablo Manuel Rizzo"
---

# Feature: Game Loop Observability

## Descripción funcional

Agrega mecanismos de observabilidad al game loop para que agentes IA puedan diagnosticar problemas sin ejecutar el código manualmente. La feature instrumenta los puntos críticos del flujo `input → update → render` con logging estructurado a archivos en `./logs/`, activado por la variable de entorno `DEBUG=1`. Además, captura snapshots del tablero en eventos de error/crash para facilitar diagnóstico post-mortem. Incluye tests de integración que verifican el flujo real completo, dado que los tests unitarios pasan pero el ejecutable real no responde a inputs.

Status: Approved by Pablo Manuel Rizzo on 2026-05-05
Extended with integration postconditions (PC8-PC13) approved by Pablo Manuel Rizzo on 2026-05-05

## Contrato

### Firma pública

```go
// Inicialización del sistema de logging (llamar al inicio de main)
func InitLogging() error

// Log de evento en el game loop (usa JSON por línea)
func LogEvent(event string, data map[string]interface{})

// Snapshot del estado del tablero en error/crash
func SnapshotBoard(g *Game, reason string) error

// Lectura de input no bloqueante usando tcell (reemplaza exec.Command bash)
func ReadDirectionNonBlocking(screen tcell.Screen) (Direction, error)

// Variable de entorno: DEBUG=1 habilita, cualquier otro valor deshabilita
```

### Postcondiciones

1. **PC1**: Si la variable de entorno `DEBUG=1` está seteada, `InitLogging()` crea el directorio `./logs/` si no existe y abre el archivo de log.
   - Input: `os.Setenv("DEBUG", "1")` antes de iniciar el juego
   - Output: Existe `./logs/vivorita2-debug.log` con entrada inicial
   - Error: Si no puede crear directorio o archivo, retorna error

2. **PC2**: Con `DEBUG=1`, cada lectura exitosa de `input.ReadDirectionNonBlocking()` loggea el char crudo y la `Direction` convertida.
   - Input: Tecla presionada (ej: 'd' para derecha)
   - Output: Log entries con `event:"input_raw"` (char leído) y `event:"input_converted"` (Direction resultante)
   - Error: Si `ReadDirectionNonBlocking()` falla, log con `event:"input_error"`

3. **PC3**: Con `DEBUG=1`, cada llamada a `game.Update(dir)` loggea la dirección recibida y el estado después de actualizar.
   - Input: `game.Update(game.DirRight)`
   - Output: Log entry con `event:"update"`, `direction`, `snake_head:{X,Y}`, `score`, `over`, `paused`
   - Error: N/A (Update no retorna error)

4. **PC4**: Con `DEBUG=1`, cada llamada a `render.RenderBoard()` loggea que se ejecutó.
   - Input: Llamada a `render.RenderBoard(screen, snake, food, score, highScore)`
   - Output: Log entry con `event:"render"`, `timestamp`
   - Error: N/A

5. **PC5**: En eventos de game over o error, se guarda un snapshot del tablero en `./logs/`.
   - Input: `game.IsOver() == true` o panic
   - Output: Archivo `./logs/board-snapshot-<timestamp>.json` con `snake_segments`, `food_position`, `score`, `high_score`
   - Error: Si no puede escribir snapshot, log de error

6. **PC6**: Si `DEBUG` no está seteado o es distinto de "1", NO se escribe nada en `./logs/`.
   - Input: Sin variable DEBUG, o `DEBUG=0`
   - Output: `./logs/` permanece sin cambios
   - Error: N/A

7. **PC7**: Test de integración verifica el flujo completo `input → update → render` simulando inputs y verificando cambios de estado.
   - Input: Secuencia simulada de inputs que activen el flujo real
   - Output: Snake se movió correctamente, logs registran el flujo completo
   - Error: N/A

8. **PC8**: `main()` importa `"vivorita2/src/observability"` y llama `observability.InitLogging()` tras crear el juego (`src/main.go:24`).
   - Input: Inicio de `main()`, después de `game.NewGameWithHighScore(...)`
   - Output: `InitLogging()` se ejecuta antes del game loop (línea 31). Si retorna error, `panic(err)`.
   - Error: Sin `InitLogging()`, no se crean logs ni snapshots.
   - Verificación: `go build` compila sin errores de import; ejecutar con `DEBUG=1` produce `./logs/vivorita2-debug.log` con entrada `event:"init"`.

9. **PC9**: `input.ReadDirectionNonBlocking()` usa `tcell.Screen.PollEvent()` en lugar de `exec.Command("bash", ...)`.
   - Input: `screen.PollEvent(10 * time.Millisecond)` en `src/input/input.go`, reemplazando la línea 33.
   - Output: `ReadDirectionNonBlocking(screen tcell.Screen) (Direction, error)`. Mapeo: `tcell.KeyRune` + `ev.Rune()` → `w/s/a/d/q/p` → `Direction`; `tcell.KeyESC` / `tcell.KeyCtrlC` → `DirQuit`. Timeout (`nil`) → `DirUp, nil` (sin input). Tecla no mapeable → `DirUp, nil` con `event:"input_error"` loggeado.
   - Error: Si `PollEvent()` retorna un evento no-key, se ignora silenciosamente.
   - Verificación: `go build` compila; `strace -f -e trace=execve ./vivorita2` NO muestra spawns de bash; `go test` pasa.

10. **PC10**: `ReadDirectionNonBlocking()` loggea eventos `input_raw` / `input_converted` / `input_error`.
    - Input: Cada llamada a `ReadDirectionNonBlocking(screen)` en `src/input/input.go`.
    - Output:
      - Tecla reconocida: `event:"input_raw"` con `data:{"key": "<KeyRune|KeyESC|...>", "rune": "<carácter>"}` seguido de `event:"input_converted"` con `data:{"direction": "<DirUp|DirDown|DirLeft|DirRight|DirPause|DirQuit>"}`.
      - Timeout (sin input): NO se loggea nada.
      - Tecla no mapeable: `event:"input_error"` con `data:{"key": "<desconocida>", "rune": "<carácter>"}`.
    - Error: Si se lee una tecla pero no se loggea `input_raw` + `input_converted`, la postcondición falla.
    - Verificación: Con `DEBUG=1`, presionar 'd' genera dos entradas consecutivas: `input_raw` con `rune:"d"` e `input_converted` con `direction:"DirRight"`.

11. **PC11**: `game.Update()` loggea evento `update`.
    - Input: Cada llamada a `g.Update(gameDir)` en `src/main.go:55` (dentro del bloque `default:` del game loop).
    - Output: Inmediatamente después de `g.Update(gameDir)`, se llama `observability.LogEvent("update", data)` con `data:{"direction": "<DirUp|DirDown|DirLeft|DirRight>", "snake_head": {"X": int, "Y": int}, "score": int, "over": bool, "paused": bool}`.
    - Error: Si `Update()` se ejecuta pero no hay entrada `event:"update"` en el log, la postcondición falla.
    - Verificación: Con `DEBUG=1`, cada tick que procesa un input genera una entrada `update` con los campos especificados.

12. **PC12**: `render.RenderBoard()` loggea evento `render`.
    - Input: Cada llamada a `render.RenderBoard()` en tres puntos de `src/main.go`: línea 35 (`case <-ticker.C`), línea 57 (`default:` tras `g.Update()`), línea 66 (game over).
    - Output: Inmediatamente después de cada `render.RenderBoard(...)`, se llama `observability.LogEvent("render", data)` con `data:{"source": "ticker|update|gameover"}`. El campo `source` distingue cuál de los tres call sites invocó el render.
    - Error: Si `RenderBoard()` se ejecuta pero no hay entrada `event:"render"` correspondiente, la postcondición falla.
    - Verificación: Con `DEBUG=1`, el log contiene entradas `render` con `source` diferenciando los tres call sites.

13. **PC13**: Al detectar game over en `main()`, se llama `SnapshotBoard()`.
    - Input: Bloque de detección de game over en `src/main.go:64-70` (`if g.IsOver() { ... }`).
    - Output: Antes de `render.RenderBoard()` (línea 66) y `time.Sleep()` (línea 68), se llama `observability.SnapshotBoard(g, "game_over")`. Si retorna error, se loggea pero NO se interrumpe el flujo.
    - Error: Si `g.IsOver() == true` pero no se genera `./logs/board-snapshot-<timestamp>.json`, la postcondición falla.
    - Verificación: Con `DEBUG=1`, provocar game over genera el archivo snapshot con campos `snake_segments`, `food_position`, `score`, `high_score`, `reason:"game_over"`.

## Invariantes verificables

1. **I1**: Los logs en `./logs/vivorita2-debug.log` son JSON válido (un objeto JSON por línea).
2. **I2**: Cada evento de log tiene los campos obligatorios: `timestamp` (ISO 8601), `event` (string), `data` (object).
3. **I3**: El snapshot del tablero contiene: `snake_segments []Position`, `food_position Position`, `score int`, `high_score int`, `reason string`.
4. **I4**: `Game.Run()` en `src/game/game.go:62-97` permanece sin cambios (código muerto, el loop real es `main()`).
5. **I5**: `src/input/input.go` NO importa `"os/exec"` ni `"runtime"` tras la migración a tcell.

## Criterios de aceptación

1. Al ejecutar con `DEBUG=1 ./vivorita2`, el archivo `./logs/vivorita2-debug.log` se crea y contiene entradas JSON para cada input, update y render del ciclo.
2. Al provocar game over (chocar pared), se genera `./logs/board-snapshot-<timestamp>.json` con estado completo.
3. Sin `DEBUG=1`, la ejecución no crea ni modifica archivos en `./logs/`.
4. El test de integración pasa: simula secuencia de inputs → verifica que `Update()` cambió el estado → verifica que `RenderBoard()` fue llamado (mediantes logs).
5. Un agente IA leyendo `./logs/vivorita2-debug.log` puede determinar por qué un input no fue procesado (ej: tecla no mapeada, o dirección no llega a `Update()`).
6. El input system (`ReadDirectionNonBlocking()`) queda instrumentado: se loggea la tecla tcell leída, el carácter, y la conversión a `Direction`.
7. El juego NO spawnea subprocess bash durante la ejecución (verificable con `strace -f -e trace=execve`).
8. `ReadDirectionNonBlocking()` acepta `tcell.Screen` como parámetro y usa `PollEvent()` con timeout de 10ms.
