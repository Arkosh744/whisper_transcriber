package main

import (
	"context"
	"fmt"

	"whisper-transcriber/pkg/models"
	"whisper-transcriber/internal/service"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx            context.Context
	transcriber    models.Transcriber
	modelManager   models.ModelManager
	ffmpeg         models.FFmpegService
	formatter      models.Formatter
	queue          models.FileQueue
	batch          *service.BatchProcessor
	batchCancel    context.CancelFunc
	downloadCancel context.CancelFunc
}

func NewApp(
	transcriber models.Transcriber,
	modelManager models.ModelManager,
	ffmpeg models.FFmpegService,
	formatter models.Formatter,
	queue models.FileQueue,
	batch *service.BatchProcessor,
) *App {
	return &App{
		transcriber:  transcriber,
		modelManager: modelManager,
		ffmpeg:       ffmpeg,
		formatter:    formatter,
		queue:        queue,
		batch:        batch,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(_ context.Context) {
	a.transcriber.Close()
}

func (a *App) BrowseFiles() ([]models.FileItem, error) {
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

	return a.queue.Add(selection), nil
}

func (a *App) AddFiles(paths []string) ([]models.FileItem, error) {
	return a.queue.Add(paths), nil
}

func (a *App) ClearFiles() {
	a.queue.Clear()
}

func (a *App) RemoveFile(id string) {
	a.queue.Remove(id)
}

func (a *App) GetLanguages() []models.LangOption {
	return []models.LangOption{
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
	return a.ffmpeg.IsAvailable()
}

func (a *App) DownloadFFmpeg() {
	ctx, cancel := context.WithCancel(a.ctx)
	a.downloadCancel = cancel
	go func() {
		defer func() { a.downloadCancel = nil }()
		if err := a.ffmpeg.Download(ctx, downloadProgressCb(a.ctx, "ffmpeg:download:progress")); err != nil {
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
	ctx, cancel := context.WithCancel(a.ctx)
	a.downloadCancel = cancel
	go func() {
		defer func() { a.downloadCancel = nil }()
		if err := a.modelManager.DownloadModel(ctx, downloadProgressCb(a.ctx, "model:download:progress")); err != nil {
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

func (a *App) StartTranscription(config models.TranscriptionConfig) error {
	if !a.modelManager.IsModelAvailable() {
		return fmt.Errorf("model not found — download it first")
	}

	if !a.ffmpeg.IsAvailable() {
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

	go a.batch.Run(
		batchCtx,
		config,
		fileStatusCb(a.ctx),
		func(fileID, outputPath string) {
			wailsRuntime.EventsEmit(a.ctx, "transcription:complete", map[string]interface{}{
				"fileID":     fileID,
				"outputPath": outputPath,
			})
		},
		func() {
			wailsRuntime.EventsEmit(a.ctx, "batch:complete", nil)
		},
	)
	return nil
}

func (a *App) CancelTranscription() {
	if a.batchCancel != nil {
		a.batchCancel()
	}
}

