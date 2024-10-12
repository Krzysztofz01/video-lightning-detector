package statistics

import (
	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

type DescriptiveStatistics struct {
	BrightnessMean                             float64   `json:"brightness-mean"`
	BrightnessMovingMean                       []float64 `json:"brightness-moving-mean"`
	BrightnessStandardDeviation                float64   `json:"brightness-standard-deviation"`
	BrightnessMax                              float64   `json:"brightness-max"`
	ColorDifferenceMean                        float64   `json:"color-difference-mean"`
	ColorDifferenceMovingMean                  []float64 `json:"color-difference-moving-mean"`
	ColorDifferenceStandardDeviation           float64   `json:"color-difference-standard-deviation"`
	ColorDifferenceMax                         float64   `json:"color-difference-max"`
	BinaryThresholdDifferenceMean              float64   `json:"binary-threshold-difference-mean"`
	BinaryThresholdDifferenceMovingMean        []float64 `json:"binary-threshold-difference-moving-mean"`
	BinaryThresholdDifferenceStandardDeviation float64   `json:"binary-threshold-difference-standard-deviation"`
	BinaryThresholdDifferenceMax               float64   `json:"binary-threshold-difference-max"`
}

func CreateDescriptiveStatistics(fc frame.FrameCollection, movingMeanResolution int) DescriptiveStatistics {
	frames := fc.GetAll()

	var (
		movingMeanBias                int       = movingMeanResolution / 2
		brightness                    []float64 = make([]float64, 0, len(frames))
		colorDiff                     []float64 = make([]float64, 0, len(frames))
		binaryThresholdDiff           []float64 = make([]float64, 0, len(frames))
		brightnessMovingMean          []float64 = make([]float64, 0, len(frames))
		colorDiffMovingMean           []float64 = make([]float64, 0, len(frames))
		binaryThresholdDiffMovingMean []float64 = make([]float64, 0, len(frames))
	)

	for _, frame := range frames {
		brightness = append(brightness, frame.Brightness)
		colorDiff = append(colorDiff, frame.ColorDifference)
		binaryThresholdDiff = append(binaryThresholdDiff, frame.BinaryThresholdDifference)
	}

	for index := range frames {
		brightnessMovingMean = append(brightnessMovingMean, utils.MovingMean(brightness, index, movingMeanBias))
		colorDiffMovingMean = append(colorDiffMovingMean, utils.MovingMean(colorDiff, index, movingMeanBias))
		binaryThresholdDiffMovingMean = append(binaryThresholdDiffMovingMean, utils.MovingMean(binaryThresholdDiff, index, movingMeanBias))
	}

	return DescriptiveStatistics{
		BrightnessMean:                             utils.Mean(brightness),
		BrightnessMovingMean:                       brightnessMovingMean,
		BrightnessStandardDeviation:                utils.StandardDeviation(brightness),
		BrightnessMax:                              utils.Max(brightness),
		ColorDifferenceMean:                        utils.Mean(colorDiff),
		ColorDifferenceMovingMean:                  colorDiffMovingMean,
		ColorDifferenceStandardDeviation:           utils.StandardDeviation(colorDiff),
		ColorDifferenceMax:                         utils.Max(colorDiff),
		BinaryThresholdDifferenceMean:              utils.Mean(binaryThresholdDiff),
		BinaryThresholdDifferenceMovingMean:        binaryThresholdDiffMovingMean,
		BinaryThresholdDifferenceStandardDeviation: utils.StandardDeviation(binaryThresholdDiff),
		BinaryThresholdDifferenceMax:               utils.Max(binaryThresholdDiff),
	}
}
