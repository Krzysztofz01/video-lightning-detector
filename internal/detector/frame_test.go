package detector

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
)

func TestCreateFrameStrikeDetectorShouldReturnErrorForInvalidParams(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	opt := options.GetDefaultStreamDetectorOptions()
	opt.FrameDetectionPlotResolution = 1
	opt.FrameDetectionPlotThreshold = 1
	_, err := CreateFrameStrikeDetector(0, 0, opt)
	assert.NotNil(t, err)

	opt = options.GetDefaultStreamDetectorOptions()
	opt.FrameDetectionPlotResolution = -1
	opt.FrameDetectionPlotThreshold = 0.5
	_, err = CreateFrameStrikeDetector(img.Bounds().Dx(), img.Bounds().Dy(), opt)
	assert.NotNil(t, err)

	opt = options.GetDefaultStreamDetectorOptions()
	opt.FrameDetectionPlotResolution = 0
	opt.FrameDetectionPlotThreshold = 0.5
	_, err = CreateFrameStrikeDetector(img.Bounds().Dx(), img.Bounds().Dy(), opt)
	assert.NotNil(t, err)

	opt = options.GetDefaultStreamDetectorOptions()
	opt.FrameDetectionPlotResolution = 2
	opt.FrameDetectionPlotThreshold = -1
	_, err = CreateFrameStrikeDetector(img.Bounds().Dx(), img.Bounds().Dy(), opt)
	assert.NotNil(t, err)

	opt = options.GetDefaultStreamDetectorOptions()
	opt.FrameDetectionPlotResolution = 2
	opt.FrameDetectionPlotThreshold = 2
	_, err = CreateFrameStrikeDetector(img.Bounds().Dx(), img.Bounds().Dy(), opt)
	assert.NotNil(t, err)

	opt = options.GetDefaultStreamDetectorOptions()
	opt.FrameDetectionPlotResolution = 1
	opt.FrameDetectionPlotThreshold = 1
	opt.DetectionBoundsExpression = "nil"
	_, err = CreateFrameStrikeDetector(img.Bounds().Dx(), img.Bounds().Dy(), opt)
	assert.NotNil(t, err)

	opt = options.GetDefaultStreamDetectorOptions()
	opt.FrameDetectionPlotResolution = 1
	opt.FrameDetectionPlotThreshold = 1
	opt.DetectionBoundsExpression = fmt.Sprintf("%d:%d:%d:%d", 0, 0, 20, 20)
	_, err = CreateFrameStrikeDetector(img.Bounds().Dx(), img.Bounds().Dy(), opt)
	assert.NotNil(t, err)
}

func TestGetDetectionPlotShouldRetrunErrorForInvalidParams(t *testing.T) {
	width := 16
	height := 16

	opt := options.GetDefaultStreamDetectorOptions()
	opt.FrameDetectionPlotResolution = 4
	opt.FrameDetectionPlotThreshold = 0.9

	frameStrikeDetector, err := CreateFrameStrikeDetector(width, height, opt)
	assert.Nil(t, err)
	assert.NotNil(t, frameStrikeDetector)

	_, err = frameStrikeDetector.GetDetectionPlot(nil)
	assert.NotNil(t, err)

	_, err = frameStrikeDetector.GetDetectionPlot(image.NewRGBA(image.Rect(0, 0, width/2, height*2)))
	assert.NotNil(t, err)

	opt = options.GetDefaultStreamDetectorOptions()
	opt.FrameDetectionPlotResolution = 4
	opt.FrameDetectionPlotThreshold = 0.9
	opt.DetectionBoundsExpression = fmt.Sprintf("%d:%d:%d:%d", 0, 0, width/2, height/2)

	frameStrikeDetector, err = CreateFrameStrikeDetector(width, height, opt)
	assert.Nil(t, err)
	assert.NotNil(t, frameStrikeDetector)

	_, err = frameStrikeDetector.GetDetectionPlot(image.NewRGBA(image.Rect(0, 0, width*2, height*2)))
	assert.NotNil(t, err)
}

func TestGetDetectionPlotShouldReturnCorrectPlotDataForValidParamsWithoutBbox(t *testing.T) {
	const (
		size       int     = 12
		resolution int     = 6
		threshold  float64 = 225.0 / 255.0
	)

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var c color.Color
			if x < size/2 {
				c = color.Black
			} else {
				c = color.White
			}

			img.Set(x, y, c)
		}
	}

	opt := options.GetDefaultStreamDetectorOptions()
	opt.FrameDetectionPlotResolution = resolution
	opt.FrameDetectionPlotThreshold = threshold

	frameStrikeDetector, err := CreateFrameStrikeDetector(img.Bounds().Dx(), img.Bounds().Dy(), opt)
	assert.Nil(t, err)
	assert.NotNil(t, frameStrikeDetector)

	var (
		expectedHorizontal []float64 = []float64{0, 0, 0, 1, 1, 1}
		expectedVertical   []float64 = []float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5}
	)

	plots, err := frameStrikeDetector.GetDetectionPlot(img)
	assert.Nil(t, err)
	assert.Len(t, plots, 2)
	assert.Equal(t, expectedHorizontal, plots[0])
	assert.Equal(t, expectedVertical, plots[1])
}

func TestGetDetectionPlotShouldReturnCorrectPlotDataForValidParamsWithBbox(t *testing.T) {
	const (
		sizeFull   int     = 4
		sizeBbox   int     = 2
		resolution int     = 4
		threshold  float64 = 225.0 / 255.0
	)

	imgFull := image.NewRGBA(image.Rect(0, 0, sizeFull, sizeFull))
	imgBbox := image.NewRGBA(image.Rect(0, 0, sizeBbox, sizeBbox))

	for y := 0; y < sizeBbox; y++ {
		for x := 0; x < sizeBbox; x++ {
			var c color.Color
			if x < sizeBbox/2 {
				c = color.Black
			} else {
				c = color.White
			}

			imgBbox.Set(x, y, c)
		}
	}

	opt := options.GetDefaultStreamDetectorOptions()
	opt.DetectionBoundsExpression = fmt.Sprintf("%d:%d:%d:%d", sizeBbox/2, sizeBbox/2, sizeBbox, sizeBbox)
	opt.FrameDetectionPlotResolution = resolution
	opt.FrameDetectionPlotThreshold = threshold

	frameStrikeDetector, err := CreateFrameStrikeDetector(imgFull.Bounds().Dx(), imgFull.Bounds().Dy(), opt)
	assert.Nil(t, err)
	assert.NotNil(t, frameStrikeDetector)

	var (
		expectedHorizontal []float64 = []float64{0, 0.0, 0.5, 0}
		expectedVertical   []float64 = []float64{0, 0.25, 0.25, 0}
	)

	plots, err := frameStrikeDetector.GetDetectionPlot(imgBbox)
	assert.Nil(t, err)
	assert.Len(t, plots, 2)
	assert.Equal(t, expectedHorizontal, plots[0])
	assert.Equal(t, expectedVertical, plots[1])
}
