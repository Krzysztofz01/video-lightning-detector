package denoise

import (
	"github.com/stretchr/testify/assert"
	"image"
	"image/color"
	"testing"

	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
)

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
	originalImage := image.NewRGBA(image.Rect(0, 0, 3, 3))
	originalImage.Set(0, 0, color.RGBA{123, 45, 67, 0xff})
	originalImage.Set(1, 0, color.RGBA{200, 34, 89, 0xff})
	originalImage.Set(2, 0, color.RGBA{12, 240, 78, 0xff})
	originalImage.Set(0, 1, color.RGBA{90, 12, 230, 0xff})
	originalImage.Set(1, 1, color.RGBA{45, 200, 134, 0xff})
	originalImage.Set(2, 1, color.RGBA{250, 100, 50, 0xff})
	originalImage.Set(0, 2, color.RGBA{10, 180, 210, 0xff})
	originalImage.Set(1, 2, color.RGBA{255, 220, 100, 0xff})
	originalImage.Set(2, 2, color.RGBA{75, 30, 190, 0xff})

	blurredImage := image.NewRGBA(originalImage.Rect)

	for _, a := range algorithms {
		err := Denoise(originalImage, blurredImage, a)
		assert.Nil(t, err)

		for x := 0; x < originalImage.Rect.Dx(); x += 1 {
			for y := 0; y < originalImage.Rect.Dy(); y += 1 {
				oR, oG, oB, _ := originalImage.At(x, y).RGBA()
				bR, bG, bB, _ := blurredImage.At(x, y).RGBA()

				if a == options.NoDenoise {
					assert.Equal(t, oR, bR)
					assert.Equal(t, oG, bG)
					assert.Equal(t, oB, bB)
				} else {
					assert.NotEqual(t, oR, bR)
					assert.NotEqual(t, oG, bG)
					assert.NotEqual(t, oB, bB)
				}
			}
		}
	}
}

var algorithms []options.DenoiseAlgorithm = []options.DenoiseAlgorithm{
	options.NoDenoise,
	options.StackBlur8,
	options.StackBlur16,
	options.StackBlur32,
}
