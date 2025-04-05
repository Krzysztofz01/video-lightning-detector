package denoise

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/esimov/stackblur-go"
)

func IsValidAlgorithm(s Algorithm) bool {
	switch s {
	case NoDenoise, StackBlur8, StackBlur16, StackBlur32:
		return true
	default:
		return false
	}
}

func Denoise(src image.Image, dst *image.RGBA, s Algorithm) error {
	if src == nil {
		return fmt.Errorf("denoise: the source image reference is nil")
	}

	if dst == nil {
		return fmt.Errorf("denoise: the destination image pointer is nil")
	}

	if src.Bounds().Dx() != dst.Bounds().Dx() || src.Bounds().Dy() != dst.Bounds().Dy() {
		return fmt.Errorf("denoise: source and destination images bounds missmatch")
	}

	switch s {
	case NoDenoise:
		return noDenoise(src, dst)
	case StackBlur8:
		return stackBlurDenoise(src, dst, 8)
	case StackBlur16:
		return stackBlurDenoise(src, dst, 16)
	case StackBlur32:
		return stackBlurDenoise(src, dst, 32)
	default:
		return fmt.Errorf("denoise: invalid denoise strategy specified")
	}
}

func noDenoise(src image.Image, dst *image.RGBA) error {
	srcRGBA := image.NewRGBA(image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy()))
	draw.Draw(srcRGBA, srcRGBA.Bounds(), src, src.Bounds().Min, draw.Src)

	copy(dst.Pix, srcRGBA.Pix)
	return nil
}

func stackBlurDenoise(src image.Image, dst *image.RGBA, radius int) error {
	imgBlur, err := stackblur.Process(src, uint32(radius))
	if err != nil {
		return fmt.Errorf("utils: external image bluring utility failed: %w", err)
	}

	imgBlurRgba := image.NewRGBA(image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy()))
	draw.Draw(imgBlurRgba, imgBlurRgba.Bounds(), imgBlur, imgBlur.Bounds().Min, draw.Src)

	copy(dst.Pix, imgBlurRgba.Pix)
	return nil
}
