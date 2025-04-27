package denoise

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/esimov/stackblur-go"
)

// TODO: Research and implement bilateral filter

func Denoise(src image.Image, dst *image.RGBA, a options.DenoiseAlgorithm) error {
	if src == nil {
		return fmt.Errorf("denoise: the source image reference is nil")
	}

	if dst == nil {
		return fmt.Errorf("denoise: the destination image pointer is nil")
	}

	if src.Bounds().Dx() != dst.Bounds().Dx() || src.Bounds().Dy() != dst.Bounds().Dy() {
		return fmt.Errorf("denoise: source and destination images bounds missmatch")
	}

	switch a {
	case options.NoDenoise:
		return noDenoise(src, dst)
	case options.StackBlur8:
		return stackBlurDenoise(src, dst, 8)
	case options.StackBlur16:
		return stackBlurDenoise(src, dst, 16)
	case options.StackBlur32:
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
