package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct bound to the Wails frontend.
type App struct {
	ctx          context.Context
	transcriber  *Transcriber
	modelManager *ModelManager
	files        []FileItem
	mu           sync.Mutex
	batchCancel  context.CancelFunc
}

// NewApp creates a new App.
func NewApp() *App {
	return &App{
		transcriber:  NewTranscriber(),
		modelManager: NewModelManager(),
	}
}

// startup is called when the Wails app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.transcriber.SetContext(ctx)
	a.modelManager.SetContext(ctx)
}

// shutdown is called when the app closes.
func (a *App) shutdown(ctx context.Context) {
	a.transcriber.Close()
}

// --- Bound methods (called from JS) ---

// BrowseFiles opens a native file dialog and returns selected files.
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

// ClearFiles removes all files from the queue.
func (a *App) ClearFiles() {
	a.mu.Lock()
	a.files = nil
	a.mu.Unlock()
}

// RemoveFile removes a file from the queue by ID.
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

// GetLanguages returns the list of available languages.
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

// IsFFmpegAvailable checks if ffmpeg is reachable (bundled or in PATH).
func (a *App) IsFFmpegAvailable() bool {
	return IsFFmpegAvailable()
}

// DownloadFFmpeg downloads a static ffmpeg build. Runs in a goroutine.
func (a *App) DownloadFFmpeg() {
	go func() {
		if err := DownloadFFmpeg(a.ctx); err != nil {
			wailsRuntime.EventsEmit(a.ctx, "ffmpeg:download:error", err.Error())
			return
		}
		wailsRuntime.EventsEmit(a.ctx, "ffmpeg:download:done", nil)
	}()
}

// IsModelAvailable checks if the GGML model exists locally.
func (a *App) IsModelAvailable() bool {
	return a.modelManager.IsModelAvailable()
}

// DownloadModel downloads the GGML model. Runs in a goroutine.
func (a *App) DownloadModel() {
	go func() {
		if err := a.modelManager.DownloadModel(); err != nil {
			wailsRuntime.EventsEmit(a.ctx, "model:download:error", err.Error())
			return
		}
		wailsRuntime.EventsEmit(a.ctx, "model:download:done", a.modelManager.ModelPath())
	}()
}

// StartTranscription begins batch processing all pending files.
func (a *App) StartTranscription(config TranscriptionConfig) error {
	// Ensure model
	if !a.modelManager.IsModelAvailable() {
		return fmt.Errorf("model not found — download it first")
	}

	// Load model if not already loaded
	if !a.transcriber.IsLoaded() {
		wailsRuntime.EventsEmit(a.ctx, "model:loading", nil)
		if err := a.transcriber.LoadModel(a.modelManager.ModelPath()); err != nil {
			return fmt.Errorf("failed to load model: %w", err)
		}
		wailsRuntime.EventsEmit(a.ctx, "model:loaded", nil)
	}

	// Create cancellable context for the batch
	batchCtx, cancel := context.WithCancel(context.Background())
	a.batchCancel = cancel

	go a.runBatch(batchCtx, config)
	return nil
}

// CancelTranscription stops the current batch.
func (a *App) CancelTranscription() {
	if a.batchCancel != nil {
		a.batchCancel()
	}
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

		// Step 1: Convert any audio/video to 16kHz mono WAV via FFmpeg
		// whisper.cpp only accepts raw PCM WAV, so we always convert
		wavPath, err := ExtractAudio(fileItem.Path)
		if err != nil {
			a.emitStatus(fileItem.ID, "error", 0, err.Error())
			continue
		}
		audioPath := wavPath

		// Step 2: Transcribe
		result, err := a.transcriber.TranscribeFile(ctx, fileItem.ID, audioPath, config.Language)

		// Cleanup temp WAV
		os.Remove(audioPath)

		if err != nil {
			a.emitStatus(fileItem.ID, "error", 0, err.Error())
			continue
		}

		// Step 3: Write output
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
