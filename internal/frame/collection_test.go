package frame

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFramesCollectionShouldCreate(t *testing.T) {
	collection := NewFrameCollection(5)
	assert.NotNil(t, collection)

	assert.Panics(t, func() {
		NewFrameCollection(-1)
	})

	assert.Panics(t, func() {
		NewFrameCollection(0)
	})
}

func TestFramesCollectionShouldPushValidFrames(t *testing.T) {
	frames := []*Frame{
		CreateNewFrame(mockImage(color.White), mockImage(color.White), 1, BinaryThresholdParam),
		CreateNewFrame(mockImage(color.White), mockImage(color.White), 2, BinaryThresholdParam),
		CreateNewFrame(mockImage(color.White), mockImage(color.White), 3, BinaryThresholdParam),
		CreateNewFrame(mockImage(color.White), mockImage(color.White), 4, BinaryThresholdParam),
		CreateNewFrame(mockImage(color.White), mockImage(color.White), 5, BinaryThresholdParam),
	}

	collection := NewFrameCollection(len(frames))

	for _, frame := range frames {
		err := collection.Push(frame)
		assert.Nil(t, err)
	}
}

func TestFramesCollectionShouldNotPushInvalidFrame(t *testing.T) {
	collection := NewFrameCollection(1)

	// NOTE: nil frame
	err := collection.Push(nil)
	assert.NotNil(t, err)

	// NOTE: frame with invalid ordinal number
	frame := CreateNewFrame(mockImage(color.White), mockImage(color.White), 2, BinaryThresholdParam)
	err = collection.Push(frame)
	assert.NotNil(t, err)

	frame = CreateNewFrame(mockImage(color.White), mockImage(color.White), 1, BinaryThresholdParam)
	err = collection.Push(frame)
	assert.Nil(t, err)

	// NOTE: frame out of capacity range
	frame = CreateNewFrame(mockImage(color.White), mockImage(color.White), 2, BinaryThresholdParam)
	err = collection.Push(frame)
	assert.NotNil(t, err)
}

func TestFramesCollectionShouldCorrectlyHandleAccessBasedOnLength(t *testing.T) {
	collection := NewFrameCollection(1)

	assert.Panics(t, func() {
		collection.Count()
	})

	assert.Panics(t, func() {
		collection.GetAll()
	})

	frame := CreateNewFrame(mockImage(color.White), mockImage(color.White), 1, BinaryThresholdParam)
	err := collection.Push(frame)
	assert.Nil(t, err)

	frames := collection.GetAll()
	assert.NotNil(t, frames)
	assert.Len(t, frames, 1)

	count := collection.Count()
	assert.Equal(t, 1, count)
}
