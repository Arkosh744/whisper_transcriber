"""Output format writers for transcription segments."""

import json


def format_txt(segments, output_path: str) -> None:
    """[MM:SS] text — plain text with timestamps."""
    with open(output_path, "w", encoding="utf-8") as f:
        for seg in segments:
            mm = int(seg.start // 60)
            ss = int(seg.start % 60)
            f.write(f"[{mm:02d}:{ss:02d}] {seg.text.strip()}\n")


def format_srt(segments, output_path: str) -> None:
    """Standard SRT subtitle format."""
    with open(output_path, "w", encoding="utf-8") as f:
        for i, seg in enumerate(segments, 1):
            start = _srt_time(seg.start)
            end = _srt_time(seg.end)
            f.write(f"{i}\n{start} --> {end}\n{seg.text.strip()}\n\n")


def format_json(segments, output_path: str) -> None:
    """JSON array of {start, end, text}."""
    data = [
        {
            "start": round(seg.start, 3),
            "end": round(seg.end, 3),
            "text": seg.text.strip(),
        }
        for seg in segments
    ]
    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)


def format_markdown(segments, output_path: str) -> None:
    """Markdown with timestamp headers."""
    with open(output_path, "w", encoding="utf-8") as f:
        f.write("# Transcription\n\n")
        for seg in segments:
            mm = int(seg.start // 60)
            ss = int(seg.start % 60)
            f.write(f"**[{mm:02d}:{ss:02d}]** {seg.text.strip()}\n\n")


def _srt_time(seconds: float) -> str:
    """Convert seconds to HH:MM:SS,mmm format."""
    h = int(seconds // 3600)
    m = int((seconds % 3600) // 60)
    s = int(seconds % 60)
    ms = int((seconds % 1) * 1000)
    return f"{h:02d}:{m:02d}:{s:02d},{ms:03d}"


# Registry: format name -> (formatter function, file extension)
FORMATTERS = {
    "TXT": (format_txt, ".txt"),
    "SRT": (format_srt, ".srt"),
    "JSON": (format_json, ".json"),
    "Markdown": (format_markdown, ".md"),
}
