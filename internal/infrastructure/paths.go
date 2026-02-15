package infrastructure

import (
	"os"
	"path/filepath"
)

func AppDataDir() string {
	exePath, _ := os.Executable()
	dir := filepath.Dir(exePath)

	testPath := filepath.Join(dir, ".writetest")
	if err := os.WriteFile(testPath, []byte{}, 0644); err != nil {
		if configDir, err := os.UserConfigDir(); err == nil {
			dir = filepath.Join(configDir, "WhisperTranscriber")
			os.MkdirAll(dir, 0755)
		}
	} else {
		os.Remove(testPath)
	}

	return dir
}
