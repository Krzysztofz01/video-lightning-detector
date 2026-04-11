package utils

import (
	"github.com/stretchr/testify/assert"
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestCopyAsRgbaShouldNotCopyNilImage(t *testing.T) {
	i, err := CopyAsRgba(nil)
	assert.NotNil(t, err)
	assert.Nil(t, i)
}

func TestCopyAsRgbaShouldCopy(t *testing.T) {
	var (
		size   int          = 4
		rect                = image.Rect(0, 0, size, size)
		images []draw.Image = []draw.Image{
			image.NewRGBA(rect),
			image.NewNRGBA(rect),
			image.NewGray(rect),
		}
	)

	for _, i := range images {
		for y := 0; y < i.Bounds().Dy(); y += 1 {
			for x := 0; x < i.Bounds().Dx(); x += 1 {
				i.Set(x, y, color.White)
			}
		}
	}

	for _, i := range images {
		imageRgba, err := CopyAsRgba(i)
		assert.Nil(t, err)
		assert.NotNil(t, imageRgba)

		for y := 0; y < i.Bounds().Dy(); y += 1 {
			for x := 0; x < i.Bounds().Dx(); x += 1 {
				eR, eG, eB, eA := i.At(x, y).RGBA()
				aR, aG, aB, aA := imageRgba.At(x, y).RGBA()

				assert.Equal(t, eR, aR)
				assert.Equal(t, eG, aG)
				assert.Equal(t, eB, aB)
				assert.Equal(t, eA, aA)
			}
		}

		imageRgba.Set(0, 0, color.Black)
		eR, eG, eB, _ := i.At(0, 0).RGBA()
		aR, aG, aB, _ := imageRgba.At(0, 0).RGBA()

		assert.NotEqual(t, eR, aR)
		assert.NotEqual(t, eG, aG)
		assert.NotEqual(t, eB, aB)
	}
}
