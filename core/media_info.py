"""Media file duration extraction via PyAV."""

import av


def get_duration(file_path: str) -> float:
    """Return file duration in seconds. Returns 0.0 on failure."""
    try:
        with av.open(file_path, mode="r", metadata_errors="ignore") as container:
            if container.duration is not None:
                return container.duration / 1_000_000  # microseconds -> seconds
            for stream in container.streams.audio:
                if stream.duration and stream.time_base:
                    return float(stream.duration * stream.time_base)
    except Exception:
        pass
    return 0.0
