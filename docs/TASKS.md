# Tasks — Whisper Transcriber Go Rewrite

## In Progress

- [x] [100%] Clean Architecture рефакторинг — 3 слоя (domain/service/infrastructure), DI, callbacks вместо EventsEmit (2026-02-15)

## Completed

- [x] [100%] Bugfixes + polish (14 items) — critical bugs, UX, polish across all tiers (2026-02-15)
- [x] [100%] Go→root migration — moved Go from go-version/ to root, deleted Python, updated go.mod/wails.json/.gitignore/README (2026-02-15)

- [x] [100%] Vulkan GPU build — 56 MB PE32+ with SPIR-V shaders, static libstdc++ (2026-02-15)
- [x] [100%] FFmpeg on-demand download — download button in UI, auto-extract (2026-02-15)

- [x] [100%] Go backend code — types, model, ffmpeg, transcriber, formatter, app, main (2026-02-14)
- [x] [100%] Svelte frontend — App, FileList, Controls, ProgressPanel, style (2026-02-14)
- [x] [100%] whisper.cpp static lib (Linux) — CPU-only build, cmake (2026-02-15)
- [x] [100%] whisper.cpp static lib (Windows) — cross-compile with mingw, CPU-only (2026-02-15)
- [x] [100%] Go CGo compilation — verified with whisper.cpp bindings (2026-02-15)
- [x] [100%] Wails JS bindings — regenerated 8 methods + 3 models (2026-02-15)
- [x] [100%] Windows .exe — 12 MB cross-compiled PE32+ GUI (2026-02-15)
- [x] [100%] Makefile — 12 targets for build automation (2026-02-15)
- [x] [100%] Fix transcriber.go API — Process() 4 args, NextSegment() iterator (2026-02-15)

## Backlog

- [ ] Unit tests for formatters and WAV reader
- [ ] Integration test with tiny model
- [ ] Windows GUI testing
- [ ] NSIS installer
