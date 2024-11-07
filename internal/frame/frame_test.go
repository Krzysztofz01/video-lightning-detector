package frame

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldCreateNewFirstFrame(t *testing.T) {
	a := mockImage(color.White)
	b := mockImage(color.Black)

	frame := CreateNewFrame(a, b, 1, BinaryThresholdParam)

	assert.NotNil(t, frame)
	assert.Equal(t, 1.0, frame.Brightness)
	assert.Equal(t, 0.0, frame.ColorDifference)
	assert.Equal(t, 0.0, frame.BinaryThresholdDifference)
}

func TestShouldCreateNewFrameWithDifferentNeighbour(t *testing.T) {
	a := mockImage(color.White)
	b := mockImage(color.Black)

	frame := CreateNewFrame(a, b, 2, BinaryThresholdParam)

	assert.NotNil(t, frame)
	assert.Equal(t, 1.0, frame.Brightness)
	assert.Equal(t, 1.0, frame.ColorDifference)
	assert.Equal(t, 1.0, frame.BinaryThresholdDifference)
}

func TestShouldCreateNewFrameWithIdenticalNeighbour(t *testing.T) {
	a := mockImage(color.White)
	b := mockImage(color.White)

	frame := CreateNewFrame(a, b, 2, BinaryThresholdParam)

	assert.NotNil(t, frame)
	assert.Equal(t, 1.0, frame.Brightness)
	assert.Equal(t, 0.0, frame.ColorDifference)
	assert.Equal(t, 0.0, frame.BinaryThresholdDifference)
}

func mockImage(c color.Color) image.Image {
	width := 4
	height := 4

	image := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x += 1 {
		for y := 0; y < height; y += 1 {
			image.Set(x, y, c)
		}
	}

	return image
}
