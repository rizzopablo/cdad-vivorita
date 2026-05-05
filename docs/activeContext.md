# Active Context

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

### Próxima feature en cola
Pendiente de priorización.
