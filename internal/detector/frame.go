package detector

import (
	"fmt"
	"image"

	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

// Get a set of two slices with length specified via resolution which values incidate the amount of pixels after
// the binary-thresholiding (with threshold specified by the provided value) as a mean. The image is divided into
// resolution parts horizontally for the first slice and vertically for the second one.
func getFrameStrikePlot(frame *image.RGBA, resolution int, threshold float64) ([2][]float64, error) {
	if frame == nil {
		return [2][]float64{}, fmt.Errorf("detector: the specified frame image reference is nil")
	}

	if resolution <= 0 {
		return [2][]float64{}, fmt.Errorf("detector: the resolution must be greater than zero")
	}

	if threshold < 0.0 || threshold > 1.0 {
		return [2][]float64{}, fmt.Errorf("detector: the threshold for binary-threshold must be between zero and one")
	}

	var (
		w    int       = frame.Bounds().Dx()
		wf   float64   = float64(frame.Bounds().Dx())
		hf   float64   = float64(frame.Bounds().Dy())
		rf   float64   = float64(resolution)
		hVal []float64 = make([]float64, resolution)
		hSum []float64 = make([]float64, resolution)
		vVal []float64 = make([]float64, resolution)
		vSum []float64 = make([]float64, resolution)
	)

	var (
		x, y     float64
		hhi, vhi int
		r, g, b  uint8
	)

	for index := 0; index < len(frame.Pix); index += 4 {
		y = float64((index / 4) / w)
		x = float64((index / 4) % w)

		hhi = int(x / wf * rf)
		vhi = int(y / hf * rf)

		hSum[hhi] += 1
		vSum[vhi] += 1

		r = frame.Pix[index+0]
		g = frame.Pix[index+1]
		b = frame.Pix[index+2]

		if utils.BinaryThreshold(r, g, b, threshold) == 0x00 {
			continue
		}

		hVal[hhi] += 1
		vVal[vhi] += 1

	}

	for index := 0; index < resolution; index += 1 {
		hVal[index] /= hSum[index]
		vVal[index] /= vSum[index]
	}

	return [2][]float64{
		hVal,
		vVal,
	}, nil
}
