package utils

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlurImageShouldReturnErrorOnNilSource(t *testing.T) {
	originalImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	originalImage.Set(0, 0, color.White)
	originalImage.Set(0, 1, color.White)
	originalImage.Set(1, 0, color.Black)
	originalImage.Set(1, 1, color.Black)

	err := BlurImage(originalImage, nil, 5)

	assert.NotNil(t, err)
}

func TestBlurImageShouldReturnErrorOnNilDestination(t *testing.T) {
	blurredImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	err := BlurImage(nil, blurredImage, 5)

	assert.NotNil(t, err)
}

func TestBlurImageShouldReturnErrorOnSourceDestinationImageSizeMissmatch(t *testing.T) {
	originalImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	originalImage.Set(0, 0, color.White)
	originalImage.Set(0, 1, color.White)
	originalImage.Set(1, 0, color.Black)
	originalImage.Set(1, 1, color.Black)

	blurredImage := image.NewRGBA(image.Rect(0, 0, 3, 3))
	err := BlurImage(originalImage, blurredImage, 5)

	assert.NotNil(t, err)
}

func TestBlurImageShouldReturnErrorOnInvalidParam(t *testing.T) {
	originalImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	originalImage.Set(0, 0, color.White)
	originalImage.Set(0, 1, color.White)
	originalImage.Set(1, 0, color.Black)
	originalImage.Set(1, 1, color.Black)

	blurredImage := image.NewRGBA(originalImage.Rect)
	err := BlurImage(originalImage, blurredImage, 0)

	assert.NotNil(t, err)
}

func TestBlurImageShouldBlurImage(t *testing.T) {
	originalImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	originalImage.Set(0, 0, color.White)
	originalImage.Set(0, 1, color.White)
	originalImage.Set(1, 0, color.Black)
	originalImage.Set(1, 1, color.Black)

	blurredImage := image.NewRGBA(originalImage.Rect)
	err := BlurImage(originalImage, blurredImage, 5)

	assert.Nil(t, err)

	for x := 0; x < originalImage.Rect.Dx(); x += 1 {
		for y := 0; y < originalImage.Rect.Dy(); y += 1 {
			assert.NotEqual(t, originalImage.At(x, y), blurredImage.At(x, y))
		}
	}
}
