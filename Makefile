SHELL := /bin/bash

# Paths
WHISPER_DIR     := $(CURDIR)/third_party/whisper.cpp
WHISPER_BUILD   := $(WHISPER_DIR)/build
WHISPER_WIN     := $(WHISPER_DIR)/build-win
WHISPER_VULKAN  := $(WHISPER_DIR)/build-win-vulkan
WHISPER_COMMIT  := 764482c3175d

# CGo flags (Linux)
export CGO_CFLAGS  = -I$(WHISPER_DIR)/include -I$(WHISPER_DIR)/ggml/include
export CGO_LDFLAGS = -L$(WHISPER_BUILD)/src -L$(WHISPER_BUILD)/ggml/src

# Windows cross-compiler
WIN_CC  := x86_64-w64-mingw32-gcc
WIN_CXX := x86_64-w64-mingw32-g++

.PHONY: help whisper-lib whisper-lib-win whisper-lib-win-vulkan \
        bindings build-check build-win build-win-vulkan \
        model ffmpeg-win dev clean

help:
	@echo "Targets:"
	@echo "  whisper-lib           Build whisper.cpp (Linux, CPU)"
	@echo "  whisper-lib-win       Build whisper.cpp (Windows, CPU)"
	@echo "  whisper-lib-win-vulkan Build whisper.cpp (Windows, Vulkan GPU)"
	@echo "  bindings              Regenerate Wails JS/TS bindings"
	@echo "  build-check           Verify Go compilation (Linux)"
	@echo "  build-win             Cross-compile Windows .exe (CPU)"
	@echo "  build-win-vulkan      Cross-compile Windows .exe (Vulkan)"
	@echo "  dev                   Run Wails dev server"
	@echo "  model                 Download GGML model (~574 MB)"
	@echo "  ffmpeg-win            Download static ffmpeg.exe"
	@echo "  clean                 Clean build artifacts"

# --- whisper.cpp ---

$(WHISPER_DIR)/CMakeLists.txt:
	git clone https://github.com/ggml-org/whisper.cpp.git $(WHISPER_DIR)
	cd $(WHISPER_DIR) && git checkout $(WHISPER_COMMIT)

whisper-lib: $(WHISPER_DIR)/CMakeLists.txt
	@if [ ! -f $(WHISPER_BUILD)/src/libwhisper.a ]; then \
		echo "Building whisper.cpp (Linux, CPU)..."; \
		mkdir -p $(WHISPER_BUILD) && cd $(WHISPER_BUILD) && \
		cmake .. -DBUILD_SHARED_LIBS=OFF -DCMAKE_BUILD_TYPE=Release \
			-DGGML_CUDA=OFF -DGGML_VULKAN=OFF -DGGML_BLAS=OFF \
			-DWHISPER_BUILD_EXAMPLES=OFF -DWHISPER_BUILD_TESTS=OFF && \
		cmake --build . -j$$(nproc); \
	else \
		echo "whisper.cpp (Linux) already built"; \
	fi

whisper-lib-win: $(WHISPER_DIR)/CMakeLists.txt
	@if [ ! -f $(WHISPER_WIN)/src/libwhisper.a ]; then \
		echo "Building whisper.cpp (Windows, CPU)..."; \
		mkdir -p $(WHISPER_WIN) && cd $(WHISPER_WIN) && \
		cmake .. -DCMAKE_SYSTEM_NAME=Windows \
			-DCMAKE_C_COMPILER=$(WIN_CC) \
			-DCMAKE_CXX_COMPILER=$(WIN_CXX) \
			-DCMAKE_C_FLAGS="-D_WIN32_WINNT=0x0601 -DNDEBUG" \
			-DCMAKE_CXX_FLAGS="-D_WIN32_WINNT=0x0601 -DNDEBUG" \
			-DBUILD_SHARED_LIBS=OFF -DCMAKE_BUILD_TYPE=Release \
			-DGGML_CUDA=OFF -DGGML_VULKAN=OFF -DGGML_BLAS=OFF \
			-DGGML_OPENMP=OFF \
			-DWHISPER_BUILD_EXAMPLES=OFF -DWHISPER_BUILD_TESTS=OFF && \
		cmake --build . -j$$(nproc) && \
		cd ggml/src && for f in ggml*.a; do [ ! -f "lib$$f" ] && ln -s "$$f" "lib$$f" || true; done; \
	else \
		echo "whisper.cpp (Windows, CPU) already built"; \
	fi

VULKAN_WIN := $(CURDIR)/third_party/vulkan-win64

whisper-lib-win-vulkan: $(WHISPER_DIR)/CMakeLists.txt
	@if [ ! -f $(WHISPER_VULKAN)/src/libwhisper.a ]; then \
		echo "Building whisper.cpp (Windows, Vulkan)..."; \
		mkdir -p $(WHISPER_VULKAN) && cd $(WHISPER_VULKAN) && \
		cmake .. -DCMAKE_SYSTEM_NAME=Windows \
			-DCMAKE_C_COMPILER=$(WIN_CC) \
			-DCMAKE_CXX_COMPILER=$(WIN_CXX) \
			-DCMAKE_C_FLAGS="-D_WIN32_WINNT=0x0601 -DNDEBUG" \
			-DCMAKE_CXX_FLAGS="-D_WIN32_WINNT=0x0601 -DNDEBUG" \
			-DBUILD_SHARED_LIBS=OFF -DCMAKE_BUILD_TYPE=Release \
			-DGGML_VULKAN=ON -DGGML_OPENMP=OFF -DGGML_BLAS=OFF -DGGML_CUDA=OFF \
			-DVulkan_LIBRARY=$(VULKAN_WIN)/libvulkan-1.a \
			-DVulkan_INCLUDE_DIR=$(VULKAN_WIN)/include \
			-DVulkan_GLSLC_EXECUTABLE=$$(which glslc) \
			-DWHISPER_BUILD_EXAMPLES=OFF -DWHISPER_BUILD_TESTS=OFF && \
		cmake --build . -j$$(nproc) && \
		cd ggml/src && \
		for f in ggml*.a; do [ ! -f "lib$$f" ] && ln -s "$$f" "lib$$f" || true; done && \
		ln -sf ggml-vulkan/ggml-vulkan.a libggml-vulkan.a; \
	else \
		echo "whisper.cpp (Windows, Vulkan) already built"; \
	fi

# --- Go / Wails ---

bindings: whisper-lib
	wails generate module

build-check: whisper-lib
	@echo "Verifying Go compilation..."
	go build -v ./...
	@echo "OK"

dev: whisper-lib bindings
	wails dev

build-win: whisper-lib-win bindings frontend
	@echo "Cross-compiling Windows .exe (CPU)..."
	@mkdir -p build/bin
	CGO_CFLAGS="-I$(WHISPER_DIR)/include -I$(WHISPER_DIR)/ggml/include" \
	CGO_LDFLAGS="-L$(WHISPER_WIN)/src -L$(WHISPER_WIN)/ggml/src" \
	GOOS=windows GOARCH=amd64 \
	CC=$(WIN_CC) CXX=$(WIN_CXX) CGO_ENABLED=1 \
		go build -tags desktop,production -ldflags "-w -s -H windowsgui -extldflags '-static'" \
		-o build/bin/whisper_transcriber.exe .
	@echo "Built: build/bin/whisper_transcriber.exe ($$(du -h build/bin/whisper_transcriber.exe | cut -f1))"

build-win-vulkan: whisper-lib-win-vulkan bindings frontend
	@echo "Cross-compiling Windows .exe (Vulkan GPU)..."
	@mkdir -p build/bin
	CGO_CFLAGS="-I$(WHISPER_DIR)/include -I$(WHISPER_DIR)/ggml/include" \
	CGO_LDFLAGS="-L$(WHISPER_VULKAN)/src -L$(WHISPER_VULKAN)/ggml/src -L$(VULKAN_WIN) -lggml-vulkan -lvulkan-1 -lstdc++" \
	GOOS=windows GOARCH=amd64 \
	CC=$(WIN_CC) CXX=$(WIN_CXX) CGO_ENABLED=1 \
		go build -tags desktop,production -ldflags "-w -s -H windowsgui -extldflags '-static'" \
		-o build/bin/whisper_transcriber.exe .
	@echo "Built: build/bin/whisper_transcriber.exe Vulkan ($$(du -h build/bin/whisper_transcriber.exe | cut -f1))"

frontend:
	@cd frontend && npm install --silent && npm run build
	@echo "Frontend built"

# --- Assets ---

model:
	@mkdir -p models
	@if [ ! -f models/ggml-large-v3-turbo-q5_0.bin ]; then \
		echo "Downloading GGML model (~574 MB)..."; \
		curl -L --progress-bar -o models/ggml-large-v3-turbo-q5_0.bin \
			"https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3-turbo-q5_0.bin"; \
	else \
		echo "Model already exists"; \
	fi

ffmpeg-win:
	@mkdir -p build/bin
	@if [ ! -f build/bin/ffmpeg.exe ]; then \
		echo "Downloading FFmpeg static build (~200 MB)..."; \
		curl -L --retry 3 --retry-delay 5 -o /tmp/ffmpeg-win.zip \
			"https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip" && \
		python3 -c "import zipfile,shutil,os; z=zipfile.ZipFile('/tmp/ffmpeg-win.zip'); \
			f=[n for n in z.namelist() if n.endswith('bin/ffmpeg.exe')][0]; \
			z.extract(f,'/tmp/ffmpeg-ext'); \
			shutil.copy2('/tmp/ffmpeg-ext/'+f,'build/bin/ffmpeg.exe')" && \
		rm -rf /tmp/ffmpeg-win.zip /tmp/ffmpeg-ext; \
		echo "FFmpeg: build/bin/ffmpeg.exe ($$(du -h build/bin/ffmpeg.exe | cut -f1))"; \
	else \
		echo "ffmpeg.exe already exists"; \
	fi

# --- Cleanup ---

clean:
	rm -rf $(WHISPER_BUILD) $(WHISPER_WIN) $(WHISPER_VULKAN)
	rm -rf build/bin
