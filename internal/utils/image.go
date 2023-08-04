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
