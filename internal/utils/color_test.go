package utils

import (
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldConvertColorToGrayscale(t *testing.T) {
	cases := map[color.Color]float64{
		color.RGBA{0, 0, 0, 0}:          0.0,
		color.RGBA{0, 0, 0, 255}:        0.0,
		color.RGBA{255, 255, 255, 0}:    1.0,
		color.RGBA{255, 255, 255, 255}:  1.0,
		color.RGBA{50, 100, 200, 0}:     0.376470588,
		color.RGBA{50, 100, 200, 255}:   0.376470588,
		color.NRGBA{0, 0, 0, 0}:         0,
		color.NRGBA{0, 0, 0, 255}:       0,
		color.NRGBA{255, 255, 255, 0}:   0,
		color.NRGBA{255, 255, 255, 255}: 1.0,
		color.NRGBA{50, 100, 200, 0}:    0,
		color.NRGBA{50, 100, 200, 255}:  0.376470588,
		color.White:                     1.0,
		color.Black:                     0,
	}

	const delta float64 = 1e-2

	for color, expected := range cases {
		actual := ColorToGrayscale(color)

		assert.InDelta(t, expected, actual, delta)
	}
}

func TestShouldConvertColorToRgba(t *testing.T) {
	cases := map[color.Color]color.RGBA{
		color.RGBA{0, 0, 0, 0}:          {0, 0, 0, 0},
		color.RGBA{0, 0, 0, 254}:        {0, 0, 0, 254},
		color.RGBA{0, 0, 0, 255}:        {0, 0, 0, 255},
		color.RGBA{255, 255, 255, 0}:    {255, 255, 255, 0},
		color.RGBA{255, 255, 255, 254}:  {255, 255, 255, 254},
		color.RGBA{255, 255, 255, 255}:  {255, 255, 255, 255},
		color.NRGBA{0, 0, 0, 0}:         {0, 0, 0, 0},
		color.NRGBA{0, 0, 0, 254}:       {0, 0, 0, 254},
		color.NRGBA{0, 0, 0, 255}:       {0, 0, 0, 255},
		color.NRGBA{255, 255, 255, 0}:   {0, 0, 0, 0},
		color.NRGBA{255, 255, 255, 254}: {254, 254, 254, 254},
		color.NRGBA{255, 255, 255, 255}: {255, 255, 255, 255},
		color.Gray{Y: 244}:              {244, 244, 244, 255},
		color.White:                     {255, 255, 255, 255},
		color.Black:                     {0, 0, 0, 255},
	}

	for color, expected := range cases {
		actual := ColorToRgba(color)

		assert.Equal(t, expected, actual)
	}
}

func TestShouldGetColorDifference(t *testing.T) {
	cases := []struct {
		a        color.Color
		b        color.Color
		expected float64
	}{
		{color.Black, color.Black, 0.0},
		{color.Black, color.White, 1.0},
		{color.RGBA{255, 100, 10, 0xff}, color.RGBA{20, 200, 255, 0xff}, 0.758169935},
	}

	const delta float64 = 1e-2

	for _, c := range cases {
		actual := GetColorDifference(c.a, c.b)

		assert.InDelta(t, c.expected, actual, delta)
	}
}

func TestShouldPerformBinaryThreshold(t *testing.T) {
	cases := map[color.Color]color.Color{
		color.White:                     color.White,
		color.Black:                     color.Black,
		color.RGBA{50, 50, 50, 0xff}:    color.Black,
		color.RGBA{180, 180, 180, 0xff}: color.White,
	}

	const threshold float64 = 0.5
	for color, expected := range cases {
		actual := BinaryThreshold(color, threshold)

		assert.Equal(t, expected, actual)
	}
}

func TestShouldGetColorBrightness(t *testing.T) {
	cases := map[color.Color]float64{
		color.Black: 0.0,
		color.White: 1.0,
	}

	const delta float64 = 1e-7

	for color, expected := range cases {
		actual := GetColorBrightness(color)

		assert.InDelta(t, expected, actual, delta)
	}
}

func TestLuminanceRangeCubicRootShouldCalculatePreciseValuesInCorrectRange(t *testing.T) {
	const (
		min        = 16.0 / 116.0
		max        = 1.0
		iterations = 10000
		step       = (max - min) / float64(iterations)
		delta      = 1e-11
	)

	for x := min; x < max; x += step {
		expected := math.Cbrt(x)
		actual := luminanceRangeCubeRoot(x)

		assert.InDelta(t, expected, actual, delta)
	}
}
