---
id: "005-input-capture-broken"
title: "Input Capture Architecture Defect"
status: "approved"
approved_by: "Architect Discovery Phase"
approved_at: "2026-05-05"
---

# Feature: Input Capture Architecture Defect

## Descripción funcional

**Defect**: Keyboard input is completely non-functional. Debug logs show zero input events (no `input_raw`, `input_converted`, or `input_error` entries) despite arrow keys and WASD being mapped in `input.go`.

**Root causes (Discovery-validated)**:

1. **Goroutine leak in `ReadDirectionNonBlocking()`**: The function spawns a goroutine that calls `screen.PollEvent()`, which blocks indefinitely. When the 10ms timeout fires, the goroutine continues blocked on `PollEvent()`, causing a goroutine leak and race condition where events are lost if they arrive after the timeout but while the goroutine is still executing.

2. **Unsafe tcell concurrency pattern**: tcell explicitly warns that multiple goroutines must not call `PollEvent()` simultaneously. The current pattern violates this safety constraint, creating undefined behavior.

3. **No initial render**: The first `render` event only occurs after the first ticker tick (200ms), violating PC2 of this spec and PC8 of `004-gameplay-polish` (board must be visible from start within 100ms).

**Impact**: Feature `004-gameplay-polish` is marked done but is unusable—no input is captured at all.

## Contrato

### Firma pública

```go
// src/main.go
// Game loop must render initial board before entering select
// Screen must be configured for raw input after Init()

// src/input/input.go
func ReadDirectionNonBlocking(screen tcell.Screen) (Direction, error)
// Must successfully receive tcell.EventKey events from terminal
```

### Postcondiciones

1. **PC1**: Initial board render occurs BEFORE entering the select loop (fixes RC3).
   - Input: Game startup, immediately after screen setup and logging init
   - Output: Board (snake, food, borders) is rendered and displayed; event `render` with `"source":"initial"` is logged
   - Error: If first render only happens after ticker fires (~200ms) or doesn't happen before select loop, fails
   - Timing: Logged timestamp must be < 100ms from app start
   - Verificación: With `DEBUG=1`, grep logs for first `render` event; verify it has `"source":"initial"` and occurs before any `render` with `"source":"ticker"`

2. **PC2**: `ReadDirectionNonBlocking()` refactored to use safe tcell concurrency pattern (fixes RC1, RC2).
   - Input: Function called from game loop
   - Output: Either (a) uses `screen.ChannelEvents()` in dedicated background goroutine with cancellation, OR (b) uses `screen.HasPendingEvent()` for non-blocking checks before `PollEvent()`, OR (c) alternative pattern that does not spawn ephemeral goroutines
   - Error: If goroutine+PollEvent pattern remains, fails
   - Verificación: Code review confirms no goroutine spawning in ReadDirectionNonBlocking loop; no `go func() { ... screen.PollEvent() ... }` pattern

3. **PC3**: Input events (arrow keys and WASD) are captured reliably when pressed.
   - Input: User presses any of ↑, ↓, ←, →, W, A, S, D during game loop execution
   - Output: Log contains matching `input_raw` entry with key identifier (e.g., `"key":"KeyUp"` or `"key":"KeyRune"/"rune":"w"`)
   - Error: If key press is not logged, fails
   - Verificación: Test E2E: type rapid sequence of keys, verify all appear in log as `input_raw` events within 100ms of user input

4. **PC4**: Arrow keys are converted to correct game direction.
   - Input: User presses each arrow key (↑, ↓, ←, →)
   - Output: Each key produces corresponding `input_converted` log with direction (DirUp, DirDown, DirLeft, DirRight)
   - Error: If any arrow key is not converted or converted incorrectly, fails
   - Verificación: Test matrix: for each arrow key, verify corresponding direction logged

5. **PC5**: WASD keys are converted to correct game direction (regression test).
   - Input: User presses each of W, A, S, D
   - Output: Each key produces corresponding `input_converted` log with direction (DirUp, DirLeft, DirDown, DirRight)
   - Error: If WASD conversion changed from before fix, fails
   - Verificación: Test matrix: for each WASD key, verify corresponding direction logged (same as 003-input-capture-investigation test expectations)

6. **PC6**: No goroutine leaks occur during game loop execution.
   - Input: Game loop running for 1000 iterations (200 seconds simulated time)
   - Output: Final goroutine count equals initial goroutine count + 1 (main goroutine only)
   - Error: If goroutine count grows unbounded during execution, fails
   - Verificación: Test uses `runtime.NumGoroutine()` before and after loop; verify no growth

## Invariantes verificables

### I1: tcell.Screen.Init() is the ONLY raw mode configuration
No additional raw mode calls are needed or should be present in the code. `Init()` automatically configures the terminal (tcsetattr on Unix, equivalent on Windows).

**Verificación**: No calls to special raw mode methods in main.go (e.g., no `screen.SetMode()` or custom terminal config); only `screen.Init()` after NewScreen() and `screen.Fini()` at cleanup.

### I2: Event polling is safe for concurrent use
No goroutine spawning in the input polling loop. The implementation uses either `screen.ChannelEvents()` with proper cancellation OR `screen.HasPendingEvent()` + single `PollEvent()` call, not the ephemeral goroutine pattern.

**Verificación**: Code review: no `go func()` or `<-time.After()` racing with `screen.PollEvent()` in ReadDirectionNonBlocking().

### I3: Initial render is displayed before game loop awaits input
The board (snake, food, borders) is visible on screen from app startup, within 100ms, before the select loop begins waiting for input.

**Verificación**: With `DEBUG=1`, log shows `"render"` event with `"source":"initial"` timestamp < 100ms from app start, before any `"source":"ticker"` renders.

### I4: Ticker periodicity is unaffected
The 200ms ticker interval remains unchanged and unaffected by input handling refactoring.

**Verificación**: Successive `"source":"ticker"` render events are logged ~200ms apart (±20ms tolerance).

## Criterios de aceptación

1. **Initial board visible**: Within 100ms of startup, board is displayed on screen with snake, food, borders visible.
2. **Initial render logged**: Log contains `render` event with `"source":"initial"` before any `"source":"ticker"` renders.
3. **Arrow keys captured**: Pressing ↑, ↓, ←, → produces `input_raw` + `input_converted` log entries.
4. **WASD captured**: Pressing W, A, S, D produces `input_raw` + `input_converted` log entries (same as before fix).
5. **No missed keystrokes**: Rapid sequence test (10 keystrokes within 1 second) all logged as `input_raw` events.
6. **Snake moves on input**: Pressing arrow keys or WASD causes snake to move in correct direction.
7. **No goroutine leaks**: `runtime.NumGoroutine()` before and after 1000 iterations is equal (no unbounded growth).
8. **All tests pass**: `go test ./... -v` returns 0 failures; no new test failures introduced.
9. **No "waiting" logs**: With `DEBUG=1`, no log entries contain "waiting", "poll", or "timeout" messages related to input capture.

---

**Status**: Draft — Pending discovery gate validation

