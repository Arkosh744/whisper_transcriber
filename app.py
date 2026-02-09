"""Main application window."""

from __future__ import annotations

import os
from tkinter import messagebox

import customtkinter as ctk

from config import APP_NAME, WINDOW_HEIGHT, WINDOW_WIDTH
from core.transcriber import TranscriptionEngine
from ui.controls import Controls
from ui.file_list import FileList
from ui.progress import ProgressPanel
from workers.transcribe_worker import TranscribeWorker


class TranscriberApp(ctk.CTk):
    def __init__(self):
        super().__init__()

        self.title(APP_NAME)
        self.geometry(f"{WINDOW_WIDTH}x{WINDOW_HEIGHT}")
        self.minsize(600, 450)
        self.resizable(True, True)

        ctk.set_appearance_mode("dark")
        ctk.set_default_color_theme("blue")

        self._engine = TranscriptionEngine()
        self._worker: TranscribeWorker | None = None
        self._last_output_folder: str | None = None

        self._build_ui()
        self.protocol("WM_DELETE_WINDOW", self._on_close)

    def _build_ui(self) -> None:
        ctk.CTkLabel(
            self, text=APP_NAME, font=ctk.CTkFont(size=20, weight="bold"),
        ).pack(pady=(12, 5))

        self._file_list = FileList(self, on_files_changed=self._on_files_changed)
        self._file_list.pack(fill="both", expand=True, padx=10, pady=(0, 5))

        ctk.CTkFrame(self, height=2, fg_color=("gray80", "gray25")).pack(fill="x", padx=15, pady=2)

        self._controls = Controls(self, on_start=self._on_start, on_cancel=self._on_cancel)
        self._controls.pack(fill="x", padx=5, pady=2)

        ctk.CTkFrame(self, height=2, fg_color=("gray80", "gray25")).pack(fill="x", padx=15, pady=2)

        self._progress = ProgressPanel(self)
        self._progress.pack(fill="x", padx=5, pady=(2, 8))

        self._controls._start_btn.configure(state="disabled")

    def _on_files_changed(self) -> None:
        has_files = self._file_list.file_count > 0
        self._controls._start_btn.configure(state="normal" if has_files else "disabled")

    def _on_start(self) -> None:
        files = self._file_list.files
        if not files:
            return
        self._progress.reset()
        for i in range(len(files)):
            self._file_list.set_status(i, "pending")
        self._controls.set_running(True)
        self._file_list.set_enabled(False)
        self._worker = TranscribeWorker(
            engine=self._engine, files=files,
            language=self._controls.selected_language,
            output_format=self._controls.selected_format,
            on_progress=self._on_progress, on_file_done=self._on_file_done,
            on_error=self._on_error, on_all_done=self._on_all_done,
            on_model_loading=self._on_model_loading,
            on_model_loaded=self._on_model_loaded, schedule=self.after,
        )
        self._worker.start()

    def _on_cancel(self) -> None:
        if self._worker and self._worker.is_running:
            self._worker.cancel()
            self._progress.set_status("Cancelling...")

    def _on_model_loading(self) -> None:
        self._progress.set_indeterminate(True)
        self._progress.set_status("Loading model... (first time takes ~15 seconds)")

    def _on_model_loaded(self) -> None:
        self._progress.set_indeterminate(False)
        self._progress.set_status("Model loaded. Starting transcription...")

    def _on_progress(self, file_idx, seg_end, total_dur, status) -> None:
        self._file_list.set_status(file_idx, "processing")
        total_files = self._file_list.file_count
        file_progress = seg_end / total_dur if total_dur > 0 else 0
        overall = (file_idx + file_progress) / total_files
        self._progress.set_progress(overall)
        self._progress.set_status(status)

    def _on_file_done(self, file_idx, output_path) -> None:
        self._file_list.set_status(file_idx, "done")
        self._last_output_folder = os.path.dirname(output_path)

    def _on_error(self, file_idx, error_msg) -> None:
        if file_idx >= 0:
            self._file_list.set_status(file_idx, "error")
        self._progress.set_status(f"Error: {error_msg}")

    def _on_all_done(self) -> None:
        self._controls.set_running(False)
        self._file_list.set_enabled(True)
        if self._last_output_folder:
            self._progress.show_complete(self._last_output_folder)
        else:
            self._progress.set_status("Finished (no files processed)")

    def _on_close(self) -> None:
        if self._worker and self._worker.is_running:
            if messagebox.askyesno("Transcription in progress", "Cancel and exit?"):
                self._worker.cancel()
                self.after(500, self.destroy)
        else:
            self.destroy()
