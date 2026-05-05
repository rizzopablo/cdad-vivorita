# Landscape: Snake Game (Terminal)

## Feature
001-snake-game — Clon clásico del juego de la víbora en terminal.

## Stack
- **Lenguaje**: Go
- **Interfaz**: Terminal (texto/unicode)
- **Controles**: Teclado (flechas / WASD)

## Componentes identificados

### Game Engine
- Loop principal con tick de tiempo
- Control de dirección por teclado
- Detección de colisiones (bordes, cuerpo propio)

### Game State
- Posición de la víbora (lista de coordenadas)
- Dirección actual
- Posición de la comida
- Puntaje actual
- Estado del juego (jugando, pausado, game over)

### Render
- Dibujar tablero en terminal (caracteres ASCII/unicode)
- Render de víbora, comida, bordes, puntaje

### Score Persistence
- Guardar/cargar high score (archivo local, ej: JSON)

## Decisiones
- **Lenguaje**: Go (se elimina `pyproject.toml`, proyecto 100% Go)
- **Tablero**: Tamaño clásico original (40x20 celdas)
- **Velocidad**: Fija

## APIs/Hooks externos
- Ninguno. Todo local.
