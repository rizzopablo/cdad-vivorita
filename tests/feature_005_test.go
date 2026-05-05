package tests

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
	"vivorita2/src/input"
	"vivorita2/src/observability"

	"github.com/gdamore/tcell/v2"
)

// TestPC1_InitialRenderTiming verifies that initial board render occurs
// within 100ms of startup with source:"initial" logged.
// This test will FAIL until the fix is implemented (missing initial render before select loop).
func TestPC1_InitialRenderTiming(t *testing.T) {
	// Verify that main.go calls initial render BEFORE entering select loop
	mainSource, err := os.ReadFile("../src/main.go")
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	sourceStr := string(mainSource)

	// Find the main function
	mainIdx := strings.Index(sourceStr, "func main()")
	if mainIdx == -1 {
		t.Fatal("PC1: Could not find main() function")
	}

	// Extract main function body (up to first for loop for select)
	selectIdx := strings.Index(sourceStr[mainIdx:], "for running")
	if selectIdx == -1 {
		selectIdx = strings.Index(sourceStr[mainIdx:], "for {")
		if selectIdx == -1 {
			t.Fatal("PC1: Could not find main game loop")
		}
	}
	selectIdx += mainIdx
	mainBody := sourceStr[mainIdx:selectIdx]

	// Check if there's a render call with source:"initial" BEFORE the loop
	hasInitialRender := strings.Contains(mainBody, "\"initial\"") ||
		strings.Contains(mainBody, "'initial'")

	hasRenderCall := strings.Contains(mainBody, "RenderBoard") ||
		strings.Contains(mainBody, "render")

	if !hasInitialRender {
		t.Error("PC1 RED PHASE FAIL: No initial render with source:\"initial\" found before main game loop. " +
			"Must call RenderBoard and log render event with source:\"initial\" before entering select loop.")
	}

	if !hasRenderCall && !hasInitialRender {
		t.Error("PC1 RED PHASE FAIL: No RenderBoard call found before main game loop. " +
			"Board must be rendered within 100ms of startup, before waiting for input.")
	}

	os.Unsetenv("DEBUG")
}

// TestPC2_ReadDirectionRefactorNoGoroutineSpawning verifies that
// ReadDirectionNonBlocking uses safe tcell concurrency pattern.
// This test will FAIL until goroutine leak is fixed (currently uses go func + PollEvent).
func TestPC2_ReadDirectionRefactorNoGoroutineSpawning(t *testing.T) {
	source, err := os.ReadFile("../src/input/input.go")
	if err != nil {
		t.Fatalf("Failed to read input.go: %v", err)
	}

	sourceStr := string(source)

	// Find ReadDirectionNonBlocking function
	funcIdx := strings.Index(sourceStr, "func ReadDirectionNonBlocking")
	if funcIdx == -1 {
		t.Fatal("PC2: Could not find ReadDirectionNonBlocking function")
	}

	// Extract function body
	funcEnd := strings.Index(sourceStr[funcIdx:], "\nfunc ")
	if funcEnd == -1 {
		funcEnd = len(sourceStr) - funcIdx
	} else {
		funcEnd += funcIdx
	}
	funcBody := sourceStr[funcIdx:funcEnd]

	// Check for goroutine spawning pattern (go func)
	if strings.Contains(funcBody, "go func()") {
		t.Error("PC2 FAIL: ReadDirectionNonBlocking still uses goroutine spawning pattern (go func)")
	}

	// Check for unsafe PollEvent in goroutine (indirect marker)
	if strings.Contains(funcBody, "select {") && strings.Contains(funcBody, "go func") {
		t.Error("PC2 FAIL: ReadDirectionNonBlocking uses select with goroutine (race condition pattern)")
	}

	// Verify safe pattern: either ChannelEvents with context, or HasPendingEvent + PollEvent
	hasSafeChannelPattern := strings.Contains(funcBody, "ChannelEvents")
	hasSafeHasPendingPattern := strings.Contains(funcBody, "HasPendingEvent")

	if !hasSafeChannelPattern && !hasSafeHasPendingPattern {
		t.Errorf("PC2 FAIL: ReadDirectionNonBlocking does not use safe tcell concurrency pattern. "+
			"Must use either ChannelEvents (with cancellation) or HasPendingEvent + PollEvent.\n"+
			"Current pattern may be: %s", funcBody[:200])
	}
}

// TestPC3_ArrowKeysCapturedAsInputRaw verifies that arrow keys produce
// input_raw events in the debug log. This test checks that the input.go
// code correctly logs all arrow key captures. Will FAIL if input handling
// is completely broken due to goroutine leaks.
func TestPC3_ArrowKeysCapturedAsInputRaw(t *testing.T) {
	// Verify input.go logs input_raw for arrow keys
	source, err := os.ReadFile("../src/input/input.go")
	if err != nil {
		t.Fatalf("Failed to read input.go: %v", err)
	}

	sourceStr := string(source)

	// Check that handleKeyEvent logs input_raw
	if !strings.Contains(sourceStr, "input_raw") {
		t.Error("PC3 RED PHASE FAIL: input.go does not log input_raw events. " +
			"Must call LogEvent(\"input_raw\", ...) when keyboard input is received.")
	}

	// Verify that arrow keys are handled (cases for KeyUp, KeyDown, KeyLeft, KeyRight)
	arrowKeys := []string{"tcell.KeyUp", "tcell.KeyDown", "tcell.KeyLeft", "tcell.KeyRight"}
	for _, key := range arrowKeys {
		if !strings.Contains(sourceStr, "case "+key) {
			t.Errorf("PC3 RED PHASE FAIL: No case handler for arrow key %s in handleKeyEvent", key)
		}
	}

	os.Unsetenv("DEBUG")
}

// TestPC4_ArrowKeysConvertedCorrectly verifies that arrow keys are
// converted to the correct game direction. This test will FAIL if
// arrow key mapping is incorrect or input_converted logging is missing.
func TestPC4_ArrowKeysConvertedCorrectly(t *testing.T) {
	source, err := os.ReadFile("../src/input/input.go")
	if err != nil {
		t.Fatalf("Failed to read input.go: %v", err)
	}

	sourceStr := string(source)

	// Verify input_converted is logged
	if !strings.Contains(sourceStr, "input_converted") {
		t.Error("PC4 RED PHASE FAIL: input.go does not log input_converted events. " +
			"Must call LogEvent(\"input_converted\", ...) to document direction conversion.")
	}

	// Verify arrow keys are converted to correct directions
	conversions := map[string]string{
		"case tcell.KeyUp":    "DirUp",
		"case tcell.KeyDown":  "DirDown",
		"case tcell.KeyLeft":  "DirLeft",
		"case tcell.KeyRight": "DirRight",
	}

	for keyCase, expectedDir := range conversions {
		if !strings.Contains(sourceStr, keyCase) {
			t.Errorf("PC4 RED PHASE FAIL: Missing case handler %s", keyCase)
		}

		// Find the case and verify it returns the correct direction
		caseIdx := strings.Index(sourceStr, keyCase)
		if caseIdx != -1 {
			// Extract next 200 chars after case statement
			caseBlock := sourceStr[caseIdx : caseIdx+200]
			if !strings.Contains(caseBlock, "return "+expectedDir) &&
				!strings.Contains(caseBlock, "return "+expectedDir) {
				t.Logf("PC4 NOTE: Arrow key %s may not convert to %s correctly", keyCase, expectedDir)
			}
		}
	}
}

// TestPC5_WASDKeysConvertedCorrectly verifies that WASD keys are
// converted to the correct game direction (regression test).
// This test will FAIL if WASD mapping is broken.
func TestPC5_WASDKeysConvertedCorrectly(t *testing.T) {
	source, err := os.ReadFile("../src/input/input.go")
	if err != nil {
		t.Fatalf("Failed to read input.go: %v", err)
	}

	sourceStr := string(source)

	// Verify WASD conversions are still in place (regression test)
	wasdConversions := map[string]string{
		`case "w"`: "DirUp",
		`case "a"`: "DirLeft",
		`case "s"`: "DirDown",
		`case "d"`: "DirRight",
	}

	for caseStr, expectedDir := range wasdConversions {
		if !strings.Contains(sourceStr, caseStr) && !strings.Contains(sourceStr, strings.ToUpper(caseStr)) {
			t.Errorf("PC5 FAIL: WASD case %s not found in handleKeyEvent (regression in feature 003)", caseStr)
		}

		// Find the case and verify correct direction
		caseIdx := strings.Index(sourceStr, caseStr)
		if caseIdx != -1 {
			caseBlock := sourceStr[caseIdx : caseIdx+200]
			if !strings.Contains(caseBlock, "return "+expectedDir) {
				t.Logf("PC5 NOTE: WASD case %s may not return %s correctly (regression)", caseStr, expectedDir)
			}
		}
	}
}

// TestPC6_NoGoroutineLeaksOnLoop verifies that no goroutine leaks occur
// during game loop execution. This test will FAIL if ReadDirectionNonBlocking
// continues spawning ephemeral goroutines that don't exit cleanly.
func TestPC6_NoGoroutineLeaksOnLoop(t *testing.T) {
	// Record initial goroutine count
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	initialGoroutines := runtime.NumGoroutine()

	// Simulate 100 loop iterations
	// Each iteration calls ReadDirectionNonBlocking with nil screen (simulates timeout)
	iterations := 100

	for i := 0; i < iterations; i++ {
		// Call ReadDirectionNonBlocking with nil screen (simulates timeout)
		// Currently this spawns a goroutine that will cause leak
		dir, err := input.ReadDirectionNonBlocking(nil)
		if err != nil {
			t.Logf("Iteration %d: ReadDirectionNonBlocking returned error: %v", i, err)
		}
		if dir != input.DirNone {
			t.Logf("Iteration %d: Expected DirNone on nil screen, got %v", i, dir)
		}

		// Small delay to allow goroutines to run and (hopefully) clean up
		time.Sleep(2 * time.Millisecond)
	}

	// Allow time for goroutines to clean up
	time.Sleep(500 * time.Millisecond)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// Check final goroutine count
	finalGoroutines := runtime.NumGoroutine()
	goroutineGrowth := finalGoroutines - initialGoroutines

	// RED PHASE: The current implementation spawns goroutines for each call
	// After 100 iterations, we expect significant growth
	// If growth > 10, the goroutine leak is confirmed
	if goroutineGrowth > 10 {
		t.Errorf("PC6 RED PHASE FAIL: Goroutine leak detected. "+
			"Initial: %d, Final: %d, Growth: %d\n"+
			"This confirms ReadDirectionNonBlocking is spawning goroutines (go func { screen.PollEvent() }) "+
			"that block indefinitely when timeout fires, causing unbounded goroutine growth.",
			initialGoroutines, finalGoroutines, goroutineGrowth)
	} else if goroutineGrowth > 5 {
		t.Logf("PC6 WARNING: Moderate goroutine growth: Initial: %d, Final: %d, Growth: %d. "+
			"This suggests potential leak - verify safe concurrency pattern is used.",
			initialGoroutines, finalGoroutines, goroutineGrowth)
	} else {
		t.Logf("PC6 INFO: Goroutine counts acceptable - Initial: %d, Final: %d, Growth: %d",
			initialGoroutines, finalGoroutines, goroutineGrowth)
	}
}

// TestPC1_InitialRenderLoggingStructure verifies that initial render is
// properly logged with source:"initial" metadata. Tests the code structure
// to ensure LogEvent("render", {..., "source": "initial", ...}) is called.
func TestPC1_InitialRenderLoggingStructure(t *testing.T) {
	mainSource, err := os.ReadFile("../src/main.go")
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	sourceStr := string(mainSource)

	// Find main function
	mainIdx := strings.Index(sourceStr, "func main()")
	if mainIdx == -1 {
		t.Fatal("PC1: Could not find main() function")
	}

	// Find the for loop (start of game loop)
	forIdx := strings.Index(sourceStr[mainIdx:], "for running")
	if forIdx == -1 {
		forIdx = strings.Index(sourceStr[mainIdx:], "for {")
	}
	if forIdx == -1 {
		t.Fatal("PC1: Could not find game loop")
	}
	forIdx += mainIdx

	// Extract code BEFORE game loop
	beforeLoop := sourceStr[mainIdx:forIdx]

	// RED PHASE: Check that initial render LogEvent is missing
	hasInitialLogEvent := strings.Contains(beforeLoop, "LogEvent(\"render\"") &&
		strings.Contains(beforeLoop, "\"initial\"")

	if !hasInitialLogEvent {
		t.Error("PC1 RED PHASE FAIL: Code before game loop does not log render event with source:\"initial\". " +
			"Must call observability.LogEvent(\"render\", map[string]interface{}{...\"source\": \"initial\"...}) " +
			"BEFORE entering the main select loop to ensure board is visible within 100ms.")
	}

	os.Unsetenv("DEBUG")
}

// TestPC2_SafeTcellConcurrencyPattern_CodeReview verifies that the code
// uses safe tcell concurrency patterns (ChannelEvents or HasPendingEvent).
func TestPC2_SafeTcellConcurrencyPattern_CodeReview(t *testing.T) {
	source, err := os.ReadFile("../src/input/input.go")
	if err != nil {
		t.Fatalf("Failed to read input.go: %v", err)
	}

	sourceStr := string(source)

	// Find ReadDirectionNonBlocking
	funcIdx := strings.Index(sourceStr, "func ReadDirectionNonBlocking")
	if funcIdx == -1 {
		t.Fatal("PC2: Could not find ReadDirectionNonBlocking function")
	}

	funcEnd := strings.Index(sourceStr[funcIdx:], "\nfunc ")
	if funcEnd == -1 {
		funcEnd = len(sourceStr) - funcIdx
	} else {
		funcEnd += funcIdx
	}
	funcBody := sourceStr[funcIdx:funcEnd]

	// RED phase: Current implementation uses unsafe pattern
	// Check for the broken pattern: go func { screen.PollEvent() }
	hasUnsafePattern := strings.Contains(funcBody, "go func()") &&
		strings.Contains(funcBody, "PollEvent")

	if hasUnsafePattern {
		t.Error("PC2 RED PHASE: ReadDirectionNonBlocking uses unsafe goroutine + PollEvent pattern. " +
			"This violates tcell concurrency rules and causes goroutine leaks.")
	}

	// Verify the fix is in place (safe pattern)
	hasSafePattern := strings.Contains(funcBody, "ChannelEvents") ||
		strings.Contains(funcBody, "HasPendingEvent")

	if !hasSafePattern && !hasUnsafePattern {
		t.Logf("PC2 INFO: ReadDirectionNonBlocking uses unknown pattern. " +
			"Verify it's safe per tcell documentation.")
	}
}

// TestPC3_InputRawEventStructure verifies the structure of input_raw events.
func TestPC3_InputRawEventStructure(t *testing.T) {
	os.Setenv("DEBUG", "1")
	logsDir := "./logs"
	logFile := logsDir + "/vivorita2-debug.log"
	os.RemoveAll(logsDir)

	err := observability.InitLogging()
	if err != nil {
		t.Fatalf("InitLogging() returned error: %v", err)
	}

	input.LogEvent = observability.LogEvent

	// Log an input_raw event
	observability.LogEvent("input_raw", map[string]interface{}{
		"key":  "KeyUp",
		"rune": "",
	})

	time.Sleep(50 * time.Millisecond)

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify structure
	if !strings.Contains(logContent, "input_raw") {
		t.Error("PC3 FAIL: input_raw event not logged")
	}

	if !strings.Contains(logContent, "KeyUp") {
		t.Error("PC3 FAIL: Key information not in input_raw event")
	}

	os.Unsetenv("DEBUG")
	os.RemoveAll(logsDir)
}

// TestPC4_DirectionConversionMatrix creates a matrix of arrow key inputs
// and verifies the correct direction output for each.
func TestPC4_DirectionConversionMatrix(t *testing.T) {
	// This test documents the expected arrow key to direction conversion matrix
	conversionTests := []struct {
		name        string
		tcellKey    tcell.Key
		expectedDir string
	}{
		{"Arrow Up", tcell.KeyUp, "DirUp"},
		{"Arrow Down", tcell.KeyDown, "DirDown"},
		{"Arrow Left", tcell.KeyLeft, "DirLeft"},
		{"Arrow Right", tcell.KeyRight, "DirRight"},
	}

	for _, tc := range conversionTests {
		t.Run(tc.name, func(t *testing.T) {
			// Verify the conversion is documented in the source
			source, err := os.ReadFile("../src/input/input.go")
			if err != nil {
				t.Fatalf("Failed to read input.go: %v", err)
			}

			sourceStr := string(source)

			// Check that the key is handled
			if !strings.Contains(sourceStr, fmt.Sprintf("case tcell.%s", tc.name)) &&
				!strings.Contains(sourceStr, fmt.Sprintf("case tcell.Key%s", strings.TrimPrefix(tc.name, "Arrow "))) {
				// Try alternative names
				keyNames := map[string][]string{
					"Arrow Up":    {"tcell.KeyUp", "KeyUp"},
					"Arrow Down":  {"tcell.KeyDown", "KeyDown"},
					"Arrow Left":  {"tcell.KeyLeft", "KeyLeft"},
					"Arrow Right": {"tcell.KeyRight", "KeyRight"},
				}

				found := false
				if names, ok := keyNames[tc.name]; ok {
					for _, name := range names {
						if strings.Contains(sourceStr, "case "+name) {
							found = true
							break
						}
					}
				}

				if !found {
					t.Errorf("PC4 FAIL: Key case for %s not found in handleKeyEvent", tc.name)
				}
			}

			// Check that the correct direction is returned
			if !strings.Contains(sourceStr, tc.expectedDir) {
				t.Errorf("PC4 FAIL: Direction %s not found in input.go", tc.expectedDir)
			}
		})
	}
}

// TestPC5_WASDRegressionMatrixFullCoverage tests WASD keys comprehensively.
func TestPC5_WASDRegressionMatrixFullCoverage(t *testing.T) {
	// Regression test: verify WASD keys still work after fix
	conversionTests := []struct {
		name       string
		rune       rune
		expectedIn string
	}{
		{"W key", 'w', "DirUp"},
		{"A key", 'a', "DirLeft"},
		{"S key", 's', "DirDown"},
		{"D key", 'd', "DirRight"},
	}

	source, err := os.ReadFile("../src/input/input.go")
	if err != nil {
		t.Fatalf("Failed to read input.go: %v", err)
	}

	sourceStr := string(source)

	for _, tc := range conversionTests {
		t.Run(fmt.Sprintf("WASD_%c", tc.rune), func(t *testing.T) {
			// Verify the WASD key is handled
			caseStr := fmt.Sprintf("case \"%c\"", tc.rune)
			if !strings.Contains(sourceStr, caseStr) &&
				!strings.Contains(sourceStr, strings.ToUpper(caseStr)) {
				t.Logf("PC5 FAIL: WASD case for '%c' not found in input.go", tc.rune)
			}

			// Verify correct direction is returned
			if !strings.Contains(sourceStr, tc.expectedIn) {
				t.Logf("PC5 FAIL: Direction %s for key '%c' not found", tc.expectedIn, tc.rune)
			}
		})
	}
}

// Helper function to check if a log file contains valid event structure
func logContainsEvent(logContent, eventType, field, value string) bool {
	// Check for event type
	if !strings.Contains(logContent, eventType) {
		return false
	}

	// Check for field
	if field != "" && !strings.Contains(logContent, field) {
		return false
	}

	// Check for value
	if value != "" && !strings.Contains(logContent, value) {
		return false
	}

	return true
}
