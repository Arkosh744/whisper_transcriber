package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"whisper-transcriber/pkg/models"
)

type Formatter struct{}

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (f *Formatter) WriteOutput(result *models.TranscriptionResult, sourcePath, format string) (string, error) {
	valid := map[string]bool{"txt": true, "srt": true, "json": true, "md": true}
	if !valid[format] {
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	base := strings.TrimSuffix(sourcePath, filepath.Ext(sourcePath))
	outPath := base + "." + format

	var content string
	var err error

	switch format {
	case "txt":
		content = formatTXT(result)
	case "srt":
		content = formatSRT(result)
	case "json":
		content, err = formatJSON(result)
		if err != nil {
			return "", err
		}
	case "md":
		content = formatMarkdown(result)
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}

	return outPath, os.WriteFile(outPath, []byte(content), 0644)
}

func formatTXT(r *models.TranscriptionResult) string {
	var sb strings.Builder
	for _, seg := range r.Segments {
		mm := int(seg.Start) / 60
		ss := int(seg.Start) % 60
		sb.WriteString(fmt.Sprintf("[%02d:%02d] %s\n", mm, ss, strings.TrimSpace(seg.Text)))
	}
	return sb.String()
}

func formatSRT(r *models.TranscriptionResult) string {
	var sb strings.Builder
	for i, seg := range r.Segments {
		sb.WriteString(fmt.Sprintf("%d\n", i+1))
		sb.WriteString(fmt.Sprintf("%s --> %s\n", srtTime(seg.Start), srtTime(seg.End)))
		sb.WriteString(strings.TrimSpace(seg.Text) + "\n\n")
	}
	return sb.String()
}

func srtTime(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	ms := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}

func formatJSON(r *models.TranscriptionResult) (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func formatMarkdown(r *models.TranscriptionResult) string {
	var sb strings.Builder
	sb.WriteString("# Transcription\n\n")
	for _, seg := range r.Segments {
		mm := int(seg.Start) / 60
		ss := int(seg.Start) % 60
		sb.WriteString(fmt.Sprintf("**[%02d:%02d]** %s\n\n", mm, ss, strings.TrimSpace(seg.Text)))
	}
	return sb.String()
}
