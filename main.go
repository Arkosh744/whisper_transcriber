package main

import (
	"embed"
	"log"

	"whisper-transcriber/internal/infrastructure"
	"whisper-transcriber/internal/service"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	appDir := infrastructure.AppDataDir()

	transcriber := service.NewTranscriber()
	modelMgr := service.NewModelManager(appDir)
	ffmpeg := service.NewFFmpegService(appDir)
	formatter := service.NewFormatter()
	queue := service.NewFileQueue()
	batch := service.NewBatchProcessor(transcriber, ffmpeg, formatter, queue)

	app := NewApp(transcriber, modelMgr, ffmpeg, formatter, queue, batch)

	err := wails.Run(&options.App{
		Title:     "Whisper Transcriber",
		Width:     900,
		Height:    640,
		MinWidth:  700,
		MinHeight: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 24, G: 24, B: 27, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			Theme:                windows.Dark,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal("Error:", err.Error())
	}
}
