---
id: "007-continuous-input-response"
title: "Continuous Input Response"
status: "approved"
approved_by: "Pablo Manuel Rizzo"
approved_at: "2026-05-05"
---

# Feature: Continuous Input Response

## Descripción

El bug actual causa una desincronización entre la captura del primer input y el game loop cuando el usuario presiona la primera tecla de dirección. Específicamente, `g.Update()` se invoca **inmediatamente en el default branch del select** (línea 72 de `main.go`) cuando captura el primer input direccional, en lugar de bufferizarse y ejecutarse en el siguiente ciclo del ticker de 200ms. Esto causa que el primer movimiento sea "uno a uno" sin continuidad visual con movimientos posteriores, rompiendo la UX esperada donde el game loop mantiene un tempo consistente.

El problema es aún más crítico cuando el usuario presiona múltiples direcciones rápidamente: el primer input dispara un Update inmediato, el segundo input espera al ticker, causando movimientos de velocidad variable y pérdida de continuidad.

**Por qué es un problema**:
- El usuario espera que todos los movimientos (incluyendo el primero) ocurran en sincronía con el ticker de 200ms.
- La ejecución asincrónica de Update en el default branch crea una ventana donde el estado del juego es inconsistente.
- Input buffer y Update están desacoplados, violando el Invariante I1 (ticker es el único orchestrador de cambios de estado).

## Contrato

### Postcondiciones

1. **PC1: Primer input dispara movimiento en ciclo de ticker, no en default branch**
   - Input: Usuario presiona tecla de dirección válida (DirUp, DirDown, DirLeft, DirRight) cuando `firstInputReceived == false`
   - Output: `g.Update()` NO se ejecuta en ese ciclo del select (default branch); en su lugar, se bufferiza la dirección en `pendingDirection`
   - Error: Si `g.Update()` se ejecuta en línea dentro del default branch, el PC falla
   - Verificación: Log de observabilidad NO contiene evento `update` en el tick donde se captura el primer input; SÍ contiene evento `update` en el siguiente tick donde se procesa `pendingDirection`

2. **PC2: Game loop realiza Update() una sola vez por ciclo de 200ms (no duplicado)**
   - Input: Primer input presionado; juego en ejecución normal
   - Output: En cualquier periodo de 200ms, `g.Update()` se invoca exactamente una vez (no cero, no dos)
   - Error: Si `g.Update()` se invoca cero veces o más de una vez en una ventana de 200ms, falla
   - Verificación: Análisis de logs con timestamp; contar eventos `update` en ventanas de 200ms consecutivas; todas deben tener count = 1

3. **PC3: Inputs posteriores mantienen continuidad de movimiento**
   - Input: Usuario presiona múltiples direcciones consecutivas (ej: Up, Right, Up en ticks 0, 1, 2) después del primer input
   - Output: Cada input se procesa exactamente una vez en el ticker inmediatamente posterior; no hay saltos de inputs ni duplicación
   - Error: Si se pierde un input, se duplica un movimiento, o se procesa fuera de orden, falla
   - Verificación: Secuencia de eventos `update` en logs correspond exactamente a la secuencia de inputs (en orden, sin duplicación, sin omisión); `snake.Head()` en cada tick corresponde a la posición esperada tras aplicar la dirección del ciclo anterior

4. **PC4: No hay doble movimiento en transición primer→segundo input**
   - Input: Usuario presiona primer input en tick T0, segundo input en tick T1 (T1 = T0 + 200ms)
   - Output: En el periodo [T0, T0+200ms), snake se mueve exactamente una vez (por el primer input); en el periodo [T0+200ms, T0+400ms), snake se mueve exactamente una vez (por el segundo input)
   - Error: Si snake se mueve dos veces en una ventana de 200ms, falla
   - Verificación: Test E2E simula dos inputs con temporización exacta; captura posiciones de `snake.Head()` en cada tick; cuenta transiciones (cambios de posición); debe haber exactamente una transición por input

5. **PC5: firstInputReceived flag previene movimiento antes de primer input**
   - Input: Game loop ejecutándose, `firstInputReceived == false`, sin input del usuario
   - Output: `g.Update()` nunca se invoca; `snake.Head()` permanece en su posición inicial
   - Error: Si snake se mueve o `g.Update()` se ejecuta antes del primer input, falla
   - Verificación: Test E2E que ejecuta 10 ticks (2 segundos) sin presionar ninguna tecla; verifica que `snake.Head()` es idéntica en todos los ticks; verifica logs para confirmar cero eventos `update`

### Postcondiciones de invariantes (derivadas)

6. **PC6: pendingDirection buffer almacena solo la dirección más reciente**
   - Input: Usuario presiona Up, Up, Right en ticks 0, 1, 2 antes de que se procese el ticker
   - Output: `pendingDirection` después del tick 2 contiene solo DirRight (la dirección más reciente), no un queue
   - Error: Si pendingDirection es un queue o acumula múltiples direcciones, falla (aunque esto sería una optimización futura)
   - Verificación: Código en main.go muestra que `pendingDirection` se sobrescribe, no se encola

7. **PC7: Condicional de Update se ejecuta solo desde ticker.C case**
   - Input: Cualquier secuencia de inputs
   - Output: Todas las invocaciones de `g.Update()` ocurren dentro del `case <-ticker.C:` block, nunca en el `default:` block (después de PC1-PC5 fijas)
   - Error: Si hay una línea `g.Update()` en el default block, falla
   - Verificación: Code review de main.go líneas 45-105; búsqueda de `g.Update()` debe encontrar solo una ocurrencia dentro del ticker.C case

## Invariantes verificables

### I1: Ticker es el único orchestrador de cambios de estado (game state mutations)

**Descripción**: Todos los cambios de estado del juego (posición de snake, comida, score, over, paused) se originan en el `case <-ticker.C:` del select. El default branch nunca debe causar cambios de estado (solo bufferizaciones).

**Verificación**:
- Code review: No hay `g.Update()`, `g.Pause()`, `g.Resume()`, `g.SetDirection()` u otras mutaciones en el default branch después de capturar el input
- Property test: En 1000 ticks con inputs aleatorios, todas las transiciones de estado ocurren exactamente una vez por tick, sincronizado con ticker.C

### I2: Congruencia entre input capture y input processing

**Descripción**: Si un input se captura en el default branch del tick T, se procesa (mutación) exactamente en el ticker.C del tick T+1 o T (si ambos ocurren en el mismo ciclo del select), pero nunca dos veces en ticks diferentes.

**Verificación**:
- Test E2E que inyecta inputs con timestamps precisos; verifica que cada input aparece exactamente una vez en los logs `update` después de ser capturado
- Invariante monitoreable: `count(update logs) == count(input logs)` a lo largo de 100 ticks

### I3: firstInputReceived es monótono (una vez true, siempre true)

**Descripción**: El flag `firstInputReceived` se asigna a true cuando se captura el primer input direccional válido y nunca revierte a false.

**Verificación**:
- Code review: No existe `firstInputReceived = false` después de su inicialización
- Test E2E: Simula 50 inputs consecutivos; verifica que el flag sigue siendo true durante y después de todos ellos

### I4: Ticker permanece síncrono a 200ms durante toda la ejecución

**Descripción**: Independientemente de inputs, pauses, o cambios de estado, el ticker mantiene un intervalo consistente de 200ms entre ticks consecutivos.

**Verificación**:
- Property test: Mide deltas de tiempo entre eventos `render` consecutivos (loggeados por ticker.C); todos están dentro del rango [190ms, 210ms]
- Invariante crítica: El ticker nunca se resetea, re-inicia, o se modifica durante la ejecución

## Criterios de aceptación

1. **Primer input → Update en ticker, no en default**: Log del primer input muestra evento `update` en el siguiente tick (después del ticker), no en el mismo tick

2. **Update() invocado una sola vez por 200ms**: Análisis de logs; contar eventos `update` en ventanas de 200ms; todos tienen count = 1

3. **Secuencia de inputs preservada**: Suite de tests compara evento `update` logs con inputs en el orden esperado; todos coinciden

4. **Sin doble movimiento en transición**: Test E2E con dos inputs de 200ms de separación; conteo de transiciones en `snake.Head()` = 2 (una por input)

5. **Sin movimiento antes de primer input**: Test E2E de 10 ticks sin inputs; `snake.Head()` idéntica en todos; zero eventos `update` en logs

6. **pendingDirection buffer implementado**: Code review de main.go muestra variable `pendingDirection` de tipo `game.Direction` inicializada antes del select loop

7. **Update solo en ticker.C case**: Búsqueda de `g.Update()` en main.go encuentra una sola ocurrencia dentro del `case <-ticker.C:` block

8. **Suite tests verde**: `go test ./... -v` retorna 0 fallos; todos los tests existentes + nuevos pasan

9. **Compilación correcta**: `go build` compila sin errores

10. **Invariantes de logs**: Con `DEBUG=1`, logs analizados no muestran violaciones de I1-I4 (ticker síncrono, single update per 200ms, inputs procesados una sola vez)

## Testing Strategy

### Auditoría de Suite Existente

**Tests a revisar** (test-writer):
1. `feature_005_test.go`: Verificar que inputs se capturan correctamente; asegurarse de que no hay regresiones en `ReadDirectionNonBlocking()`
2. `game_test.go`: Revisar tests de movimiento; deben pasar sin modificación post-fix (I4 - comportamiento post-primer-input idéntico)
3. `input_test.go`: Validar que input capture es determinístico; sin cambios esperados aquí

**Puntos de riesgo de regresión**:
- PC4 de 004-gameplay-polish (firstInputReceived flag): Asegurarse de que el flag sigue siendo monótono y no interfiere con el nuevo pendingDirection
- Timing de render (ticker.C): Verificar que los renders siguen ocurriendo cada 200ms, no acelerados por updates en default branch

### Tests nuevos a escribir

#### Test Unit: `TestGameLoop_PendingDirectionBuffer`
- **Qué valida**: PC1, PC6
- **Setup**: Game loop mock, select de dos iteraciones (input en default, ticker en siguiente)
- **Acción**: Simular input "Up" en default branch
- **Aserción**: `pendingDirection` contiene DirUp después de default; `g.Update()` no se ha llamado; en siguiente ticker.C, `g.Update(DirUp)` se invoca una sola vez

#### Test E2E: `TestContinuousInputResponse_FirstInputTiming`
- **Qué valida**: PC1, PC2
- **Setup**: Game loop real, logging activo con timestamps
- **Acción**: Presionar "d" (DirRight) en tick 0
- **Aserción**:
  - Evento `update` NO está en logs del tick 0 (donde se captura el input)
  - Evento `update` con direction=DirRight está en logs del tick 1 (próximo ticker.C)
  - Exactamente una ocurrencia de `update` en ventana [tick0, tick1]

#### Test E2E: `TestContinuousInputResponse_ConsecutiveInputs`
- **Qué valida**: PC3, PC4, I2
- **Setup**: Game loop real, logging activo
- **Acción**: Presionar Up (T=0ms), Right (T=200ms), Up (T=400ms) en ticks específicos
- **Aserción**:
  - Logs muestran eventos `update` en orden: Up, Right, Up
  - snake.Head() transiciones correspond exactamente a las direcciones
  - Cero duplicación de movimientos
  - Entre cada par de inputs hay exactamente un `update`

#### Test E2E: `TestContinuousInputResponse_NoMovementBeforeFirstInput`
- **Qué valida**: PC5, I3
- **Setup**: Game loop real, logging activo, sin input inyectado
- **Acción**: Dejar ejecutar game loop durante 2 segundos (10 ticks de 200ms) sin presionar teclas
- **Aserción**:
  - snake.Head() es idéntica en todos los ticks
  - Zero eventos `update` en logs
  - firstInputReceived permanece false (si es observable)

#### Property Test: `TestContinuousInputResponse_TickerSynchrony`
- **Qué valida**: I4, PC2
- **Setup**: Game loop real, random inputs inyectados, 100 ticks
- **Acción**: Ejecutar con inputs en posiciones aleatorias; loggear cada evento `render` con timestamp
- **Aserción**:
  - Todos los deltas entre eventos `render` consecutivos están en [190ms, 210ms]
  - Eventos `update` son monotónicos en tiempo (no hay viajes al pasado)

### Invariantes a probar con properties

1. **I1 (Ticker orchestrador único)**:
   - Property: `count(state_mutations) == count(ticker.C_events)` en cualquier secuencia de inputs aleatorios
   - Tool: Property-based testing (quickcheck-style) con 1000 casos de entrada

2. **I2 (Input → single processing)**:
   - Property: `count(update_logs with dir=D) == count(input_logs with dir=D)` para cada dirección D
   - Tool: Comparación de logs, assertion en test

3. **I3 (firstInputReceived monotónicity)**:
   - Property: Una vez que aparece primer `update` en logs, todos los `update` posteriores ocurren regularmente (no hay "dormidas")
   - Tool: Secuencial check en logs

4. **I4 (Ticker synchrony)**:
   - Property: Todos los deltas de tiempo entre `ticker.C` eventos están en [200ms ± 10%]
   - Tool: Medición de timestamps en logs, histograma

### Artefactos de auditoría esperados

Test-writer debe producir:
- Documento `test-audit.md` documentando qué tests pasaron/fallaron pre-fix
- Listado de tests nuevos con cobertura de PC1-PC7 y I1-I4
- Reporte de regresiones (idealmente cero)
- Confirmación de que logging de observabilidad captura las condiciones verificables

---

**Status**: Approved by Pablo Manuel Rizzo on 2026-05-05
