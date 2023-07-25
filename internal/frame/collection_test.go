package frame

import (
	"bytes"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFramesCollectionShouldCreate(t *testing.T) {
	collection := CreateNewFramesCollection(5)

	assert.NotNil(t, collection)
}

func TestFramesCollectionShouldAppendFrame(t *testing.T) {
	frame := CreateNewFrame(mockImage(color.White), mockImage(color.White), 1)
	collection := CreateNewFramesCollection(5)

	err := collection.Append(frame)

	assert.Nil(t, err)
}

func TestFramesCollectionShouldNotAppendNilFrame(t *testing.T) {
	collection := CreateNewFramesCollection(5)

	err := collection.Append(nil)

	assert.NotNil(t, err)
}

func TestFramesCollectionShouldNotAppendFrameWithSameOrdinalNumber(t *testing.T) {
	frame1 := CreateNewFrame(mockImage(color.White), mockImage(color.White), 2)
	frame2 := CreateNewFrame(mockImage(color.Black), mockImage(color.Black), 2)
	collection := CreateNewFramesCollection(5)

	err := collection.Append(frame1)
	assert.Nil(t, err)

	err = collection.Append(frame2)
	assert.NotNil(t, err)
}

func TestFramesCollectionShouldGetFrame(t *testing.T) {
	frameNumber := 2
	frame := CreateNewFrame(mockImage(color.White), mockImage(color.White), frameNumber)
	collection := CreateNewFramesCollection(5)

	err := collection.Append(frame)
	assert.Nil(t, err)

	actualFrame, err := collection.Get(frameNumber)
	assert.Nil(t, err)
	assert.NotNil(t, actualFrame)

	assert.Equal(t, frame, actualFrame)
}

func TestFramesCollectionShouldNotGetNotExistingFrame(t *testing.T) {
	collection := CreateNewFramesCollection(5)

	frame, err := collection.Get(3)
	assert.NotNil(t, err)
	assert.Nil(t, frame)
}

func TestFramesCollectionShouldCalculateStatistics(t *testing.T) {
	frame1 := CreateNewFrame(mockImage(color.White), mockImage(color.Black), 1)
	frame2 := CreateNewFrame(mockImage(color.Black), mockImage(color.White), 2)
	collection := CreateNewFramesCollection(5)

	err := collection.Append(frame1)
	assert.Nil(t, err)

	err = collection.Append(frame2)
	assert.Nil(t, err)

	statistics := collection.CalculateStatistics()

	assert.Equal(t, statistics.BrightnessMean, 0.5)
	assert.Equal(t, statistics.BrightnessStandardDeviation, 0.5)
	assert.Equal(t, statistics.BrightnessMax, 1.0)
	assert.Equal(t, statistics.ColorDifferenceMean, 1.0)
	assert.Equal(t, statistics.ColorDifferenceStandardDeviation, 0.0)
	assert.Equal(t, statistics.ColorDifferenceMax, 1.0)
	assert.Equal(t, statistics.BinaryThresholdDifferenceMean, 1.0)
	assert.Equal(t, statistics.BinaryThresholdDifferenceStandardDeviation, 0.0)
	assert.Equal(t, statistics.BinaryThresholdDifferenceMax, 1.0)
}

func TestFramesCollectionShouldExportJsonReport(t *testing.T) {
	buffer := &bytes.Buffer{}
	assert.Zero(t, buffer.Len())

	collection := CreateNewFramesCollection(5)
	assert.NotNil(t, collection)

	collection.Append(CreateNewFrame(mockImage(color.White), mockImage(color.White), 1))

	err := collection.ExportJsonReport(buffer)
	assert.Nil(t, err)

	assert.NotZero(t, buffer.Len())
}

func TestFramesCollectionShouldExportCsvReport(t *testing.T) {
	buffer := &bytes.Buffer{}
	assert.Zero(t, buffer.Len())

	collection := CreateNewFramesCollection(5)
	assert.NotNil(t, collection)

	collection.Append(CreateNewFrame(mockImage(color.White), mockImage(color.White), 1))

	err := collection.ExportCsvReport(buffer)
	assert.Nil(t, err)

	assert.NotZero(t, buffer.Len())
}
