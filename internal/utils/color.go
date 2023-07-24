package utils

import (
	"image/color"
	"math"
)

// Convert the provided color represented by the color.Color interface to the color.RGBA struct instance.
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

// Convert a color to grayscale represented as a value from zero to one.
func GetColorGrayscale(c color.Color) float64 {
	rgba := ColorToRgba(c)
	return ((float64(rgba.R) * 0.299) + (float64(rgba.G) * 0.587) + (float64(rgba.B) * 0.114)) / 255.0
}

// Calculate the brightness of the color represented as a value from zero to one.
func GetColorBrightness(c color.Color) float64 {
	rgba := ColorToRgba(c)

	return ((float64(rgba.R) * 0.299) + (float64(rgba.G) * 0.587) + (float64(rgba.B) * 0.114)) / 255.0
}

// Calculate the difference between two colors represented as a value from zero to one using the mean of RGB components difference.
func GetColorDifference(a, b color.Color) float64 {
	aRgba := ColorToRgba(a)
	bRgba := ColorToRgba(b)

	rDiff := math.Abs(float64(aRgba.R) - float64(bRgba.R))
	gDiff := math.Abs(float64(aRgba.G) - float64(bRgba.G))
	bDiff := math.Abs(float64(aRgba.B) - float64(bRgba.B))

	return (rDiff + gDiff + bDiff) / (255.0 * 3.0)
}

// Perform a binary threshold on a given color with specfied cutoff threshold and returna black or white color.
func BinaryThreshold(c color.Color, t float64) color.Color {
	grayscale := GetColorGrayscale(c)

	if grayscale < t {
		return color.Black
	} else {
		return color.White
	}
}
