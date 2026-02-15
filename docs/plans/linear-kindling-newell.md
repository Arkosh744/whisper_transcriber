# План: удалить комменты из Go, заменить Python на Go, переписать README

## Context

Go rewrite завершён на 100% и живёт в `go-version/`. Python-версия осталась в корне. Нужно:
1. Убрать все комментарии из Go-кода (doc-comments, inline, section separators)
2. Переместить Go-код из `go-version/` в корень, удалить Python-файлы
3. Переписать README под Go/Wails вместо Python/CustomTkinter

## Task 1: Удалить комменты из Go-кода

7 файлов, ~40 doc-comment строк + ~20 inline.

**Сохранить**: `//go:embed all:frontend/dist` в main.go (компилятор-директива, не комментарий)

| Файл | Что убрать |
|------|-----------|
| `types.go` | `// FileItem represents...`, `// pending \| processing...`, `// 0-100`, `// "auto"...`, `// seconds`, все doc-comments на типах |
| `model.go` | `// ModelManager handles...`, `// NewModelManager creates...`, `// Model is stored...`, `// ModelPath returns...`, `// IsModelAvailable...`, `// DownloadModel...`, `// 64 KB chunks`, `// cleanup on error` |
| `ffmpeg.go` | `// ffmpegBin returns...`, `// 1. Check next to executable`, `// 2. Fall back to system PATH`, `// ffmpegLocalPath...`, `// IsFFmpegAvailable...`, `// DownloadFFmpeg...`, `// Emits ffmpeg:download...`, `// ExtractAudio converts...`, `// whisper.cpp only accepts...`, `// Returns the path...`, `// Download zip to temp file`, `// Extract ffmpeg.exe from zip`, `// 16 kHz sample rate...`, `// mono`, `// 16-bit PCM`, `// overwrite`, `// extractFFmpegFromZip...` |
| `transcriber.go` | `// Transcriber wraps...`, `// NewTranscriber...`, `// SetContext...`, `// LoadModel...`, `// IsLoaded...`, `// TranscribeFile processes...`, `// Emits "transcription:progress"...`, `// Read WAV samples...`, `// Create whisper context`, `// Configure language`, `// fallback to auto...`, `// Use EncoderBeginCallback...`, `// ProgressCallback for...`, `nil, // segment callback...`, `// Collect segments via NextSegment`, `// Close releases...`, `// readWavSamples reads...`, `// normalized to...`, `// safe because...`, `// Skip WAV header...`, `// Read 16-bit samples` |
| `formatter.go` | `// WriteOutput writes...`, `// Returns the output...` |
| `app.go` | Все doc-comments на структурах и методах, `// --- Bound methods ---`, `// Ensure model`, `// Load model if not already loaded`, `// Create cancellable context for the batch`, `// Step 1:...`, `// Step 2:...`, `// Step 3:...`, `// Cleanup temp WAV`, `// File transcribed successfully` |
| `main.go` | Ничего — только `//go:embed` (сохранить) |

## Task 2: Переместить Go в корень, удалить Python

### 2.1 Удалить Python-файлы (git rm)

```
main.py
app.py
config.py
requirements.txt
build.spec
core/           (transcriber.py, model_downloader.py, formatters.py, media_info.py, __init__.py)
workers/        (transcribe_worker.py, __init__.py)
ui/             (file_list.py, controls.py, progress.py, __init__.py)
Makefile        (Python Makefile — заменится Go Makefile)
```

Также удалить с диска (не в git): `venv/`, `__pycache__/`

### 2.2 Переместить go-version/* в корень

```bash
# Файлы Go
git mv go-version/app.go go-version/ffmpeg.go go-version/formatter.go \
       go-version/model.go go-version/transcriber.go go-version/types.go \
       go-version/main.go go-version/go.mod go-version/go.sum \
       go-version/wails.json go-version/Makefile .

# Директории
git mv go-version/frontend .
git mv go-version/build .
```

Файлы не в git (скопировать вручную + удалить go-version/):
- `go-version/third_party/` → `third_party/`

### 2.3 Обновить пути и имена

**go.mod**: `module go-version` → `module whisper-transcriber`

**wails.json**: `"name": "go-version"` → `"name": "whisper-transcriber"`, `"outputfilename": "go-version"` → `"outputfilename": "whisper-transcriber"`

**Makefile**: путей обновлять не нужно — используют `$(CURDIR)` (автоматически корень)

### 2.4 Обновить .gitignore

Заменить Python-ориентированный .gitignore на Go:

```gitignore
# Build
build/bin/
frontend/dist/
frontend/node_modules/

# whisper.cpp (cloned + built at build time)
third_party/

# Models (downloaded at runtime)
models/

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db
```

## Task 3: Переписать README

Новый README.md:
- Заголовок: Whisper Transcriber
- Описание: Desktop GUI + whisper.cpp (Go/Wails v2), Vulkan GPU
- Бейджи: Go, Wails, Vulkan, License
- Features: Vulkan GPU (~10-50x realtime), 12-56 MB .exe, on-demand model/FFmpeg download, batch processing, 16 languages, 4 formats, dark theme
- ASCII screenshot (как был, но обновить)
- Quick Start: `make whisper-lib-win && make build-win`
- Requirements: Go 1.23+, Wails CLI, mingw-w64 (для кросс-компиляции)
- Project Structure: корневой layout с Go-файлами
- How It Works: Video → FFmpeg → WAV → whisper.cpp → segments → formatter → file
- Makefile targets (из текущего Makefile)
- License: MIT

## Критические файлы

| Файл | Действие |
|------|----------|
| `go-version/*.go` (7 шт.) | убрать комменты, переместить в корень |
| `go-version/go.mod` | переименовать module, переместить |
| `go-version/wails.json` | обновить name, переместить |
| `.gitignore` | заменить Python → Go |
| `README.md` | полностью переписать |
| `main.py, app.py, config.py, ...` | удалить |
| `core/, workers/, ui/` | удалить |

## Порядок выполнения

1. Убрать комменты из Go-файлов (пока они ещё в `go-version/`)
2. `git rm` Python-файлов
3. `git mv` Go-файлов в корень
4. Обновить go.mod, wails.json, .gitignore
5. Написать новый README.md
6. Удалить пустой `go-version/`
7. `rm -rf venv/ __pycache__/`

## Верификация

1. `go build ./...` — компиляция из корня (с CGo flags)
2. Проверить что `//go:embed` сохранён в main.go
3. `grep -r "^//" *.go` — убедиться что остался только `//go:embed`
4. `ls *.py` — Python-файлов нет
5. `cat go.mod` — module whisper-transcriber
6. `cat README.md` — Go-ориентированный README
