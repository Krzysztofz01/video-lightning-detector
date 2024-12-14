package statistics

import (
	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

type DescriptiveStatistics struct {
	BrightnessMean                             float64   `json:"brightness-mean"`
	BrightnessMovingMean                       []float64 `json:"brightness-moving-mean"`
	BrightnessMovingStdDev                     []float64 `json:"brightness-moving-standard-deviation"`
	BrightnessStandardDeviation                float64   `json:"brightness-standard-deviation"`
	BrightnessMin                              float64   `json:"brightness-min"`
	BrightnessMax                              float64   `json:"brightness-max"`
	ColorDifferenceMean                        float64   `json:"color-difference-mean"`
	ColorDifferenceMovingMean                  []float64 `json:"color-difference-moving-mean"`
	ColorDifferenceMovingStdDev                []float64 `json:"color-difference-moving-standard-deviation"`
	ColorDifferenceStandardDeviation           float64   `json:"color-difference-standard-deviation"`
	ColorDifferenceMin                         float64   `json:"color-difference-min"`
	ColorDifferenceMax                         float64   `json:"color-difference-max"`
	BinaryThresholdDifferenceMean              float64   `json:"binary-threshold-difference-mean"`
	BinaryThresholdDifferenceMovingMean        []float64 `json:"binary-threshold-difference-moving-mean"`
	BinaryThresholdDifferenceMovingStdDev      []float64 `json:"binary-threshold-difference-moving-standard-deviation"`
	BinaryThresholdDifferenceStandardDeviation float64   `json:"binary-threshold-difference-standard-deviation"`
	BinaryThresholdDifferenceMin               float64   `json:"binary-threshold-difference-min"`
	BinaryThresholdDifferenceMax               float64   `json:"binary-threshold-difference-max"`
}

func CreateDescriptiveStatistics(fc frame.FrameCollection, movingMeanResolution int) DescriptiveStatistics {
	frames := fc.GetAll()

	var (
		movingMeanBias                  int       = movingMeanResolution / 2
		brightness                      []float64 = make([]float64, 0, len(frames))
		colorDiff                       []float64 = make([]float64, 0, len(frames))
		binaryThresholdDiff             []float64 = make([]float64, 0, len(frames))
		brightnessMovingMean            []float64 = make([]float64, 0, len(frames))
		brightnessMovingStdDev          []float64 = make([]float64, 0, len(frames))
		colorDiffMovingMean             []float64 = make([]float64, 0, len(frames))
		colorDiffMovingStdDev           []float64 = make([]float64, 0, len(frames))
		binaryThresholdDiffMovingMean   []float64 = make([]float64, 0, len(frames))
		binaryThresholdDiffMovingStdDev []float64 = make([]float64, 0, len(frames))
	)

	for _, frame := range frames {
		brightness = append(brightness, frame.Brightness)
		colorDiff = append(colorDiff, frame.ColorDifference)
		binaryThresholdDiff = append(binaryThresholdDiff, frame.BinaryThresholdDifference)
	}

	var (
		movingMean   float64
		movingStdDev float64
	)

	for index := range frames {
		movingMean, movingStdDev = utils.MovingMeanStdDev(brightness, index, movingMeanBias)
		brightnessMovingMean = append(brightnessMovingMean, movingMean)
		brightnessMovingStdDev = append(brightnessMovingStdDev, movingStdDev)

		movingMean, movingStdDev = utils.MovingMeanStdDev(colorDiff, index, movingMeanBias)
		colorDiffMovingMean = append(colorDiffMovingMean, movingMean)
		colorDiffMovingStdDev = append(colorDiffMovingStdDev, movingStdDev)

		movingMean, movingStdDev = utils.MovingMeanStdDev(binaryThresholdDiff, index, movingMeanBias)
		binaryThresholdDiffMovingMean = append(binaryThresholdDiffMovingMean, movingMean)
		binaryThresholdDiffMovingStdDev = append(binaryThresholdDiffMovingStdDev, movingStdDev)
	}

	brightnessMin, brightnessMax := utils.MinMax(brightness)
	colorDiffMin, colorDiffMax := utils.MinMax(colorDiff)
	btDiffMin, btDiffMax := utils.MinMax(binaryThresholdDiff)

	return DescriptiveStatistics{
		BrightnessMean:                             utils.Mean(brightness),
		BrightnessMovingMean:                       brightnessMovingMean,
		BrightnessMovingStdDev:                     brightnessMovingStdDev,
		BrightnessStandardDeviation:                utils.StandardDeviation(brightness),
		BrightnessMin:                              brightnessMin,
		BrightnessMax:                              brightnessMax,
		ColorDifferenceMean:                        utils.Mean(colorDiff),
		ColorDifferenceMovingMean:                  colorDiffMovingMean,
		ColorDifferenceMovingStdDev:                colorDiffMovingStdDev,
		ColorDifferenceStandardDeviation:           utils.StandardDeviation(colorDiff),
		ColorDifferenceMin:                         colorDiffMin,
		ColorDifferenceMax:                         colorDiffMax,
		BinaryThresholdDifferenceMean:              utils.Mean(binaryThresholdDiff),
		BinaryThresholdDifferenceMovingMean:        binaryThresholdDiffMovingMean,
		BinaryThresholdDifferenceMovingStdDev:      binaryThresholdDiffMovingStdDev,
		BinaryThresholdDifferenceStandardDeviation: utils.StandardDeviation(binaryThresholdDiff),
		BinaryThresholdDifferenceMin:               btDiffMin,
		BinaryThresholdDifferenceMax:               btDiffMax,
	}
}
