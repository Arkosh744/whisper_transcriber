package main

import (
	"crypto/rand"
	"fmt"
)

// FileItem represents a single file in the transcription queue.
type FileItem struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Name     string `json:"name"`
	SizeMB   int    `json:"sizeMb"`
	Status   string `json:"status"`   // pending | processing | done | error | cancelled
	Progress int    `json:"progress"` // 0-100
	Error    string `json:"error"`
}

// TranscriptionConfig holds user-selected options for a batch run.
type TranscriptionConfig struct {
	Language     string `json:"language"`     // "auto", "en", "ru", etc.
	OutputFormat string `json:"outputFormat"` // "txt", "srt", "json", "md"
}

// Segment is a single transcription segment with timestamps.
type Segment struct {
	Index int     `json:"index"`
	Start float64 `json:"start"` // seconds
	End   float64 `json:"end"`   // seconds
	Text  string  `json:"text"`
}

// TranscriptionResult holds all segments for one file.
type TranscriptionResult struct {
	FilePath string    `json:"filePath"`
	Language string    `json:"language"`
	Segments []Segment `json:"segments"`
}

// LangOption is a language choice for the UI dropdown.
type LangOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
