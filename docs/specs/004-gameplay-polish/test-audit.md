# Test Audit Report — 004-gameplay-polish

**Status**: APPROVED
**Approved by**: Pablo Manuel Rizzo
**Approved at**: 2026-05-05
**Auditor**: test-writer (modalidad Test Audit)
**Date**: 2026-05-05

---

## Resumen de comportamiento que cambia

La feature 004-gameplay-polish introduce **tres cambios estructurales** al comportamiento del juego y la configuración de red:

1. **Ticker acelerado**: El game loop cambia de 150ms a 200ms (PC1, I1), lo que afecta el timing de render events loggeados en observability_test.go
2. **Constante InputPollTimeout**: El magic number 10ms se extrae a `const InputPollTimeout` en src/input/input.go (PC2, PC3). Sin impacto en tests existentes de comportamiento.
3. **Snake estática al inicio**: Más importante — la serpiente **NO se mueve** hasta recibir el primer input de dirección válida (PC4-PC7, I2, I3). Un nuevo flag `firstInputReceived` controla cuándo se invoca `g.Update()`. Esto afecta game_test.go y observability_test.go que asumen movimiento inmediato tras Update().

---

## Tests modificados

| Test | Archivo | Cambio | Justificación (spec ref) |
|------|---------|--------|--------------------------|
| `TestGame_MaintainsDirectionWithoutInput` | tests/game_test.go | **SE MODIFICA**: Actualmente verifica que la estructura del código puede mantener dirección. Con 004, el comportamiento cambia: sin input válido, g.Update() NO se invoca. Test debe validar que la dirección se mantiene NO porque se llama Update(DirNone), sino porque Update() se SALTA. | PC6: mientras `firstInputReceived == false`, g.Update() NO se invoca |
| `TestGame_DirectionPreservationBehavior` | tests/game_test.go | **SE MODIFICA**: Actual comportamiento es ejecutar Update(DirRight) 5 veces. Con 004, antes del primer input válido, ningún Update se ejecuta. Test debe primero enviar un input válido (simular primer input), luego verificar que direcciones posteriores se preservan. | PC5, PC7: primer input de dirección válida habilita Update(); PC9: post-primer-input comportamiento idéntico al anterior |
| `TestPostcondition3_LogUpdateEvents` | tests/observability_test.go | **SE MODIFICA**: Línea 145 ejecuta `g.Update(game.DirRight)` sin simular "primer input". Con 004, este Update se ejecutaría SOLO después del primer input. Test debe usar un setup que marque `firstInputReceived = true` primero, O integrar la lógica de "primer input" en el test E2E. | PC6, PC7: Update() solo se invoca post-primer-input |
| `TestPostcondition7_IntegrationFullFlow` | tests/observability_test.go | **SE MODIFICA**: Línea 317 ejecuta `g.Update(game.DirRight)` sin contexto de "primer input". Test debe simular: (1) inicio del game, (2) primer input válido recibido, (3) luego Update() y logging. O puede omitirse Update() si el test es puramente de logging structure. | PC6, PC7, PC8: Update() solo post-primer-input; render debe ocurrir sin input (tablero visible) |

---

## Tests nuevos a escribir

| Test | Postcondición(es) | Archivo | Descripción |
|------|------------------|---------|-------------|
| `TestGameLoop_TickerIs200ms` | PC1, I1 | tests/game_test.go | Verifica que el ticker en src/main.go usa 200ms, no 150ms. Busca línea `time.NewTicker(200 * time.Millisecond)`. |
| `TestInput_InputPollTimeoutConstantDefined` | PC2, I1 | tests/input_test.go | Verifica que `const InputPollTimeout = 10 * time.Millisecond` existe en src/input/input.go. |
| `TestInput_ReadDirectionUsesInputPollTimeout` | PC3, I1 | tests/input_test.go | Verifica que ReadDirectionNonBlocking() usa `InputPollTimeout` en `time.After()`, no el magic number 10ms literal. |
| `TestGameLoop_FirstInputReceivedFlagExists` | PC4, I2 | tests/game_test.go | Code review + behavioral: verifica que src/main.go declara `firstInputReceived := false` antes del game loop. |
| `TestGameLoop_FirstInputReceivedSetOnValidDirection` | PC5, I2 | tests/game_test.go | **E2E**: simula game loop sin input, envía primer input de dirección válida (DirRight), verifica que flag se asigna true. Puede requerir exposición del flag o medición indirecta (g.Update() se invoca). |
| `TestGameLoop_SnakeFrozenUntilFirstInput` | PC6, I3 | tests/game_test.go | **E2E + integration**: ejecuta game loop 10 ticks sin presionar teclas (input poll retorna DirNone), verifica que snake.Head() permanece en posición inicial. CRÍTICO: mide directamente el comportamiento observable. |
| `TestGameLoop_SnakeMoveAfterFirstInput` | PC7, I2, I3 | tests/game_test.go | **E2E**: game loop tick 0-4 sin input (snake estática), tick 5 presiona "d" (DirRight), tick 6 verifica snake.Head().X incrementó. Valida que primer input dispara Update(). |
| `TestGameLoop_BoardVisibleWithoutInput` | PC8, I3 | tests/observability_test.go | Con `DEBUG=1`, verifica que evento `render` ocurre en los primeros 200ms, ANTES de cualquier input. Prueba que RenderBoard() se ejecuta desde inicio. |
| `TestGameLoop_PostFirstInputBehaviorUnchanged` | PC9, I4 | tests/game_test.go | Compara: (1) snake movements, (2) food eating, (3) pause/quit entre versión "post-primer-input" y tests existentes sin cambios. Verifica que lógica posterior al primer input no cambió. |
| `TestGameLoop_NoWaitingForInputLogs` | PC10 | tests/observability_test.go | Con `DEBUG=1`, ejecuta 2 segundos sin input, verifica que log NO contiene "waiting_for_input", "polling", u otros eventos de espera. Observa que observability.LogEvent() NO loguea la espera. |

---

## Tests sin cambios (untouched)

| Test | Archivo | Razón |
|------|---------|-------|
| `TestConvertInputToGameDirection_DirNoneReturnsCurrentDir` | tests/game_test.go | Valida que convertInputToGameDirection(DirNone, currentDir) retorna currentDir. Feature 003 agregó currentDir param. **No impactado por 004**: la lógica de mapping DirNone→currentDir permanece idéntica. |
| `TestDirUp_NotFallbackForNoInput` | tests/game_test.go | Verifica que DirUp no es fallback por ausencia de input (feature 003). **No impactado por 004**: 004 solo agrega un flag booleano anterior a Update(). Conversion logic feature 003 no cambia. |
| `TestGame_DirNoneConstantExists` | tests/game_test.go | Verifica que game.DirNone existe. **No impactado**: constante no cambia por 004. |
| `TestGameRun_UsesDirNoneCorrectly` | tests/game_test.go | Verifica que Run() maneja DirNone preservando dirección (feature 003 contract). **No impactado por 004**: post-primer-input, lógica de DirNone permanece igual. |
| `TestReadDirectionNonBlocking_TimeoutReturnsDirNone` | tests/input_test.go | Valida que timeout retorna DirNone (feature 003). **No impactado**: 004 solo cambia donde se usa el timeout (constante vs magic number) pero returnValue permanece DirNone. |
| `TestReadDirectionNonBlocking_UnmappedKeyReturnsDirNoneWithLog` | tests/input_test.go | Verifica que teclas unmapped retornan DirNone + log input_error (feature 003). **No impactado por 004**: 004 no cambia input handling, solo ticker y primer-input flag. |
| `TestInput_DirNoneConstantExists` | tests/input_test.go | Verifica input.DirNone existe y tiene valor 6. **No impactado**: constante no cambia. |
| `TestPostcondition1_InitLogging` | tests/observability_test.go | Verifica que InitLogging() crea logs/ y vivorita2-debug.log (feature 002). **No impactado por 004**: logging infrastructure no cambia. |
| `TestPostcondition6_NoLogsWithoutDebug` | tests/observability_test.go | Verifica que sin DEBUG=1, logs/ no existe (feature 002). **No impactado**: logging flag behavior no cambia. |
| `TestPostcondition2_LogInputEvents` | tests/observability_test.go | Verifica que input_raw e input_converted se loguean (feature 002, feature 003). **No impactado por 004**: logging behavior pre-primer-input y post permanece igual. Input events se loguean siempre (incluso DirNone). |
| `TestPostcondition4_LogRenderEvents` | tests/observability_test.go | Verifica que evento render se loguea (feature 002). **No impactado por 004**: render sigue ocurriendo, solo timing cambia (200ms vs 150ms) pero evento log persiste. |
| `TestPostcondition5_SnapshotOnGameOver` | tests/observability_test.go | Verifica que SnapshotBoard() crea archivo en logs/ (feature 002). **No impactado por 004**: snapshot behavior no cambia. |
| `TestPostcondition8_MainImportsObservability` | tests/observability_test.go | Verifica que main.go importa observability y llama InitLogging() (feature 002). **No impactado**: imports y InitLogging() llamada permanecen, solo agregan líneas de manejo de firstInputReceived. |
| `TestPostcondition9_ReadDirectionUsesTcell` | tests/observability_test.go | Verifica que input.go usa tcell.PollEvent (feature 003). **No impactado por 004**: tcell integration no cambia. |
| `TestPostcondition10_InputLoggingInternal` | tests/observability_test.go | Verifica que input.go loguea input_raw, input_converted, input_error (feature 002, feature 003). **No impactado por 004**: logging behavior de input preservado. |
| `TestPostcondition11_UpdateLoggingInMain` | tests/observability_test.go | Verifica que main.go loguea evento "update" con direction, snake_head (feature 002). **Parcialmente impactado pero NO requiere cambio**: Update() ahora se ejecuta DESPUÉS del primer input. Test debe funcionar como-es porque solo verifica que **cuando** Update() ocurre, se loguea. Si el test no presimula input, fallará... revísase en sección siguiente. |
| `TestPostcondition12_RenderLoggingWithSource` | tests/observability_test.go | Verifica que main.go loguea "render" con fuentes ticker, update, gameover (feature 002). **No impactado directamente**: render eventos siguen ocurriendo con esos sources, solo timing de ticker cambia de 150ms a 200ms. |
| `TestPostcondition13_SnapshotOnGameoverInMain` | tests/observability_test.go | Verifica que main.go llama SnapshotBoard() en game over (feature 002). **No impactado**: SnapshotBoard() behavior no cambia. |

---

## Análisis de risk — Tests que parecen untouched pero requieren verificación

**Importante**: Los siguientes tests de observability pueden FALLAR tras implementación de 004 si no se revisan cuidadosamente:

- **TestPostcondition11_UpdateLoggingInMain** (línea 512): Ejecuta `g.Update(game.DirRight)` sin simular input. Con 004, si g.Update() es controlado por firstInputReceived, este test falla a menos que:
  - (a) El test sea unit-level y solo verifique que LogEvent() es **capaz** de loguear (no que se ejecutó en game loop), o
  - (b) El game loop exposure permite setear firstInputReceived = true antes de Update().

  **Acción recomendada**: Revisar si el test fallaría tras 004. Si es UNIT-level (verifica observability.LogEvent), marcar como untouched. Si es INTEGRATION-level (requiere g.Update() real), marcar para modificación.

---

## Regression risk assessment

### Riesgos DETECTADOS

1. **Timing regressions** (moderado): El cambio de ticker 150ms → 200ms puede afectar tests que esperan ciertos tiempos. Búsqueda de hardcoded timeouts en tests debe realizarse. Mitigación: Tests nuevos PC1 verifican explícitamente 200ms.

2. **Integration flow breakage** (moderado): Tests de observability que hacen Update() sin simular "primer input" fallarán si el flag `firstInputReceived` bloquea Update(). Mitigación: PC6 test verifica que snake está congelada sin input; test audit ya identificó 4 tests a modificar.

3. **Behavioral assumptions** (bajo): Tests existentes de game_test.go (feature 003) asumen que convertInputToGameDirection() + DirNone handling funciona siempre. Con 004, esa lógica solo se ejecuta **post-primer-input**. Si tests no contextualizan input, pueden pasar incorrectamente (positivos falsos). Mitigación: PC9 test verifica que comportamiento post-primer-input es idéntico.

4. **Logging behavioral drift** (bajo): PC10 especifica que NO hay logs de "waiting for input". Tests de observability que asumen logs específicos pueden sorprenderse si nuevo código loguea más. Mitigación: PC10 test verifica explícitamente que logs de espera no existen.

### Conclusión

**Severity**: MODERATE
**Confidence**: HIGH

La feature introduce cambios estructurales significativos (flag firstInputReceived). Tests existentes son mayormente **isolados de este cambio** (testing input mapping, observability infrastructure), pero:

- **4 tests requieren modificación** (observability_test líneas 145, 317 + game_test líneas 215-225)
- **10 tests nuevos deben escribirse** para cubrir PC1-PC10
- **1 test requiere auditoría de segunda pasada** (TestPostcondition11) luego de implementación

**No significant architectural risks detected**. El flag es interno al game loop (main.go). Interfaces públicas de `game.Game`, `input.ReadDirectionNonBlocking()`, `observability` no cambian.

---

## Apéndice — Matriz de cobertura

| Postcondición | Test | Estado |
|---------------|------|--------|
| PC1 | TestGameLoop_TickerIs200ms | NUEVO |
| PC2 | TestInput_InputPollTimeoutConstantDefined | NUEVO |
| PC3 | TestInput_ReadDirectionUsesInputPollTimeout | NUEVO |
| PC4 | TestGameLoop_FirstInputReceivedFlagExists | NUEVO |
| PC5 | TestGameLoop_FirstInputReceivedSetOnValidDirection | NUEVO |
| PC6 | TestGameLoop_SnakeFrozenUntilFirstInput | NUEVO |
| PC7 | TestGameLoop_SnakeMoveAfterFirstInput | NUEVO |
| PC8 | TestGameLoop_BoardVisibleWithoutInput | NUEVO |
| PC9 | TestGameLoop_PostFirstInputBehaviorUnchanged | NUEVO |
| PC10 | TestGameLoop_NoWaitingForInputLogs | NUEVO |
| I1 | PC1, PC2, PC3 tests | CUBIERTO |
| I2 | PC5 test | CUBIERTO |
| I3 | PC6, PC7, PC8 tests | CUBIERTO |
| I4 | PC9 test | CUBIERTO |

---

**AUDIT REPORT COMPLETE**

Próximo paso: Aprobación humana de este audit. Una vez aprobado, procede RED phase (test-writer escribe tests nuevos en rojo).
