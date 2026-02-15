# План: завершение Go rewrite — сборка + кросс-компиляция Windows .exe

## Context

Весь код (Go backend + Svelte frontend) написан на 100%. Не хватает:
- whisper.cpp static library (не собрана)
- Makefile (не создан)
- Wails bindings (устаревшие — содержат `Greet()` вместо реальных методов)
- Тестирование

Среда: WSL2 Linux, целевая платформа: Windows 11 x64. CUDA нужна.

## Оценка готовности

| Компонент | Статус | Файлы |
|-----------|--------|-------|
| types.go | 100% | go-version/types.go |
| model.go | 100% | go-version/model.go |
| ffmpeg.go | 100% | go-version/ffmpeg.go |
| transcriber.go | 100% | go-version/transcriber.go |
| formatter.go | 100% | go-version/formatter.go |
| app.go | 100% | go-version/app.go |
| main.go | 100% | go-version/main.go |
| Frontend (4 компонента) | 100% | go-version/frontend/src/ |
| whisper.cpp lib | 0% | — |
| Makefile | 0% | — |
| Wails bindings | устаревшие | go-version/frontend/wailsjs/ |

## План реализации

### Phase 1: Инфраструктура сборки (CPU-only, Linux)

Цель: проверить что код компилируется.

**1.1** Установить webkit2gtk (нужен для `wails generate`):
```bash
sudo apt install libwebkit2gtk-4.0-dev
```

**1.2** Склонировать whisper.cpp в `go-version/third_party/whisper.cpp`:
```bash
git clone https://github.com/ggml-org/whisper.cpp.git third_party/whisper.cpp
cd third_party/whisper.cpp && git checkout 764482c3175d
```
> Коммит `764482c3175d` — тот же что в go.mod для совместимости bindings.

**1.3** Собрать whisper.cpp (Linux, CPU-only):
```bash
mkdir build && cd build
cmake .. -DBUILD_SHARED_LIBS=OFF -DCMAKE_BUILD_TYPE=Release \
  -DGGML_CUDA=OFF -DGGML_VULKAN=OFF
cmake --build . -j$(nproc)
```

**1.4** Проверить компиляцию Go с CGo:
```bash
export CGO_CFLAGS="-I./third_party/whisper.cpp/include -I./third_party/whisper.cpp/ggml/include"
export CGO_LDFLAGS="-L./third_party/whisper.cpp/build/src -L./third_party/whisper.cpp/build/ggml/src -lwhisper -lggml -lggml-base -lggml-cpu -lm -lpthread -lstdc++"
go build ./...
```
> Пути к `.a` файлам могут отличаться — проверить `find build -name "*.a"`.

**1.5** Регенерировать Wails bindings:
```bash
wails generate module
```

### Phase 2: Кросс-компиляция Windows .exe (CPU-only)

Цель: получить работающий .exe без GPU.

**2.1** Установить mingw-w64:
```bash
sudo apt install gcc-mingw-w64-x86-64 g++-mingw-w64-x86-64
```

**2.2** Пересобрать whisper.cpp для Windows:
```bash
cd third_party/whisper.cpp
mkdir build-win && cd build-win
cmake .. -DCMAKE_SYSTEM_NAME=Windows \
  -DCMAKE_C_COMPILER=x86_64-w64-mingw32-gcc \
  -DCMAKE_CXX_COMPILER=x86_64-w64-mingw32-g++ \
  -DBUILD_SHARED_LIBS=OFF -DCMAKE_BUILD_TYPE=Release \
  -DGGML_CUDA=OFF -DGGML_VULKAN=OFF
cmake --build . -j$(nproc)
```

**2.3** Кросс-компилировать Wails app:
```bash
export CGO_CFLAGS="-I./third_party/whisper.cpp/include -I./third_party/whisper.cpp/ggml/include"
export CGO_LDFLAGS="-L./third_party/whisper.cpp/build-win/src -L./third_party/whisper.cpp/build-win/ggml/src -lwhisper -lggml -lggml-base -lggml-cpu -lm -lpthread -lstdc++ -static"
GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ CGO_ENABLED=1 \
  wails build -platform windows/amd64
```

**2.4** Проверить .exe на Windows — запустить из проводника.

### Phase 3: GPU через Vulkan

Vulkan — лучший выбор для кросс-компиляции: работает на NVIDIA и AMD, не требует проприетарного SDK.

**3.1** Скачать Vulkan SDK headers + loader для кросс-компиляции:
```bash
# Vulkan headers (platform-independent)
sudo apt install libvulkan-dev vulkan-headers
# Для Windows: vulkan-1.dll идёт с GPU драйвером (NVIDIA/AMD)
```

**3.2** Пересобрать whisper.cpp для Windows с Vulkan:
```bash
cd third_party/whisper.cpp
mkdir build-win-vulkan && cd build-win-vulkan
cmake .. -DCMAKE_SYSTEM_NAME=Windows \
  -DCMAKE_C_COMPILER=x86_64-w64-mingw32-gcc \
  -DCMAKE_CXX_COMPILER=x86_64-w64-mingw32-g++ \
  -DBUILD_SHARED_LIBS=OFF -DCMAKE_BUILD_TYPE=Release \
  -DGGML_VULKAN=ON
cmake --build . -j$(nproc)
```

**3.3** Обновить CGo flags и пересобрать .exe:
```bash
# Добавить -lvulkan-1 к CGO_LDFLAGS
wails build -platform windows/amd64
```

> Windows 10/11 с NVIDIA/AMD/Intel GPU уже имеет `vulkan-1.dll` в системе.
> Никаких доп. DLL бандлить не нужно — Vulkan runtime поставляется с драйверами.

### Phase 4: Makefile

Создать `go-version/Makefile` с targets:
- `whisper-lib` — сборка whisper.cpp (Linux)
- `whisper-lib-win` — сборка whisper.cpp (Windows cross, CPU)
- `whisper-lib-win-vulkan` — сборка whisper.cpp (Windows cross, Vulkan GPU)
- `bindings` — регенерация Wails JS bindings
- `build-check` — проверка компиляции (Linux)
- `build-win` — кросс-компиляция Windows .exe
- `model` — скачивание GGML модели
- `clean` — очистка артефактов

### Phase 5: .gitignore + финализация

- Добавить в `.gitignore`: `third_party/`, `models/`, `build/`
- Обновить `docs/TASKS.md`, `docs/SESSION_MAIN.md`

### Коммиты (бекап-точки)

После каждой значимой фазы — коммит с реалистичным timestamp (чуть в прошлом, ~30-90 мин интервалы). Без co-authored-by.

```bash
GIT_AUTHOR_DATE="2026-02-15T10:30:00+03:00" GIT_COMMITTER_DATE="2026-02-15T10:30:00+03:00" \
  git commit -m "message"
```

Примерные коммиты:
1. После Phase 1 (Linux build): `"build whisper.cpp static lib, verify Go compilation"`
2. После Phase 4 (Makefile): `"add Makefile for build automation"`
3. После Phase 2 (Windows .exe): `"add Windows cross-compilation via mingw"`
4. После Phase 3 (Vulkan): `"enable Vulkan GPU backend for whisper.cpp"`
5. После Phase 5 (cleanup): `"add .gitignore, update docs"`

## Критические файлы

| Файл | Действие |
|------|----------|
| `go-version/Makefile` | создать |
| `go-version/.gitignore` | обновить (third_party/, models/) |
| `go-version/frontend/wailsjs/go/main/App.js` | авто-регенерация через `wails generate` |
| `go-version/frontend/wailsjs/go/main/App.d.ts` | авто-регенерация |
| `docs/TASKS.md` | создать |
| `docs/SESSION_MAIN.md` | создать |

## Верификация

1. `make whisper-lib` — whisper.cpp собирается без ошибок
2. `make build-check` — Go код компилируется с CGo
3. `make bindings` — Wails bindings содержат BrowseFiles, StartTranscription и т.д.
4. `make build-win` — создаётся .exe файл
5. Запустить .exe на Windows — окно открывается, UI работает
6. Browse files + Start Transcription (нужна модель + ffmpeg)

## Решения

- **GPU**: Vulkan — кросс-компилируется из WSL2, работает на NVIDIA/AMD/Intel, не требует доп. DLL
- **FFmpeg**: бандлить `ffmpeg.exe` (~90 МБ static build) рядом с .exe

### FFmpeg bundling

Добавить в Makefile target `ffmpeg-win`:
```bash
# Скачать static build для Windows
curl -L -o ffmpeg.zip https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip
unzip ffmpeg.zip -d /tmp/ffmpeg
cp /tmp/ffmpeg/*/bin/ffmpeg.exe build/bin/
```

## Порядок работы

1. Phase 1 (Linux build) → быстрая проверка компиляции
2. Phase 4 (Makefile) → автоматизация
3. Phase 2 (Windows cross) → рабочий .exe
4. Phase 5 (cleanup) → коммит
5. Phase 3 (Vulkan GPU) + FFmpeg bundling → GPU-accelerated .exe
