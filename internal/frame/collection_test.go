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

		factory := CreateFrameFactory(BinaryThresholdParam)

		for index := 0; index < c.Count; index += 1 {
			var (
				f   *Frame
				err error
			)

			if index == 0 {
				f, err = factory.CreateNewFrame(mockImage(color.White), nil)
			} else {
				f, err = factory.CreateNewFrame(mockImage(color.White), mockImage(color.White))
			}

			assert.NotNil(t, f)
			assert.Nil(t, err)

			err = collection.Push(f)
			assert.Nil(t, err)

			frames = append(frames, f)
		}

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

	ff := CreateFrameFactory(BinaryThresholdParam)
	f1, _ := ff.CreateNewFrame(mockImage(color.White), nil)
	f2, _ := ff.CreateNewFrame(mockImage(color.White), mockImage(color.White))

	// NOTE: nil frame
	err := collection.Push(nil)
	assert.NotNil(t, err)

	// NOTE: frame with invalid ordinal number
	err = collection.Push(f2)
	assert.NotNil(t, err)

	err = collection.Push(f1)
	assert.Nil(t, err)

	err = collection.Push(f1)
	assert.NotNil(t, err)
}
