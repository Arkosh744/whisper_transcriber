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

// Transcriber wraps whisper.cpp model and provides transcription with progress.
type Transcriber struct {
	wailsCtx context.Context
	model    whisper.Model
	mu       sync.Mutex
}

// NewTranscriber creates an uninitialized transcriber.
func NewTranscriber() *Transcriber {
	return &Transcriber{}
}

// SetContext stores the Wails context for event emission.
func (t *Transcriber) SetContext(ctx context.Context) {
	t.wailsCtx = ctx
}

// LoadModel initializes the whisper model from a GGML file.
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

// IsLoaded returns true if a model is currently loaded.
func (t *Transcriber) IsLoaded() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.model != nil
}

// TranscribeFile processes a single WAV file and returns segments.
// Emits "transcription:progress" events with {fileID, progress}.
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

	// Read WAV samples as []float32
	samples, err := readWavSamples(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio: %w", err)
	}

	// Create whisper context
	wCtx, err := t.model.NewContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create context: %w", err)
	}

	// Configure language
	if language != "" && language != "auto" {
		if err := wCtx.SetLanguage(language); err != nil {
			// fallback to auto if language not supported
			_ = wCtx.SetLanguage("auto")
		}
	}

	// Use EncoderBeginCallback for context cancellation,
	// ProgressCallback for native whisper.cpp progress reporting.
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
		nil, // segment callback (unused — we collect after)
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

	// Collect segments via NextSegment iterator
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

// Close releases the whisper model.
func (t *Transcriber) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.model != nil {
		t.model.Close()
		t.model = nil
	}
}

// readWavSamples reads a 16-bit PCM WAV file and returns float32 samples
// normalized to [-1.0, 1.0]. Assumes standard 44-byte WAV header
// (safe because we control FFmpeg output format exactly).
func readWavSamples(path string) ([]float32, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Skip WAV header (44 bytes for standard PCM WAV from FFmpeg)
	header := make([]byte, 44)
	if _, err := io.ReadFull(f, header); err != nil {
		return nil, fmt.Errorf("invalid WAV header: %w", err)
	}

	// Read 16-bit samples
	var samples []float32
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
