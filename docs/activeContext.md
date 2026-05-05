# Active Context

## 2026-05-05 — Feature: 003-input-capture-investigation

**Estado**: Done — Ciclo CDAD completo (Etapa 5 cerrada)

### Resumen
Fix del bug de captura de input: la serpiente se movía sola hacia abajo porque `ReadDirectionNonBlocking()` retornaba `DirUp` en timeout. Se introdujo `DirNone` en ambos packages (`input` y `game`) y se modificó `convertInputToGameDirection` para usar la dirección actual como fallback.

### Decisiones relevantes
- **`DirNone` como sentinel**: Ambos packages (`input.Direction` y `game.Direction`) definen `DirNone` como valor de "sin dirección".
- **Fallback a dirección actual**: `convertInputToGameDirection` recibe `currentDir` y lo retorna cuando `inputDir == DirNone`.
- **`snake.Direction()` expuesto**: Nuevo método en `Snake` para exponer la dirección actual al game loop.

### Deuda técnica detectada
- **I5 pendiente**: `InputPollTimeout` como constante (magic number 10ms en `input.go:37`).
- **Duplicación de `convertInputToGameDirection`**: Existe en `game.go` (código muerto) y `main.go`.

### Próxima feature en cola: 004-gameplay-polish
1. Snake estática al inicio (no se mueve hasta primer input del jugador)
2. Velocidad ajustada (más lenta que 150ms, pero no excesivamente lenta)
3. Constante `InputPollTimeout` (I5 pendiente)

---

## 2026-05-05 — Feature: 002-game-loop-observability

**Estado**: Investigando — Datos de observabilidad recolectados, root cause pendiente de confirmación

### Indicios del problema (evidencia de logs)

**Patrón observado en 2 ejecuciones con DEBUG=1:**
- Snake se mueve automáticamente con dirección `DirUp` sin input del usuario
- No hay eventos `input_raw` ni `input_converted` en los logs (0 capturas)
- Game loop ejecuta normalmente: `init` → múltiples `update` → `game_over` en Y=20
- Snapshot generado correctamente al game over

**Hipótesis pendientes de verificación:**
1. `tcell.Screen.PollEvent()` no está recibiendo eventos KeyRune
2. Timeout de 10ms en PollEvent vs ticker de 150ms — posible race condition
3. Screen no configurado para raw mode / eventos no procesados
4. `select` en main.go prioriza `ticker.C` sobre input channel
5. Problema en cómo main.go consume `ReadDirectionNonBlocking()`

**Próximos pasos de investigación:**
- [ ] Agregar logging más granular en `input.go` (antes y después de PollEvent)
- [ ] Verificar si `PollEvent` retorna `nil` o eventos no-Key
- [ ] Testear `ReadDirectionNonBlocking` aislado con screen simulado
- [ ] Revisar select statement en `main.go` — ¿prioriza ticker sobre input?

**Nota**: Los logs prueban que el sistema de observabilidad (PC1-PC13) funciona correctamente. El problema está en la captura de input, no en el logging.

---

## 2026-05-05 — Feature: 002-game-loop-observability

**Estado**: Done — Ciclo CDAD completo (Etapa 5 cerrada)

### Decisiones relevantes
- **Package `observability`**: Nuevo módulo para logging estructurado JSON y snapshots de tablero, activado por variable de entorno `DEBUG=1`.
- **Migración de input a tcell.PollEvent()**: Reemplazo de `exec.Command("bash", ...)` por `tcell.Screen.PollEvent()` con timeout de 10ms, eliminando spawn de subprocesos bash.
- **Dependency injection para logging**: `input.LogEvent` se inyecta desde `main.go` (`input.LogEvent = observability.LogEvent`) para evitar import circular y mantener boundaries de módulo.
- **Integración en main.go**: `InitLogging()` al inicio, `LogEvent` en puntos críticos (input, update, render), `SnapshotBoard()` en game over.

### Deuda técnica detectada
- **Error silencioso en JSON marshaling** (`src/observability/observability.go:100`): `writeJSONLog()` ignora errores de `json.Marshal()` sin logging ni propagación. Si falla el marshal, la entrada no se escribe y no hay diagnóstico.
- **Duplicación de mapeo input → game direction**: `src/input/input.go:64-105` y `src/main.go:99-111` tienen lógica similar. `ReadDirectionNonBlocking()` ya mapea internamente, pero main.go re-mapea.
- **Divergencia spec-código en dependency injection**: El spec documenta `observability.LogEvent` como función global, pero la implementación usa variable injectable. El spec no documenta este mecanismo.
- **Patrón repetitivo en `handleKeyEvent`** (`src/input/input.go:54-123`): Switch con casos similares que loggean y retornan. Podría refactorizarse con helper `logAndReturn()`.

### Próxima feature en cola
Pendiente de priorización.

---

## 2026-05-05 — Feature: 001-snake-game

**Estado**: Done — Ciclo CDAD completo (Etapa 5 cerrada)

### Decisiones relevantes
- Estructura de paquetes Go: `game/`, `render/`, `input/`, `score/` separados por responsabilidad
- Sistema de coordenadas: cartesiano (Y crece hacia arriba), no convención de terminal
- Spawn de comida: `math/rand` con selección aleatoria entre posiciones libres
- High score: persiste en `~/.vivorita2_highscore.json`, se carga en `NewGame()`
- Biblioteca de UI: `tcell` para renderizado en terminal

### Deuda técnica detectada
- **Coordenada Y invertida**: `DirUp` incrementa Y, `DirDown` decrementa Y (líneas 37-40 en snake.go). Esto invierte el movimiento visual respecto a convenciones de terminal. render.go debe transformar coordenadas o el usuario percibe dirección invertida.
- **Magic numbers**: dimensiones 40x20 hardcodeadas en game.go:124 (`head.X > 39 || head.Y > 19`). Deberían ser constantes (`BoardWidth`, `BoardHeight`).
- **PC4 coverage parcial**: tests solo cubren colisión en bordes X, no en Y (ver review.md).
- **main.go usa `~` sin expandir**: línea 24 pasa `~/.vivorita2_highscore.json` literal a `NewGameWithHighScore`, que no expande el tilde. `NewGame()` sí funciona correctamente.