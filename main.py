"""Entry point: CUDA setup and app launch."""

import os
import sys

# Add project root to path so imports work
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))


def setup_cuda_paths():
    """Add NVIDIA CUDA DLL directories to PATH."""
    nvidia_libs = os.path.join(sys.prefix, "Lib", "site-packages", "nvidia")
    for lib in ["cublas", "cudnn"]:
        bin_dir = os.path.join(nvidia_libs, lib, "bin")
        if os.path.isdir(bin_dir):
            os.environ["PATH"] = bin_dir + os.pathsep + os.environ.get("PATH", "")


def main():
    setup_cuda_paths()

    from app import TranscriberApp

    app = TranscriberApp()
    app.mainloop()


if __name__ == "__main__":
    main()
