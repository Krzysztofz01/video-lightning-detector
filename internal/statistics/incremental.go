package statistics

import (
	"fmt"
	"sync"

	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

type IncrementalDescriptiveStatistics interface {
	Push(f *frame.Frame)
	Peek() DescriptiveStatisticsEntry
}

type incrementalDescriptiveStatistics struct {
	Bias        int
	Previous    DescriptiveStatisticsEntry
	FrameBuffer utils.CircularBuffer[*frame.Frame]
}

func (stat *incrementalDescriptiveStatistics) Push(f *frame.Frame) {
	var (
		length         int     = stat.FrameBuffer.GetTotalCount()
		nextBrightness float64 = f.Brightness
		nextColorDiff  float64 = f.ColorDifference
		nextBtDiff     float64 = f.BinaryThresholdDifference
	)

	if length == 0 {
		stat.FrameBuffer.Push(f)
		stat.Previous = DescriptiveStatisticsEntry{
			BrightnessMean:                               nextBrightness,
			BrightnessMovingMeanAtPoint:                  nextBrightness,
			BrightnessMovingStdDevAtPoint:                0,
			BrightnessStandardDeviation:                  0,
			BrightnessMin:                                nextBrightness,
			BrightnessMax:                                nextBrightness,
			ColorDifferenceMean:                          nextColorDiff,
			ColorDifferenceMovingMeanAtPoint:             nextColorDiff,
			ColorDifferenceMovingStdDevAtPoint:           0,
			ColorDifferenceStandardDeviation:             0,
			ColorDifferenceMin:                           nextColorDiff,
			ColorDifferenceMax:                           nextColorDiff,
			BinaryThresholdDifferenceMean:                nextBtDiff,
			BinaryThresholdDifferenceMovingMeanAtPoint:   nextBtDiff,
			BinaryThresholdDifferenceMovingStdDevAtPoint: 0,
			BinaryThresholdDifferenceStandardDeviation:   0,
			BinaryThresholdDifferenceMin:                 nextBtDiff,
			BinaryThresholdDifferenceMax:                 nextBtDiff,
		}

		return
	}

	var (
		movingDeltaBrightness float64 = 0.0
		movingDeltaColorDiff  float64 = 0.0
		movingDeltaBtDiff     float64 = 0.0
	)

	if length >= stat.Bias {
		discardFrame, err := stat.FrameBuffer.GetTail(0)
		if err != nil {
			panic(fmt.Errorf("statistics: failed to access the discard frame: %w", err))
		}

		movingDeltaBrightness = discardFrame.Brightness
		movingDeltaColorDiff = discardFrame.ColorDifference
		movingDeltaBtDiff = discardFrame.BinaryThresholdDifference
	}

	var (
		brightnessMin, brightnessMax                 float64
		brightnessMean, brightnessStdDev             float64
		brightnessMovingMean, brightnessMovingStdDev float64
		colorDiffMin, colorDiffMax                   float64
		colorDiffMean, colorDiffStdDev               float64
		colorDiffMovingMean, colorDiffMovingStdDev   float64
		btDiffMin, btDiffMax                         float64
		btDiffMean, btDiffStdDev                     float64
		btDiffMovingMean, btDiffMovingStdDev         float64
		previous                                     DescriptiveStatisticsEntry = stat.Previous
		bias                                         int                        = stat.Bias
	)

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()

		brightnessMin, brightnessMax = utils.MinMaxInc(nextBrightness, previous.BrightnessMin, previous.BrightnessMax)
		brightnessMean, brightnessStdDev = utils.MeanStdDevInc(nextBrightness, previous.BrightnessMean, previous.BrightnessStandardDeviation, length)
		brightnessMovingMean, brightnessMovingStdDev = utils.MovingMeanStdDevInc(nextBrightness, movingDeltaBrightness, previous.BrightnessMovingMeanAtPoint, previous.BrightnessMovingStdDevAtPoint, length, bias)
	}()

	go func() {
		defer wg.Done()

		colorDiffMin, colorDiffMax = utils.MinMaxInc(nextColorDiff, previous.ColorDifferenceMin, previous.ColorDifferenceMax)
		colorDiffMean, colorDiffStdDev = utils.MeanStdDevInc(nextColorDiff, previous.ColorDifferenceMean, previous.ColorDifferenceStandardDeviation, length)
		colorDiffMovingMean, colorDiffMovingStdDev = utils.MovingMeanStdDevInc(nextColorDiff, movingDeltaColorDiff, previous.ColorDifferenceMovingMeanAtPoint, previous.ColorDifferenceMovingStdDevAtPoint, length, bias)
	}()

	go func() {
		defer wg.Done()

		btDiffMin, btDiffMax = utils.MinMaxInc(nextBtDiff, previous.BinaryThresholdDifferenceMin, previous.BinaryThresholdDifferenceMax)
		btDiffMean, btDiffStdDev = utils.MeanStdDevInc(nextBtDiff, previous.BinaryThresholdDifferenceMean, previous.BinaryThresholdDifferenceStandardDeviation, length)
		btDiffMovingMean, btDiffMovingStdDev = utils.MovingMeanStdDevInc(nextBtDiff, movingDeltaBtDiff, previous.BinaryThresholdDifferenceMovingMeanAtPoint, previous.BinaryThresholdDifferenceMovingStdDevAtPoint, length, bias)
	}()

	wg.Wait()

	stat.FrameBuffer.Push(f)
	stat.Previous = DescriptiveStatisticsEntry{
		BrightnessMean:                               brightnessMean,
		BrightnessMovingMeanAtPoint:                  brightnessMovingMean,
		BrightnessMovingStdDevAtPoint:                brightnessMovingStdDev,
		BrightnessStandardDeviation:                  brightnessStdDev,
		BrightnessMin:                                brightnessMin,
		BrightnessMax:                                brightnessMax,
		ColorDifferenceMean:                          colorDiffMean,
		ColorDifferenceMovingMeanAtPoint:             colorDiffMovingMean,
		ColorDifferenceMovingStdDevAtPoint:           colorDiffMovingStdDev,
		ColorDifferenceStandardDeviation:             colorDiffStdDev,
		ColorDifferenceMin:                           colorDiffMin,
		ColorDifferenceMax:                           colorDiffMax,
		BinaryThresholdDifferenceMean:                btDiffMean,
		BinaryThresholdDifferenceMovingMeanAtPoint:   btDiffMovingMean,
		BinaryThresholdDifferenceMovingStdDevAtPoint: btDiffMovingStdDev,
		BinaryThresholdDifferenceStandardDeviation:   btDiffStdDev,
		BinaryThresholdDifferenceMin:                 btDiffMin,
		BinaryThresholdDifferenceMax:                 btDiffMax,
	}
}

func (stat *incrementalDescriptiveStatistics) Peek() DescriptiveStatisticsEntry {
	return stat.Previous
}

func NewIncrementalDescriptiveStatistics(movingMeanResolution int) IncrementalDescriptiveStatistics {
	// NOTE: The incremental statistics take only the "left" part under account
	// which is the half of the resolution length plus one for the position.
	bias := movingMeanResolution/2 + 1

	return &incrementalDescriptiveStatistics{
		Bias:        bias,
		Previous:    DescriptiveStatisticsEntry{},
		FrameBuffer: utils.NewCircularBuffer[*frame.Frame](bias),
	}
}
