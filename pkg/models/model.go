package models

import (
	"crypto/rand"
	"fmt"
)

type FileItem struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Name     string `json:"name"`
	SizeMB   int    `json:"sizeMb"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	Error    string `json:"error"`
}

type TranscriptionConfig struct {
	Language     string `json:"language"`
	OutputFormat string `json:"outputFormat"`
}

type Segment struct {
	Index int     `json:"index"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

type TranscriptionResult struct {
	FilePath string    `json:"filePath"`
	Language string    `json:"language"`
	Segments []Segment `json:"segments"`
}

type LangOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func GenerateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
