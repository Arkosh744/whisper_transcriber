package models

import "context"

type Transcriber interface {
	LoadModel(modelPath string) error
	IsLoaded() bool
	TranscribeFile(ctx context.Context, fileID, audioPath, language string, onProgress ProgressFunc) (*TranscriptionResult, error)
	Close()
}

type ModelManager interface {
	ModelPath() string
	IsModelAvailable() bool
	DownloadModel(ctx context.Context, onProgress ProgressFunc) error
}

type FFmpegService interface {
	IsAvailable() bool
	Download(ctx context.Context, onProgress ProgressFunc) error
	ExtractAudio(ctx context.Context, inputPath string) (wavPath string, err error)
}

type Formatter interface {
	WriteOutput(result *TranscriptionResult, sourcePath, format string) (outputPath string, err error)
}

type FileQueue interface {
	Add(paths []string) []FileItem
	Remove(id string)
	Clear()
	Snapshot() []FileItem
	UpdateStatus(id, status string, progress int, errMsg string)
}
