"""Control panel: language selector, format selector, start/cancel buttons."""

from __future__ import annotations

from typing import Callable

import customtkinter as ctk

from config import LANGUAGES, OUTPUT_FORMATS


class Controls(ctk.CTkFrame):
    """Language/format dropdowns and start/cancel buttons."""

    def __init__(
        self,
        master,
        on_start: Callable,
        on_cancel: Callable,
        **kwargs,
    ):
        super().__init__(master, fg_color="transparent", **kwargs)
        self._on_start = on_start
        self._on_cancel = on_cancel

        # --- Row 1: Dropdowns ---
        row1 = ctk.CTkFrame(self, fg_color="transparent")
        row1.pack(fill="x", padx=10, pady=(5, 8))

        ctk.CTkLabel(row1, text="Language:", font=ctk.CTkFont(size=13)).pack(
            side="left"
        )
        self._lang_display = list(LANGUAGES.values())
        self._lang_codes = list(LANGUAGES.keys())
        self._lang_var = ctk.StringVar(value=self._lang_display[0])
        self._lang_menu = ctk.CTkOptionMenu(
            row1,
            values=self._lang_display,
            variable=self._lang_var,
            width=150,
            height=30,
        )
        self._lang_menu.pack(side="left", padx=(5, 20))

        ctk.CTkLabel(row1, text="Format:", font=ctk.CTkFont(size=13)).pack(
            side="left"
        )
        self._format_var = ctk.StringVar(value="TXT")
        self._format_menu = ctk.CTkOptionMenu(
            row1,
            values=list(OUTPUT_FORMATS.keys()),
            variable=self._format_var,
            width=120,
            height=30,
        )
        self._format_menu.pack(side="left", padx=5)

        # --- Row 2: Buttons ---
        row2 = ctk.CTkFrame(self, fg_color="transparent")
        row2.pack(fill="x", padx=10, pady=(0, 5))

        self._start_btn = ctk.CTkButton(
            row2,
            text="\u25b6  Start Transcription",
            width=220,
            height=38,
            font=ctk.CTkFont(size=14, weight="bold"),
            command=self._on_start,
        )
        self._start_btn.pack(side="left")

        self._cancel_btn = ctk.CTkButton(
            row2,
            text="Cancel",
            width=100,
            height=38,
            fg_color=("gray70", "gray30"),
            hover_color=("red3", "red4"),
            state="disabled",
            command=self._on_cancel,
        )
        self._cancel_btn.pack(side="left", padx=(10, 0))

    @property
    def selected_language(self) -> str | None:
        """Return language code or None for auto-detect."""
        display = self._lang_var.get()
        idx = self._lang_display.index(display)
        code = self._lang_codes[idx]
        return None if code == "auto" else code

    @property
    def selected_format(self) -> str:
        return self._format_var.get()

    def set_running(self, running: bool) -> None:
        if running:
            self._start_btn.configure(state="disabled")
            self._cancel_btn.configure(state="normal")
            self._lang_menu.configure(state="disabled")
            self._format_menu.configure(state="disabled")
        else:
            self._start_btn.configure(state="normal")
            self._cancel_btn.configure(state="disabled")
            self._lang_menu.configure(state="normal")
            self._format_menu.configure(state="normal")
