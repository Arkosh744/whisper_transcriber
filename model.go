package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	modelFileName = "ggml-large-v3-turbo-q5_0.bin"
	modelURL      = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3-turbo-q5_0.bin"
)

type ModelManager struct {
	ctx      context.Context
	modelDir string
}

func NewModelManager() *ModelManager {
	exePath, _ := os.Executable()
	return &ModelManager{
		modelDir: filepath.Join(filepath.Dir(exePath), "models"),
	}
}

func (m *ModelManager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

func (m *ModelManager) ModelPath() string {
	return filepath.Join(m.modelDir, modelFileName)
}

func (m *ModelManager) IsModelAvailable() bool {
	info, err := os.Stat(m.ModelPath())
	return err == nil && info.Size() > 0
}

func (m *ModelManager) DownloadModel() error {
	if err := os.MkdirAll(m.modelDir, 0755); err != nil {
		return fmt.Errorf("cannot create models dir: %w", err)
	}

	tmpPath := m.ModelPath() + ".tmp"
	defer os.Remove(tmpPath)

	resp, err := http.Get(modelURL)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	total := resp.ContentLength
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("cannot create temp file: %w", err)
	}
	defer out.Close()

	var downloaded int64
	buf := make([]byte, 64*1024)

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)
			if total > 0 && m.ctx != nil {
				pct := int(float64(downloaded) / float64(total) * 100)
				downloadedMB := float64(downloaded) / (1024 * 1024)
				totalMB := float64(total) / (1024 * 1024)
				wailsRuntime.EventsEmit(m.ctx, "model:download:progress", map[string]interface{}{
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
			return readErr
		}
	}

	out.Close()
	return os.Rename(tmpPath, m.ModelPath())
}
