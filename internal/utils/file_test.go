package utils

import (
	"image"
	"image/color"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testPath = "test"

func testCleanupPath() {
	_ = os.RemoveAll(testPath)
}

func TestShouldCreateFileWithTree(t *testing.T) {
	testCleanupPath()
	defer testCleanupPath()

	cases := []string{
		path.Join(testPath, "test_file.test"),
		path.Join(testPath, "test/test_file.txt"),
		path.Join(testPath, "test/test/test_file.txt"),
	}

	for _, c := range cases {
		file, err := CreateFileWithTree(c)

		assert.NotNil(t, file)
		assert.Nil(t, err)

		_ = file.Close()
	}
}

func TestShouldNotExportImageAsPngForEmptyPath(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)

	err := ExportImageAsPng("", img)

	assert.NotNil(t, err)
}

func TestShouldNotExportImageAsPngForNilImage(t *testing.T) {
	testCleanupPath()
	defer testCleanupPath()

	imagePath := path.Join(testPath, "test/test_image.png")
	err := ExportImageAsPng(imagePath, nil)

	assert.NotNil(t, err)
}

func TestShouldExportImageAsPngForValidPathAndImage(t *testing.T) {
	testCleanupPath()
	defer testCleanupPath()

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)

	imagePath := path.Join(testPath, "test/test_image.png")

	assert.False(t, FileExists(imagePath))

	err := ExportImageAsPng(imagePath, img)
	assert.Nil(t, err)

	assert.True(t, FileExists(imagePath))
}
