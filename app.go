package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx            context.Context
	transcriber    *Transcriber
	modelManager   *ModelManager
	files          []FileItem
	mu             sync.Mutex
	batchCancel    context.CancelFunc
	downloadCancel context.CancelFunc
}

func NewApp() *App {
	return &App{
		transcriber:  NewTranscriber(),
		modelManager: NewModelManager(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.transcriber.SetContext(ctx)
	a.modelManager.SetContext(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	a.transcriber.Close()
}

func (a *App) BrowseFiles() ([]FileItem, error) {
	selection, err := wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Video Files",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Video Files", Pattern: "*.mp4;*.mkv;*.avi;*.mov;*.webm"},
			{DisplayName: "Audio Files", Pattern: "*.wav;*.mp3;*.flac;*.ogg;*.m4a"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, err
	}

	var items []FileItem
	for _, path := range selection {
		sizeMB := 0
		if info, err := os.Stat(path); err == nil {
			sizeMB = int(info.Size() / (1024 * 1024))
		}
		items = append(items, FileItem{
			ID:     generateID(),
			Path:   path,
			Name:   filepath.Base(path),
			SizeMB: sizeMB,
			Status: "pending",
		})
	}

	a.mu.Lock()
	a.files = append(a.files, items...)
	a.mu.Unlock()

	return items, nil
}

func (a *App) ClearFiles() {
	a.mu.Lock()
	a.files = nil
	a.mu.Unlock()
}

func (a *App) RemoveFile(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, f := range a.files {
		if f.ID == id {
			a.files = append(a.files[:i], a.files[i+1:]...)
			return
		}
	}
}

func (a *App) GetLanguages() []LangOption {
	return []LangOption{
		{Code: "auto", Name: "Auto-detect"},
		{Code: "ru", Name: "Russian"},
		{Code: "en", Name: "English"},
		{Code: "de", Name: "German"},
		{Code: "fr", Name: "French"},
		{Code: "es", Name: "Spanish"},
		{Code: "zh", Name: "Chinese"},
		{Code: "ja", Name: "Japanese"},
		{Code: "ko", Name: "Korean"},
		{Code: "uk", Name: "Ukrainian"},
		{Code: "pl", Name: "Polish"},
		{Code: "it", Name: "Italian"},
		{Code: "pt", Name: "Portuguese"},
		{Code: "tr", Name: "Turkish"},
		{Code: "ar", Name: "Arabic"},
		{Code: "hi", Name: "Hindi"},
	}
}

func (a *App) IsFFmpegAvailable() bool {
	return IsFFmpegAvailable()
}

func (a *App) DownloadFFmpeg() {
	downloadCtx, cancel := context.WithCancel(a.ctx)
	a.downloadCancel = cancel
	go func() {
		defer func() { a.downloadCancel = nil }()
		if err := DownloadFFmpeg(downloadCtx); err != nil {
			wailsRuntime.EventsEmit(a.ctx, "ffmpeg:download:error", err.Error())
			return
		}
		wailsRuntime.EventsEmit(a.ctx, "ffmpeg:download:done", nil)
	}()
}

func (a *App) IsModelAvailable() bool {
	return a.modelManager.IsModelAvailable()
}

func (a *App) DownloadModel() {
	downloadCtx, cancel := context.WithCancel(a.ctx)
	a.downloadCancel = cancel
	go func() {
		defer func() { a.downloadCancel = nil }()
		if err := a.modelManager.DownloadModel(downloadCtx); err != nil {
			wailsRuntime.EventsEmit(a.ctx, "model:download:error", err.Error())
			return
		}
		wailsRuntime.EventsEmit(a.ctx, "model:download:done", a.modelManager.ModelPath())
	}()
}

func (a *App) CancelDownload() {
	if a.downloadCancel != nil {
		a.downloadCancel()
	}
}

func (a *App) StartTranscription(config TranscriptionConfig) error {
	if !a.modelManager.IsModelAvailable() {
		return fmt.Errorf("model not found — download it first")
	}

	if !IsFFmpegAvailable() {
		return fmt.Errorf("FFmpeg not found — download it first")
	}

	if !a.transcriber.IsLoaded() {
		wailsRuntime.EventsEmit(a.ctx, "model:loading", nil)
		if err := a.transcriber.LoadModel(a.modelManager.ModelPath()); err != nil {
			return fmt.Errorf("failed to load model: %w", err)
		}
		wailsRuntime.EventsEmit(a.ctx, "model:loaded", nil)
	}

	batchCtx, cancel := context.WithCancel(a.ctx)
	a.batchCancel = cancel

	go a.runBatch(batchCtx, config)
	return nil
}

func (a *App) CancelTranscription() {
	if a.batchCancel != nil {
		a.batchCancel()
	}
}

func (a *App) AddFiles(paths []string) ([]FileItem, error) {
	var items []FileItem
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}
		sizeMB := int(info.Size() / (1024 * 1024))
		items = append(items, FileItem{
			ID:     generateID(),
			Path:   path,
			Name:   filepath.Base(path),
			SizeMB: sizeMB,
			Status: "pending",
		})
	}

	a.mu.Lock()
	a.files = append(a.files, items...)
	a.mu.Unlock()

	return items, nil
}

func (a *App) runBatch(ctx context.Context, config TranscriptionConfig) {
	a.mu.Lock()
	filesToProcess := make([]FileItem, len(a.files))
	copy(filesToProcess, a.files)
	a.mu.Unlock()

	for _, fileItem := range filesToProcess {
		select {
		case <-ctx.Done():
			a.emitStatus(fileItem.ID, "cancelled", 0, "")
			wailsRuntime.EventsEmit(a.ctx, "batch:complete", nil)
			return
		default:
		}

		a.emitStatus(fileItem.ID, "processing", 0, "")

		wavPath, err := ExtractAudio(ctx, fileItem.Path)
		if err != nil {
			a.emitStatus(fileItem.ID, "error", 0, err.Error())
			continue
		}
		audioPath := wavPath

		result, err := a.transcriber.TranscribeFile(ctx, fileItem.ID, audioPath, config.Language)

		os.Remove(audioPath)

		if err != nil {
			a.emitStatus(fileItem.ID, "error", 0, err.Error())
			continue
		}

		outPath, err := WriteOutput(result, fileItem.Path, config.OutputFormat)
		if err != nil {
			a.emitStatus(fileItem.ID, "error", 0, err.Error())
			continue
		}

		a.emitStatus(fileItem.ID, "done", 100, "")
		wailsRuntime.EventsEmit(a.ctx, "transcription:complete", map[string]interface{}{
			"fileID":     fileItem.ID,
			"outputPath": outPath,
		})
	}

	wailsRuntime.EventsEmit(a.ctx, "batch:complete", nil)
}

func (a *App) emitStatus(fileID, status string, progress int, errMsg string) {
	wailsRuntime.EventsEmit(a.ctx, "file:status", map[string]interface{}{
		"fileID":   fileID,
		"status":   status,
		"progress": progress,
		"error":    errMsg,
	})
}
