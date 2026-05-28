package utils

import (
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldConvertColorToGrayscale(t *testing.T) {
	cases := map[color.RGBA]float64{
		{0, 0, 0, 0}:         0.0,
		{0, 0, 0, 255}:       0.0,
		{255, 255, 255, 0}:   1.0,
		{255, 255, 255, 255}: 1.0,
		{50, 100, 200, 0}:    0.376470588,
		{50, 100, 200, 255}:  0.376470588,
	}

	const delta float64 = 1e-2

	for c, expected := range cases {
		actual := ColorToGrayscale(c.R, c.G, c.B)

		assert.InDelta(t, expected, actual, delta)
	}
}

func TestShouldGetColorDifference(t *testing.T) {
	cases := []struct {
		a        color.RGBA
		b        color.RGBA
		expected float64
	}{
		{color.RGBA{0x00, 0x00, 0x00, 0xff}, color.RGBA{0x00, 0x00, 0x00, 0xff}, 0.0},
		{color.RGBA{0x00, 0x00, 0x00, 0xff}, color.RGBA{0xff, 0xff, 0xff, 0xff}, 1.0},
		{color.RGBA{255, 100, 10, 0xff}, color.RGBA{20, 200, 255, 0xff}, 0.758169935},
	}

	const delta float64 = 1e-2

	for _, c := range cases {
		actual := GetColorDifference(c.a.R, c.a.G, c.a.B, c.b.R, c.b.G, c.b.B)

		assert.InDelta(t, c.expected, actual, delta)
	}
}

func TestShouldPerformBinaryThreshold(t *testing.T) {
	cases := map[color.RGBA]uint8{
		{0x00, 0x00, 0x00, 0xff}: 0x00,
		{0xff, 0xff, 0xff, 0xff}: 0xff,
		{50, 50, 50, 0xff}:       0x00,
		{180, 180, 180, 0xff}:    0xff,
	}

	const threshold float64 = 0.5
	for c, expected := range cases {
		actual := BinaryThreshold(c.R, c.G, c.B, threshold)

		assert.Equal(t, expected, actual)
	}
}

func TestShouldGetColorBrightness(t *testing.T) {
	cases := map[color.RGBA]float64{
		{0x00, 0x00, 0x00, 0xff}: 0.0,
		{0xff, 0xff, 0xff, 0xff}: 1.0,
		{0x28, 0x4d, 0x27, 0xff}: 0.29172414,
		{0xde, 0xe1, 0x8d, 0xff}: 0.87632063,
		{0x4a, 0xcd, 0x1f, 0xff}: 0.73034825,
		{0xd0, 0xb0, 0x60, 0xff}: 0.73092332,
		{0x22, 0x20, 0x74, 0xff}: 0.18514360,
		{0xa6, 0xb5, 0x85, 0xff}: 0.71451814,
		{0xba, 0x97, 0xe6, 0xff}: 0.68230503,
		{0x88, 0xa8, 0x27, 0xff}: 0.64472235,
		{0xc5, 0x37, 0x50, 0xff}: 0.45882309,
		{0xeb, 0x95, 0x34, 0xff}: 0.69043620,
		{0x2f, 0xfe, 0xb6, 0xff}: 0.89329888,
		{0x0a, 0x45, 0xf6, 0xff}: 0.39537648,
		{0x02, 0x84, 0x97, 0xff}: 0.50393189,
		{0x99, 0x93, 0xa9, 0xff}: 0.62086941,
		{0x7f, 0x9d, 0x55, 0xff}: 0.61029568,
		{0x02, 0xf5, 0xb7, 0xff}: 0.86373459,
		{0x63, 0x0c, 0x8b, 0xff}: 0.26097604,
		{0x97, 0x11, 0x60, 0xff}: 0.33615176,
		{0x4d, 0x2c, 0x3a, 0xff}: 0.22600186,
		{0xb8, 0x32, 0x9c, 0xff}: 0.45458176,
	}

	const delta float64 = 1e-7

	for c, expected := range cases {
		actual := GetColorBrightness(c.R, c.G, c.B)

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
