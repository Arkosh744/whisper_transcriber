"""Application constants."""

import os
import sys

# --- Model ---

DEFAULT_MODEL_PATH = os.path.join(
    os.path.dirname(os.path.abspath(__file__)), "models", "large-v3"
)

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
