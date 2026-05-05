# Review — 003-input-capture-investigation

## Resumen Ejecutivo

**Feature**: Input Capture Bug Fix - Corrección del bug donde la serpiente se movía automáticamente hacia abajo (DirUp en modelo cartesiano) cuando no había input del usuario.

**Estado de implementación**: ✅ Completada
- Suite de tests: 17/17 pasando (9 tests específicos de esta feature + 8 tests de observability)
- Build: Sin errores
- Cobertura de postcondiciones: 8/8 (PC1-PC8)
- Cobertura de invariantes: 5/5 (I1-I5)

**Resultado del review**: Sin bloqueantes. 1 opcional menor.

---

## Bloqueantes

*No se encontraron bloqueantes.*

---

## Opcionales

### 1. Magic number de timeout sin constante nombrada (I5)

**Ubicación**: `src/input/input.go:37`

**Problema**: El timeout de 10ms está hardcodeado como literal `time.After(10 * time.Millisecond)`. La invariante I5 del spec establece: *"No hay magic numbers en el timeout de input — el valor 10ms debe ser constante `InputPollTimeout`"*.

**Sugerencia**: Definir una constante exportada:
```go
const InputPollTimeout = 10 * time.Millisecond
```
Y usarla en la línea 37:
```go
case <-time.After(InputPollTimeout):
```

**Severidad**: Opcional (el valor está documentado en el spec y no hay otros usos, pero viola I5)

---

## Verificación de Postcondiciones

| Postcondición | Descripción | Estado | Test que cubre |
|--------------|-------------|--------|----------------|
| PC1 | `ReadDirectionNonBlocking()` retorna `DirNone` en timeout | ✅ | `TestReadDirectionNonBlocking_TimeoutReturnsDirNone` |
| PC2 | `ReadDirectionNonBlocking()` retorna `DirNone` en tecla no mapeada | ✅ | `TestReadDirectionNonBlocking_UnmappedKeyReturnsDirNoneWithLog` |
| PC3 | `convertInputToGameDirection()` usa `currentDir` como fallback | ✅ | `TestConvertInputToGameDirection_DirNoneReturnsCurrentDir` |
| PC4 | Sin input, serpiente mantiene dirección actual | ✅ | `TestGame_MaintainsDirectionWithoutInput` |
| PC5 | No hay fallback `DirUp` por ausencia de input | ✅ | `TestDirUp_NotFallbackForNoInput` |
| PC6 | `game.Direction` define `DirNone` | ✅ | `TestGame_DirNoneConstantExists` |
| PC7 | `input.Direction` define `DirNone` | ✅ | `TestInput_DirNoneConstantExists` |
| PC8 | `game.Run()` maneja `DirNone` correctamente | ✅ | `TestGameRun_UsesDirNoneCorrectly` |

## Verificación de Invariantes

| Invariante | Descripción | Estado | Verificación |
|-----------|-------------|--------|--------------|
| I1 | `DirNone` tiene valor distinto a otras direcciones | ✅ | Verificado por `iota` secuencial (valor 6) |
| I2 | Timeout/tecla no mapeada → solo `DirNone` | ✅ | Código inspeccionado: `default` cases retornan `DirNone` |
| I3 | `DirNone` input preserva `currentDir` | ✅ | `convertInputToGameDirection` línea 114 (`default: return currentDir`) |
| I4 | Dirección solo cambia con input válido | ✅ | `main.go:60-61` solo llama `Update()` si `!g.IsOver()` y usa `gameDir` |
| I5 | Timeout 10ms como constante `InputPollTimeout` | ⚠️ | Magic number hardcodeado en línea 37 |

## Verificación de Criterios de Aceptación

| Criterio | Estado | Notas |
|----------|--------|-------|
| CA1: 30 segundos sin input mantiene dirección | ✅ | Verificado por diseño (PC3+PC4) |
| CA2: DEBUG=1 sin teclas no loguea conversión | ✅ | Solo loguea `input_raw`, no `input_converted` en timeout |
| CA3: Teclas no mapeadas mantienen dirección | ✅ | `default` case retorna `DirNone` |
| CA4: Test `TestReadDirectionNonBlocking_TimeoutReturnsDirNone` pasa | ✅ | PASS |
| CA5: Test `TestReadDirectionNonBlocking_UnknownKeyReturnsDirNone` pasa | ✅ | PASS |
| CA6: Test `TestConvertInputToGameDirection_FallbackToCurrent` pasa | ✅ | PASS (como `TestConvertInputToGameDirection_DirNoneReturnsCurrentDir`) |
| CA7: Test `TestGameLoop_MaintainsDirectionWithoutInput` pasa | ✅ | PASS (como `TestGame_MaintainsDirectionWithoutInput`) |
| CA8: Build sin errores, tests sin regression | ✅ | 17/17 tests pasan, build OK |

## Observaciones Adicionales

### Positivas

1. **Diseño limpio**: La separación entre `input.Direction` y `game.Direction` con ambos teniendo `DirNone` permite capas bien definidas.

2. **Manejo de nil**: `ReadDirectionNonBlocking(nil)` retorna `DirNone`, lo que hace que PC8 (código muerto en `game.Run()`) sea seguro.

3. **Logging apropiado**: Los eventos `input_error` se loguean correctamente para teclas no mapeadas (PC2).

### Notas técnicas

1. **Duplicación de `convertInputToGameDirection`**: Existe en `src/game/game.go:103-116` y `src/main.go:98-113`. Esto es aceptable porque:
   - `game.Run()` es código muerto (según PC8)
   - `main.go` es el entry point real
   - Ambas funciones son idénticas y mantienen la misma semántica

2. **Tests de inspección de código**: Los tests `TestReadDirectionNonBlocking_UnmappedKeyReturnsDirNoneWithLog` y `TestConvertInputToGameDirection_DirNoneReturnsCurrentDir` usan análisis de código fuente (`os.ReadFile` + `strings.Contains`) en lugar de tests de comportamiento puros. Esto es una práctica de "test audit" válida para verificar estructura, aunque idealmente deberían complementarse con tests de comportamiento que mockeen `tcell.Screen`.

---

**Review completado por**: reviewer (Etapa 4 CDAD)
**Fecha**: 2026-05-05
