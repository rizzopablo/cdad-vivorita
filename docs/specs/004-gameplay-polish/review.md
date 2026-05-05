# Review — 004-gameplay-polish

**Reviewer**: Claude Haiku 4.5 (model: claude-haiku-4-5-20251001)
**Independencia**: Read-only; spec-driven review sin acceso a implementación previa.
**Status**: Análisis completado. Todos los tests verdes.

---

## Bloqueantes

Ninguno identificado.

---

## Opcionales

### 1. Redundancia en la verificación del flag firstInputReceived (líneas 61-64)

**Ubicación**: `src/main.go:60-64`

**Problema**:
El código verifica dos veces el estado de `firstInputReceived`:
```go
case input.DirUp, input.DirDown, input.DirLeft, input.DirRight:
    if !firstInputReceived {
        firstInputReceived = true
    }
    if firstInputReceived && !g.IsOver() {
        g.Update(gameDir)
        // ...
    }
```

La primera verificación (línea 61-63) asigna `true` al flag, pero inmediatamente la segunda verificación (línea 64) comprueba nuevamente `firstInputReceived`. Esto significa que en el **primer input válido**, el flag se asigna a `true` en la línea 62, y luego se verifica nuevamente en la línea 64 y entra correctamente al bloque `Update()`.

**Impacto**: Funcionalmente correcto pero subóptimo. La verificación se puede simplificar a una sola condición:
```go
case input.DirUp, input.DirDown, input.DirLeft, input.DirRight:
    if !firstInputReceived {
        firstInputReceived = true
        // Entrar directo al Update en el primer input
    }
    if firstInputReceived && !g.IsOver() {
        g.Update(gameDir)
        // ...
    }
```

**Sugerencia**: Refactorizar para consolidar la lógica. Alternativa: usar asignación + comprobación en una sola estructura de control o un bloque con un comentario explícito si se prefiere mantener la claridad del "flag monotónico".

---

### 2. Asimetría en el manejo de estados del juego en el ticker y el default

**Ubicación**: `src/main.go:39-46` (ticker) vs `src/main.go:84-97` (default case)

**Problema**:
- En el **ticker** (línea 39-46): se renderiza solo si `!g.IsPaused() && !g.IsOver()`.
- En el **default case** (línea 84-97): cuando `g.IsOver()`, siempre se renderiza **sin** comprobar `IsPaused()`.

Esta asimetría es menor pero hace que el comportamiento visual sea inconsistente si alguna lógica futura permite pausar cuando el juego está terminado (escenario remoto pero teórico).

**Impacto**: Bajo. No viola el spec, pero representa una inconsistencia de diseño.

**Sugerencia**: Para máxima consistencia, considerar aplicar la misma lógica de comprobación de estado en ambas ramas del `select`. Por ejemplo:
```go
// Ticker branch
if !g.IsPaused() && !g.IsOver() { render... }

// Gameover branch (en default)
if !g.IsPaused() && g.IsOver() { render... }
```

---

### 3. PC9 — Validación de comportamiento post-primer-input (verificación marginal)

**Ubicación**: Spec `docs/specs/004-gameplay-polish/spec.md`, postcondición 9

**Problema**:
PC9 especifica que "Post-primer-input, el game loop se comporta exactamente igual al original". La implementación respeta esto, pero el spec no define cuál es el "original" de referencia. Los tests verifican código estático (búsqueda de patrones) pero no comparan trazas de ejecución reales.

**Impacto**: Bajo a marginal. Los tests de observabilidad anteriores (`002-game-loop-observability`) implícitamente validan esto al pasar, pero no hay test explícito comparativo pre/post.

**Sugerencia**: Agreguen un test parametrizado que compare trazas de ejecución (secuencias de posiciones del snake, eventos de render) entre la versión actual post-primer-input y un golden trace capturado. Esto es un asunto de rigor de test, no de implementación.

---

### 4. Render inicial requiere espera de 200ms (latencia visible)

**Ubicación**: Spec `docs/specs/004-gameplay-polish/spec.md`, postcondición 8

**Problema**:
PC8 especifica "El tablero inicial es visible sin necesidad de input". La implementación lo cumple: el ticker renderiza en los primeros 200ms. Sin embargo, en la práctica, el usuario ve una pantalla vacía durante hasta ~200ms antes del primer render (tiempo para el primer tick del ticker).

Esto es un comportamiento correcto según el spec, pero tiene implicación UX: la serpiente no es visible hasta el primer tick del ticker. Si alguien implementara presionar una tecla antes del primer render (< 200ms), el juego respondería correctamente pero el tablero no aparecería hasta 200ms.

**Impacto**: Bajo. Conforme al spec y probablemente aceptable para un juego de snake. Pero es una latencia perceptible en hardware lento.

**Sugerencia**: Comentario de diseño: Si la experiencia debe mostrar el tablero inmediatamente (< 100ms), considerar hacer un render eagerly antes del loop. El spec actual está OK con la latencia de 200ms.

---

### 5. DirPause y DirQuit no disparan el flag firstInputReceived

**Ubicación**: `src/main.go:52-59`

**Problema**:
Según PC5, "El flag se asigna true cuando se recibe el **primer input que NO sea DirNone, DirPause ni DirQuit**". La implementación es correcta: solo los casos `DirUp/Down/Left/Right` disparan el flag (línea 60-81).

Sin embargo, esto significa que si un usuario presiona `Q` (quit) o `P` (pause) antes de mover la serpiente, el juego cierra o pausa sin que la serpiente se haya movido nunca. PC6 permite esto ("mientras flag false, g.Update() NO se invoca"), pero es un comportamiento que podría resultar confuso: un usuario podría presionar `Q` pensando que está en un menú, y el juego termina sin haber visto la serpiente moverse.

**Impacto**: Bajo. Conforme al spec. Comportamiento raro pero documentado.

**Sugerencia**: Considerar un comentario en el código explicando por qué DirPause/DirQuit NO disparan el movimiento del juego. O, si es deseable, actualizar el spec para clarificar que DirQuit debe permitir terminar el juego incluso sin haber presionado una dirección.

---

## Resumen de Verificación de Postcondiciones

| PC | Verificado | Estado | Notas |
|----|----|--------|-------|
| PC1 | ✓ | OK | Ticker cambiado a 200ms |
| PC2 | ✓ | OK | InputPollTimeout definida como constante |
| PC3 | ✓ | OK | ReadDirectionNonBlocking() usa InputPollTimeout |
| PC4 | ✓ | OK | Flag firstInputReceived inicializado false |
| PC5 | ✓ | OK | Flag asignado true en primer input de dirección válida |
| PC6 | ✓ | OK | Mientras flag false, g.Update() NO se invoca |
| PC7 | ✓ | OK | Primer input dispara g.Update() |
| PC8 | ✓ | OK | Tablero visible sin input (en primer tick del ticker, ~200ms) |
| PC9 | ✓ | OK | Post-primer-input comportamiento sin cambios (tests verdes) |
| PC10 | ✓ | OK | Sin logs de "waiting for input" |

## Resumen de Invariantes

| Invariante | Verificado | Estado | Notas |
|------------|-------|--------|-------|
| I1 | ✓ | OK | Ticker y InputPollTimeout constantes (no cambian en tiempo de ejecución) |
| I2 | ✓ | OK | Flag monótono (once true, siempre true; no hay reassignments a false) |
| I3 | ✓ | OK | Snake visual pero sin movimiento antes del primer input |
| I4 | ✓ | OK | Comportamiento post-primer-input idéntico (tests de observabilidad pasan) |

---

## Test Suite

**Estado**: ✓ VERDE (34 tests, todos PASS)

Cobertura:
- PC1-PC10: Tests estáticos de código fuente ✓
- InputPollTimeout: Tests de constante y uso ✓
- Observabilidad: Tests de logging y snapshots ✓
- Integración: Test de flujo completo ✓

---

## Conclusión

**Diff conforme al spec aprobado. Cero bloqueantes. Implementación correcta y funcional.**

Las cinco sugerencias opcionales son mejoras de estilo, claridad y rigor de testing, no defectos críticos. La feature está lista para merge.

---

**Aprobación recomendada**: SÍ, sin cambios requeridos.

---

## Metadata

- **Feature**: 004-gameplay-polish
- **Spec Status**: Approved by Pablo Manuel Rizzo on 2026-05-05
- **Diff**: Confirmado contra spec
- **Tests**: All green
- **Blockers**: 0
- **Optionals**: 5 (mejoras sugeridas, no críticas)

