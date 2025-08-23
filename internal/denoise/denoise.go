package denoise

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/esimov/stackblur-go"

	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
)

// TODO: Research and implement bilateral filter

// Denoise the src image using the specified denoise algorithm (blurs and low-pass filters) and store the result to the dst image pointer.
// This function is not thread-safe.
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
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)

	return nil
}

var dstNrgbaBuffer *image.NRGBA = nil

func stackBlurDenoise(src image.Image, dst *image.RGBA, radius int) error {
	if dstNrgbaBuffer == nil || dstNrgbaBuffer.Bounds() != dst.Bounds() {
		dstNrgbaBuffer = image.NewNRGBA(dst.Bounds())
	}

	if err := stackblur.Process(dstNrgbaBuffer, src, uint32(radius)); err != nil {
		return fmt.Errorf("denoise: external image bluring utility failed: %w", err)
	}

	draw.Draw(dst, dst.Bounds(), dstNrgbaBuffer, dstNrgbaBuffer.Bounds().Min, draw.Src)

	return nil
}
