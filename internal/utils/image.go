package utils

import (
	"fmt"
	"image"
	"image/draw"
)

// Create a RGBA copy of the image represented with a image.Image interface
func CopyAsRgba(i image.Image) (*image.RGBA, error) {
	if i == nil {
		return nil, fmt.Errorf("utils: invalid image reference provided")
	}

	rgba := image.NewRGBA(i.Bounds())
	draw.Draw(rgba, rgba.Bounds(), i, i.Bounds().Min, draw.Src)

	return rgba, nil
}
