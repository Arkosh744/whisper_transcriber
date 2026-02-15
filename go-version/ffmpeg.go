package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ffmpegBin returns the path to ffmpeg, checking bundled location first.
func ffmpegBin() (string, error) {
	// 1. Check next to executable
	exePath, _ := os.Executable()
	bundled := filepath.Join(filepath.Dir(exePath), "ffmpeg")
	if _, err := os.Stat(bundled); err == nil {
		return bundled, nil
	}
	// 2. Fall back to system PATH
	p, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", fmt.Errorf("ffmpeg not found: install it or place ffmpeg binary next to the app")
	}
	return p, nil
}

// ExtractAudio converts any audio/video file to 16kHz mono WAV suitable for whisper.cpp.
// whisper.cpp only accepts raw PCM WAV, so we always run ffmpeg regardless of input format.
// Returns the path to the temporary WAV file. Caller must os.Remove() it when done.
func ExtractAudio(inputPath string) (string, error) {
	ff, err := ffmpegBin()
	if err != nil {
		return "", err
	}

	tmpFile, err := os.CreateTemp("", "whisper-*.wav")
	if err != nil {
		return "", err
	}
	outPath := tmpFile.Name()
	tmpFile.Close()

	cmd := exec.Command(ff,
		"-i", inputPath,
		"-ar", "16000",      // 16 kHz sample rate (whisper requirement)
		"-ac", "1",           // mono
		"-c:a", "pcm_s16le", // 16-bit PCM
		"-y",                 // overwrite
		outPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(outPath)
		return "", fmt.Errorf("ffmpeg failed: %s\n%s", err, string(output))
	}
	return outPath, nil
}
