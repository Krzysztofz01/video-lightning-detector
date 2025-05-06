package detector

import (
	"fmt"
	"time"

	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/printer"
	"github.com/Krzysztofz01/video-lightning-detector/internal/statistics"
)

type AutoThresholdStrategy int

const (
	AboveMeanOfDeviations AutoThresholdStrategy = iota
)

type AutoThreshold interface {
	// Perform the auto-threshold calculation with the given strategy and applying the results to the provided options and returning the altered copy
	ApplyToOptions(opt options.DetectorOptions, s AutoThresholdStrategy) (options.DetectorOptions, error)
}

type autoThreshold struct {
	Frames     frame.FrameCollection
	Statistics statistics.DescriptiveStatistics
	Printer    printer.Printer
}

func (at *autoThreshold) ApplyToOptions(opt options.DetectorOptions, s AutoThresholdStrategy) (options.DetectorOptions, error) {
	autoThresholdTime := time.Now()
	at.Printer.Debug("Starting the auto thresholds calculation stage.")

	var (
		thresholds thresholdSet
		err        error
	)

	switch s {
	case AboveMeanOfDeviations:
		if thresholds, err = at.CalculateAboveMeanOfDeviations(); err != nil {
			return options.DetectorOptions{}, fmt.Errorf("detector: failed to calculate the thresholds using about mean of deviations: %w", err)
		}
	default:
		return options.DetectorOptions{}, fmt.Errorf("detector: invalid auto threshold strategy specified")
	}

	var (
		defaultOptions options.DetectorOptions = options.GetDefaultDetectorOptions()
		copyOptions    options.DetectorOptions = opt.Clone()
	)

	if opt.BrightnessDetectionThreshold == defaultOptions.BrightnessDetectionThreshold {
		copyOptions.BrightnessDetectionThreshold = thresholds.BrightnessThreshold
		at.Printer.Debug("Auto calculated brightness detection threshold: %g", thresholds.BrightnessThreshold)
	} else {
		at.Printer.Warning("The brightness detection threshold (%f) value was explicitly specified and would not be replace by the auto-calculated one (%f)",
			opt.BrightnessDetectionThreshold,
			thresholds.BrightnessThreshold)
	}

	if opt.ColorDifferenceDetectionThreshold == defaultOptions.ColorDifferenceDetectionThreshold {
		copyOptions.ColorDifferenceDetectionThreshold = thresholds.ColorDifferenceThreshold
		at.Printer.Debug("Auth calculated color difference detection threshold: %g", thresholds.ColorDifferenceThreshold)
	} else {
		at.Printer.Warning("The color difference detection threshold (%f) value was explicitly specified and would not be replace by the auto-calculated one (%f)",
			opt.ColorDifferenceDetectionThreshold,
			thresholds.ColorDifferenceThreshold)
	}

	if opt.BinaryThresholdDifferenceDetectionThreshold == defaultOptions.BinaryThresholdDifferenceDetectionThreshold {
		copyOptions.BinaryThresholdDifferenceDetectionThreshold = thresholds.BinaryThresholdDifferenceThreshold
		at.Printer.Debug("Auto calculated binary threshold difference threshold: %g", thresholds.BinaryThresholdDifferenceThreshold)
	} else {
		at.Printer.Warning("The binary threshold detection threshold (%f) value was explicitly specified and would not be replace by the auto-calculated one (%f)",
			opt.BinaryThresholdDifferenceDetectionThreshold,
			thresholds.BinaryThresholdDifferenceThreshold)
	}

	at.Printer.Debug("Auto thresholds calculation stage finished. Stage took: %s", time.Since(autoThresholdTime))
	return copyOptions, nil
}

func (at *autoThreshold) CalculateAboveMeanOfDeviations() (thresholdSet, error) {
	frames := at.Frames.GetAll()

	// TODO: Those values can be either controlled via different strategies or be fine-tuned.
	const (
		brightnessDiffCoefficient   float64 = 1.0
		brightnessStdDevCoefficient float64 = 0.0
		colorDiffDiffCoefficient    float64 = 1.0
		colorDiffStdDevCoefficient  float64 = 0.0
		btDiffDiffCoefficient       float64 = 0.25
		btDiffStdDevCoefficient     float64 = 0.15
	)

	var (
		brightnessMeanDiffSum float64 = 0.0
		brightnessStdDevSum   float64 = 0.0
		brightnessCount       int     = 0
		colorDiffMeanDiffSum  float64 = 0.0
		colorDiffStdDevSum    float64 = 0.0
		colorDiffCount        int     = 0
		btDiffMeanDiffSum     float64 = 0.0
		btDiffStdDevSum       float64 = 0.0
		btDiffCount           int     = 0
	)

	for frameIndex, frame := range frames {
		if brightnessDiff := frame.Brightness - at.Statistics.BrightnessMovingMean[frameIndex]; brightnessDiff > 0 {
			brightnessMeanDiffSum += brightnessDiff
			brightnessStdDevSum += at.Statistics.BrightnessMovingStdDev[frameIndex]
			brightnessCount += 1
		}

		if colorDiff := frame.ColorDifference - at.Statistics.ColorDifferenceMovingMean[frameIndex]; colorDiff > 0 {
			colorDiffMeanDiffSum += colorDiff
			colorDiffStdDevSum += at.Statistics.ColorDifferenceMovingStdDev[frameIndex]
			colorDiffCount += 1
		}

		if btDiff := frame.BinaryThresholdDifference - at.Statistics.BinaryThresholdDifferenceMovingMean[frameIndex]; btDiff > 0 {
			btDiffMeanDiffSum += btDiff
			btDiffStdDevSum += at.Statistics.BinaryThresholdDifferenceMovingStdDev[frameIndex]
			btDiffCount += 1
		}
	}

	var brightnessThreshold float64
	if brightnessCount == 0 {
		brightnessThreshold = 0
	} else {
		countf := float64(brightnessCount)
		brightnessThreshold = brightnessDiffCoefficient*brightnessMeanDiffSum/countf + brightnessStdDevCoefficient*brightnessStdDevSum/countf
	}

	var colorDifferenceThreshold float64
	if colorDiffCount == 0 {
		colorDifferenceThreshold = 0
	} else {
		countf := float64(colorDiffCount)
		colorDifferenceThreshold = colorDiffDiffCoefficient*colorDiffMeanDiffSum/countf + colorDiffStdDevCoefficient*colorDiffStdDevSum/countf
	}

	var binaryThresholdDifferenceThreshold float64
	if btDiffCount == 0 {
		binaryThresholdDifferenceThreshold = 0
	} else {
		countf := float64(btDiffCount)
		binaryThresholdDifferenceThreshold = btDiffDiffCoefficient*btDiffMeanDiffSum/countf + btDiffStdDevCoefficient*btDiffStdDevSum/countf
	}

	return thresholdSet{
		BrightnessThreshold:                brightnessThreshold,
		ColorDifferenceThreshold:           colorDifferenceThreshold,
		BinaryThresholdDifferenceThreshold: binaryThresholdDifferenceThreshold,
	}, nil
}

func NewAutoThreshold(fc frame.FrameCollection, ds statistics.DescriptiveStatistics, p printer.Printer) AutoThreshold {
	return &autoThreshold{
		Frames:     fc,
		Statistics: ds,
		Printer:    p,
	}
}

type thresholdSet struct {
	BrightnessThreshold                float64
	ColorDifferenceThreshold           float64
	BinaryThresholdDifferenceThreshold float64
}
