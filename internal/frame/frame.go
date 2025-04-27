package frame

import (
	"image"
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
	frame := processFrame(currentFrame, previousFrame, ordinalNumber, binaryThresholdParam)

	return &Frame{
		OrdinalNumber:             ordinalNumber,
		ColorDifference:           frame.ColorDifference,
		BinaryThresholdDifference: frame.BinaryThresholdDifference,
		Brightness:                frame.Brightness,
	}
}
