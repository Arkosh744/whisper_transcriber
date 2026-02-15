package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"whisper-transcriber/pkg/models"
	"whisper-transcriber/internal/infrastructure"
)

const (
	modelFileName = "ggml-large-v3-turbo-q5_0.bin"
	modelURL      = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3-turbo-q5_0.bin"
)

type ModelMgr struct {
	modelDir string
}

func NewModelManager(appDir string) *ModelMgr {
	return &ModelMgr{
		modelDir: filepath.Join(appDir, "models"),
	}
}

func (m *ModelMgr) ModelPath() string {
	return filepath.Join(m.modelDir, modelFileName)
}

func (m *ModelMgr) IsModelAvailable() bool {
	info, err := os.Stat(m.ModelPath())
	return err == nil && info.Size() > 0
}

func (m *ModelMgr) DownloadModel(ctx context.Context, onProgress models.ProgressFunc) error {
	if err := os.MkdirAll(m.modelDir, 0755); err != nil {
		return fmt.Errorf("cannot create models dir: %w", err)
	}

	tmpPath := m.ModelPath() + ".tmp"
	defer os.Remove(tmpPath)

	resp, err := infrastructure.HTTPGetWithRetry(ctx, modelURL, 3)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	total := resp.ContentLength
	out, err := os.Create(tmpPath)
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
	return os.Rename(tmpPath, m.ModelPath())
}
