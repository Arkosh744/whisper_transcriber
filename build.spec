# -*- mode: python ; coding: utf-8 -*-
"""PyInstaller spec for Whisper Transcriber."""

import os
import sys
from PyInstaller.utils.hooks import collect_data_files, collect_all

site_packages = os.path.join(sys.prefix, "Lib", "site-packages")

# CustomTkinter theme assets
ctk_datas, ctk_binaries, ctk_hiddenimports = collect_all("customtkinter")

# CUDA DLLs from nvidia packages
cuda_bins = []
for lib in ["cublas", "cudnn"]:
    bin_dir = os.path.join(site_packages, "nvidia", lib, "bin")
    if os.path.isdir(bin_dir):
        for dll in os.listdir(bin_dir):
            if dll.endswith(".dll"):
                cuda_bins.append((os.path.join(bin_dir, dll), "."))

# ctranslate2 native libraries
ct2_datas, ct2_binaries, ct2_hiddenimports = collect_all("ctranslate2")

# faster-whisper assets
fw_datas, fw_binaries, fw_hiddenimports = collect_all("faster_whisper")

# tokenizers
tok_datas, tok_binaries, tok_hiddenimports = collect_all("tokenizers")

a = Analysis(
    ["main.py"],
    pathex=[],
    binaries=cuda_bins + ctk_binaries + ct2_binaries + fw_binaries + tok_binaries,
    datas=ctk_datas + ct2_datas + fw_datas + tok_datas,
    hiddenimports=(
        ctk_hiddenimports
        + ct2_hiddenimports
        + fw_hiddenimports
        + tok_hiddenimports
        + [
            "av",
            "huggingface_hub",
            "numpy",
        ]
    ),
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    noarchive=False,
)

pyz = PYZ(a.pure)

exe = EXE(
    pyz,
    a.scripts,
    [],
    exclude_binaries=True,
    name="WhisperTranscriber",
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=False,
    console=False,  # No console window (GUI app)
)

coll = COLLECT(
    exe,
    a.binaries,
    a.datas,
    strip=False,
    upx=False,
    name="WhisperTranscriber",
)
