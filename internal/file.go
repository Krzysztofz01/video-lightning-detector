package internal

import (
	"os"
	"path/filepath"
)

func CreateFileWithTree(path string) (*os.File, error) {
	directoryPath := filepath.Dir(path)

	if _, err := os.Stat(directoryPath); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(directoryPath, 0770); err != nil {
			return nil, err
		}
	}

	return os.Create(path)
}
