package models

import "errors"

var (
	ErrModelNotLoaded = errors.New("model not loaded")
	ErrFFmpegNotFound = errors.New("ffmpeg not found: download it via the app or install system-wide")
)
