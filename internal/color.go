package internal

import (
	"image/color"
	"math"
)

func ColorToRgba(c color.Color) color.RGBA {
	if rgba, ok := c.(color.RGBA); ok {
		return rgba
	}

	r32, g32, b32, a32 := c.RGBA()
	return color.RGBA{
		R: uint8(r32 >> 8),
		G: uint8(g32 >> 8),
		B: uint8(b32 >> 8),
		A: uint8(a32 >> 8),
	}
}

func GetGrayscaleBasedBrightness(c color.Color) float64 {
	rgba := ColorToRgba(c)

	return ((float64(rgba.R) * 0.299) + (float64(rgba.G) * 0.587) + (float64(rgba.B) * 0.114)) / 255.0
}

func GetColorDifference(a, b color.Color) float64 {
	aRgba := ColorToRgba(a)
	bRgba := ColorToRgba(b)

	rDiff := math.Abs(float64(aRgba.R) - float64(bRgba.R))
	gDiff := math.Abs(float64(aRgba.G) - float64(bRgba.G))
	bDiff := math.Abs(float64(aRgba.B) - float64(bRgba.B))

	return (rDiff + gDiff + bDiff) / (255.0 * 3.0)
}
