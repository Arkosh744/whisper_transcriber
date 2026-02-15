# Research Notes — Lightweight Transcription App

## Варианты архитектуры (исследовано)

### 1. Go + Wails + whisper.cpp ✅ ВЫБРАНО
- **Бинарник:** ~8-15 МБ
- **whisper.cpp Go bindings:** `github.com/ggml-org/whisper.cpp/bindings/go` (официальные, обновлены Feb 2026)
- **Wails v2:** Go backend + HTML/CSS/JS frontend, использует системный WebView2 (~8 МБ)
- **GPU:** CUDA, Vulkan, Metal — всё поддерживается whisper.cpp
- **Модели:** GGML формат, квантизация Q5_0 уменьшает large-v3 с 2.9 ГБ до ~1 ГБ
- **Видео:** whisper.cpp не читает видео напрямую → нужен FFmpeg subprocess

### 2. Tauri + whisper-rs (Rust)
- **Бинарник:** ~10-15 МБ
- **Плюсы:** самый лёгкий вариант, whisper-rs активно поддерживается
- **Минусы:** нужен Rust, кривая обучения
- **Пример:** Handy (2026) — < 10 МБ, Tauri + whisper.cpp

### 3. Python + UPX сжатие
- **Бинарник:** 700→~300 МБ (UPX 50-70% compression)
- **Плюсы:** минимум работы, всё уже написано
- **Минусы:** всё равно жирно, CUDA DLLs не сжимаются хорошо
- **Вывод:** bulk unavoidable — ~600 МБ это CUDA + CTranslate2, не зависит от бандлера

### 4. whisper.cpp CLI обёртка
- **Бинарник:** ~5 МБ
- **Минусы:** GUI примитивный (Win32 API / FLTK)

## Модели GGML — размеры и качество

| Модель | Оригинал | Q5_0 | Сжатие |
|--------|----------|------|--------|
| tiny | 75 МБ | 31 МБ | 59% |
| base | 142 МБ | 57 МБ | 60% |
| small | 466 МБ | 182 МБ | 61% |
| medium | 1.5 ГБ | 515 МБ | 66% |
| large-v3 | 2.9 ГБ | 1.08 ГБ | 63% |
| **large-v3-turbo** | — | **574 МБ** | — |

**large-v3-turbo:** 6x быстрее large-v3, 809M параметров, 1-2% потеря точности.
Q5_0 квантизация — минимальная потеря качества, рекомендуется.

## Python bundlers сравнение

| Инструмент | Размер | CUDA | Startup |
|-----------|--------|------|---------|
| PyInstaller | baseline | ✅ | стандарт |
| Nuitka | ~такой же | ✅ (не тестировано) | быстрее |
| cx_Freeze | больше | ✅ | быстрее |
| PyOxidizer | чуть меньше | ✅ | самый быстрый |

**Вывод:** смена бандлера даёт <10% разницы. Bulk = CUDA DLLs.

## Go GUI фреймворки

| Фреймворк | Бинарник | Подход | Зрелость |
|-----------|----------|--------|----------|
| **Wails** | ~8 МБ | Go + WebView2 | активно развивается |
| Fyne | ~15-30 МБ | нативный Go | самый зрелый |
| Gio | минимальный | immediate mode | растущий |

## Ссылки

- [whisper.cpp](https://github.com/ggml-org/whisper.cpp)
- [whisper.cpp Go bindings](https://pkg.go.dev/github.com/ggml-org/whisper.cpp/bindings/go)
- [Wails](https://wails.io/)
- [whisper-rs](https://github.com/tazz4843/whisper-rs)
- [GGML models](https://huggingface.co/ggerganov/whisper.cpp)
- [Handy (Tauri + whisper)](https://github.com/cjpais/Handy)
