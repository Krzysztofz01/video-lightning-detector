package utils

import (
	"image"
)

func Otsu(i image.RGBA) float64 {
	var (
		histogram [256]int = [256]int{}
		width     int      = i.Bounds().Dx()
		offset    int
		r, g, b   byte
		gsf       float64
		gs        int
	)

	for y := 0; y < i.Bounds().Dy(); y += 1 {
		for x := 0; x < i.Bounds().Dx(); x += 1 {
			offset = 4*y*width + 4*x

			r = i.Pix[offset+0]
			g = i.Pix[offset+1]
			b = i.Pix[offset+2]

			gsf = ColorToGrayscale(r, g, b)
			gs = int(gsf * 255.0)

			histogram[gs] += 1
		}
	}

	var (
		size             int     = i.Bounds().Dx() * i.Bounds().Dy()
		histogramSum     float64 = 0.0
		backgroundSum    float64 = 0.0
		backgroundWeight int     = 0
		foregroundWeight int     = 0
		maxVariance      float64 = 0.0
		threshold        float64 = 0.0
	)

	for i, bin := range histogram {
		histogramSum += float64(i * bin)
	}

	for i, bin := range histogram {
		backgroundWeight += bin
		if backgroundWeight == 0 {
			continue
		}

		foregroundWeight = size - backgroundWeight
		if foregroundWeight == 0 {
			break
		}

		backgroundSum += float64(i * bin)

		backgroundMean := backgroundSum / float64(backgroundWeight)
		foregroundMean := (histogramSum - backgroundSum) / float64(foregroundWeight)
		meanDiff := backgroundMean - foregroundMean

		variance := meanDiff * meanDiff * float64(backgroundWeight) * float64(foregroundWeight)

		if variance > maxVariance {
			maxVariance = variance
			threshold = float64(i) / 255.0
		}
	}

	return threshold
}
