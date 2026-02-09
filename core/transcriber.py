"""Wrapper around faster-whisper WhisperModel with cancel support."""

from __future__ import annotations

from typing import Callable, Optional

from faster_whisper import WhisperModel

from config import BEAM_SIZE, VAD_FILTER, VAD_MIN_SILENCE_MS


class CancelledException(Exception):
    pass


class TranscriptionEngine:
    """Loads the whisper model and transcribes files with progress callbacks."""

    def __init__(self, model_path: Optional[str] = None):
        self.model_path = model_path
        self.model: Optional[WhisperModel] = None

    def set_model_path(self, path: str) -> None:
        """Update model path (e.g. after download). Resets loaded model."""
        self.model_path = path
        self.model = None

    def load_model(self) -> None:
        """Load model into GPU. Slow (~10-20s). Call from background thread."""
        if not self.model_path:
            raise RuntimeError("Model path not set.")
        self.model = WhisperModel(
            self.model_path,
            device="cuda",
            compute_type="float16",
        )

    @property
    def is_loaded(self) -> bool:
        return self.model is not None

    def transcribe(
        self,
        file_path: str,
        language: Optional[str],
        on_segment: Callable,
        cancel_flag: Callable[[], bool],
    ) -> tuple[list, float]:
        """
        Transcribe a file.

        Args:
            file_path: Path to video/audio file.
            language: Language code or None for auto-detect.
            on_segment: Callback(segment, total_duration) called per segment.
            cancel_flag: Callable returning True if cancellation requested.

        Returns:
            (list of segments, total duration in seconds)

        Raises:
            CancelledException: If cancel_flag returns True during transcription.
        """
        if not self.is_loaded:
            raise RuntimeError("Model not loaded. Call load_model() first.")

        segments_gen, info = self.model.transcribe(
            file_path,
            language=language,
            beam_size=BEAM_SIZE,
            vad_filter=VAD_FILTER,
            vad_parameters=dict(min_silence_duration_ms=VAD_MIN_SILENCE_MS),
        )

        collected = []
        for seg in segments_gen:
            if cancel_flag():
                raise CancelledException()
            on_segment(seg, info.duration)
            collected.append(seg)

        return collected, info.duration
