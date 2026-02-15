package main

import (
	"context"

	"whisper-transcriber/pkg/models"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func downloadProgressCb(ctx context.Context, event string) models.ProgressFunc {
	return func(percent int, downloadedMB, totalMB string) {
		wailsRuntime.EventsEmit(ctx, event, map[string]interface{}{
			"percent":    percent,
			"downloaded": downloadedMB,
			"total":      totalMB,
		})
	}
}

func fileStatusCb(ctx context.Context) models.StatusFunc {
	return func(fileID, status string, progress int, errMsg string) {
		wailsRuntime.EventsEmit(ctx, "file:status", map[string]interface{}{
			"fileID":   fileID,
			"status":   status,
			"progress": progress,
			"error":    errMsg,
		})
	}
}
