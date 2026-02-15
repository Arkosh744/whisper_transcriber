# Whisper Transcriber

Desktop GUI app for transcribing video/audio files using [whisper.cpp](https://github.com/ggml-org/whisper.cpp) with Vulkan GPU acceleration.

Drop files → get timestamped transcripts. Single `.exe`, no Python, no CUDA drivers.

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)
![Wails](https://img.shields.io/badge/Wails-v2-red)
![Vulkan](https://img.shields.io/badge/Vulkan-GPU%20accelerated-blue)
![License](https://img.shields.io/badge/license-MIT-gray)

## Features

- **Vulkan GPU acceleration** — whisper.cpp with Vulkan backend, ~10-50x realtime speed
- **Tiny portable binary** — 12-56 MB `.exe` depending on build (CPU / Vulkan)
- **On-demand downloads** — model (~574 MB) and FFmpeg fetched at first launch, not bundled
- **Direct video input** — MP4, MKV, AVI, MOV, WebM, plus audio formats
- **Multiple output formats** — TXT, SRT, JSON, Markdown
- **16 languages** — auto-detect or manual selection
- **Batch processing** — queue multiple files, per-file progress, cancel anytime
- **Dark theme** — native look via Wails/WebView2

## Screenshot

```
┌─────────────────────────────────────────────┐
│  Whisper Transcriber                        │
├─────────────────────────────────────────────┤
│  [Browse Files]  [Add More]     [Clear All] │
│  ▶ video_01.mp4              1420 MB        │
│  ✅ video_02.mkv              890 MB        │
│  ⏳ video_03.mov              650 MB        │
├─────────────────────────────────────────────┤
│  Language: [Auto-detect]  Format: [TXT]     │
│  [▶ Start Transcription]  [Cancel]          │
├─────────────────────────────────────────────┤
│  ████████████████░░░░░░  68%                │
│  [2/3] video_03.mov                         │
└─────────────────────────────────────────────┘
```

## Quick Start

```bash
# Build whisper.cpp with Vulkan support (requires mingw-w64 for cross-compilation)
make whisper-lib-win-vulkan

# Cross-compile Windows .exe with Vulkan GPU backend
make build-win-vulkan
```

Output: `build/bin/whisper_transcriber.exe`

On first launch the app downloads the GGML model (~574 MB) and FFmpeg automatically.

## Requirements

- Go 1.23+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2
- Node.js (for frontend build)
- mingw-w64 (for Windows cross-compilation from Linux)
- `glslc` (for Vulkan shader compilation, part of Vulkan SDK)

## Project Structure

```
whisper-transcriber/
├── main.go              # Wails app entry, embed frontend
├── app.go               # App struct, bound methods for JS
├── transcriber.go       # whisper.cpp wrapper, progress events
├── model.go             # GGML model discovery + download
├── ffmpeg.go            # FFmpeg discovery, download, audio extraction
├── formatter.go         # TXT / SRT / JSON / Markdown output
├── types.go             # Shared types: FileItem, Segment, Config
├── go.mod
├── wails.json
├── Makefile             # Build targets for whisper.cpp + Wails
├── frontend/            # Svelte frontend (Wails WebView2)
│   ├── src/
│   │   ├── App.svelte
│   │   └── lib/         # FileList, Controls, ProgressPanel
│   └── wailsjs/         # Auto-generated Go bindings for JS
├── build/               # Platform build configs (icon, manifests)
└── third_party/         # whisper.cpp (cloned at build time)
```

## How It Works

```
Video/Audio → FFmpeg → 16kHz mono WAV → whisper.cpp → segments → formatter → file
```

1. **Audio extraction** — FFmpeg converts any input to 16kHz 16-bit mono PCM WAV (whisper.cpp requirement)
2. **Transcription** — whisper.cpp processes WAV samples via Go bindings, reports progress per encoder step
3. **Output** — formatter writes the chosen format next to the source file (`video.mp4` → `video.txt`)

## Makefile Targets

```
make help                 Show all targets
make whisper-lib          Build whisper.cpp (Linux, CPU)
make whisper-lib-win      Build whisper.cpp (Windows, CPU)
make whisper-lib-win-vulkan Build whisper.cpp (Windows, Vulkan GPU)
make bindings             Regenerate Wails JS/TS bindings
make build-check          Verify Go compilation (Linux)
make build-win            Cross-compile Windows .exe (CPU)
make build-win-vulkan     Cross-compile Windows .exe (Vulkan)
make dev                  Run Wails dev server
make model                Download GGML model (~574 MB)
make ffmpeg-win           Download static ffmpeg.exe
make clean                Clean build artifacts
```

## License

MIT
