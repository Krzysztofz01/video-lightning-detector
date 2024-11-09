package frame

import (
	"image"
	"image/color"

	"github.com/Krzysztofz01/pimit"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
	"go.uber.org/atomic"
)

const (
	// TODO: Implement different threshold for day/night. The brightness value can be used as the determinant.
	BinaryThresholdParam float64 = 200.0 / 255.0
)

// Strucutre representing a single video frame and its calculated parameters.
// TODO: When it coms to BinaryThreshold we need to test which approach gives better results.
// Currently we are comparing the BT of the previous and current frame and than calcualte the white_pixels / all_pixels
// Alternatively we can just count the occurance of white pixels and return the non-normalized result
type Frame struct {
	OrdinalNumber             int     `json:"ordinal-number"`
	ColorDifference           float64 `json:"color-difference"`
	BinaryThresholdDifference float64 `json:"binary-threshold-difference"`
	Brightness                float64 `json:"brightness"`
}

// Create a new frame instance by providing the current and previous frame images and the ordinal number of the frame.
func CreateNewFrame(currentFrame, previousFrame image.Image, ordinalNumber int, binaryThresholdParam float64) *Frame {
	var (
		brightness    *atomic.Float64 = atomic.NewFloat64(0.0)
		cDifference   *atomic.Float64 = atomic.NewFloat64(0)
		btcDifference *atomic.Int32   = atomic.NewInt32(0)
	)

	pimit.ParallelRead(currentFrame, func(x, y int, currentColor color.Color) {
		brightness.Add(utils.GetColorBrightness(currentColor))

		if ordinalNumber > 1 {
			previousColor := previousFrame.At(x, y)

			cDifference.Add(utils.GetColorDifference(currentColor, previousColor))

			thresholdCurrent := utils.BinaryThreshold(currentColor, binaryThresholdParam)
			thresholdPrevious := utils.BinaryThreshold(previousColor, binaryThresholdParam)
			if thresholdCurrent != thresholdPrevious {
				btcDifference.Add(1)
			}
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
