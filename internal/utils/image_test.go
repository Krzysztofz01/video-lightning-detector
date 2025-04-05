package utils

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScaleShouldReturnErrorForNilSource(t *testing.T) {
	destinationImage := image.NewRGBA(image.Rect(0, 0, 1, 1))
	destinationImage.Set(0, 0, color.Black)

	err := ScaleImage(nil, destinationImage, 0.5)
	assert.NotNil(t, err)
}

func TestScaleShouldReturnErrorForNilDestination(t *testing.T) {
	sourceImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	sourceImage.Set(0, 0, color.White)
	sourceImage.Set(0, 1, color.White)
	sourceImage.Set(1, 0, color.Black)
	sourceImage.Set(1, 1, color.Black)

	err := ScaleImage(sourceImage, nil, 0.5)
	assert.NotNil(t, err)
}

func TestScaleShouldReturnErrorForInvalidScaleFactor(t *testing.T) {
	sourceImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	sourceImage.Set(0, 0, color.White)
	sourceImage.Set(0, 1, color.White)
	sourceImage.Set(1, 0, color.Black)
	sourceImage.Set(1, 1, color.Black)

	destinationImage := image.NewRGBA(image.Rect(0, 0, 1, 1))
	destinationImage.Set(0, 0, color.Black)

	err := ScaleImage(sourceImage, destinationImage, -0.5)
	assert.NotNil(t, err)
}

func TestScaleShouldReturnErrorForImageSizeScaleFactorMissmatch(t *testing.T) {
	sourceImage := image.NewRGBA(image.Rect(0, 0, 4, 4))
	sourceImage.Set(0, 0, color.White)
	sourceImage.Set(0, 1, color.White)
	sourceImage.Set(1, 0, color.Black)
	sourceImage.Set(1, 1, color.Black)

	destinationImage := image.NewRGBA(image.Rect(0, 0, 1, 1))
	destinationImage.Set(0, 0, color.Black)

	err := ScaleImage(sourceImage, destinationImage, 0.5)
	assert.NotNil(t, err)
}

func TestScaleShouldScaleGivenImage(t *testing.T) {
	sourceImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	sourceImage.Set(0, 0, color.White)
	sourceImage.Set(0, 1, color.White)
	sourceImage.Set(1, 0, color.Black)
	sourceImage.Set(1, 1, color.Black)

	destinationImage := image.NewRGBA(image.Rect(0, 0, 1, 1))
	destinationImage.Set(0, 0, color.Black)

	err := ScaleImage(sourceImage, destinationImage, 0.5)
	assert.Nil(t, err)
}

func TestScaleShouldScaleGivenImageWithSameSize(t *testing.T) {
	sourceImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	sourceImage.Set(0, 0, color.White)
	sourceImage.Set(0, 1, color.White)
	sourceImage.Set(1, 0, color.Black)
	sourceImage.Set(1, 1, color.Black)

	destinationImage := image.NewRGBA(image.Rect(0, 0, 2, 2))

	err := ScaleImage(sourceImage, destinationImage, 1.0)
	assert.Nil(t, err)
}
