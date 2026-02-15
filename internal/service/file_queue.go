package service

import (
	"os"
	"path/filepath"
	"sync"

	"whisper-transcriber/pkg/models"
)

type FileQueue struct {
	mu    sync.Mutex
	files []models.FileItem
}

func NewFileQueue() *FileQueue {
	return &FileQueue{}
}

func (q *FileQueue) Add(paths []string) []models.FileItem {
	var items []models.FileItem
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		items = append(items, models.FileItem{
			ID:     models.GenerateID(),
			Path:   path,
			Name:   filepath.Base(path),
			SizeMB: int(info.Size() / (1024 * 1024)),
			Status: "pending",
		})
	}

	q.mu.Lock()
	q.files = append(q.files, items...)
	q.mu.Unlock()

	return items
}

func (q *FileQueue) Remove(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, f := range q.files {
		if f.ID == id {
			q.files = append(q.files[:i], q.files[i+1:]...)
			return
		}
	}
}

func (q *FileQueue) Clear() {
	q.mu.Lock()
	q.files = nil
	q.mu.Unlock()
}

func (q *FileQueue) Snapshot() []models.FileItem {
	q.mu.Lock()
	defer q.mu.Unlock()
	cp := make([]models.FileItem, len(q.files))
	copy(cp, q.files)
	return cp
}

func (q *FileQueue) UpdateStatus(id, status string, progress int, errMsg string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i := range q.files {
		if q.files[i].ID == id {
			q.files[i].Status = status
			q.files[i].Progress = progress
			q.files[i].Error = errMsg
			return
		}
	}
}
