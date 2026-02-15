package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Transcriber struct {
	wailsCtx context.Context
	model    whisper.Model
	mu       sync.Mutex
}

func NewTranscriber() *Transcriber {
	return &Transcriber{}
}

func (t *Transcriber) SetContext(ctx context.Context) {
	t.wailsCtx = ctx
}

func (t *Transcriber) LoadModel(modelPath string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.model != nil {
		t.model.Close()
	}

	model, err := whisper.New(modelPath)
	if err != nil {
		return fmt.Errorf("failed to load model: %w", err)
	}
	t.model = model
	return nil
}

func (t *Transcriber) IsLoaded() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.model != nil
}

func (t *Transcriber) TranscribeFile(
	ctx context.Context,
	fileID string,
	audioPath string,
	language string,
) (*TranscriptionResult, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.model == nil {
		return nil, fmt.Errorf("model not loaded")
	}

	samples, err := readWavSamples(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio: %w", err)
	}

	wCtx, err := t.model.NewContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create context: %w", err)
	}

	if language != "" && language != "auto" {
		if err := wCtx.SetLanguage(language); err != nil {
			_ = wCtx.SetLanguage("auto")
		}
	}

	cancelled := false
	if err := wCtx.Process(samples,
		func() bool {
			select {
			case <-ctx.Done():
				cancelled = true
				return false
			default:
				return true
			}
		},
		nil,
		func(progress int) {
			if t.wailsCtx != nil {
				wailsRuntime.EventsEmit(t.wailsCtx, "transcription:progress", map[string]interface{}{
					"fileID":   fileID,
					"progress": progress,
				})
			}
		},
	); err != nil {
		if cancelled {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("transcription failed: %w", err)
	}

	var segments []Segment
	for i := 0; ; i++ {
		seg, err := wCtx.NextSegment()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading segment: %w", err)
		}
		segments = append(segments, Segment{
			Index: seg.Num,
			Start: seg.Start.Seconds(),
			End:   seg.End.Seconds(),
			Text:  seg.Text,
		})
	}

	return &TranscriptionResult{
		FilePath: audioPath,
		Language: language,
		Segments: segments,
	}, nil
}

func (t *Transcriber) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.model != nil {
		t.model.Close()
		t.model = nil
	}
}

func readWavSamples(path string) ([]float32, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	header := make([]byte, 44)
	if _, err := io.ReadFull(f, header); err != nil {
		return nil, fmt.Errorf("invalid WAV header: %w", err)
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	dataSize := fi.Size() - 44 // subtract WAV header
	if dataSize < 0 {
		dataSize = 0
	}
	n := dataSize / 2 // 2 bytes per int16 sample
	samples := make([]float32, 0, n)
	for {
		var sample int16
		err := binary.Read(f, binary.LittleEndian, &sample)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		samples = append(samples, float32(sample)/32768.0)
	}
	return samples, nil
}
