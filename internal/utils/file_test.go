package utils

import (
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldCreateFileWithTree(t *testing.T) {
	cases := []string{
		"test_file.test",
		"test/test_file.txt",
		"test/test/test_file.txt",
	}

	for _, c := range cases {
		file, err := CreateFileWithTree(c)

		assert.NotNil(t, file)
		assert.Nil(t, err)

		// NOTE: Cleanup
		file.Close()
		os.RemoveAll(c)
	}
}

func TestShouldNotExportImageAsPngForEmptyPath(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)

	err := ExportImageAsPng("", img)

	assert.NotNil(t, err)
}

func TestShouldNotExportImageAsPngForNilImage(t *testing.T) {
	err := ExportImageAsPng("test/test_image.png", nil)

	assert.NotNil(t, err)
}

func TestShouldExportImageAsPngForValidPathAndImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)

	path := "test/test_image.png"

	err := ExportImageAsPng(path, img)

	assert.Nil(t, err)

	// NOTE: Cleanup
	os.RemoveAll(path)
}
