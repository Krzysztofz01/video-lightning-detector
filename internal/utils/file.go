package utils

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// Create the whole path directory tree and the final file.
func CreateFileWithTree(path string) (*os.File, error) {
	directoryPath := filepath.Dir(path)

	if _, err := os.Stat(directoryPath); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(directoryPath, 0770); err != nil {
			return nil, err
		}
	}

	return os.Create(path)
}

// Create a new png file at the given path and encode the specified image into it.
// TODO: Add unit tests
func ExportImageAsPng(path string, img image.Image) error {
	if len(path) == 0 {
		return errors.New("utils: invalid image path specified")
	}

	if img == nil {
		return errors.New("utils: the provided image reference is nil")
	}

	file, err := CreateFileWithTree(path)
	if err != nil {
		return fmt.Errorf("utils: failed to create the png image file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("utils: failed to encode the image as png: %w", err)
	}

	return nil
}
