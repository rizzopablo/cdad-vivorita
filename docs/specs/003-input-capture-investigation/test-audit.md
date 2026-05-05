# Test Audit Report: Input Capture Bug Fix

**Feature:** 003-input-capture-investigation  
**Audit Date:** 2026-05-05  
**Auditor:** test-writer (CDAD Phase 3.0)

---

## Resumen de comportamiento que cambia

El spec introduce una nueva constante `DirNone` tanto en `input` como en `game` packages para representar la ausencia de input. Esto cambia el comportamiento actual donde:

1. **Cambio de fallback**: Actualmente `ReadDirectionNonBlocking` retorna `DirUp` en timeout y teclas no mapeadas. El nuevo comportamiento debe retornar `DirNone`.
2. **Cambio en conversión**: `convertInputToGameDirection` actualmente retorna `DirUp` para inputs no reconocidos. Debe retornar la dirección actual cuando recibe `DirNone`.
3. **Nueva invariante**: La serpiente debe mantener su dirección cuando no hay input válido (PC4).

---

## Tests existentes afectados

### Tests en `tests/observability_test.go`:

**Análisis**: Los tests existentes verifican observabilidad (logging) y estructura del código. Ninguno de ellos verifica el comportamiento de fallback de dirección ni el game loop de input.

| Test | Relación con spec | Decisión | Justificación |
|------|------------------|----------|---------------|
| `TestPostcondition6_NoLogsWithoutDebug` | No relacionado | **Untouched** | Solo verifica directorio de logs, no toca input/game loop |
| `TestPostcondition1_InitLogging` | No relacionado | **Untouched** | Solo verifica inicialización de logging |
| `TestPostcondition2_LogInputEvents` | Parcialmente relacionado | **Untouched** | Verifica logging de input pero no el valor de retorno de dirección. El test sigue siendo válido porque verifica que se loguee, no qué dirección se retorna. |
| `TestPostcondition3_LogUpdateEvents` | No relacionado | **Untouched** | Verifica logging de update events |
| `TestPostcondition4_LogRenderEvents` | No relacionado | **Untouched** | Verifica logging de render events |
| `TestPostcondition5_SnapshotOnGameOver` | No relacionado | **Untouched** | Verifica snapshots en game over |
| `TestPostcondition7_IntegrationFullFlow` | Parcialmente relacionado | **Untouched** | Verifica flujo integrado pero no verifica el valor específico de dirección fallback. Sigue siendo válido. |
| `TestPostcondition8_MainImportsObservability` | No relacionado | **Untouched** | Verifica imports |
| `TestPostcondition9_ReadDirectionUsesTcell` | No relacionado | **Untouched** | Verifica uso de tcell |
| `TestPostcondition10_InputLoggingInternal` | Parcialmente relacionado | **Untouched** | Verifica logging interno, no comportamiento de fallback |
| `TestPostcondition11_UpdateLoggingInMain` | No relacionado | **Untouched** | Verifica logging de update |
| `TestPostcondition12_RenderLoggingWithSource` | No relacionado | **Untouched** | Verifica logging de render |
| `TestPostcondition13_SnapshotOnGameoverInMain` | No relacionado | **Untouched** | Verifica snapshot en game over |
| `convertInputToGameDirection` (helper en test) | **AFECTADO** | **Must be ignored** | Esta función helper en el test NO debe modificarse según reglas del rol. Es solo para el test, no código de producción. |

**Conclusión**: **0 tests existentes necesitan modificación**. Todos los tests de observabilidad siguen siendo válidos porque no verifican el comportamiento específico de fallback de dirección que está cambiando.

---

## Tests nuevos a escribir

| Postcondición | Nombre del test | Archivo | Descripción |
|---------------|-----------------|---------|-------------|
| PC1 | `TestReadDirectionNonBlocking_TimeoutReturnsDirNone` | tests/input_test.go | Verifica que timeout retorne DirNone, no DirUp |
| PC2 | `TestReadDirectionNonBlocking_UnmappedKeyReturnsDirNoneWithLog` | tests/input_test.go | Verifica que tecla no mapeada retorne DirNone y loguee input_error |
| PC3 | `TestConvertInputToGameDirection_DirNoneReturnsCurrentDir` | tests/game_test.go | Verifica que DirNone preserve dirección actual |
| PC4 | `TestGame_MaintainsDirectionWithoutInput` | tests/game_test.go | Test de integración: 200 ticks sin input mantienen dirección |
| PC5 | `TestDirUp_NotFallbackForNoInput` | tests/game_test.go | Verifica que DirUp no sea el fallback en ausencia de input |
| PC6 | `TestGame_DirNoneConstantExists` | tests/game_test.go | Compile-time check: game.DirNone existe |
| PC7 | `TestInput_DirNoneConstantExists` | tests/input_test.go | Compile-time check: input.DirNone existe |
| PC8 | `TestGameRun_UsesDirNoneCorrectly` | tests/game_test.go | Verifica que Run() use DirNone correctamente |

**Total: 8 tests nuevos**

---

## Regression Risk Assessment

| Riesgo | Nivel | Mitigación |
|--------|-------|------------|
| Tests existentes de observabilidad fallan si cambia logging | Bajo | Los tests de observabilidad no verifican valores de dirección, solo que se loguee |
| Alguien modifica código de implementación antes de que tests estén en RED | Medio | Tests deben escribirse y commitearse ANTES de cualquier cambio en src/ |
| Constantes DirNone no definidas causando compilation error en tests | Esperado | Es correcto - tests en RED deben fallar, pueden fallar por compile error inicial hasta que se implemente |

---

## Gate de Test Audit Checklist

- [x] Suite existente mapeada y analizada
- [x] Tests modificados identificados (0 en este caso)
- [x] Tests untouched listados explícitamente
- [x] Tests nuevos listados con mapeo a postcondiciones
- [x] Justificación escrita para cada decisión
- [x] Regression risks documentados
- [x] Ready para aprobación humana

---

## Aprobación

**Estado:** Pendiente aprobación humana para proceder a fase RED.

**Nota para implementer futuro**: Los tests nuevos verificarán el nuevo comportamiento con `DirNone`. Los tests existentes de observabilidad seguirán pasando porque no dependen del valor específico de dirección fallback.
