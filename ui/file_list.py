"""Scrollable file list widget with per-file status."""

from __future__ import annotations

import os
from typing import Callable

import customtkinter as ctk


class FileList(ctk.CTkFrame):
    """Displays a list of added files with status indicators."""

    STATUS_ICONS = {
        "pending": "\u23f3",    # hourglass
        "processing": "\u25b6",  # play
        "done": "\u2705",       # check
        "error": "\u274c",      # cross
    }

    def __init__(self, master, on_files_changed: Callable, **kwargs):
        super().__init__(master, **kwargs)
        self._on_files_changed = on_files_changed
        self._files: list[str] = []
        self._statuses: list[str] = []
        self._row_frames: list[ctk.CTkFrame] = []

        # Header with Browse + Clear buttons
        self._header = ctk.CTkFrame(self, fg_color="transparent")
        self._header.pack(fill="x", padx=5, pady=(5, 2))

        self._browse_btn = ctk.CTkButton(
            self._header,
            text="Browse Files",
            width=130,
            height=32,
            command=self._browse,
        )
        self._browse_btn.pack(side="left")

        self._add_btn = ctk.CTkButton(
            self._header,
            text="Add More",
            width=100,
            height=32,
            fg_color="transparent",
            border_width=1,
            command=self._browse,
        )
        self._add_btn.pack(side="left", padx=(8, 0))

        self._clear_btn = ctk.CTkButton(
            self._header,
            text="Clear All",
            width=90,
            height=32,
            fg_color="transparent",
            border_width=1,
            text_color=("gray40", "gray70"),
            command=self.clear,
        )
        self._clear_btn.pack(side="right")

        # Scrollable file list
        self._scroll = ctk.CTkScrollableFrame(self, height=180)
        self._scroll.pack(fill="both", expand=True, padx=5, pady=5)

        # Empty state label
        self._empty_label = ctk.CTkLabel(
            self._scroll,
            text="No files added.\nClick 'Browse Files' to select video files.",
            text_color=("gray50", "gray60"),
            font=ctk.CTkFont(size=13),
        )
        self._empty_label.pack(pady=30)

    @property
    def files(self) -> list[str]:
        return list(self._files)

    @property
    def file_count(self) -> int:
        return len(self._files)

    def _browse(self) -> None:
        from tkinter import filedialog
        from config import SUPPORTED_EXTENSIONS, FILE_TYPE_LABEL

        ext_pattern = " ".join(f"*{e}" for e in SUPPORTED_EXTENSIONS)
        paths = filedialog.askopenfilenames(
            title="Select video files",
            filetypes=[(FILE_TYPE_LABEL, ext_pattern), ("All files", "*.*")],
        )
        if paths:
            self.add_files(list(paths))

    def add_files(self, paths: list[str]) -> None:
        for p in paths:
            if p not in self._files:
                self._files.append(p)
                self._statuses.append("pending")
        self._rebuild_list()
        self._on_files_changed()

    def clear(self) -> None:
        self._files.clear()
        self._statuses.clear()
        self._rebuild_list()
        self._on_files_changed()

    def set_status(self, index: int, status: str) -> None:
        if 0 <= index < len(self._statuses):
            self._statuses[index] = status
            self._update_row(index)

    def _rebuild_list(self) -> None:
        for frame in self._row_frames:
            frame.destroy()
        self._row_frames.clear()

        if not self._files:
            self._empty_label.pack(pady=30)
            return

        self._empty_label.pack_forget()

        for i, (path, status) in enumerate(zip(self._files, self._statuses)):
            row = self._create_row(i, path, status)
            self._row_frames.append(row)

    def _create_row(self, index: int, path: str, status: str) -> ctk.CTkFrame:
        row = ctk.CTkFrame(self._scroll, height=30, fg_color=("gray90", "gray17"))
        row.pack(fill="x", pady=1)

        icon = self.STATUS_ICONS.get(status, "")
        icon_label = ctk.CTkLabel(row, text=icon, width=30, font=ctk.CTkFont(size=14))
        icon_label.pack(side="left", padx=(5, 0))

        name = os.path.basename(path)
        name_label = ctk.CTkLabel(
            row, text=name, anchor="w", font=ctk.CTkFont(size=13)
        )
        name_label.pack(side="left", fill="x", expand=True, padx=5)

        size_mb = os.path.getsize(path) / (1024 * 1024) if os.path.exists(path) else 0
        size_label = ctk.CTkLabel(
            row,
            text=f"{size_mb:.0f} MB",
            width=60,
            text_color=("gray50", "gray60"),
            font=ctk.CTkFont(size=12),
        )
        size_label.pack(side="right", padx=(0, 5))

        return row

    def _update_row(self, index: int) -> None:
        if 0 <= index < len(self._row_frames):
            row = self._row_frames[index]
            icon = self.STATUS_ICONS.get(self._statuses[index], "")
            # Update the icon label (first child)
            children = row.winfo_children()
            if children:
                children[0].configure(text=icon)

    def set_enabled(self, enabled: bool) -> None:
        state = "normal" if enabled else "disabled"
        self._browse_btn.configure(state=state)
        self._add_btn.configure(state=state)
        self._clear_btn.configure(state=state)
