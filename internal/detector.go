package internal

import (
	"fmt"
	"image"
	"image/color"
	"sync"

	vidio "github.com/AlexEidt/Vidio"
	"github.com/Krzysztofz01/pimit"
	"go.uber.org/atomic"
)

type Detector interface {
	Run(inputVideoPath, outputDirectoryPath string) ([]*Frame, error)
}

type detector struct {
	options DetectorOptions
}

func CreateDetector(options DetectorOptions) (Detector, error) {
	if ok, msg := options.AreValid(); !ok {
		return nil, fmt.Errorf("detector: invalid options %s", msg)
	}

	return &detector{
		options: options,
	}, nil
}

func (detector *detector) Run(inputVideoPath, outputDirectoryPath string) ([]*Frame, error) {
	video, err := vidio.NewVideo(inputVideoPath)
	if err != nil {
		return nil, fmt.Errorf("detector: failed to open the video file: %w", err)
	}

	framePrevious := image.NewRGBA(image.Rect(0, 0, video.Width(), video.Height()))
	frameCurrent := image.NewRGBA(image.Rect(0, 0, video.Width(), video.Height()))

	frames := make([]*Frame, 0, video.Frames())

	video.SetFrameBuffer(frameCurrent.Pix)

	frameNumber := 1
	frameCount := video.Frames()

	for video.Read() {
		wg := sync.WaitGroup{}
		wg.Add(2)

		frame := NewFrame(frameNumber)

		go func() {
			defer wg.Done()

			brightness := atomic.NewFloat64(0)
			pimit.ParallelColumnColorRead(frameCurrent, func(color color.Color) {
				brightness.Add(GetGrayscaleBasedBrightness(color))
			})

			frame.SetBrightness(brightness.Load() / float64(frameCount))
		}()

		go func() {
			defer wg.Done()
			if frameNumber == 1 {
				frame.SetDifference(0)
			}

			difference := atomic.NewFloat64(0)
			pimit.ParallelColumnRead(frameCurrent, func(xIndex, yIndex int, color color.Color) {
				colorPrevious := framePrevious.At(xIndex, yIndex)

				difference.Add(GetColorDifference(color, colorPrevious))
			})

			frame.SetDifference(difference.Load() / float64(frameCount))
		}()

		wg.Wait()

		frames = append(frames, frame)

		frameNumber += 1
		copy(framePrevious.Pix, frameCurrent.Pix)
	}

	for a, b := range frames {

	}

	return frames, nil
}

type DetectorOptions struct {
	FrameDifferenceThreshold float64
	FrameBrightnessThreshold float64
}

func (options *DetectorOptions) AreValid() (bool, string) {
	return false, ""
}

func GetDefaultDetectorOptions() DetectorOptions {
	return DetectorOptions{
		FrameDifferenceThreshold: 0,
		FrameBrightnessThreshold: 0,
	}
}
