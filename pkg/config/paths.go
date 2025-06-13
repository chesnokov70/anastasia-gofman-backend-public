package config

import (
	"os"
	"path/filepath"
)

func GetUploadsPath() string {
	workDir, err := os.Getwd()
	if err != nil {
		return "uploads"
	}
	return filepath.Join(workDir, "uploads")
}

func GetUploadFilePath(subdir, filename string) string {
	return filepath.Join(GetUploadsPath(), subdir, filename)
}
