package frame

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFramesCollectionShouldCreate(t *testing.T) {
	collection := CreateNewFrameCollection(5)

	assert.NotNil(t, collection)
}

func TestFramesCollectionShouldAppendFrame(t *testing.T) {
	frame := CreateNewFrame(mockImage(color.White), mockImage(color.White), 1)
	collection := CreateNewFrameCollection(5)

	err := collection.Append(frame)

	assert.Nil(t, err)
}

func TestFramesCollectionShouldNotAppendNilFrame(t *testing.T) {
	collection := CreateNewFrameCollection(5)

	err := collection.Append(nil)

	assert.NotNil(t, err)
}

func TestFramesCollectionShouldNotAppendFrameWithSameOrdinalNumber(t *testing.T) {
	frame1 := CreateNewFrame(mockImage(color.White), mockImage(color.White), 2)
	frame2 := CreateNewFrame(mockImage(color.Black), mockImage(color.Black), 2)
	collection := CreateNewFrameCollection(5)

	err := collection.Append(frame1)
	assert.Nil(t, err)

	err = collection.Append(frame2)
	assert.NotNil(t, err)
}

func TestFramesCollectionShouldGetFrame(t *testing.T) {
	frameNumber := 2
	frame := CreateNewFrame(mockImage(color.White), mockImage(color.White), frameNumber)
	collection := CreateNewFrameCollection(5)

	err := collection.Append(frame)
	assert.Nil(t, err)

	actualFrame, err := collection.Get(frameNumber)
	assert.Nil(t, err)
	assert.NotNil(t, actualFrame)

	assert.Equal(t, frame, actualFrame)
}

func TestFramesCollectionShouldNotGetNotExistingFrame(t *testing.T) {
	collection := CreateNewFrameCollection(5)

	frame, err := collection.Get(3)
	assert.NotNil(t, err)
	assert.Nil(t, frame)
}
