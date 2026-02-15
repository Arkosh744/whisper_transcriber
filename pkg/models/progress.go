package models

type ProgressFunc func(percent int, downloadedMB, totalMB string)

type StatusFunc func(fileID, status string, progress int, errMsg string)
