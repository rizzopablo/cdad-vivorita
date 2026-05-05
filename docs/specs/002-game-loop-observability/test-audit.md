# Test Audit Report — 002-game-loop-observability (EXTENDED)

**Feature**: Game Loop Observability (vivorita2)
**Spec version**: Extended with PC8-PC13, Approved 2026-05-05
**Test Writer**: test-writer (CDAD AUDIT phase)
**Audit completed**: 2026-05-05
**Status**: Approved by Pablo Manuel Rizzo on 2026-05-05

---

## Comportamiento que cambia (resumen)

El spec original (PC1-PC7) definía el sistema de observabilidad como un package `observability` con funciones puras. El spec extendido agrega 6 postcondiciones de integración (PC8-PC13) que requieren modificar código existente del juego:

1. **PC8**: `main()` debe importar `"vivorita2/src/observability"` y llamar `observability.InitLogging()` tras crear el juego (`src/main.go:24`).
2. **PC9**: `input.ReadDirectionNonBlocking()` cambia de firma: de `ReadDirectionNonBlocking() (Direction, error)` a `ReadDirectionNonBlocking(screen tcell.Screen) (Direction, error)`. Reemplaza `exec.Command("bash", ...)` por `tcell.Screen.PollEvent(10 * time.Millisecond)`.
3. **PC10**: `ReadDirectionNonBlocking()` debe loggear eventos `input_raw` / `input_converted` / `input_error` internamente (no solo el caller).
4. **PC11**: `src/main.go:55` — tras `g.Update(gameDir)`, se llama `observability.LogEvent("update", data)`.
5. **PC12**: `src/main.go` — tras cada `render.RenderBoard()` (3 call sites: línea 35, 57, 66), se llama `observability.LogEvent("render", data)` con campo `source`.
6. **PC13**: `src/main.go:64-70` — al detectar `g.IsOver()`, se llama `observability.SnapshotBoard(g, "game_over")` antes del render final.

**Impacto en tests existentes**: Los tests de PC1-PC7 en `tests/observability_test.go` NO llaman directamente a `ReadDirectionNonBlocking()`, por lo que el cambio de firma NO los rompe directamente. Sin embargo, el test de integración PC7 y el futuro test de PC9 necesitarán adaptación.

---

## Tests modificados (PC1-PC7)

### TestPostcondition2_LogInputEvents
- **Cambio**: Actualizar para verificar que `ReadDirectionNonBlocking(screen)` loggea internamente (PC10), no solo que `LogEvent()` puede escribir.
- **Spec ref**: PC10 (líneas 89-96 del spec) — "ReadDirectionNonBlocking() loggea eventos input_raw / input_converted / input_error"
- **Nueva expectativa**: El test debe verificar que al llamar `ReadDirectionNonBlocking(screen)` con una tecla reconocida, se generan DOS entradas consecutivas en el log: `input_raw` con `rune:"d"` e `input_converted` con `direction:"DirRight"`.
- **Porqué**: PC10 exige que el logging sea interno a `ReadDirectionNonBlocking()`, no externo. El test actual simula el logging manualmente con `observability.LogEvent()`, lo cual no verifica el comportamiento real de la función.
- **Nota**: Requiere mock de `tcell.Screen` o uso de `testscreen` para simular input.

### TestPostcondition7_IntegrationFullFlow
- **Cambio**: Actualizar para usar `ReadDirectionNonBlocking(screen)` con firma nueva y verificar flujo real `input → update → render` con logging automático.
- **Spec ref**: PC7 (líneas 72-75) + PC9 (líneas 87-88) + PC10 (líneas 89-96)
- **Nueva expectativa**: El test debe crear un screen simulado, llamar `ReadDirectionNonBlocking(screen)`, verificar que el snake se mueve, y que los logs contienen `input_raw`, `input_converted`, `update`, y `render` generados automáticamente por el código integrado.
- **Porqué**: El test actual simula todos los logs manualmente. Con PC8-PC13, el logging es automático en los puntos de integración. El test debe verificar el flujo real, no la simulación.

---

## Tests untouched (PC1-PC7)

Estos tests se mantienen EXACTAMENTE como están porque validan comportamiento que no cambia con la integración:

1. `TestPostcondition1_InitLogging` — PC1: `InitLogging()` crea directorio y archivo de log. No cambia.
2. `TestPostcondition3_LogUpdateEvents` — PC3: Verifica formato de log de `update`. El formato de datos no cambia con PC11, solo se agrega el call site en main.go.
3. `TestPostcondition4_LogRenderEvents` — PC4: Verifica formato de log de `render`. El formato no cambia con PC12, solo se agrega campo `source` y call sites en main.go.
4. `TestPostcondition5_SnapshotOnGameOver` — PC5: `SnapshotBoard()` genera archivo snapshot. No cambia con PC13, solo se agrega el call site en main.go.
5. `TestPostcondition6_NoLogsWithoutDebug` — PC6: Sin `DEBUG=1`, no se escribe en `./logs/`. No cambia.

**Importancia**: Estos 5 tests validan el package `observability` en aislamiento. Las PC8-PC13 agregan call sites en `main.go` e `input.go`, pero no cambian el comportamiento de las funciones del package observability.

---

## Tests nuevos a escribir (PC8-PC13)

Cada postcondición de integración requiere un test nuevo:

- `test_postcondition_8_main_imports_observability` — PC8: Verificar que `main()` importa observability y llama `InitLogging()` antes del game loop. Test de compilación + ejecución con `DEBUG=1` verificando entrada `event:"init"` en log.
- `test_postcondition_9_read_direction_uses_tcell` — PC9: Verificar que `ReadDirectionNonBlocking(screen tcell.Screen)` usa `PollEvent()` con timeout de 10ms. Test con screen simulado: enviar tecla 'd', verificar que retorna `DirRight`. Verificar que `src/input/input.go` NO importa `"os/exec"` ni `"runtime"` (Invariante I5).
- `test_postcondition_10_input_logging_internal` — PC10: Verificar que `ReadDirectionNonBlocking()` loggea internamente `input_raw` + `input_converted` para tecla reconocida, y `input_error` para tecla no mapeable. Con `DEBUG=1`, presionar 'd' genera dos entradas consecutivas.
- `test_postcondition_11_update_logging_in_main` — PC11: Verificar que tras `g.Update(gameDir)` en `main.go`, se loggea evento `update` con campos `direction`, `snake_head`, `score`, `over`, `paused`. Test de integración con screen simulado.
- `test_postcondition_12_render_logging_with_source` — PC12: Verificar que cada `render.RenderBoard()` en los 3 call sites de `main.go` genera log `render` con campo `source` diferenciando `ticker`, `update`, `gameover`.
- `test_postcondition_13_snapshot_on_gameover_in_main` — PC13: Verificar que al detectar `g.IsOver()` en `main.go`, se llama `SnapshotBoard(g, "game_over")` y se genera archivo `./logs/board-snapshot-<timestamp>.json` con campos requeridos.

---

## Regression Risk Assessment

### Cambio de firma pública — `ReadDirectionNonBlocking()`

| Aspecto | Riesgo | Mitigación |
|---------|--------|------------|
| **Firma cambia de `()` a `(screen tcell.Screen)`** | **ALTO** — Cualquier caller existente se rompe. El único caller actual es `src/main.go:40`. | El caller en main.go se actualiza como parte de PC9. Tests de PC1-PC7 NO llaman esta función directamente. |
| **`os/exec` y `runtime` se eliminan de input.go** | **MEDIO** — Invariante I5 lo exige, pero si algún otro módulo importa estas funciones desde input.go, se rompe. | Verificar que ningún otro archivo importa desde `vivorita2/src/input` funciones que dependan de `os/exec`/`runtime`. |
| **`ReadDirectionNonBlocking` ahora requiere screen no-nil** | **MEDIO** — Tests que pasen `nil` como screen panicarán. | Los tests nuevos deben usar screen simulado. Tests untouched no llaman esta función. |
| **Logging interno en ReadDirectionNonBlocking** | **BAJO** — Si `DEBUG` no está seteado, el logging es no-op. Comportamiento existente preservado. | PC6 ya verifica que sin `DEBUG=1` no se escribe nada. |
| **3 nuevos call sites de LogEvent en main.go** | **BAJO** — Son adiciones, no reemplazos. El game loop existente sigue funcionando. | Tests untouched de PC3, PC4, PC5 verifican formato de log. Tests nuevos PC11-PC13 verifican call sites. |
| **SnapshotBoard llamado en main.go** | **BAJO** — Es adición antes del render final. Si retorna error, se loggea pero no interrumpe. | PC5 ya verifica formato de snapshot. PC13 verifica call site. |

### Cobertura de tests

- ✅ PC1-PC7: Cubiertas por tests existentes (5 untouched + 2 modificados)
- ✅ PC8-PC13: 6 tests nuevos propuestos
- ⚠️ **Riesgo residual**: Los tests de PC8, PC11, PC12, PC13 requieren ejecutar `main()` o simular el game loop completo. Esto puede ser frágil si el test framework no soporta bien la simulación de `tcell.Screen`. Se recomienda usar `testscreen` o mock de la interface `tcell.Screen`.

### Invariantes

- ✅ I1 (JSON válido): Cubierto implícitamente por todos los tests que leen logs
- ✅ I2 (campos obligatorios): Cubierto por tests de formato (PC2, PC3, PC4)
- ✅ I3 (snapshot fields): Cubierto por PC5 y PC13
- ✅ I4 (Game.Run() sin cambios): Cubierto por tests untouched en `src/game/snake_test.go`
- ✅ I5 (sin os/exec ni runtime en input.go): Cubierto por test nuevo PC9

---

## Gate de Test Audit

- [x] Cada test modificado está justificado en spec.md con referencia explícita (2 tests: PC2, PC7)
- [x] No hay test modificado sin justificación documentada
- [x] Tests untouched están listados explícitamente (5 tests de PC1-PC7)
- [x] Tests nuevos listados para cada postcondición PC8-PC13 (6 tests)
- [x] Regression risk assessment completado (cambio de firma pública documentado)
- [x] Humano aprobó este report antes de pasar a RED — Approved by Pablo Manuel Rizzo on 2026-05-05
