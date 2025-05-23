package detector

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStrikePlotShouldReturnErrorForInvalidParams(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	_, err := getFrameStrikePlot(nil, 2, 2)
	assert.NotNil(t, err)

	_, err = getFrameStrikePlot(img, -1, 0.5)
	assert.NotNil(t, err)

	_, err = getFrameStrikePlot(img, 0, 0.5)
	assert.NotNil(t, err)

	_, err = getFrameStrikePlot(img, 2, -1)
	assert.NotNil(t, err)

	_, err = getFrameStrikePlot(img, 2, 2)
	assert.NotNil(t, err)
}

func TestGetStrikePlotShouldReturnCorrectPlotDataForValidParams(t *testing.T) {
	const (
		size       int     = 12
		resolution int     = 6
		threshold  float64 = 225.0 / 255.0
	)

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var c color.Color
			if x < size/2 {
				c = color.Black
			} else {
				c = color.White
			}

			img.Set(x, y, c)
		}
	}

	var (
		expectedHorizontal []float64 = []float64{0, 0, 0, 1, 1, 1}
		expectedVertical   []float64 = []float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5}
	)

	plots, err := getFrameStrikePlot(img, resolution, threshold)
	assert.Nil(t, err)
	assert.Len(t, plots, 2)
	assert.Equal(t, expectedHorizontal, plots[0])
	assert.Equal(t, expectedVertical, plots[1])
}
