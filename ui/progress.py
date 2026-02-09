"""Progress bar and status label widget."""

from __future__ import annotations

import os
import subprocess
from typing import Optional

import customtkinter as ctk


class ProgressPanel(ctk.CTkFrame):
    """Shows transcription progress and status."""

    def __init__(self, master, **kwargs):
        super().__init__(master, fg_color="transparent", **kwargs)

        self._bar = ctk.CTkProgressBar(self, height=18, corner_radius=5)
        self._bar.pack(fill="x", padx=10, pady=(5, 3))
        self._bar.set(0)

        self._status = ctk.CTkLabel(
            self,
            text="Ready",
            font=ctk.CTkFont(size=12),
            text_color=("gray40", "gray60"),
            anchor="w",
        )
        self._status.pack(fill="x", padx=12, pady=(0, 2))

        self._open_btn = ctk.CTkButton(
            self,
            text="Open Output Folder",
            width=160,
            height=30,
            fg_color="transparent",
            border_width=1,
            command=self._open_folder,
        )
        self._open_btn.pack(pady=(0, 5))
        self._open_btn.pack_forget()  # hidden until done

        self._output_folder: Optional[str] = None

    def set_progress(self, value: float) -> None:
        """Set progress bar value (0.0 to 1.0)."""
        self._bar.set(max(0.0, min(1.0, value)))

    def set_status(self, text: str) -> None:
        self._status.configure(text=text)

    def set_indeterminate(self, active: bool) -> None:
        if active:
            self._bar.configure(mode="indeterminate")
            self._bar.start()
        else:
            self._bar.stop()
            self._bar.configure(mode="determinate")

    def reset(self) -> None:
        self._bar.set(0)
        self._status.configure(text="Ready")
        self._open_btn.pack_forget()
        self._output_folder = None

    def show_complete(self, output_folder: str) -> None:
        self._output_folder = output_folder
        self._bar.set(1.0)
        self._status.configure(text="Complete!")
        self._open_btn.pack(pady=(0, 5))

    def _open_folder(self) -> None:
        if self._output_folder and os.path.isdir(self._output_folder):
            os.startfile(self._output_folder)
