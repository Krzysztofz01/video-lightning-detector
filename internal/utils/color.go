package utils

import (
	"image/color"
	"math"
)

var linearRgbComponentLookup [256]float64

func init() {
	for x := 0; x < len(linearRgbComponentLookup); x += 1 {
		xNorm := float64(x) / 255.0

		if xNorm <= 0.04045 {
			linearRgbComponentLookup[x] = xNorm / 12.92
		} else {
			linearRgbComponentLookup[x] = math.Pow((xNorm+0.055)/1.055, 2.4)
		}
	}
}

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
func ColorToGrayscale(c color.Color) float64 {
	rgba := ColorToRgba(c)
	return ((float64(rgba.R) * 0.299) + (float64(rgba.G) * 0.587) + (float64(rgba.B) * 0.114)) / 255.0
}

// Calculate the brightness of the color represented as a value from zero to one.
func GetColorBrightness(c color.Color) float64 {
	rgba := ColorToRgba(c)
	lR := linearRgbComponentLookup[rgba.R]
	lG := linearRgbComponentLookup[rgba.G]
	lB := linearRgbComponentLookup[rgba.B]

	luminance := 0.2126*lR + 0.7152*lG + 0.0722*lB
	if luminance <= 0.008856 {
		return (luminance * 903.3) / 100.0
	} else {
		return (math.Pow(luminance, 1.0/3.0)*116.0 - 16.0) / 100.0
	}
}

func GetColorBrightnessApprox(c color.Color) float64 {
	rgba := ColorToRgba(c)
	lR := linearRgbComponentLookup[rgba.R]
	lG := linearRgbComponentLookup[rgba.G]
	lB := linearRgbComponentLookup[rgba.B]

	luminance := 0.2126*lR + 0.7152*lG + 0.0722*lB
	if luminance <= 0.008856 {
		return (luminance * 903.3) / 100.0
	} else {
		return (luminanceRangeCubeRoot(luminance)*116.0 - 16.0) / 100.0
	}
}

func luminanceRangeCubeRoot(x float64) float64 {
	reg := (-0.358955950652834 * x * x) + (0.934309346877746 * x) + 0.414814427166639

	for i := 0; i < 3; i += 1 {
		regp2 := reg * reg
		reg = reg - ((regp2*reg)-x)/(3*regp2)
	}

	return reg
}

// Calculate the difference between two colors represented as a value frocwdm zero to one using the mean of RGB components difference.
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
	grayscale := ColorToGrayscale(c)

	if grayscale < t {
		return color.Black
	} else {
		return color.White
	}
}
