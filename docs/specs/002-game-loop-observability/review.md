# Review — 002-game-loop-observability

## Bloqueantes

Sin hallazgos bloqueantes.

Todas las postcondiciones (PC1-PC13) se cumplen correctamente. Las invariantes I1-I5 están satisfechas. El contrato público coincide con la implementación observable.

---

## Opcionales

### 1. Error silencioso en JSON marshaling

**Ubicación:** `src/observability/observability.go:100`

**Problema:** La función `writeJSONLog()` ignora errores de `json.Marshal()` sin logging ni propagación. Si el marshal falla, la entrada JSON no se escriba en el log, pero la función retorna silenciosamente.

**Severidad:** Opcional (no bloqueante)

**Impacto:**
- Invariante I1 ("logs son JSON válido") se violaría si marshal falla.
- PC1 requiere entrada inicial en log; si marshal de `event:"init"` falla, el archivo queda vacío sin diagnostic.
- Escenario real: marshal de tipos simples (string, int, struct) casi nunca falla, pero cuando `data` contiene tipos no serializables, el error se pierde.

**Sugerencia:**
```go
if jsonData, err := json.Marshal(entry); err == nil {
    file.Write(append(jsonData, '\n'))
} else {
    logFile.Write([]byte(fmt.Sprintf("{\"timestamp\":\"%s\",\"event\":\"marshal_error\",\"data\":{\"error\":\"%s\"}}\n", 
        time.Now().Format(time.RFC3339), err.Error())))
}
```

---

### 2. Duplicación de mapeo input → game direction

**Ubicación:** 
- `src/input/input.go:64-105` (mapeo interno w/s/a/d → Direction)
- `src/main.go:99-111` (función `convertInputToGameDirection` replica mapeo)

**Problema:** Code smell: dos lugares mapean input.Direction → game.Direction con lógica similar. La responsabilidad de conversión está difusa.

**Severidad:** Opcional (code smell, no funcional)

**Impacto:**
- Mantenimiento: si se agregan nuevas direcciones (ej: diagonales), hay dos puntos de cambio.
- Confusión: `ReadDirectionNonBlocking()` ya mapea internamente teclas → Direction, pero main.go re-mapea Direction → game.Direction.

**Sugerencia:**
1. Mantener conversion en input.go (ya existe internamente).
2. Eliminar `convertInputToGameDirection` de main.go si no aporta valor.
3. Alternativa: ajustar spec para explicitar que input devuelve `game.Direction` directamente, eliminando la conversión en main.go.

---

### 3. Divergencia arquitectural: dependency injection vs import directo

**Ubicación:** 
- Spec PC10: "ReadDirectionNonBlocking() loggea eventos con `observability.LogEvent`" (implícito)
- Implementación: `src/input/input.go:11` (variable `LogEvent` injectable)
- `src/main.go:30` (asigna `input.LogEvent = observability.LogEvent`)

**Problema:** La firma pública del spec (línea 29) define `observability.LogEvent` como función global. La implementación usa dependency injection para que input pueda loggear sin importar observability directamente. El spec no documenta este mecanismo.

**Severidad:** Opcional (divergencia menor)

**Impacto:**
- Spec dice "input loggea con observability.LogEvent", pero código usa input.LogEvent (variable).
- Diseño técnico correcto (evita import circular, mantiene boundaries), pero spec no refleja arquitectura.
- Postcondiciones se cumplen (PC10, PC2): el logging se ejecuta, los tests pasan.

**Sugerencia:**
- Actualizar spec sección "Firma pública" o "Contrato" para documentar dependency injection explícitamente.
- Agregar nota: "input module recibe LogEvent como variable injectable; main.go asigna observability.LogEvent al inicio."

---

### 4. Patrón repetitivo en switch de handleKeyEvent

**Ubicación:** `src/input/input.go:54-123`

**Problema:** Switch con 6-7 casos similares, cada uno:
1. Loggea input_converted
2. Retorna Direction

**Severidad:** Opcional (estilo)

**Sugerencia:** Helper para reducir repetición:
```go
func logAndReturn(dir string, d Direction) (Direction, error) {
    if LogEvent != nil {
        LogEvent("input_converted", map[string]interface{}{"direction": dir})
    }
    return d, nil
}
```

---

## Resumen

- **Bloqueantes:** 0
- **Opcionales:** 4

**Estado:** Implementación cumple spec. Hallazgos opcionales son mejoras de robustez, mantenibilidad y claridad documental. No impiden merge.