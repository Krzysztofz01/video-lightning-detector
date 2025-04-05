package denoise

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidAlgorithmShouldReturnCorrectBoolean(t *testing.T) {
	cases := map[Algorithm]bool{
		NoDenoise:   true,
		StackBlur8:  true,
		StackBlur16: true,
		StackBlur32: true,
		-1:          false,
	}

	for algorithm, expected := range cases {
		actual := IsValidAlgorithm(algorithm)

		assert.Equal(t, expected, actual)
	}
}

func TestDenoiseShouldReturnErrorOnNilSource(t *testing.T) {
	originalImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	originalImage.Set(0, 0, color.White)
	originalImage.Set(0, 1, color.White)
	originalImage.Set(1, 0, color.Black)
	originalImage.Set(1, 1, color.Black)

	for _, a := range algorithms {
		err := Denoise(originalImage, nil, a)

		assert.NotNil(t, err)
	}
}

func TestDenoiseShouldReturnErrorOnNilDestination(t *testing.T) {
	blurredImage := image.NewRGBA(image.Rect(0, 0, 2, 2))

	for _, a := range algorithms {
		err := Denoise(nil, blurredImage, a)

		assert.NotNil(t, err)
	}
}

func TestDenoiseShouldReturnErrorOnSourceDestinationImageSizeMissmatch(t *testing.T) {
	originalImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	originalImage.Set(0, 0, color.White)
	originalImage.Set(0, 1, color.White)
	originalImage.Set(1, 0, color.Black)
	originalImage.Set(1, 1, color.Black)

	blurredImage := image.NewRGBA(image.Rect(0, 0, 3, 3))

	for _, a := range algorithms {
		err := Denoise(originalImage, blurredImage, a)

		assert.NotNil(t, err)
	}
}

func TestDenoiseShouldReturnErrorOnInvalidAlgorithm(t *testing.T) {
	originalImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	originalImage.Set(0, 0, color.White)
	originalImage.Set(0, 1, color.White)
	originalImage.Set(1, 0, color.Black)
	originalImage.Set(1, 1, color.Black)

	blurredImage := image.NewRGBA(originalImage.Rect)

	err := Denoise(originalImage, blurredImage, -1)

	assert.NotNil(t, err)
}

func TestDenoiseShouldDenoiseImage(t *testing.T) {
	originalImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	originalImage.Set(0, 0, color.White)
	originalImage.Set(0, 1, color.White)
	originalImage.Set(1, 0, color.Black)
	originalImage.Set(1, 1, color.Black)

	blurredImage := image.NewRGBA(originalImage.Rect)

	for _, a := range algorithms {
		err := Denoise(originalImage, blurredImage, a)

		assert.Nil(t, err)

		if a == NoDenoise {
			continue
		}

		for x := 0; x < originalImage.Rect.Dx(); x += 1 {
			for y := 0; y < originalImage.Rect.Dy(); y += 1 {
				assert.NotEqual(t, originalImage.At(x, y), blurredImage.At(x, y))
			}
		}
	}
}

var algorithms []Algorithm = []Algorithm{
	NoDenoise,
	StackBlur8,
	StackBlur16,
	StackBlur32,
}
