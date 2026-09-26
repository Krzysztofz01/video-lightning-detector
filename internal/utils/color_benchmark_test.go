package utils

import (
	"math"
	"testing"
)

var result float64

var colorToGrayscaleTestColors = [][3]uint8{
	{0, 0, 0},
	{255, 255, 255},
	{255, 0, 0},
	{0, 255, 0},
	{0, 0, 255},
	{255, 255, 0},
	{255, 0, 255},
	{0, 255, 255},
	{128, 128, 128},
	{64, 128, 192},
	{10, 20, 30},
	{50, 100, 150},
	{123, 45, 200},
	{17, 234, 91},
	{200, 150, 75},
}

func colorToGrayscaleNaive(x1, x2, x3 uint8) float64 {
	return ((float64(x1) * 0.299) +
		(float64(x2) * 0.587) +
		(float64(x3) * 0.114)) / 255.0
}

func BenchmarkColorToGrayscale(b *testing.B) {
	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := colorToGrayscaleTestColors[i%len(colorToGrayscaleTestColors)]
			result = ColorToGrayscale(c[0], c[1], c[2])
		}
	})

	b.Run("Naive", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := colorToGrayscaleTestColors[i%len(colorToGrayscaleTestColors)]
			result = colorToGrayscaleNaive(c[0], c[1], c[2])
		}
	})
}

var colorDifferenceTestColors = [][6]uint8{
	{0, 0, 0, 0, 0, 0},
	{255, 255, 255, 0, 0, 0},
	{255, 0, 0, 0, 255, 0},
	{0, 255, 0, 0, 0, 255},
	{0, 0, 255, 255, 255, 0},
	{128, 128, 128, 64, 64, 64},
	{10, 20, 30, 40, 50, 60},
	{123, 45, 200, 17, 234, 91},
	{200, 150, 75, 25, 100, 220},
	{255, 128, 64, 32, 16, 8},
}

func getColorDifferenceNaive(aR, aG, aB, bR, bG, bB uint8) float64 {
	rDiff := math.Abs(float64(aR) - float64(bR))
	gDiff := math.Abs(float64(aG) - float64(bG))
	bDiff := math.Abs(float64(aB) - float64(bB))

	return (rDiff + gDiff + bDiff) / (255.0 * 3.0)
}

func BenchmarkColorDifference(b *testing.B) {
	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := colorDifferenceTestColors[i%len(colorDifferenceTestColors)]

			result = GetColorDifference(
				c[0], c[1], c[2],
				c[3], c[4], c[5],
			)
		}
	})

	b.Run("Naive", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := colorDifferenceTestColors[i%len(colorDifferenceTestColors)]

			result = getColorDifferenceNaive(
				c[0], c[1], c[2],
				c[3], c[4], c[5],
			)
		}
	})
}

var colorBrightnessTestColors = [][3]uint8{
	{0, 0, 0},
	{255, 255, 255},
	{255, 0, 0},
	{0, 255, 0},
	{0, 0, 255},
	{255, 255, 0},
	{255, 0, 255},
	{0, 255, 255},
	{128, 128, 128},
	{64, 64, 64},
	{192, 192, 192},
	{10, 20, 30},
	{50, 100, 150},
	{123, 45, 200},
	{17, 234, 91},
	{200, 150, 75},
}

func BenchmarkGetColorBrightness(b *testing.B) {
	b.Run("NaiveWithOptimizations", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := colorBrightnessTestColors[i%len(colorBrightnessTestColors)]

			result = GetColorBrightness(
				c[0], c[1], c[2],
			)
		}
	})

	b.Run("Approximation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := colorBrightnessTestColors[i%len(colorBrightnessTestColors)]

			result = GetColorBrightnessApprox(
				c[0], c[1], c[2],
			)
		}
	})
}
