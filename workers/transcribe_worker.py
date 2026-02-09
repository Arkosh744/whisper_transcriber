"""Background thread for transcription with progress callbacks."""

from __future__ import annotations

import os
import threading
from typing import Callable, Optional

from core.formatters import FORMATTERS
from core.transcriber import CancelledException, TranscriptionEngine


class TranscribeWorker:
    """Runs transcription in a daemon thread, reporting progress to the GUI."""

    def __init__(
        self,
        engine: TranscriptionEngine,
        files: list[str],
        language: Optional[str],
        output_format: str,
        on_progress: Callable,     # (file_idx, seg_end, total_dur, status_text)
        on_file_done: Callable,    # (file_idx, output_path)
        on_error: Callable,        # (file_idx, error_str)
        on_all_done: Callable,     # ()
        on_model_loading: Callable,  # ()
        on_model_loaded: Callable,   # ()
        schedule: Callable,        # tkinter root.after
    ):
        self._engine = engine
        self._files = files
        self._language = language
        self._format = output_format
        self._on_progress = on_progress
        self._on_file_done = on_file_done
        self._on_error = on_error
        self._on_all_done = on_all_done
        self._on_model_loading = on_model_loading
        self._on_model_loaded = on_model_loaded
        self._schedule = schedule
        self._cancel = threading.Event()
        self._thread: Optional[threading.Thread] = None

    def start(self) -> None:
        self._cancel.clear()
        self._thread = threading.Thread(target=self._run, daemon=True)
        self._thread.start()

    def cancel(self) -> None:
        self._cancel.set()

    @property
    def is_running(self) -> bool:
        return self._thread is not None and self._thread.is_alive()

    def _run(self) -> None:
        try:
            # Load model if needed
            if not self._engine.is_loaded:
                self._schedule(0, self._on_model_loading)
                try:
                    self._engine.load_model()
                except Exception as e:
                    self._schedule(0, self._on_error, -1, f"Model loading failed: {e}")
                    return
                self._schedule(0, self._on_model_loaded)

            if self._cancel.is_set():
                return

            formatter_fn, ext = FORMATTERS[self._format]

            for idx, file_path in enumerate(self._files):
                if self._cancel.is_set():
                    break
                try:
                    self._process_file(idx, file_path, formatter_fn, ext)
                except CancelledException:
                    break
                except Exception as e:
                    self._schedule(0, self._on_error, idx, str(e))
        finally:
            self._schedule(0, self._on_all_done)

    def _process_file(
        self, idx: int, file_path: str, formatter_fn: Callable, ext: str
    ) -> None:
        base = os.path.splitext(file_path)[0]
        output_path = base + ext

        def on_segment(seg, total_duration):
            status = (
                f"[{idx + 1}/{len(self._files)}] "
                f"{os.path.basename(file_path)}  "
                f"[{_fmt_time(seg.end)} / {_fmt_time(total_duration)}]"
            )
            self._schedule(
                0, self._on_progress, idx, seg.end, total_duration, status
            )

        segments, duration = self._engine.transcribe(
            file_path,
            language=self._language,
            on_segment=on_segment,
            cancel_flag=self._cancel.is_set,
        )

        formatter_fn(segments, output_path)
        self._schedule(0, self._on_file_done, idx, output_path)


def _fmt_time(seconds: float) -> str:
    m = int(seconds // 60)
    s = int(seconds % 60)
    return f"{m:02d}:{s:02d}"
