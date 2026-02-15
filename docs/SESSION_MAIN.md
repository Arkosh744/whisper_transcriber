# Session Log — Main

## [2026-02-15 10:30:00] Project assessment

- Evaluated all Go backend files (7) and Svelte frontend (4 components)
- All code 100% written, needs build infrastructure
- Identified: whisper.cpp not built, Makefile missing, Wails bindings stale

## [2026-02-15 11:00:00] Phase 1 — Linux build infrastructure

- Cloned whisper.cpp at commit 764482c3175d (matching go.mod)
- Built static libs: libwhisper.a (859K), libggml-cpu.a (1.3M), libggml-base.a (1015K), libggml.a (65K)
- **Fixed transcriber.go**: Process() 3→4 args, NextSegment() iterator, EncoderBeginCallback for cancellation
- Verified Go compilation with absolute CGo paths
- Regenerated Wails bindings: 8 methods + 3 TS models

## [2026-02-15 11:45:00] Phase 4 — Makefile

- Created go-version/Makefile with 12 targets
- Verified `make build-check` passes

## [2026-02-15 12:15:00] Phase 2 — Windows cross-compilation

- Installed mingw-w64, libvulkan-dev
- Built whisper.cpp for Windows: _WIN32_WINNT=0x0601 (mingw headers fix), OPENMP=OFF
- Library naming fix: symlinks ggml.a → libggml.a
- Wails CLI doesn't pass CC to Go — switched to direct `go build` with `-tags desktop,production -ldflags "-w -s -H windowsgui"`
- **Result: 12 MB .exe** — PE32+ GUI x86-64 for MS Windows

## [2026-02-15 13:00:00] Phase 3 — Vulkan GPU

- Generated Windows Vulkan import library (623 functions, 513K) via dlltool
- Installed glslc (shaderc 2023.8) for SPIR-V shader compilation
- Isolated Vulkan headers to avoid /usr/include poisoning mingw cross-compiler
- Built whisper.cpp with Vulkan: 100+ GLSL shaders compiled to SPIR-V
- Fixed link order: `-lggml-vulkan -lstdc++` needed in CGO_LDFLAGS (Go duplicates flags)
- Created symlink chain: ggml-vulkan/ggml-vulkan.a → libggml-vulkan.a in ggml/src/
- Updated Makefile: whisper-lib-win-vulkan + build-win-vulkan targets with correct paths
- **Result: 56 MB .exe** — PE32+ GUI x86-64 with Vulkan GPU support, static libstdc++

## [2026-02-15 14:00:00] FFmpeg bundling + fixes

- Fixed ffmpeg.go: added `runtime.GOOS == "windows"` check for `.exe` extension
- Downloaded static ffmpeg.exe (200 MB GPL build) from BtbN/FFmpeg-Builds
- Updated Makefile: ffmpeg-win uses python3 zipfile (unzip not available in WSL2)
- Added `-extldflags '-static'` to both CPU and Vulkan build targets
- DLL dependencies verified: only vulkan-1.dll + Windows system DLLs

## [2026-02-15 14:30:00] FFmpeg on-demand download (replace bundling)

- Changed approach: instead of bundling 200 MB ffmpeg.exe, download on demand via UI button
- ffmpeg.go: added `IsFFmpegAvailable()`, `DownloadFFmpeg(ctx)`, `extractFFmpegFromZip()`
- app.go: added bound methods `IsFFmpegAvailable()`, `DownloadFFmpeg()`
- Controls.svelte: added `ffmpegReady` prop, FFmpeg notice with download button
- ProgressPanel.svelte: added FFmpeg download progress bar with stats
- App.svelte: added state management, event listeners for ffmpeg:download:* events
- Regenerated Wails bindings: added `IsFFmpegAvailable`, `DownloadFFmpeg`
- Committed: `1257279 add on-demand FFmpeg download instead of bundling`

## [2026-02-15 16:00:00] Go→root migration, cleanup

- Removed all comments from 6 Go files (~60 comment lines), preserved `//go:embed` in main.go
- Deleted Python files: main.py, app.py, config.py, requirements.txt, build.spec, core/, workers/, ui/, Makefile
- Moved Go code from go-version/ to root via `git mv` (git tracks as renames)
- Renamed module: `go-version` → `whisper-transcriber` in go.mod
- Updated wails.json: name + outputfilename → `whisper-transcriber`
- Replaced .gitignore: Python → Go/Wails oriented
- Rewrote README.md: Go/Wails/Vulkan instead of Python/CustomTkinter/CUDA
- Cleaned up: venv/, __pycache__/, go-version/ directory

## [2026-02-15 17:00:00] Finalize, push, merge, release

- Ran `go mod tidy`: removed 7 unused dependencies (go-audio/*, testify, spew, difflib, yaml)
- Fixed `gh` CLI: switched git protocol from SSH to HTTPS (port 443 issue in WSL2)
- Pushed `feat/go-rewrite` branch to GitHub
- Created PR #1: "Rewrite: Python → Go/Wails with Vulkan GPU"
- Merged PR #1 into master (fast-forward), deleted branch
- Created release v1.0.0 with two Windows binaries:
  - `whisper_transcriber.exe` (12 MB, CPU)
  - `whisper_transcriber_vulkan.exe` (56 MB, Vulkan GPU)
- Release: https://github.com/Arkosh744/whisper_transcriber/releases/tag/v1.0.0

## [2026-02-15 18:00:00] Frontend UX improvements (6 changes)

- **Error display**: handleBrowse shows errors via statusMessage instead of console.error; handleDownloadModel/handleDownloadFFmpeg wrapped in try/catch
- **Output path**: transcription:complete handler saves outputPath; FileList shows path below completed files
- **Drag & drop**: OnFileDrop/OnFileDropOff from Wails runtime; AddFiles Go binding; visual drag-over state in empty FileList
- **LocalStorage**: language and outputFormat persisted via reactive statements; restored in onMount
- **Cancel download**: CancelDownload binding; cancel button in ProgressPanel progress bars; handleCancelDownload handler
- **Cancel race condition**: cancelling state flag; handleCancel sets cancelling=true; batch:complete shows correct message; Controls disables cancel button while cancelling
- Files changed: App.svelte, FileList.svelte, ProgressPanel.svelte, Controls.svelte

## [2026-02-15 19:00:00] Bugfixes + polish (14 changes across all tiers)

### Tier 1 — Critical bugs
- **#1 ExtractAudio context**: signature `ExtractAudio(ctx, inputPath)`, `exec.CommandContext` kills ffmpeg on cancel, temp WAV cleaned up
- **#2 Batch context**: `context.WithCancel(a.ctx)` instead of `context.Background()` — batch goroutine terminates on app shutdown
- **#3 Pre-allocate WAV samples**: `readWavSamples` computes `(fileSize-44)/2` and pre-allocates slice capacity
- **#4 defer out.Close()**: `DownloadFFmpeg` uses defer for file handle safety on error paths
- **#5 FFmpeg check**: `StartTranscription` validates FFmpeg availability before proceeding

### Tier 2 — Core UX
- **#6 Error display**: `handleBrowse`, `handleDownloadModel`, `handleDownloadFFmpeg` show errors via statusMessage
- **#7 Output path**: `transcription:complete` saves outputPath; FileList shows path for completed files
- **#8 Drag & drop**: `OnFileDrop` from Wails runtime + `AddFiles` Go method; visual drag-over state
- **#9 Format validation**: `WriteOutput` validates format against allowed set before switch

### Tier 3 — Polish
- **#10 Retry downloads**: `httpGetWithRetry(ctx, url, 3)` with exponential backoff; used by both model and ffmpeg downloaders
- **#11 LocalStorage settings**: language and outputFormat persisted/restored via localStorage
- **#12 Cancel downloads**: `CancelDownload()` method + cancel button in ProgressPanel; download goroutines use cancellable context
- **#13 Cancel race condition**: `cancelling` state flag prevents premature isRunning reset; Controls disables cancel button
- **#14 UAC fallback**: `appDataDir()` writes next to .exe, falls back to `%APPDATA%/WhisperTranscriber` if read-only

### New files
- `http.go` — `httpGetWithRetry` helper
- `paths.go` — `appDataDir` helper with write-test fallback

### Files changed
- Go: app.go, ffmpeg.go, model.go, transcriber.go, formatter.go
- Frontend: App.svelte, FileList.svelte, ProgressPanel.svelte, Controls.svelte
- Bindings: App.js, App.d.ts (added AddFiles, CancelDownload)
- Verified: `make build-check` passes, gopls diagnostics clean, no `context.Background` in batch code

## [2026-02-15 20:00:00] Рефакторинг: 3-слойная Clean Architecture

### Фаза 1 — Domain + Infrastructure
- Создан `internal/domain/`: model.go (FileItem, Segment, TranscriptionResult, LangOption, TranscriptionConfig, GenerateID), errors.go (ErrModelNotLoaded, ErrFFmpegNotFound), progress.go (ProgressFunc, StatusFunc), interfaces.go (Transcriber, ModelManager, FFmpegService, Formatter, FileQueue)
- Создан `internal/infrastructure/`: paths.go (AppDataDir), http.go (HTTPGetWithRetry)

### Фаза 2 — Сервисы
- `internal/service/file_queue.go` — FileQueue с mutex, Add/Remove/Clear/Snapshot/UpdateStatus
- `internal/service/formatter.go` — Formatter с WriteOutput (txt/srt/json/md)
- `internal/service/ffmpeg.go` — FFmpegSvc: IsAvailable, Download (callback вместо EventsEmit), ExtractAudio
- `internal/service/model_manager.go` — ModelMgr: ModelPath, IsModelAvailable, DownloadModel (callback)
- `internal/service/transcriber.go` — WhisperTranscriber: LoadModel, TranscribeFile (onProgress callback вместо wailsCtx)
- `internal/service/batch.go` — BatchProcessor: Run с callback-ами (onStatus, onComplete, onDone)

### Фаза 3 — Переключение
- Создан `events.go` — callback-хелперы (downloadProgressCb, fileStatusCb) с замыканиями на EventsEmit
- Переписан `app.go` — тонкий Wails-адаптер, все зависимости через конструктор (DI)
- Обновлён `main.go` — DI wiring: создание всех сервисов и BatchProcessor
- Удалены старые файлы: types.go, transcriber.go, model.go, ffmpeg.go, formatter.go, paths.go, http.go

### Фаза 4 — Верификация
- `go build ./...` — OK
- `go vet ./...` — OK
- `gopls diagnostics` на 15 файлах — 0 ошибок

### Файлы изменены
- Новые (10): internal/domain/{model,errors,progress,interfaces}.go, internal/infrastructure/{paths,http}.go, internal/service/{file_queue,formatter,ffmpeg,model_manager,transcriber,batch}.go, events.go
- Переписаны (2): app.go, main.go
- Удалены (7): types.go, transcriber.go, model.go, ffmpeg.go, formatter.go, paths.go, http.go
