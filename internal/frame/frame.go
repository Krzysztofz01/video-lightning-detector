package frame

import (
	"fmt"
	"image"
)

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

// Factory used to create frame instances. The factory controlls the ordinal numbers which are 1-indexed.
type FrameFactory interface {
	CreateNewFrame(frameImage0, frameImage1 *image.RGBA) (*Frame, error)
}

type frameFactory struct {
	BinaryThresholdParam float64
	OrdinalNumber        int
}

// Create a new frame based on the N and N-1 frame images.
func (ff *frameFactory) CreateNewFrame(frameImage0, frameImage1 *image.RGBA) (*Frame, error) {
	if frameImage0 == nil {
		return nil, fmt.Errorf("frame: invalid current frame image nil reference")
	}

	if ff.OrdinalNumber > 1 && frameImage1 == nil {
		return nil, fmt.Errorf("frame: invalid previous frame image reference in context of ordinal number")
	}

	var (
		pf = processFrame(frameImage0, frameImage1, ff.OrdinalNumber, ff.BinaryThresholdParam)
		f  = &Frame{
			OrdinalNumber:             ff.OrdinalNumber,
			ColorDifference:           pf.ColorDifference,
			BinaryThresholdDifference: pf.BinaryThresholdDifference,
			Brightness:                pf.Brightness,
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
