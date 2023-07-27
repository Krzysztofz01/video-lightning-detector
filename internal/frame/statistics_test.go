package frame

import (
	"bytes"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrameStatisticsShouldCreate(t *testing.T) {
	frames := []*Frame{
		CreateNewFrame(mockImage(color.White), mockImage(color.Black), 1),
		CreateNewFrame(mockImage(color.Black), mockImage(color.White), 2),
	}

	statistics := CreateNewFramesStatistics(frames)
	assert.NotNil(t, statistics)

	assert.Equal(t, statistics.BrightnessMean, 0.5)
	assert.Equal(t, statistics.BrightnessStandardDeviation, 0.5)
	assert.Equal(t, statistics.BrightnessMax, 1.0)
	assert.Equal(t, statistics.BrightnessMovingMean, []float64{0.5, 0.5})
	assert.Equal(t, statistics.ColorDifferenceMean, 0.5)
	assert.Equal(t, statistics.ColorDifferenceStandardDeviation, 0.5)
	assert.Equal(t, statistics.ColorDifferenceMax, 1.0)
	assert.Equal(t, statistics.ColorDifferenceMovingMean, []float64{0.5, 0.5})
	assert.Equal(t, statistics.BinaryThresholdDifferenceMean, 0.5)
	assert.Equal(t, statistics.BinaryThresholdDifferenceStandardDeviation, 0.5)
	assert.Equal(t, statistics.BinaryThresholdDifferenceMax, 1.0)
	assert.Equal(t, statistics.BinaryThresholdDifferenceMovingMean, []float64{0.5, 0.5})
}

func TestFramesStatisticsShouldExportCsvReport(t *testing.T) {
	buffer := &bytes.Buffer{}
	assert.Zero(t, buffer.Len())

	frames := []*Frame{
		CreateNewFrame(mockImage(color.White), mockImage(color.Black), 1),
		CreateNewFrame(mockImage(color.Black), mockImage(color.White), 2),
	}

	statistics := CreateNewFramesStatistics(frames)
	assert.NotNil(t, statistics)

	err := statistics.ExportCsvReport(buffer)
	assert.Nil(t, err)

	assert.NotZero(t, buffer.Len())
}
