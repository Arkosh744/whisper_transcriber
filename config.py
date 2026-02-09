"""Application constants."""

import os
import sys


# --- Model ---

MODEL_NAME = "large-v3"
HUGGINGFACE_REPO = "Systran/faster-whisper-large-v3"
MODEL_EXPECTED_SIZE_BYTES = 3_087_284_224  # ~2.88 GB (model.bin + assets)


def _app_dir() -> str:
    """Directory of the app (next to main.py or next to .exe)."""
    if getattr(sys, "frozen", False):
        return os.path.dirname(sys.executable)
    return os.path.dirname(os.path.abspath(__file__))


MODELS_DIR = os.path.join(_app_dir(), "models")

# Legacy path for backward-compatibility
_LEGACY_MODEL_PATH = r"D:\Downloads\Собеседования\Видео собесов\model-large-v3"


def find_model_path() -> str | None:
    """Find the whisper model. Returns path or None if not found.

    Search order:
      1. models/<MODEL_NAME>/ next to the app (portable)
      2. Legacy hardcoded path (backward-compat)
    """
    # 1. Local models/ folder
    local = os.path.join(MODELS_DIR, MODEL_NAME)
    if os.path.isdir(local) and os.path.isfile(os.path.join(local, "model.bin")):
        return local
    # 2. Legacy path
    if os.path.isdir(_LEGACY_MODEL_PATH) and os.path.isfile(
        os.path.join(_LEGACY_MODEL_PATH, "model.bin")
    ):
        return _LEGACY_MODEL_PATH
    return None


def get_model_download_dir() -> str:
    """Target directory for model download: models/<MODEL_NAME>/"""
    return os.path.join(MODELS_DIR, MODEL_NAME)


# --- Input ---

SUPPORTED_EXTENSIONS = (".mp4", ".mkv", ".avi", ".mov")
FILE_TYPE_LABEL = "Video files"

# --- Output ---

OUTPUT_FORMATS = {
    "TXT": ".txt",
    "SRT": ".srt",
    "JSON": ".json",
    "Markdown": ".md",
}

# --- Languages ---
# "auto" = auto-detect (language=None passed to whisper)

LANGUAGES = {
    "auto": "Auto-detect",
    "ru": "Russian",
    "en": "English",
    "de": "German",
    "fr": "French",
    "es": "Spanish",
    "zh": "Chinese",
    "ja": "Japanese",
    "ko": "Korean",
    "uk": "Ukrainian",
    "pl": "Polish",
    "it": "Italian",
    "pt": "Portuguese",
    "tr": "Turkish",
    "ar": "Arabic",
    "hi": "Hindi",
}

# --- Transcription defaults ---

BEAM_SIZE = 5
VAD_FILTER = True
VAD_MIN_SILENCE_MS = 500

# --- App ---

APP_NAME = "Whisper Transcriber"
WINDOW_WIDTH = 750
WINDOW_HEIGHT = 550
