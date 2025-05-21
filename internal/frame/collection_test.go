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

func TestFramesCollectionShouldPushValidFramesAndAccessIn(t *testing.T) {
	cases := []struct {
		Capacity int
		Count    int
	}{
		{1, 0},
		{1, 1},
		{1, 2},
		{5, 4},
		{5, 5},
		{5, 10},
	}

	for _, c := range cases {
		frames := make([]*Frame, 0, c.Count)
		collection := NewFrameCollection(c.Capacity)

		for index := 0; index < c.Count; index += 1 {
			frame := CreateNewFrame(mockImage(color.White), mockImage(color.White), index+1, BinaryThresholdParam)
			err := collection.Push(frame)
			assert.Nil(t, err)

			frames = append(frames, frame)
		}

		collection.Lock()

		assert.Equal(t, collection.Count(), c.Count)
		a := collection.GetAll()
		assert.Equal(t, a, frames)
		for _, frame := range collection.GetAll() {
			assert.NotNil(t, frame)
		}
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

	// NOTE: access before lock
	assert.Panics(t, func() {
		collection.Count()
	})

	collection.Lock()

	// NOTE: push after lock
	frame = CreateNewFrame(mockImage(color.White), mockImage(color.White), 2, BinaryThresholdParam)
	err = collection.Push(frame)
	assert.NotNil(t, err)
}

func TestFramesCollectionShouldCorrectlyHandleAccess(t *testing.T) {
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

	assert.Panics(t, func() {
		collection.Count()
	})

	assert.Panics(t, func() {
		collection.GetAll()
	})

	collection.Lock()

	frames := collection.GetAll()
	assert.NotNil(t, frames)
	assert.Len(t, frames, 1)

	count := collection.Count()
	assert.Equal(t, 1, count)
}
