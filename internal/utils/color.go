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
//
// Deprecated: Could be removed due to the refactor that lead to usage of standalone uint8 as the RGB representation
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

// Convert a RGB color to grayscale that will be represented as a value from zero to one.
func ColorToGrayscale(r, g, b uint8) float64 {
	return ((float64(r) * 0.299) + (float64(g) * 0.587) + (float64(b) * 0.114)) / 255.0
}

// Calculate the brightness of the RGB color that will be represented as a value from zero to one.
func GetColorBrightness(r, g, b uint8) float64 {
	lR := linearRgbComponentLookup[r]
	lG := linearRgbComponentLookup[g]
	lB := linearRgbComponentLookup[b]

	luminance := 0.2126*lR + 0.7152*lG + 0.0722*lB
	if luminance <= 0.008856 {
		return (luminance * 903.3) / 100.0
	} else {
		return (math.Pow(luminance, 1.0/3.0)*116.0 - 16.0) / 100.0
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

// Calculate the difference between two RGB colors that will be represented as a value from zero to one using the mean of RGB components difference.
func GetColorDifference(aR, aG, aB, bR, bG, bB uint8) float64 {
	rDiff := math.Abs(float64(aR) - float64(bR))
	gDiff := math.Abs(float64(aG) - float64(bG))
	bDiff := math.Abs(float64(aB) - float64(bB))

	return (rDiff + gDiff + bDiff) / (255.0 * 3.0)
}

// Perform a binary threshold on a given RGB color with specfied cutoff threshold and return a uint8 represented black or white color.
func BinaryThreshold(r, g, b uint8, t float64) uint8 {
	if ColorToGrayscale(r, g, b) < t {
		return 0x00
	} else {
		return 0xff
	}
}
