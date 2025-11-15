package detector

import (
	"fmt"
	"image"

	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

type FrameStrikeDetector interface {
	// Get a set of two slices with length specified via resolution which values incidate the amount of pixels after
	// the binary-thresholiding (with the specified threshold) as a mean. The image is divided into
	// resolution parts horizontally for the first slice and vertically for the second one.
	GetDetectionPlot(frame *image.RGBA) ([2][]float64, error)
}

type frameStrikeDetector struct {
	Resolution int
	Threshold  float64
	Bbox       bool
	BboxAnchor utils.Vec2i
	BboxDim    utils.Vec2i
	FrameDim   utils.Vec2i
}

func (d *frameStrikeDetector) GetDetectionPlot(frame *image.RGBA) ([2][]float64, error) {
	if frame == nil {
		return [2][]float64{}, fmt.Errorf("detector: the specified frame image reference is nil")
	}

	if d.Bbox {
		if frame.Bounds().Dx() != d.BboxDim.X || frame.Bounds().Dy() != d.BboxDim.Y {
			return [2][]float64{}, fmt.Errorf("detector: the specified frame image bounds are not matching the expected values")
		}
	} else {
		if frame.Bounds().Dx() != d.FrameDim.X || frame.Bounds().Dy() != d.FrameDim.Y {
			return [2][]float64{}, fmt.Errorf("detector: the specified frame image bounds are not matching the expected values")
		}
	}

	var (
		wf float64 = float64(d.FrameDim.X)
		hf float64 = float64(d.FrameDim.Y)
		rf float64 = float64(d.Resolution)
	)

	var (
		hVal []float64 = make([]float64, d.Resolution)
		hSum []float64 = make([]float64, d.Resolution)
		vVal []float64 = make([]float64, d.Resolution)
		vSum []float64 = make([]float64, d.Resolution)
	)

	var (
		xf, yf   float64
		hhi, vhi int
		offset   int
		r, g, b  uint8
	)

	for y := 0; y < d.FrameDim.Y; y += 1 {
		for x := 0; x < d.FrameDim.X; x += 1 {
			yf = float64(y)
			xf = float64(x)

			hhi = int(xf / wf * rf)
			vhi = int(yf / hf * rf)

			hSum[hhi] += 1
			vSum[vhi] += 1

			if d.Bbox {
				if x >= d.BboxAnchor.X && x < d.BboxAnchor.X+d.BboxDim.X && y >= d.BboxAnchor.Y && y < d.BboxAnchor.Y+d.BboxDim.Y {
					offset = 4*(y-d.BboxAnchor.Y)*d.BboxDim.X + 4*(x-d.BboxAnchor.X)
					r = frame.Pix[offset+0]
					g = frame.Pix[offset+1]
					b = frame.Pix[offset+2]
				} else {
					continue
				}
			} else {
				offset = 4*y*d.FrameDim.X + 4*x
				r = frame.Pix[offset+0]
				g = frame.Pix[offset+1]
				b = frame.Pix[offset+2]
			}

			if utils.BinaryThreshold(r, g, b, d.Threshold) == 0xff {
				hVal[hhi] += 1
				vVal[vhi] += 1
			} else {
				continue
			}
		}
	}

	for index := 0; index < d.Resolution; index += 1 {
		hVal[index] /= hSum[index]
		vVal[index] /= vSum[index]
	}

	return [2][]float64{
		hVal,
		vVal,
	}, nil
}

func CreateFrameStrikeDetector(fullFrameWidth, fullFrameHeight int, options options.StreamDetectorOptions) (FrameStrikeDetector, error) {
	if fullFrameWidth <= 0 || fullFrameHeight <= 0 {
		return nil, fmt.Errorf("detector: full frame dimensions must be greater than zero")
	}

	if options.FrameDetectionPlotResolution <= 0 {
		return nil, fmt.Errorf("detector: the resolution must be greater than zero")
	}

	if options.FrameDetectionPlotThreshold < 0.0 || options.FrameDetectionPlotThreshold > 1.0 {
		return nil, fmt.Errorf("detector: the threshold for binary-threshold must be between zero and one")
	}

	var (
		bbox       bool        = false
		bboxAnchor utils.Vec2i = utils.Vec2i{}
		bboxDim    utils.Vec2i = utils.Vec2i{}
	)

	if len(options.DetectionBoundsExpression) > 0 {
		x, y, w, h, err := utils.ParseBoundsExpression(options.DetectionBoundsExpression)
		if err != nil {
			return nil, fmt.Errorf("detector: failed to parse the bbox frame dimensions bounds expression: %w", err)
		}

		if x+w > fullFrameWidth || y+h > fullFrameHeight {
			return nil, fmt.Errorf("detector: the specified  bounds exceed the full frame dimensions")
		}

		bbox = true
		bboxAnchor = utils.Vec2i{X: x, Y: y}
		bboxDim = utils.Vec2i{X: w, Y: h}
	}

	return &frameStrikeDetector{
		Resolution: options.FrameDetectionPlotResolution,
		Threshold:  options.FrameDetectionPlotThreshold,
		Bbox:       bbox,
		BboxAnchor: bboxAnchor,
		BboxDim:    bboxDim,
		FrameDim:   utils.Vec2i{X: fullFrameWidth, Y: fullFrameHeight},
	}, nil
}
