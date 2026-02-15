package service

import (
	"context"
	"os"

	"whisper-transcriber/pkg/models"
)

type BatchCompleteFunc func(fileID, outputPath string)

type BatchDoneFunc func()

type BatchProcessor struct {
	transcriber models.Transcriber
	ffmpeg      models.FFmpegService
	formatter   models.Formatter
	queue       models.FileQueue
}

func NewBatchProcessor(
	transcriber models.Transcriber,
	ffmpeg models.FFmpegService,
	formatter models.Formatter,
	queue models.FileQueue,
) *BatchProcessor {
	return &BatchProcessor{
		transcriber: transcriber,
		ffmpeg:      ffmpeg,
		formatter:   formatter,
		queue:       queue,
	}
}

func (b *BatchProcessor) Run(
	ctx context.Context,
	config models.TranscriptionConfig,
	onStatus models.StatusFunc,
	onComplete BatchCompleteFunc,
	onDone BatchDoneFunc,
) {
	files := b.queue.Snapshot()

	for _, fileItem := range files {
		select {
		case <-ctx.Done():
			onStatus(fileItem.ID, "cancelled", 0, "")
			onDone()
			return
		default:
		}

		onStatus(fileItem.ID, "processing", 0, "")

		wavPath, err := b.ffmpeg.ExtractAudio(ctx, fileItem.Path)
		if err != nil {
			onStatus(fileItem.ID, "error", 0, err.Error())
			continue
		}

		progressCb := func(percent int, _, _ string) {
			onStatus(fileItem.ID, "processing", percent, "")
		}

		result, err := b.transcriber.TranscribeFile(ctx, fileItem.ID, wavPath, config.Language, progressCb)

		os.Remove(wavPath)

		if err != nil {
			onStatus(fileItem.ID, "error", 0, err.Error())
			continue
		}

		outPath, err := b.formatter.WriteOutput(result, fileItem.Path, config.OutputFormat)
		if err != nil {
			onStatus(fileItem.ID, "error", 0, err.Error())
			continue
		}

		onStatus(fileItem.ID, "done", 100, "")
		onComplete(fileItem.ID, outPath)
	}

	onDone()
}
