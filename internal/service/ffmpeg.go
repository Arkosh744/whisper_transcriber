package service

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"whisper-transcriber/pkg/models"
	"whisper-transcriber/internal/infrastructure"
)

const (
	ffmpegWinURL = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip"
)

type FFmpegSvc struct {
	appDir string
}

func NewFFmpegService(appDir string) *FFmpegSvc {
	return &FFmpegSvc{appDir: appDir}
}

func (s *FFmpegSvc) localPath() string {
	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name = "ffmpeg.exe"
	}
	return filepath.Join(s.appDir, name)
}

func (s *FFmpegSvc) binPath() (string, error) {
	bundled := s.localPath()
	if _, err := os.Stat(bundled); err == nil {
		return bundled, nil
	}
	p, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", models.ErrFFmpegNotFound
	}
	return p, nil
}

func (s *FFmpegSvc) IsAvailable() bool {
	_, err := s.binPath()
	return err == nil
}

func (s *FFmpegSvc) Download(ctx context.Context, onProgress models.ProgressFunc) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("auto-download only supported on Windows; install ffmpeg via package manager")
	}

	dest := s.localPath()

	resp, err := infrastructure.HTTPGetWithRetry(ctx, ffmpegWinURL, 3)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	total := resp.ContentLength

	tmpZip := dest + ".zip.tmp"
	defer os.Remove(tmpZip)

	out, err := os.Create(tmpZip)
	if err != nil {
		return fmt.Errorf("cannot create temp file: %w", err)
	}
	defer out.Close()

	var downloaded int64
	buf := make([]byte, 64*1024)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)
			if total > 0 && onProgress != nil {
				pct := int(float64(downloaded) / float64(total) * 100)
				downloadedMB := fmt.Sprintf("%.0f", float64(downloaded)/(1024*1024))
				totalMB := fmt.Sprintf("%.0f", float64(total)/(1024*1024))
				onProgress(pct, downloadedMB, totalMB)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	out.Close()

	return extractFFmpegFromZip(tmpZip, dest)
}

func (s *FFmpegSvc) ExtractAudio(ctx context.Context, inputPath string) (string, error) {
	ff, err := s.binPath()
	if err != nil {
		return "", err
	}

	tmpFile, err := os.CreateTemp("", "whisper-*.wav")
	if err != nil {
		return "", err
	}
	outPath := tmpFile.Name()
	tmpFile.Close()

	cmd := exec.CommandContext(ctx, ff,
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
		if ctx.Err() != nil {
			return "", fmt.Errorf("ffmpeg cancelled: %w", ctx.Err())
		}
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
