package frame

import (
	"image"
	"image/color"
	"strconv"
	"sync"

	"github.com/Krzysztofz01/pimit"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
	"go.uber.org/atomic"
)

// Strucutre representing a single video frame and its calculated parameters.
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
			frame.ColorDifference = 1.0
			return
		}

		frame.ColorDifference = calculateFramesColorDifference(currentFrame, previousFrame)
	}()

	go func() {
		defer wg.Done()
		if ordinalNumber == 1 {
			frame.BinaryThresholdDifference = 1.0
			return
		}

		frame.BinaryThresholdDifference = calculateFramesBinaryThresholdDifference(currentFrame, previousFrame)
	}()

	wg.Wait()
	return frame
}

func calculateFrameBrightness(currentFrame image.Image) float64 {
	brightness := atomic.NewFloat64(0.0)
	pimit.ParallelColumnColorRead(currentFrame, func(color color.Color) {
		brightness.Add(utils.GetColorBrightness(color))
	})

	return brightness.Load()
}

func calculateFramesColorDifference(currentFrame, previousFrame image.Image) float64 {
	difference := atomic.NewFloat64(0.0)
	pimit.ParallelColumnRead(currentFrame, func(x, y int, currentFrameColor color.Color) {
		previousFrameColor := previousFrame.At(x, y)

		difference.Add(utils.GetColorDifference(currentFrameColor, previousFrameColor))
	})

	frameSize := currentFrame.Bounds().Dx() * currentFrame.Bounds().Dy()
	return difference.Load() / float64(frameSize)
}

func calculateFramesBinaryThresholdDifference(currentFrame, previousFrame image.Image) float64 {
	difference := atomic.NewInt32(0)
	pimit.ParallelColumnRead(currentFrame, func(x, y int, currentFrameColor color.Color) {
		thresholdCurrent := utils.BinaryThreshold(currentFrameColor, 0.0196)
		thresholdPrevious := utils.BinaryThreshold(previousFrame.At(x, y), 0.0196)

		if thresholdCurrent != thresholdPrevious {
			difference.Add(1)
		}
	})

	frameSize := currentFrame.Bounds().Dx() * currentFrame.Bounds().Dy()
	return float64(difference.Load()) / float64(frameSize)
}

// Convert the frame string buffer format accepted by the CSV encoder.
func (frame *Frame) ToBuffer() []string {
	buffer := make([]string, 0, 3)
	buffer = append(buffer, strconv.Itoa(frame.OrdinalNumber))
	buffer = append(buffer, strconv.FormatFloat(frame.Brightness, 'f', -1, 64))
	buffer = append(buffer, strconv.FormatFloat(frame.ColorDifference, 'f', -1, 64))
	buffer = append(buffer, strconv.FormatFloat(frame.BinaryThresholdDifference, 'f', -1, 64))

	return buffer
}
