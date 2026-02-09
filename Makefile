# ─────────────────────────────────────────────────────────────
#  Whisper Transcriber — Makefile
# ─────────────────────────────────────────────────────────────

PYTHON     ?= python3
PIP        ?= pip
VENV_DIR   ?= venv
APP_ENTRY  ?= main.py
SPEC_FILE  ?= build.spec

# Activate venv — every command runs through this automatically
ifeq ($(OS),Windows_NT)
    ACTIVATE = $(VENV_DIR)\Scripts\activate &&
else
    ACTIVATE = . $(VENV_DIR)/bin/activate &&
endif

# Use SHELL=bash so `source` / `.` works correctly
SHELL := /bin/bash

.PHONY: help venv install install-dev run lint format typecheck \
        clean clean-all build download-model tree

# ─── Default target ──────────────────────────────────────────

help: ## Show this help
	@echo ""
	@echo "  Whisper Transcriber"
	@echo "  ───────────────────────────────────────"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
	@echo ""

# ─── Environment ─────────────────────────────────────────────

venv: ## Create Python virtual environment
	$(PYTHON) -m venv $(VENV_DIR)
	$(ACTIVATE) pip install --upgrade pip
	@echo ""
	@echo "  venv created & pip upgraded."
	@echo "  Manual activate:  source $(VENV_DIR)/bin/activate"
	@echo ""

install: venv ## Install production dependencies (auto-activates venv)
	$(ACTIVATE) pip install -r requirements.txt

install-dev: venv ## Install dev dependencies (auto-activates venv)
	$(ACTIVATE) pip install -r requirements.txt
	$(ACTIVATE) pip install ruff mypy

# ─── Run ─────────────────────────────────────────────────────

run: ## Launch the GUI application (auto-activates venv)
	$(ACTIVATE) $(PYTHON) $(APP_ENTRY)

# ─── Code quality ────────────────────────────────────────────

lint: ## Run ruff linter (auto-activates venv)
	$(ACTIVATE) ruff check .

format: ## Auto-format code with ruff (auto-activates venv)
	$(ACTIVATE) ruff format .

typecheck: ## Run mypy type checker (auto-activates venv)
	$(ACTIVATE) mypy --ignore-missing-imports .

# ─── Build ───────────────────────────────────────────────────

build: ## Build standalone .exe with PyInstaller (auto-activates venv)
	$(ACTIVATE) pyinstaller $(SPEC_FILE) --noconfirm
	@echo ""
	@echo "  Build output: dist/WhisperTranscriber/"
	@echo ""

# ─── Model ───────────────────────────────────────────────────

download-model: ## Download whisper large-v3 model to models/ folder
	$(ACTIVATE) $(PYTHON) -c "\
	import os; \
	from config import get_model_download_dir, HUGGINGFACE_REPO; \
	import huggingface_hub; \
	d = get_model_download_dir(); \
	os.makedirs(d, exist_ok=True); \
	print(f'Downloading to {d}...'); \
	huggingface_hub.snapshot_download( \
	    repo_id=HUGGINGFACE_REPO, \
	    local_dir=d, \
	    local_dir_use_symlinks=False, \
	    allow_patterns=['config.json','preprocessor_config.json','model.bin','tokenizer.json','vocabulary.*'] \
	); \
	print('Done!')"

# ─── Cleanup ─────────────────────────────────────────────────

clean: ## Remove __pycache__, .pyc, build artifacts
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete 2>/dev/null || true
	rm -rf build/ dist/ *.egg-info/

clean-all: clean ## Clean everything including venv and models
	rm -rf $(VENV_DIR)/
	rm -rf models/

# ─── Info ────────────────────────────────────────────────────

tree: ## Show project structure
	@echo ""
	@echo "  whisper_transcriber/"
	@find . -not -path './.git/*' \
		-not -path './.git' \
		-not -path './venv/*' \
		-not -path './venv' \
		-not -path './models/*' \
		-not -path './models' \
		-not -path './__pycache__/*' \
		-not -path './*/__pycache__/*' \
		-not -path './*/*/__pycache__/*' \
		-not -name '*.pyc' \
		-not -path './dist/*' \
		-not -path './build/*' \
		| sort \
		| sed 's|^./||' \
		| sed 's|[^/]*/|  │  |g' \
		| sed '1d'
	@echo ""
