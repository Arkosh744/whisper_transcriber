package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	ffmpegWinURL = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip"
)

func ffmpegBin() (string, error) {
	bundled := ffmpegLocalPath()
	if _, err := os.Stat(bundled); err == nil {
		return bundled, nil
	}
	p, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", fmt.Errorf("ffmpeg not found: download it via the app or install system-wide")
	}
	return p, nil
}

func ffmpegLocalPath() string {
	exePath, _ := os.Executable()
	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name = "ffmpeg.exe"
	}
	return filepath.Join(filepath.Dir(exePath), name)
}

func IsFFmpegAvailable() bool {
	_, err := ffmpegBin()
	return err == nil
}

func DownloadFFmpeg(ctx context.Context) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("auto-download only supported on Windows; install ffmpeg via package manager")
	}

	dest := ffmpegLocalPath()

	resp, err := http.Get(ffmpegWinURL)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	total := resp.ContentLength

	tmpZip := dest + ".zip.tmp"
	defer os.Remove(tmpZip)

	out, err := os.Create(tmpZip)
	if err != nil {
		return fmt.Errorf("cannot create temp file: %w", err)
	}

	var downloaded int64
	buf := make([]byte, 64*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				out.Close()
				return writeErr
			}
			downloaded += int64(n)
			if total > 0 {
				pct := int(float64(downloaded) / float64(total) * 100)
				downloadedMB := float64(downloaded) / (1024 * 1024)
				totalMB := float64(total) / (1024 * 1024)
				wailsRuntime.EventsEmit(ctx, "ffmpeg:download:progress", map[string]interface{}{
					"percent":    pct,
					"downloaded": fmt.Sprintf("%.0f", downloadedMB),
					"total":      fmt.Sprintf("%.0f", totalMB),
				})
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			out.Close()
			return readErr
		}
	}
	out.Close()

	if err := extractFFmpegFromZip(tmpZip, dest); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	return nil
}

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
		"-ar", "16000",
		"-ac", "1",
		"-c:a", "pcm_s16le",
		"-y",
		outPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(outPath)
		return "", fmt.Errorf("ffmpeg failed: %s\n%s", err, string(output))
	}
	return outPath, nil
}

func extractFFmpegFromZip(zipPath, destPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "bin/ffmpeg.exe") {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, rc)
			return err
		}
	}
	return fmt.Errorf("ffmpeg.exe not found in archive")
}
