package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
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

// Check if the given path is pointing to a existing file.
func FileExists(path string) bool {
	if file, err := os.Stat(path); (err != nil && os.IsNotExist(err)) || file.IsDir() {
		return false
	}

	return true
}

// Create a new png file at the given path and encode the specified image into it.
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

// Open a png of jpeg image with a RGBA pixel format representation
func ImportImageRgba(path string) (*image.RGBA, error) {
	var (
		file *os.File
		img  image.Image
		err  error
	)

	file, err = os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("utils: failed to open the target image file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		{
			if img, err = png.Decode(file); err != nil {
				return nil, fmt.Errorf("utils: failed to decode png file: %w", err)
			}
		}
	case ".jpg", ".jpeg":
		{
			if img, err = jpeg.Decode(file); err != nil {
				return nil, fmt.Errorf("utils: failed to decode jpeg file: %w", err)
			}
		}
	default:
		return nil, fmt.Errorf("utils: unsupported image file format")
	}

	if rgba, err := CopyAsRgba(img); err != nil {
		return nil, fmt.Errorf("utils: failed to copy image in rgba pixel format: %w", err)
	} else {
		return rgba, nil
	}
}

// Generate a SHA256 fingerprint (array of 32 bytes) of the file specified via the given path  consisting of
// the file size and three 2MB content chunks (files smaller than the chunks have different offsets)
func GetFileFingerprint(path string) ([32]byte, error) {
	const (
		chunkSize int64 = 2 * 1024 * 1024
	)

	file, err := os.Open(path)
	if err != nil {
		return [32]byte{}, fmt.Errorf("utils: failed to open the target file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	fingerprint := new(bytes.Buffer)

	fileInfo, err := file.Stat()
	if err != nil {
		return [32]byte{}, fmt.Errorf("utils: failed to read the target file stats: %w", err)
	}

	if err := binary.Write(fingerprint, binary.LittleEndian, fileInfo.Size()); err != nil {
		return [32]byte{}, fmt.Errorf("utils: failed to write the file size to fingerprint: %w", err)
	}

	offsets := []int64{
		0,
		fileInfo.Size() / 2,
		int64(math.Max(0, float64(fileInfo.Size()-chunkSize))),
	}

	buffer := make([]byte, chunkSize)
	for _, offset := range offsets {
		if _, err := file.Seek(offset, io.SeekStart); err != nil {
			return [32]byte{}, fmt.Errorf("utils: failed to seek the target file at specified offset: %w", err)
		}

		n, err := io.ReadFull(file, buffer)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			return [32]byte{}, fmt.Errorf("utils: failed to access the target file at specified offset: %w", err)
		}

		if _, err := fingerprint.Write(buffer[:n]); err != nil {
			return [32]byte{}, fmt.Errorf("utils: failed to write the file chunk to fingerprint: %w", err)
		}
	}

	hash := sha256.Sum256(fingerprint.Bytes())
	return hash, nil
}
