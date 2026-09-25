package utils

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
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

func TestImportImageRgbaShouldImportImage(t *testing.T) {
	testCleanupPath()
	defer testCleanupPath()

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)

	pngImagePath := path.Join(testPath, "test_image.png")
	jpegImagePath := path.Join(testPath, "test_image.jpeg")

	if err := os.Mkdir(testPath, 0770); err != nil {
		t.FailNow()
	}

	pngFile, err := os.Create(pngImagePath)
	if err != nil {
		t.FailNow()
	}

	if err := png.Encode(pngFile, img); err != nil {
		t.FailNow()
	}

	if err := pngFile.Close(); err != nil {
		t.FailNow()
	}

	jpegFile, err := os.Create(jpegImagePath)
	if err != nil {
		t.FailNow()
	}

	if err := jpeg.Encode(jpegFile, img, nil); err != nil {
		t.FailNow()
	}

	if err := jpegFile.Close(); err != nil {
		t.FailNow()
	}

	importedPng, err := ImportImageRgba(pngImagePath)
	assert.Nil(t, err)
	assert.NotNil(t, importedPng)

	importedJpeg, err := ImportImageRgba(jpegImagePath)
	assert.Nil(t, err)
	assert.NotNil(t, importedJpeg)
}

func TestGetFileFingerprintShouldCreateCorrectFingerprints(t *testing.T) {
	testCleanupPath()
	defer testCleanupPath()

	if err := os.Mkdir(testPath, 0770); err != nil {
		t.FailNow()
	}

	createFile := func(name string, content []byte) string {
		p := path.Join(testPath, name)
		f, err := os.Create(p)
		if err != nil {
			t.FailNow()
		}

		defer func() {
			if err := f.Close(); err != nil {
				t.FailNow()
			}
		}()

		if _, err := f.Write(content); err != nil {
			t.FailNow()
		}

		return p
	}

	aContent := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	bContent := []byte{0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F}

	aFilePath := createFile("a.bin", aContent)
	aFileFingerprint, err := GetFileFingerprint(aFilePath)
	assert.Nil(t, err)
	assert.NotEmpty(t, aFileFingerprint)

	bFilePath := createFile("b.bin", bContent)
	bFileFingerprint, err := GetFileFingerprint(bFilePath)
	assert.Nil(t, err)
	assert.NotEmpty(t, bFileFingerprint)

	cFilePath := createFile("c.bin", bContent)
	cFileFingerprint, err := GetFileFingerprint(cFilePath)
	assert.Nil(t, err)
	assert.NotEmpty(t, cFileFingerprint)

	assert.NotEqualValues(t, aFileFingerprint, bFileFingerprint)
	assert.NotEqualValues(t, aFileFingerprint, cFileFingerprint)
	assert.EqualValues(t, bFileFingerprint, cFileFingerprint)
}
