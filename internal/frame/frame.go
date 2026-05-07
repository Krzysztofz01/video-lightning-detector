package frame

import (
	"fmt"
	"image"
)

const (
	BinaryThresholdParam float64 = 200.0 / 255.0
)

// FIXME: Add support for additional weights in the file cache
// FIXME: Add support for additional weights in statistics
// FIXME: Add support for additional weights in detector and auto-threshold component
// FIXME: Fix broken unit-tests

// Strucutre representing a single video frame and its calculated parameters.
type Frame struct {
	OrdinalNumber              int     `json:"ordinal-number"`
	ColorDifference            float64 `json:"color-difference"`
	BinaryThresholdDifference  float64 `json:"binary-threshold-difference"`
	Brightness                 float64 `json:"brightness"`
	BrightnessStdDev           float64 `json:"brightness-std-dev"`
	BrightnessMin              float64 `json:"brightness-min"`
	BrightnessMax              float64 `json:"brightness-max"`
	BrightnessFirstDerivative  float64 `json:"brightness-first-derivative"`
	BrightnessSecondDerivative float64 `json:"brightness-second-derivative"`
	SaturationMean             float64 `json:"saturation-mean"`
	SaturationStdDev           float64 `json:"saturation-std-dev"`
	ColorDifferenceVariance    float64 `json:"color-difference-variance"`
	LuminanceMean              float64 `json:"luminance-mean"`
	LuminanceMin               float64 `json:"luminance-min"`
	LuminanceMax               float64 `json:"luminance-max"`
	BinaryThresholdRatio       float64 `json:"binary-threshold-ratio"`
}

// Factory used to create frame instances. The factory controlls the ordinal numbers which are 1-indexed.
type FrameFactory interface {
	CreateNewFrame(frameImage0, frameImage1 *image.RGBA, frame1, frame2 *Frame) (*Frame, error)
}

type frameFactory struct {
	BinaryThresholdParam float64
	OrdinalNumber        int
}

// Create a new frame based on the N and N-1 frame images.
func (ff *frameFactory) CreateNewFrame(frameImage0, frameImage1 *image.RGBA, frame1, frame2 *Frame) (*Frame, error) {
	if frameImage0 == nil {
		return nil, fmt.Errorf("frame: invalid current frame image nil reference")
	}

	if ff.OrdinalNumber > 1 && frameImage1 == nil {
		return nil, fmt.Errorf("frame: invalid previous frame image reference in context of ordinal number")
	}

	if ff.OrdinalNumber > 1 && frame1 == nil {
		return nil, fmt.Errorf("frame: invalid previous frame reference in context of ordinal number")
	}

	if ff.OrdinalNumber > 2 && frame2 == nil {
		return nil, fmt.Errorf("frame: invalid penultimate frame reference in context of ordinal number")
	}

	var (
		pf = processFrame(frameImage0, frameImage1, frame1, frame2, ff.OrdinalNumber, ff.BinaryThresholdParam)
		f  = &Frame{
			OrdinalNumber:              ff.OrdinalNumber,
			ColorDifference:            pf.ColorDifference,
			BinaryThresholdDifference:  pf.BinaryThresholdDifference,
			Brightness:                 pf.Brightness,
			BrightnessStdDev:           pf.BrightnessStdDev,
			BrightnessMin:              pf.BrightnessMin,
			BrightnessMax:              pf.BrightnessMax,
			BrightnessFirstDerivative:  pf.BrightnessFirstDerivative,
			BrightnessSecondDerivative: pf.BrightnessSecondDerivative,
			SaturationMean:             pf.SaturationMean,
			SaturationStdDev:           pf.SaturationStdDev,
			ColorDifferenceVariance:    pf.ColorDifferenceVariance,
			LuminanceMean:              pf.LuminanceMean,
			LuminanceMin:               pf.LuminanceMin,
			LuminanceMax:               pf.LuminanceMax,
			BinaryThresholdRatio:       pf.BinaryThresholdRatio,
		}
	)

	ff.OrdinalNumber += 1
	return f, nil
}

// Create a new frame factory instance with the specified binary threshold argument used for segmentation operations.
func CreateFrameFactory(binaryThresholdParam float64) FrameFactory {
	return &frameFactory{
		BinaryThresholdParam: binaryThresholdParam,
		OrdinalNumber:        1,
	}
}
