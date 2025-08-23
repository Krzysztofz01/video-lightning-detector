package statistics

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
)

func TestCreateDescriptiveStatisticsShouldCorrectlyCalculateDescriptiveStatistics(t *testing.T) {
	testFrames := []*frame.Frame{
		{OrdinalNumber: 1, ColorDifference: 0.5, BinaryThresholdDifference: 0.0, Brightness: 0.1},
		{OrdinalNumber: 2, ColorDifference: 0.5, BinaryThresholdDifference: 0.0, Brightness: 0.2},
		{OrdinalNumber: 3, ColorDifference: 0.0, BinaryThresholdDifference: 0.1, Brightness: 0.3},
		{OrdinalNumber: 4, ColorDifference: 0.0, BinaryThresholdDifference: 0.4, Brightness: 0.4},
		{OrdinalNumber: 5, ColorDifference: 0.4, BinaryThresholdDifference: 0.9, Brightness: 0.5},
	}

	fc := frame.NewFrameCollection(len(testFrames))
	for _, f := range testFrames {
		err := fc.Push(f)
		assert.Nil(t, err)
	}

	fc.Lock()

	cases := []struct {
		Frames                                               frame.FrameCollection
		MovingMeanResolution                                 int
		ExpectedBrightnessMean                               float64
		ExpectedBrightnessMovingMeanAtPoint                  float64
		ExpectedBrightnessMovingStdDevAtPoint                float64
		ExpectedBrightnessStandardDeviation                  float64
		ExpectedBrightnessMin                                float64
		ExpectedBrightnessMax                                float64
		ExpectedColorDifferenceMean                          float64
		ExpectedColorDifferenceMovingMeanAtPoint             float64
		ExpectedColorDifferenceMovingStdDevAtPoint           float64
		ExpectedColorDifferenceStandardDeviation             float64
		ExpectedColorDifferenceMin                           float64
		ExpectedColorDifferenceMax                           float64
		ExpectedBinaryThresholdDifferenceMean                float64
		ExpectedBinaryThresholdDifferenceMovingMeanAtPoint   float64
		ExpectedBinaryThresholdDifferenceMovingStdDevAtPoint float64
		ExpectedBinaryThresholdDifferenceStandardDeviation   float64
		ExpectedBinaryThresholdDifferenceMin                 float64
		ExpectedBinaryThresholdDifferenceMax                 float64
	}{
		{
			Frames:                                               fc,
			MovingMeanResolution:                                 4,
			ExpectedBrightnessMean:                               0.3,
			ExpectedBrightnessMovingMeanAtPoint:                  0.4,
			ExpectedBrightnessMovingStdDevAtPoint:                0.081649658092773,
			ExpectedBrightnessStandardDeviation:                  0.14142135623731,
			ExpectedBrightnessMin:                                0.1,
			ExpectedBrightnessMax:                                0.5,
			ExpectedColorDifferenceMean:                          0.28,
			ExpectedColorDifferenceMovingMeanAtPoint:             0.13333333333333,
			ExpectedColorDifferenceMovingStdDevAtPoint:           0.18856180831641,
			ExpectedColorDifferenceStandardDeviation:             0.2315167380558,
			ExpectedColorDifferenceMin:                           0.0,
			ExpectedColorDifferenceMax:                           0.5,
			ExpectedBinaryThresholdDifferenceMean:                0.28,
			ExpectedBinaryThresholdDifferenceMovingMeanAtPoint:   0.46666666666667,
			ExpectedBinaryThresholdDifferenceMovingStdDevAtPoint: 0.32998316455372,
			ExpectedBinaryThresholdDifferenceStandardDeviation:   0.34292856398964,
			ExpectedBinaryThresholdDifferenceMin:                 0.0,
			ExpectedBinaryThresholdDifferenceMax:                 0.9,
		},
	}

	const delta float64 = 1e-10

	for _, c := range cases {
		stats := CreateDescriptiveStatistics(c.Frames, c.MovingMeanResolution)
		assert.NotNil(t, stats)

		actualStats, err := stats.At(c.Frames.Count() - 1)
		assert.NotNil(t, actualStats)
		assert.Nil(t, err)

		assert.InDelta(t, c.ExpectedBrightnessMean, actualStats.BrightnessMean, delta)
		assert.InDelta(t, c.ExpectedBrightnessMovingMeanAtPoint, actualStats.BrightnessMovingMeanAtPoint, delta)
		assert.InDelta(t, c.ExpectedBrightnessMovingStdDevAtPoint, actualStats.BrightnessMovingStdDevAtPoint, delta)
		assert.InDelta(t, c.ExpectedBrightnessStandardDeviation, actualStats.BrightnessStandardDeviation, delta)
		assert.InDelta(t, c.ExpectedBrightnessMin, actualStats.BrightnessMin, delta)
		assert.InDelta(t, c.ExpectedBrightnessMax, actualStats.BrightnessMax, delta)
		assert.InDelta(t, c.ExpectedColorDifferenceMean, actualStats.ColorDifferenceMean, delta)
		assert.InDelta(t, c.ExpectedColorDifferenceMovingMeanAtPoint, actualStats.ColorDifferenceMovingMeanAtPoint, delta)
		assert.InDelta(t, c.ExpectedColorDifferenceMovingStdDevAtPoint, actualStats.ColorDifferenceMovingStdDevAtPoint, delta)
		assert.InDelta(t, c.ExpectedColorDifferenceStandardDeviation, actualStats.ColorDifferenceStandardDeviation, delta)
		assert.InDelta(t, c.ExpectedColorDifferenceMin, actualStats.ColorDifferenceMin, delta)
		assert.InDelta(t, c.ExpectedColorDifferenceMax, actualStats.ColorDifferenceMax, delta)
		assert.InDelta(t, c.ExpectedBinaryThresholdDifferenceMean, actualStats.BinaryThresholdDifferenceMean, delta)
		assert.InDelta(t, c.ExpectedBinaryThresholdDifferenceMovingMeanAtPoint, actualStats.BinaryThresholdDifferenceMovingMeanAtPoint, delta)
		assert.InDelta(t, c.ExpectedBinaryThresholdDifferenceMovingStdDevAtPoint, actualStats.BinaryThresholdDifferenceMovingStdDevAtPoint, delta)
		assert.InDelta(t, c.ExpectedBinaryThresholdDifferenceStandardDeviation, actualStats.BinaryThresholdDifferenceStandardDeviation, delta)
		assert.InDelta(t, c.ExpectedBinaryThresholdDifferenceMin, actualStats.BinaryThresholdDifferenceMin, delta)
		assert.InDelta(t, c.ExpectedBinaryThresholdDifferenceMax, actualStats.BinaryThresholdDifferenceMax, delta)
	}
}
