package tests

import (
	"os"
	"strings"
	"testing"
	"vivorita2/src/input"
)

// PC1: ReadDirectionNonBlocking retorna DirNone en timeout
func TestReadDirectionNonBlocking_TimeoutReturnsDirNone(t *testing.T) {
	// PC1: En timeout debe retornar DirNone, no DirUp
	// La implementación actual retorna DirNone en timeout (con screen nil simula timeout)

	dir, err := input.ReadDirectionNonBlocking(nil)
	if err != nil {
		t.Fatalf("ReadDirectionNonBlocking returned error: %v", err)
	}

	// PC1: En timeout debe retornar DirNone
	if dir != input.DirNone {
		t.Errorf("PC1 FAIL: ReadDirectionNonBlocking returned %v on timeout, expected DirNone", dir)
	}
}

// PC2: ReadDirectionNonBlocking retorna DirNone en tecla no mapeada (con log input_error)
func TestReadDirectionNonBlocking_UnmappedKeyReturnsDirNoneWithLog(t *testing.T) {
	// PC2: Verificar que el código maneja teclas no mapeadas retornando DirNone
	// y logueando input_error

	source, err := os.ReadFile("../src/input/input.go")
	if err != nil {
		t.Fatalf("Failed to read input.go: %v", err)
	}

	sourceStr := string(source)

	// Extraer la función handleKeyEvent para analizarla
	funcStart := strings.Index(sourceStr, "func handleKeyEvent")
	if funcStart == -1 {
		t.Fatal("PC2: Could not find handleKeyEvent function")
	}

	// Encontrar el cuerpo de la función (aproximado)
	funcEnd := strings.Index(sourceStr[funcStart:], "\nfunc keyToString")
	if funcEnd == -1 {
		funcEnd = len(sourceStr) - funcStart
	} else {
		funcEnd += funcStart
	}
	funcBody := sourceStr[funcStart:funcEnd]

	// Buscar el switch case para KeyRune
	keyRuneIdx := strings.Index(funcBody, "case tcell.KeyRune:")
	if keyRuneIdx == -1 {
		t.Fatal("PC2: Could not find KeyRune case")
	}

	// Extraer el switch interno para las runes (hasta el próximo case o default del switch externo)
	runeSwitchStart := strings.Index(funcBody[keyRuneIdx:], "switch strings.ToLower")
	if runeSwitchStart == -1 {
		t.Fatal("PC2: Could not find rune switch")
	}
	runeSwitchStart += keyRuneIdx

	// Encontrar el default del switch de runes
	runeSwitchBody := funcBody[runeSwitchStart:]
	defaultIdx := strings.Index(runeSwitchBody, "default:")
	if defaultIdx == -1 {
		t.Fatal("PC2: Could not find default case in rune switch")
	}

	// Extraer el bloque default (hasta el siguiente case o el final)
	defaultBlock := runeSwitchBody[defaultIdx:]
	blockEnd := strings.Index(defaultBlock, "\n\t\tcase ")
	if blockEnd == -1 {
		blockEnd = strings.Index(defaultBlock, "\n\t\t}")
	}
	if blockEnd == -1 {
		blockEnd = len(defaultBlock)
	}
	defaultBlock = defaultBlock[:blockEnd]

	// Verificar que el default case retorna DirNone (no DirUp)
	if !strings.Contains(defaultBlock, "return DirNone, nil") {
		t.Errorf("PC2 FAIL: default case for unmapped keys should return DirNone, found:\n%s", defaultBlock)
	}

	// Verificar que se loguea input_error en el default
	if !strings.Contains(defaultBlock, "LogEvent(\"input_error\"") {
		t.Errorf("PC2 FAIL: default case should log 'input_error' event, found:\n%s", defaultBlock)
	}

	// También verificar el default del switch externo (KeyRune)
	keyRuneDefaultIdx := strings.Index(funcBody[keyRuneIdx:], "default:")
	if keyRuneDefaultIdx != -1 {
		keyRuneDefaultIdx += keyRuneIdx
		keyRuneDefaultBlock := funcBody[keyRuneDefaultIdx:]
		blockEnd := strings.Index(keyRuneDefaultBlock, "\n\t\t}")
		if blockEnd != -1 {
			keyRuneDefaultBlock = keyRuneDefaultBlock[:blockEnd]
		}

		if !strings.Contains(keyRuneDefaultBlock, "return DirNone, nil") {
			t.Errorf("PC2 FAIL: default case for KeyRune switch should return DirNone, found:\n%s", keyRuneDefaultBlock)
		}
		if !strings.Contains(keyRuneDefaultBlock, "LogEvent(\"input_error\"") {
			t.Errorf("PC2 FAIL: default case for KeyRune switch should log 'input_error', found:\n%s", keyRuneDefaultBlock)
		}
	}
}

// PC7: input.DirNone definido y compila
func TestInput_DirNoneConstantExists(t *testing.T) {
	// PC7: Verificar que input.DirNone esté definido y tenga un valor único

	// Verificar que DirNone existe (esto fallaría en compilación si no existiera)
	var dir input.Direction = input.DirNone

	// Verificar que DirNone tiene un valor distinto a todas las demás direcciones (I1)
	if dir == input.DirUp {
		t.Error("PC7/I1 FAIL: DirNone should have a different value than DirUp")
	}
	if dir == input.DirDown {
		t.Error("PC7/I1 FAIL: DirNone should have a different value than DirDown")
	}
	if dir == input.DirLeft {
		t.Error("PC7/I1 FAIL: DirNone should have a different value than DirLeft")
	}
	if dir == input.DirRight {
		t.Error("PC7/I1 FAIL: DirNone should have a different value than DirRight")
	}
	if dir == input.DirPause {
		t.Error("PC7/I1 FAIL: DirNone should have a different value than DirPause")
	}
	if dir == input.DirQuit {
		t.Error("PC7/I1 FAIL: DirNone should have a different value than DirQuit")
	}

	// Verificar que DirNone es el valor 6 (después de DirQuit que es 5)
	// En Go, iota asigna valores secuenciales empezando desde 0
	// DirUp=0, DirDown=1, DirLeft=2, DirRight=3, DirPause=4, DirQuit=5, DirNone=6
	if int(dir) != 6 {
		t.Errorf("PC7 FAIL: DirNone should have value 6, got %d", int(dir))
	}
}
