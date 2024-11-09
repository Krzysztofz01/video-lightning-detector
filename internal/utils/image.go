package utils

import (
	"errors"
	"fmt"
	"image"

	"github.com/esimov/stackblur-go"
	"golang.org/x/image/draw"
)

// Perform a "stackblur" blurring on the source image with a specified radius parameter and store the result to the destination image pointer.
func BlurImage(src image.Image, dst *image.RGBA, radius int) error {
	if src == nil {
		return errors.New("utils: the source image reference is nil")
	}

	if dst == nil {
		return errors.New("utils: the destination image pointer is nil")
	}

	if src.Bounds().Dx() != dst.Bounds().Dx() || src.Bounds().Dy() != dst.Bounds().Dy() {
		return errors.New("utils: source and destination images bounds missmatch")
	}

	imgBlur, err := stackblur.Process(src, uint32(radius))
	if err != nil {
		return fmt.Errorf("utils: external image bluring utility failed: %w", err)
	}

	imgBlurRgba := image.NewRGBA(image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy()))
	draw.Draw(imgBlurRgba, imgBlurRgba.Bounds(), imgBlur, imgBlur.Bounds().Min, draw.Src)

	copy(dst.Pix, imgBlurRgba.Pix)
	return nil
}

// Perform a scaling process by a given factor on the RGBA image provided by the src pointer and store the result to the RGBA image specified by the dst pointer.
func ScaleImage(src, dst *image.RGBA, factor float64) error {
	if src == nil {
		return errors.New("utils: the source image reference is nil")
	}

	if dst == nil {
		return errors.New("utils: the destination image pointer is nil")
	}

	if factor < 0.0 || factor > 1.0 {
		return errors.New("utils: the scaling factor must be between zero and one")
	}

	if dst.Bounds().Dx() != int(float64(src.Bounds().Dx())*factor) || dst.Bounds().Dy() != int(float64(src.Bounds().Dy())*factor) {
		return errors.New("utils: the provided destination image size is not matching the scale factor")
	}

	if factor == 1.0 {
		copy(dst.Pix, src.Pix)
	} else {
		draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	}

	return nil
}

func Otsu(i image.Image) float64 {
	histogram := [256]int{}
	for y := 0; y < i.Bounds().Dy(); y += 1 {
		for x := 0; x < i.Bounds().Dx(); x += 1 {
			gsf := ColorToGrayscale(i.At(x, y))
			gs := int(gsf * 255.0)

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
