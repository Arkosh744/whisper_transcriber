# Whisper Transcriber

Desktop GUI app for transcribing video files using [faster-whisper](https://github.com/SYSTRAN/faster-whisper) on GPU.

Drop video files → get timestamped transcripts. That's it.

![Python](https://img.shields.io/badge/python-3.12-blue)
![CUDA](https://img.shields.io/badge/CUDA-GPU%20accelerated-green)
![License](https://img.shields.io/badge/license-MIT-gray)

## Features

- **GPU-accelerated** — faster-whisper with CUDA FP16, ~10-50x realtime
- **Direct video input** — MP4, MKV, AVI, MOV (no manual audio extraction needed)
- **Multiple output formats** — TXT, SRT, JSON, Markdown
- **16 languages** — auto-detect or manual selection
- **Model auto-download** — downloads large-v3 (~3 GB) from HuggingFace if missing
- **Dark theme** — CustomTkinter, Win11-friendly
- **Batch processing** — queue multiple files, per-file progress
- **Cancel anytime** — graceful cancellation mid-transcription

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
│  [2/3] video_03.mov [12:34 / 18:20]         │
└─────────────────────────────────────────────┘
```

## Quick Start

```bash
# clone
git clone https://github.com/arkosh/whisper_transcriber.git
cd whisper_transcriber

# setup (creates venv, installs deps)
make install

# run
make run
```

On first launch, if the model isn't found locally, a download banner appears — click "Download Model" and wait ~5 min.

## Requirements

- Python 3.10+
- NVIDIA GPU with CUDA support
- ~4 GB disk for the model
- Linux / Windows (WSL works too)

## Makefile

```
make help             Show all targets
make venv             Create virtual environment
make install          Install deps (auto-creates venv)
make install-dev      + ruff, mypy
make run              Launch GUI
make lint             Run ruff
make format           Auto-format with ruff
make typecheck        Run mypy
make build            Build .exe with PyInstaller
make download-model   Download large-v3 via CLI
make clean            Remove __pycache__, build/
make clean-all        + remove venv/ and models/
make tree             Show project structure
```

## Project Structure

```
whisper_transcriber/
├── main.py                  # entry point, CUDA setup
├── app.py                   # main window, download banner
├── config.py                # constants, model search logic
├── requirements.txt
├── build.spec               # PyInstaller config
├── Makefile
├── core/
│   ├── transcriber.py       # WhisperModel wrapper
│   ├── model_downloader.py  # HuggingFace download + progress
│   ├── formatters.py        # TXT / SRT / JSON / Markdown
│   └── media_info.py        # duration via PyAV
├── workers/
│   └── transcribe_worker.py # background thread
└── ui/
    ├── file_list.py         # scrollable file list
    ├── controls.py          # language, format, buttons
    └── progress.py          # progress bar + status
```

## How It Works

1. **Video → Audio** — faster-whisper uses PyAV internally to read audio from video containers. No FFmpeg CLI needed.
2. **Transcription** — Whisper large-v3 model runs on GPU (CUDA FP16). VAD filter skips silence.
3. **Progress** — each segment reports `seg.end / total_duration`. GUI updates via `root.after()` from worker thread.
4. **Output** — formatter writes the chosen format next to the source file (`video.mp4` → `video.txt`).

## Model

Uses `Systran/faster-whisper-large-v3` (~2.9 GB).

Search order:
1. `./models/large-v3/` next to the app (portable)
2. Legacy hardcoded path (backward-compat)
3. Not found → download banner in GUI

CLI download: `make download-model`

## Building .exe

```bash
make build
```

Output: `dist/WhisperTranscriber/` (~700 MB with CUDA DLLs). Model is **not** bundled — it lives in `models/` next to the exe.

## Roadmap

### v0.2 — UX Polish
- [ ] Drag & drop files onto the window
- [ ] Remember last used language/format between sessions
- [ ] System tray notification when batch is done

### v0.3 — Audio Support
- [ ] Direct audio input: MP3, WAV, FLAC, OGG, M4A
- [ ] Audio waveform preview in file list

### v0.4 — Output Enhancements
- [ ] Custom output directory (not just next to source)
- [ ] Word-level timestamps in SRT
- [ ] Speaker diarization (who said what)
- [ ] Auto-punctuation and paragraph splitting

### v0.5 — Model Management
- [ ] Multiple model sizes (tiny, small, medium, large-v3)
- [ ] Model size/speed comparison in UI
- [ ] CPU fallback when no GPU available
- [ ] Quantized INT8 models for lower VRAM

### v0.6 — Advanced Features
- [ ] Translation mode (any language → Russian)
- [ ] Live preview: show segments as they're transcribed
- [ ] Search across all transcripts
- [ ] Export combined transcript for multi-file batches

### v1.0
- [ ] Installer (NSIS / MSI)
- [ ] CI/CD: GitHub Actions for builds and releases

## License

MIT
