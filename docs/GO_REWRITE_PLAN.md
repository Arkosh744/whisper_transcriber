# Go Rewrite Plan — Whisper Transcriber

## Why

Python версия = ~700 МБ .exe (CUDA DLLs + CTranslate2 + Python runtime).
Go + whisper.cpp = ~15 МБ бинарник + 574 МБ модель = ~590 МБ итого, без Python.

## Stack

| Компонент | Технология | Размер |
|-----------|-----------|--------|
| Backend | Go + whisper.cpp (CGo) | ~10-15 МБ |
| Frontend | Wails v2 + Svelte | встроен в бинарник |
| Модель | GGML large-v3-turbo Q5_0 | ~574 МБ |
| Аудио | FFmpeg (subprocess) | ~90 МБ (бандл) или системный |
| GPU | whisper.cpp CUDA/Vulkan (опционально) | — |

## Architecture

```
┌──────────────────────────────────────┐
│  Wails Window (WebView2)             │
│  ┌────────────────────────────────┐  │
│  │  Svelte Frontend               │  │
│  │  FileList / Controls / Progress │  │
│  └──────────┬─────────────────────┘  │
│             │ Wails Bindings          │
│  ┌──────────▼─────────────────────┐  │
│  │  Go Backend                    │  │
│  │  app.go → transcriber.go      │  │
│  │  ffmpeg.go / formatter.go     │  │
│  │  model.go                     │  │
│  └──────────┬─────────────────────┘  │
│             │ CGo                     │
│  ┌──────────▼─────────────────────┐  │
│  │  whisper.cpp (static lib)      │  │
│  └────────────────────────────────┘  │
└──────────────────────────────────────┘
```

## Project Structure

Код Go-версии живёт в `go-version/` внутри репо (отдельная ветка `feat/go-rewrite`).

```
whisper_transcriber/
├── ...                          # Python version (master)
├── go-version/                  # Go rewrite
│   ├── main.go                  # Wails entry point
│   ├── app.go                   # Wails bindings, orchestration
│   ├── transcriber.go           # whisper.cpp wrapper
│   ├── ffmpeg.go                # video → 16kHz WAV
│   ├── formatter.go             # TXT/SRT/JSON/MD output
│   ├── model.go                 # model download + cache
│   ├── types.go                 # shared DTOs
│   ├── go.mod / go.sum
│   ├── wails.json
│   ├── Makefile
│   ├── frontend/
│   │   ├── package.json
│   │   ├── vite.config.js
│   │   ├── index.html
│   │   └── src/
│   │       ├── main.js
│   │       ├── App.svelte
│   │       └── lib/
│   │           ├── FileList.svelte
│   │           ├── Controls.svelte
│   │           ├── ProgressPanel.svelte
│   │           └── theme.css
│   └── build/
│       └── windows/
│           ├── icon.ico
│           └── info.json
└── docs/
    ├── GO_REWRITE_PLAN.md       # this file
    └── RESEARCH.md              # research notes
```

## Implementation Steps

### Step 1: Git branch + Wails scaffold (30 min)
```bash
git checkout -b feat/go-rewrite
mkdir go-version && cd go-version
wails init -n whisper_transcriber -t svelte-ts
wails dev  # verify window opens
```

### Step 2: Build whisper.cpp static lib (1-2 hours)
```bash
git clone https://github.com/ggml-org/whisper.cpp.git /tmp/whisper.cpp
cd /tmp/whisper.cpp
mkdir build && cd build
cmake .. -DBUILD_SHARED_LIBS=OFF -DCMAKE_BUILD_TYPE=Release
cmake --build . -j
```
Ключевые файлы: `libwhisper.a`, `libggml*.a`
CGo flags: `CGO_LDFLAGS="-lwhisper -lggml -lggml-base -lggml-cpu -lm -lpthread -lstdc++"`

### Step 3: types.go (15 min)
```go
type FileItem struct {
    ID, Path, Name, Status, Error string
    Progress int
}
type TranscriptionConfig struct {
    Language, OutputFormat string
}
type Segment struct {
    Index int; Start, End float64; Text string
}
```

### Step 4: model.go — download + cache (1 hour)
- Путь: `$APPDATA/whisper_transcriber/models/ggml-large-v3-turbo-q5_0.bin`
- URL: `https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3-turbo-q5_0.bin`
- Прогресс через Wails events (`model:download:progress`)
- Скачивание в `.tmp`, rename после завершения

### Step 5: ffmpeg.go — audio extraction (45 min)
```go
func ExtractAudio(input string) (wavPath string, err error)
// ffmpeg -i input -ar 16000 -ac 1 -c:a pcm_s16le -y output.wav
```
Ищет ffmpeg: 1) рядом с .exe, 2) в PATH.

### Step 6: transcriber.go — core engine (2-3 hours)
- `LoadModel(path)` — загрузка GGML модели
- `TranscribeFile(ctx, fileID, wavPath, lang)` — с прогрессом и cancel
- WAV reader: 16-bit PCM → `[]float32`
- Progress через `EncoderBeginCallback` + segment callback
- Cancel через `context.Context`

### Step 7: formatter.go — output writers (45 min)
- `WriteOutput(result, sourcePath, format)` → file next to source
- TXT: `[MM:SS] text`
- SRT: standard subtitle format
- JSON: `{segments: [{start, end, text}]}`
- MD: `**[MM:SS]** text`

### Step 8: app.go — Wails bindings (2 hours)
Exposed methods:
- `BrowseFiles()` → native file dialog
- `StartTranscription(config)` → goroutine batch
- `CancelTranscription()`
- `IsModelAvailable()`
- `GetSupportedLanguages()`

Events (Go → JS):
- `file:status` — `{fileID, status, progress, error}`
- `transcription:progress` — `{fileID, progress}`
- `model:download:progress` — `int (0-100)`
- `batch:complete`

### Step 9: main.go (15 min)
```go
wails.Run(&options.App{
    Title: "Whisper Transcriber",
    Width: 900, Height: 650,
    Windows: &windows.Options{Theme: windows.Dark},
    Bind: []interface{}{app},
})
```

### Step 10: Svelte frontend (3-4 hours)
4 компонента:
- `App.svelte` — layout + event listeners
- `FileList.svelte` — browse/add/remove files
- `Controls.svelte` — language, format, start/cancel
- `ProgressPanel.svelte` — per-file progress bars

### Step 11: Makefile (1 hour)
Targets: `whisper-lib`, `dev`, `build`, `model`, `clean`

### Step 12: Testing + polish (1-2 hours)

## Total Estimate: ~14-18 часов

## Risks

| Risk | Mitigation |
|------|-----------|
| whisper.cpp Go bindings не компилятся на Windows | Собрать whisper.cpp отдельно через CMake, вручную задать CGo flags |
| CGo линкует много GGML подбиблиотек | Проверить CMake output, перечислить все `.a` файлы |
| Cancel задержка (несколько секунд) | `EncoderBeginCallback` проверяет ctx между сегментами |
| Скачивание 574 МБ модели обрывается | Скачивание в `.tmp` + rename; в будущем HTTP Range resume |

## Size Comparison

| | Python (текущий) | Go (новый) |
|---|---|---|
| Бинарник | ~700 МБ | ~15 МБ |
| Модель | 2.9 ГБ | 574 МБ |
| FFmpeg | не нужен (PyAV) | ~90 МБ (бандл) |
| **Итого** | **~3.6 ГБ** | **~680 МБ** |
| Без модели | ~700 МБ | **~105 МБ** |
