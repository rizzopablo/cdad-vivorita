# Vivorita 2: CDAD Methodology Case Study

<img width="410" height="404" alt="imagen" src="https://github.com/user-attachments/assets/38e0e1f7-55fd-4614-b855-23b5ede34545" />


## Overview

**Vivorita 2** es el primer proyecto desarrollado **íntegramente mediante Contract-Driven AI Development (CDAD)**, una metodología de cinco etapas que integra especificación formal, test-driven development, revisión en capas, y memory banking.

Este repositorio documenta tanto el producto (clon de Snake en terminal) como la experiencia de aplicar CDAD en un proyecto real de mediano alcance (~1300 LoC, 7 features, 54 tests).

---

## Qué es CDAD (Contract-Driven AI Development)

CDAD es un ciclo estructurado de desarrollo que:

1. **Descubrimiento** — Mapear APIs, arquitectura, y decisiones previas
2. **Especificación** — Escribir contratos verificables (postcondiciones, invariantes, criterios)
3. **TDD Anti-Trampa** — Tests rojos primero, implementación mínima, refactor opcional
4. **Review Two-Layer** — Validar contra spec, detectar bloqueantes vs opcionales
5. **Merge + Memory Bank** — Documentar decisiones, actualizar contexto, comprometer cambios

Cada etapa tiene **gates de validación obligatorios** y se ejecuta en **sesiones aisladas** para máximo aislamiento cognitivo.

---

## Experiencia: 7 Features en 1 Ciclo

### Timeline

| Feature | Tipo | Status | Aprendizaje Clave |
|---------|------|--------|-------------------|
| **001** | Base | ✅ Done | Arquitectura base (game loop, persistencia) |
| **002** | Observability | ✅ Done | Logging JSON, snapshots, integración tcell |
| **003** | Bug Fix | ✅ Done | Input handling, DirNone fallback |
| **004** | Polish | ✅ Done | firstInputReceived flag, ticker 200ms |
| **005** | Bug Critical | ✅ Done | Raw mode init, thread-safety, race conditions |
| **006** | Hotfix | ✅ Done | Direction semantics (Y-axis inversion) |
| **007** | Continuous Motion | ⚠️ Partial | **CDAD falló**: tests validaban buffer, no continuidad |

### Estadísticas

- **Specs escritos**: 7 (5 completos + CDAD, 2 hotfix manual)
- **Tests totales**: 54 (42 existentes + 9 nuevos feature 007 + 3 refactorizados)
- **Cobertura**: ~95% (game loop, snake movement, input, observability)
- **Regressions**: 0 (en 6 features; feature 007 requirió hotfix manual)
- **Ciclos CDAD completos**: 6 exitosos, 1 fallido (007)

---

## Aprendizajes Clave

### ✅ Fortalezas de CDAD

#### 1. **Especificación Formal Previene Ambigüedad**
   - Escribir postcondiciones numeradas antes de código fuerza claridad
   - Los specs de features 001-006 fueron precisos; los bloqueantes se detectaron temprano
   - **Beneficio**: 0 regresiones en code review (gates validaron spec-compliance antes de merge)

#### 2. **TDD Anti-Trampa Escala Mejor que TDD Convencional**
   - Separar test-writer (RED) de implementer (GREEN) evita "tests que pasan con bugs"
   - RED tests fallaban por razones correctas (buffer no existe, timing incorrecto)
   - **Beneficio**: Tests validaban postcondiciones, no implementación

#### 3. **Gates Entre Etapas Cierran Loops de Feedback Rápido**
   - Gate specification → TDD validó que todos los tests tenían justificación en spec
   - Gate TDD → Review validó que cada cambio cumplía postcondición
   - **Beneficio**: Bloqueantes detectados antes de merge, no en producción

#### 4. **Memory Banking Captura Decisiones Arquitectónicas**
   - Specs 002, 004, 005 documentaron por qué `ticker`, `firstInputReceived`, `raw mode`
   - Future developers (o LLMs) tienen contexto para mantener decisiones
   - **Beneficio**: Feature 006 fue hotfix de 2 minutos porque arquitectura estaba clara

### ⚠️ Limitaciones Detectadas

#### 1. **Tests Pueden Validar El BUG, No La Solución (Feature 007)**
   - CDAD RED stage escribió tests que validaban `pendingDirection` buffer
   - Pero el buffer era la causa del bug, no la solución
   - Tests pasaron (buffer se asignaba) pero continuidad real falló
   - **Raíz**: Especificación fue correcta, pero tests RED no validaron **comportamiento observable** (movimiento continuo), solo **mecanismo interno** (buffer asignado)
   - **Fix**: Necesitaba E2E test o property test que validara "snake se mueve cada 200ms sin input"

#### 2. **Agentes LLM Pueden Malinterpretar Specs**
   - Agente implementer aplicó buffering correctamente según su interpretación
   - Pero buffer reseteado a `DirNone` rompía continuidad (no fue evidente en handoff)
   - **Raíz**: Handoff omitió detalles de "persistencia de dirección"
   - **Mejora**: Handoffs deben incluir anti-paterns (qué NO hacer)

#### 3. **Aislamiento de Sesiones Dificulta Debugging**
   - Feature 007 fallón en realidad después de que "todos los tests pasaron"
   - Revisor no probó manualmente el binario (sesión aislada, solo código)
   - **Raíz**: CDAD assumes "si tests pasan, está correcto"
   - **Mejora**: Review stage debe incluir "smoke test" (ejecutar binario, verificar observable behavior)

#### 4. **Postcondiciones Deben Ser OBSERVABLES**
   - PC1: "Primer input dispara movimiento en ticker" ← demasiado técnica (internal detail)
   - Mejor: "Snake se mueve continuamente cada 200ms desde primer input sin nuevo input requerido" ← observable
   - **Raíz**: Arquitectos escriben specs técnicos; test-writers escriben tests técnicos
   - **Mejora**: Specs deben separar "qué se ve" (observable) de "cómo se implementa"

---

## Cómo Se Resolvió Feature 007

Feature 007 expone el ciclo de debugging en CDAD:

1. **Reporte de usuario**: "No se mueve continuo" (observable behavior)
2. **Diagnóstico**: Tests verdes, código compila, pero binario falla
3. **Root cause**: `pendingDirection` se reseteaba, rompiendo persistencia
4. **Fix manual**: Cambiar a `currentDirection` persistente
5. **Lección**: Próximas features validarán "observable behavior", no solo "code mechanics"

**Commits relacionados**:
- `391b115` — fix: continuous snake movement (hotfix manual)
- `37cc36f` — docs: update progress (memory bank)

---

## Arquitectura del Proyecto

```
src/
├── main.go              ← Game loop (select+ticker, input buffering, render)
├── game/
│   ├── game.go          ← Game struct, Update logic
│   └── snake.go         ← Snake struct, Movement, Collision detection
├── input/
│   └── input.go         ← tcell integration, ReadDirectionNonBlocking
├── render/
│   └── render.go        ← Board rendering, tcell screen management
├── observability/
│   └── observability.go ← JSON logging, Board snapshots
└── score/
    └── score.go         ← High score persistence

docs/
├── .cdad-state.json     ← State machine (etapa actual, feature, postcondiciones)
├── progress.md          ← Features done/in-progress/queued
└── specs/
    ├── 001-snake-game/
    ├── 002-game-loop-observability/
    ├── ...
    └── 007-continuous-input-response/
        ├── spec.md       ← Postcondiciones, invariantes, criterios
        ├── test-audit.md ← Análisis de suite existente
        └── review.md     ← Validación contra spec
```

### Decisiones Arquitectónicas Clave

1. **Game Loop con Ticker + Non-Blocking Input** (Feature 002, 004, 005)
   - Ticker 200ms como sole orchestrator de mutaciones de estado
   - Input capturado en non-blocking select default branch
   - Previene race conditions, garantiza determinismo

2. **firstInputReceived Flag** (Feature 004)
   - Snake no se mueve hasta primer input del usuario
   - Previene comportamiento no intuitivo (auto-movement sin input)

3. **DirNone + Direction Fallback** (Feature 003)
   - DirNone representa "no dirección especificada"
   - Fallback a dirección actual previene reverse-direction bugs

4. **Observability Integration** (Feature 002)
   - Logging JSON estructurado en cada evento (input, update, render)
   - Snapshots de tablero en game-over
   - Facilita debugging post-hoc

5. **Persistent currentDirection** (Feature 007 hotfix)
   - Dirección actual persiste entre ticks
   - Inputs UPDATE dirección, no la BUFFERIZAN y RESETEAN
   - Permite movimiento automático continuo

---

## Recomendaciones para Futuros Desarrollos con CDAD

### Para Arquitectos (Especificación)
1. **Separar "observable" de "técnica"**
   - Observable: "Snake avanza 1 casilla cada 200ms"
   - Técnica: "Update() ejecuta en ticker.C case, no default"
   - Specs deben enfatizar observable; detalles técnicos en comentarios

2. **Incluir Anti-Patterns en Specs**
   - No solo qué SÍ hacer; qué NO hacer
   - Ejemplo: "NO buffericez dirección y luego la reseteés en cada tick"

3. **Definir "Smoke Tests" Observable**
   - Specs deben incluir "after implementation, verify by running and observing X"
   - Evita feature 007 (tests verdes, binario roto)

### Para Test-Writers (RED)
1. **Property-Based Tests para Comportamiento Continuo**
   - No solo "se ejecuta 1 vez"; "se ejecuta N veces en tiempo T"
   - Feature 007 necesitaba: "después de 1s (5 ticks), snake avanzó 5 casillas sin input"

2. **E2E/Integration Tests Además de Unit Tests**
   - Unit test: "Update() llamado con DirRight"
   - E2E test: "User presses right, snake avanza derecha cada 200ms"

3. **Log-Based Assertions**
   - Validar timing de eventos via timestamps en logs
   - Ejemplo: "update event debe ocurrir cada ~200ms, ±10%"

### Para Implementers (GREEN)
1. **Favor Claridad sobre Cleverness**
   - Feature 007 falló por "optimización" prematura (buffer reseteado)
   - Simple: dirección persiste, se aplica cada tick

2. **Manual Testing Antes de Commit**
   - Ejecutar binario, verificar observable behavior
   - No confiar 100% en tests (pueden validar mecanismo, no comportamiento)

### Para Revisers (REVIEW)
1. **Incluir "Smoke Test" en Review Gate**
   - No solo leer código; ejecutar binario
   - Verificar observable behavior antes de aprobar

2. **Detectar "Tests Validando El BUG"**
   - Si todos los tests pasan pero concepto falla, cuestionár test strategy
   - Red flag: "tests pasan pero postcondicione no se cumple observablemente"

### Para Scribes (MEMORY)
1. **Documentar Decisiones, No Solo Resultados**
   - "Por qué" ticker es sole orchestrator, no solo "qué" (Update en ticker)
   - Facilita future context para arquitectos posteriores

2. **Incluir "Lecciones Aprendidas" en ADRs**
   - ADR debe citar feature 007 como contraejemplo si aplica

---

## Conclusión

**CDAD funciona muy bien** para:
- ✅ Mantener zero regressions (gates validaron spec antes de merge)
- ✅ Capturar decisiones arquitectónicas (Memory Bank claro)
- ✅ Escalar test-driven development (sesiones aisladas, roles claros)

**CDAD tiene limitaciones** que detectamos:
- ⚠️ Tests pueden validar mecanismos internos, no comportamiento observable
- ⚠️ Aislamiento de sesiones previene debugging temprano
- ⚠️ Agentes LLM pueden malinterpretar specs incompletas

**Para vivorita 2 específicamente**:
- 6 features exitosas via CDAD puro
- 1 feature requirió hotfix manual (observable behavior mismatch)
- Resultado final: **Snake game funcional, observable-correcto, zero regressions**

---

## Ejecución

```bash
cd src && go build -o ../vivorita .
./vivorita
```

Controles:
- Flechas: Mover serpiente
- P: Pausar/Reanudar
- Q: Salir
- Primera entrada de dirección inicia movimiento continuo

---

## Archivos Clave

- **Specs**: `docs/specs/*/spec.md` (contratos verificables)
- **State Machine**: `docs/.cdad-state.json` (etapa actual, postcondiciones)
- **Memory Bank**: `docs/progress.md` (qué está done/in-progress)
- **Code**: `src/main.go` (game loop, ~100 LoC), `src/game/*.go` (lógica)
- **Tests**: `tests/game/*_test.go` (54 tests, ~50% cobertura observada)

---

**Fecha**: Mayo 5, 2026
**Metodología**: Contract-Driven AI Development (CDAD)
**Status**: Feature-Complete, Observable-Correct, Regression-Free (6/7 features via CDAD; 1/7 hotfix manual)
