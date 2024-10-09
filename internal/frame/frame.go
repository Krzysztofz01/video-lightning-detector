package frame

import (
	"image"
	"image/color"
	"sync"

	"github.com/Krzysztofz01/pimit"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
	"go.uber.org/atomic"
)

const (
	// TODO: Implement different threshold for day/night. The brightness value can be used as the determinant.
	BinaryThresholdParam float64 = 0.784313725
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
func CreateNewFrame(currentFrame, previousFrame image.Image, ordinalNumber int) *Frame {
	frame := &Frame{
		OrdinalNumber: ordinalNumber,
	}

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()

		frame.Brightness = calculateFrameBrightness(currentFrame)
	}()

	go func() {
		defer wg.Done()
		if ordinalNumber == 1 {
			frame.ColorDifference = 0.0
			return
		}

		frame.ColorDifference = calculateFramesColorDifference(currentFrame, previousFrame)
	}()

	go func() {
		defer wg.Done()
		if ordinalNumber == 1 {
			frame.BinaryThresholdDifference = 0.0
			return
		}

		frame.BinaryThresholdDifference = calculateFramesBinaryThresholdDifference(currentFrame, previousFrame)
	}()

	wg.Wait()
	return frame
}

func calculateFrameBrightness(currentFrame image.Image) float64 {
	brightness := atomic.NewFloat64(0.0)
	pimit.ParallelRead(currentFrame, func(_, _ int, c color.Color) {
		brightness.Add(utils.GetColorBrightness(c))
	})

	frameSize := currentFrame.Bounds().Dx() * currentFrame.Bounds().Dy()
	return brightness.Load() / float64(frameSize)
}

func calculateFramesColorDifference(currentFrame, previousFrame image.Image) float64 {
	difference := atomic.NewFloat64(0.0)
	pimit.ParallelRead(currentFrame, func(x, y int, currentFrameColor color.Color) {
		previousFrameColor := previousFrame.At(x, y)

		difference.Add(utils.GetColorDifference(currentFrameColor, previousFrameColor))
	})

	frameSize := currentFrame.Bounds().Dx() * currentFrame.Bounds().Dy()
	return difference.Load() / float64(frameSize)
}

func calculateFramesBinaryThresholdDifference(currentFrame, previousFrame image.Image) float64 {
	difference := atomic.NewInt32(0)
	pimit.ParallelRead(currentFrame, func(x, y int, currentFrameColor color.Color) {
		thresholdCurrent := utils.BinaryThreshold(currentFrameColor, BinaryThresholdParam)
		thresholdPrevious := utils.BinaryThreshold(previousFrame.At(x, y), BinaryThresholdParam)

		if thresholdCurrent != thresholdPrevious {
			difference.Add(1)
		}
	})

	frameSize := currentFrame.Bounds().Dx() * currentFrame.Bounds().Dy()
	return float64(difference.Load()) / float64(frameSize)
}
