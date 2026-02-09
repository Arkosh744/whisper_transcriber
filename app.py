"""Main application window."""

from __future__ import annotations

import os
from tkinter import messagebox

import customtkinter as ctk

from config import (
    APP_NAME,
    MODEL_EXPECTED_SIZE_BYTES,
    MODEL_NAME,
    WINDOW_HEIGHT,
    WINDOW_WIDTH,
    find_model_path,
)
from core.model_downloader import ModelDownloader
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
        self._downloader: ModelDownloader | None = None
        self._last_output_folder: str | None = None

        self._build_ui()
        self._check_model()

        self.protocol("WM_DELETE_WINDOW", self._on_close)

    # ── UI build ──────────────────────────────────────────────────

    def _build_ui(self) -> None:
        # Title
        self._title_label = ctk.CTkLabel(
            self,
            text=APP_NAME,
            font=ctk.CTkFont(size=20, weight="bold"),
        )
        self._title_label.pack(pady=(12, 5))

        # ── Download banner (hidden by default) ──
        self._banner = ctk.CTkFrame(self, fg_color=("gray85", "gray20"))

        self._banner_icon = ctk.CTkLabel(
            self._banner, text="\u26a0", font=ctk.CTkFont(size=22),
        )
        self._banner_icon.pack(pady=(10, 2))

        self._banner_label = ctk.CTkLabel(
            self._banner,
            text="Model not found. Download to start transcribing.",
            font=ctk.CTkFont(size=13),
        )
        self._banner_label.pack(pady=2)

        total_gb = MODEL_EXPECTED_SIZE_BYTES / (1024 ** 3)
        self._download_btn = ctk.CTkButton(
            self._banner,
            text=f"\u2b07  Download Model ({MODEL_NAME}, ~{total_gb:.1f} GB)",
            width=300, height=38,
            font=ctk.CTkFont(size=14, weight="bold"),
            command=self._start_download,
        )
        self._download_btn.pack(pady=(5, 10))

        self._dl_bar = ctk.CTkProgressBar(self._banner, height=14, corner_radius=5)
        self._dl_status = ctk.CTkLabel(
            self._banner, text="", font=ctk.CTkFont(size=12),
            text_color=("gray40", "gray60"),
        )
        self._cancel_dl_btn = ctk.CTkButton(
            self._banner, text="Cancel Download", width=150, height=30,
            fg_color="transparent", border_width=1,
            hover_color=("red3", "red4"),
            command=self._cancel_download,
        )
        # dl_bar, dl_status, cancel_dl_btn are packed dynamically when download starts

        # ── File list ──
        self._file_list = FileList(
            self, on_files_changed=self._on_files_changed
        )
        self._file_list.pack(fill="both", expand=True, padx=10, pady=(0, 5))

        # Separator
        ctk.CTkFrame(self, height=2, fg_color=("gray80", "gray25")).pack(
            fill="x", padx=15, pady=2
        )

        # Controls
        self._controls = Controls(
            self, on_start=self._on_start, on_cancel=self._on_cancel
        )
        self._controls.pack(fill="x", padx=5, pady=2)

        # Separator
        ctk.CTkFrame(self, height=2, fg_color=("gray80", "gray25")).pack(
            fill="x", padx=15, pady=2
        )

        # Progress
        self._progress = ProgressPanel(self)
        self._progress.pack(fill="x", padx=5, pady=(2, 8))

        # Initially disable start
        self._controls._start_btn.configure(state="disabled")

    # ── Model management ──────────────────────────────────────────

    def _check_model(self) -> None:
        model_path = find_model_path()
        if model_path:
            self._engine.set_model_path(model_path)
            self._banner.pack_forget()
            self._progress.set_status(f"Model: {os.path.basename(model_path)}")
            self._on_files_changed()
        else:
            # Show banner right after title
            self._banner.pack(
                fill="x", padx=10, pady=(0, 8),
                after=self._title_label,
            )
            self._progress.set_status("Model not found")
            self._controls._start_btn.configure(state="disabled")

    def _start_download(self) -> None:
        # Switch banner to download mode
        self._download_btn.pack_forget()

        self._dl_bar.set(0)
        self._dl_bar.pack(fill="x", padx=20, pady=(5, 2))
        self._dl_status.configure(text="Starting download...")
        self._dl_status.pack(pady=2)
        self._cancel_dl_btn.pack(pady=(2, 10))

        self._downloader = ModelDownloader(
            on_progress=self._on_dl_progress,
            on_done=self._on_dl_done,
            on_error=self._on_dl_error,
            schedule=self.after,
        )
        self._downloader.start()

    def _cancel_download(self) -> None:
        if self._downloader and self._downloader.is_running:
            self._downloader.cancel()
            self._dl_status.configure(text="Cancelling...")
            self.after(1500, self._reset_download_banner)

    def _reset_download_banner(self) -> None:
        self._dl_bar.pack_forget()
        self._dl_status.pack_forget()
        self._cancel_dl_btn.pack_forget()
        self._download_btn.pack(pady=(5, 10))
        self._banner_label.configure(
            text="Model not found. Download to start transcribing."
        )

    def _on_dl_progress(self, fraction: float, status: str) -> None:
        self._dl_bar.set(fraction)
        self._dl_status.configure(text=status)

    def _on_dl_done(self, model_path: str) -> None:
        self._dl_bar.set(1.0)
        self._dl_status.configure(text="Download complete!")
        self.after(1000, self._check_model)

    def _on_dl_error(self, error_msg: str) -> None:
        self._dl_status.configure(text=f"Error: {error_msg}")
        self._cancel_dl_btn.pack_forget()
        self.after(3000, self._reset_download_banner)

    # ── File management ───────────────────────────────────────────

    def _on_files_changed(self) -> None:
        has_files = self._file_list.file_count > 0
        has_model = find_model_path() is not None
        if has_files and has_model:
            self._controls._start_btn.configure(state="normal")
        else:
            self._controls._start_btn.configure(state="disabled")

    # ── Transcription ─────────────────────────────────────────────

    def _on_start(self) -> None:
        files = self._file_list.files
        if not files:
            return

        model_path = find_model_path()
        if not model_path:
            return
        self._engine.set_model_path(model_path)

        self._progress.reset()
        for i in range(len(files)):
            self._file_list.set_status(i, "pending")

        self._controls.set_running(True)
        self._file_list.set_enabled(False)

        self._worker = TranscribeWorker(
            engine=self._engine,
            files=files,
            language=self._controls.selected_language,
            output_format=self._controls.selected_format,
            on_progress=self._on_progress,
            on_file_done=self._on_file_done,
            on_error=self._on_error,
            on_all_done=self._on_all_done,
            on_model_loading=self._on_model_loading,
            on_model_loaded=self._on_model_loaded,
            schedule=self.after,
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

    def _on_progress(
        self, file_idx: int, seg_end: float, total_dur: float, status: str
    ) -> None:
        self._file_list.set_status(file_idx, "processing")

        total_files = self._file_list.file_count
        file_progress = seg_end / total_dur if total_dur > 0 else 0
        overall = (file_idx + file_progress) / total_files

        self._progress.set_progress(overall)
        self._progress.set_status(status)

    def _on_file_done(self, file_idx: int, output_path: str) -> None:
        self._file_list.set_status(file_idx, "done")
        self._last_output_folder = os.path.dirname(output_path)

    def _on_error(self, file_idx: int, error_msg: str) -> None:
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

    # ── Window close ──────────────────────────────────────────────

    def _on_close(self) -> None:
        if self._worker and self._worker.is_running:
            if messagebox.askyesno(
                "Transcription in progress",
                "Transcription is still running.\nCancel and exit?",
            ):
                self._worker.cancel()
                self.after(500, self.destroy)
        elif self._downloader and self._downloader.is_running:
            if messagebox.askyesno(
                "Download in progress",
                "Model download is in progress.\nCancel and exit?",
            ):
                self._downloader.cancel()
                self.after(500, self.destroy)
        else:
            self.destroy()
