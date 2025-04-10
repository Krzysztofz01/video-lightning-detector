package frame

import (
	"image"

	"github.com/Krzysztofz01/pimit"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
	"go.uber.org/atomic"
)

// TODO: Apporach to the binary threshold segmentation:
//       - Implementing different thresholds for day and night recordings where the birghtness would be the determinant
//       - Currently there is a comparionsing between the BT result of current na previous frame to calculate white_pixels / all_pixels,
//         alternatively just the normalized count of the occureance of white pixels could be returned as the result
//       - determine the parameter for the binary threshold via "the half of the histogram" (Otsu)

const (
	BinaryThresholdParam float64 = 200.0 / 255.0
)

// Strucutre representing a single video frame and its calculated parameters.
type Frame struct {
	OrdinalNumber             int     `json:"ordinal-number"`
	ColorDifference           float64 `json:"color-difference"`
	BinaryThresholdDifference float64 `json:"binary-threshold-difference"`
	Brightness                float64 `json:"brightness"`
}

// Create a new frame instance by providing the current and previous frame images and the ordinal number (1 indexed) of the frame.
func CreateNewFrame(currentFrame, previousFrame *image.RGBA, ordinalNumber int, binaryThresholdParam float64) *Frame {
	var (
		brightness    *atomic.Float64 = atomic.NewFloat64(0.0)
		cDifference   *atomic.Float64 = atomic.NewFloat64(0)
		btcDifference *atomic.Int32   = atomic.NewInt32(0)
	)

	width := previousFrame.Bounds().Dx()

	pimit.ParallelRgbaRead(currentFrame, func(x, y int, cR, cG, cB, _ uint8) {
		brightness.Add(utils.GetColorBrightness(cR, cG, cB))

		if ordinalNumber <= 1 {
			return
		}

		index := 4 * (y*width + x)

		pR := previousFrame.Pix[index+0]
		pG := previousFrame.Pix[index+1]
		pB := previousFrame.Pix[index+2]

		cDifference.Add(utils.GetColorDifference(cR, cG, cB, pR, pG, pB))

		thresholdCurrent := utils.BinaryThreshold(cR, cG, cB, binaryThresholdParam)
		thresholdPrevious := utils.BinaryThreshold(pR, pG, pB, binaryThresholdParam)
		if thresholdCurrent != thresholdPrevious {
			btcDifference.Add(1)
		}
	})

	frameSize := float64(currentFrame.Bounds().Dx() * currentFrame.Bounds().Dy())

	return &Frame{
		OrdinalNumber:             ordinalNumber,
		ColorDifference:           cDifference.Load() / frameSize,
		BinaryThresholdDifference: float64(btcDifference.Load()) / frameSize,
		Brightness:                brightness.Load() / frameSize,
	}
}
