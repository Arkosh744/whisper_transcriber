"""Download whisper model from HuggingFace Hub with progress tracking."""

from __future__ import annotations

import os
import shutil
import threading
import time
from typing import Callable, Optional

import huggingface_hub

from config import (
    HUGGINGFACE_REPO,
    MODEL_EXPECTED_SIZE_BYTES,
    get_model_download_dir,
)


class ModelDownloader:
    """Downloads a faster-whisper model in a background thread.

    Progress is tracked by polling the total size of downloaded files
    in the output directory every second.
    """

    def __init__(
        self,
        on_progress: Callable[[float, str], None],  # (fraction, status_text)
        on_done: Callable[[str], None],              # (model_path)
        on_error: Callable[[str], None],             # (error_message)
        schedule: Callable,                          # tkinter root.after
    ):
        self._on_progress = on_progress
        self._on_done = on_done
        self._on_error = on_error
        self._schedule = schedule
        self._cancel = threading.Event()
        self._download_thread: Optional[threading.Thread] = None
        self._poll_thread: Optional[threading.Thread] = None
        self._output_dir = get_model_download_dir()

    def start(self) -> None:
        self._cancel.clear()
        os.makedirs(self._output_dir, exist_ok=True)

        self._download_thread = threading.Thread(
            target=self._download, daemon=True
        )
        self._poll_thread = threading.Thread(
            target=self._poll_progress, daemon=True
        )
        self._download_thread.start()
        self._poll_thread.start()

    def cancel(self) -> None:
        self._cancel.set()

    @property
    def is_running(self) -> bool:
        return (
            self._download_thread is not None
            and self._download_thread.is_alive()
        )

    def _download(self) -> None:
        """Run huggingface_hub.snapshot_download in background."""
        try:
            huggingface_hub.snapshot_download(
                repo_id=HUGGINGFACE_REPO,
                local_dir=self._output_dir,
                local_dir_use_symlinks=False,
                allow_patterns=[
                    "config.json",
                    "preprocessor_config.json",
                    "model.bin",
                    "tokenizer.json",
                    "vocabulary.*",
                ],
            )

            if self._cancel.is_set():
                self._cleanup()
                return

            self._schedule(0, self._on_done, self._output_dir)

        except Exception as e:
            if self._cancel.is_set():
                self._cleanup()
                return
            self._schedule(0, self._on_error, str(e))

    def _poll_progress(self) -> None:
        """Poll download directory size every second for progress updates."""
        while not self._cancel.is_set():
            if self._download_thread and not self._download_thread.is_alive():
                break

            downloaded = self._dir_size(self._output_dir)
            fraction = min(downloaded / MODEL_EXPECTED_SIZE_BYTES, 0.99)
            downloaded_gb = downloaded / (1024 ** 3)
            total_gb = MODEL_EXPECTED_SIZE_BYTES / (1024 ** 3)
            status = f"Downloading model... {downloaded_gb:.1f} / {total_gb:.1f} GB"

            self._schedule(0, self._on_progress, fraction, status)
            time.sleep(1.0)

    def _cleanup(self) -> None:
        """Remove partially downloaded files on cancel."""
        try:
            if os.path.isdir(self._output_dir):
                shutil.rmtree(self._output_dir, ignore_errors=True)
        except Exception:
            pass

    @staticmethod
    def _dir_size(path: str) -> int:
        """Total size of all files in directory (recursive)."""
        total = 0
        try:
            for dirpath, _dirnames, filenames in os.walk(path):
                for f in filenames:
                    fp = os.path.join(dirpath, f)
                    try:
                        total += os.path.getsize(fp)
                    except OSError:
                        pass
        except OSError:
            pass
        return total
