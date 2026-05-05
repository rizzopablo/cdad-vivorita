# Progress

## Done
001-snake-game — Clon de Snake en terminal con persistencia de high score
002-game-loop-observability — Observabilidad del game loop: logging JSON estructurado, snapshots de tablero, migración de input de bash subshell a tcell.PollEvent. 13 postcondiciones implementadas y testeadas, suite completa verde (23 tests).
003-input-capture-investigation — Fix del bug de captura de input: introducción de DirNone, fallback a dirección actual, método Snake.Direction(). 8 postcondiciones, 22/22 tests verdes.
004-gameplay-polish — Snake estática al inicio (firstInputReceived flag), ticker 200ms, InputPollTimeout const. 10 postcondiciones verdes, suite completa 34 tests.
005-input-capture-broken — Fix del bug de captura de input tcell raw mode: initial render antes del loop, thread-safe concurrency. 6 postcondiciones, 20+ tests verdes.
006-vertical-direction-inversion — Fix de semántica de dirección: DirUp = Y-1 (arriba), DirDown = Y+1 (abajo). Hotfix post-merge, test semantics agregado.
007-continuous-input-response — Fix del movimiento continuo en primer input: variable currentDirection persistente en lugar de pendingDirection reseteada. Movimiento automático cada 200ms desde primer input. CDAD cycle fallido (tests validaban buffer, no continuidad); manual hotfix aplicado.

## In Progress
<!-- No active features -->

## Queued
<!-- Pendiente priorización -->
