package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCircularBufferShouldCreate(t *testing.T) {
	cbuffer := NewCircularBuffer[int](4)
	assert.NotNil(t, cbuffer)
	assert.Equal(t, 0, cbuffer.GetBufferCount())
	assert.Equal(t, 0, cbuffer.GetTotalCount())
	assert.Equal(t, false, cbuffer.IsSaturated())

	assert.Panics(t, func() {
		NewCircularBuffer[int](-1)
	})

	assert.Panics(t, func() {
		NewCircularBuffer[int](0)
	})
}

func TestSaturatedCircularBufferShouldCreate(t *testing.T) {
	values := []int{1, 2, 3, 4}

	cbuffer := NewSaturatedCircularBuffer[int](values)
	assert.NotNil(t, cbuffer)
	assert.Equal(t, len(values), cbuffer.GetBufferCount())
	assert.Equal(t, 0, cbuffer.GetTotalCount())
	assert.Equal(t, true, cbuffer.IsSaturated())

	assert.Panics(t, func() {
		NewSaturatedCircularBuffer[int]([]int{})
	})
}

func TestCircularBufferShouldPushAndAccessValues(t *testing.T) {
	cases := []struct {
		Capacity    int
		CountToPush int
	}{
		{1, 0},
		{1, 1},
		{1, 2},
		{5, 0},
		{5, 1},
		{5, 2},
		{5, 3},
		{5, 4},
		{5, 5},
		{5, 6},
		{5, 7},
		{5, 8},
		{5, 12},
	}

	for _, c := range cases {
		values := make([]float64, 0, c.CountToPush)
		cbuffer := NewCircularBuffer[float64](c.Capacity)

		for index := 0; index < c.CountToPush; index += 1 {
			value := float64(index)
			values = append(values, value)

			cbuffer.Push(value)

			expectedIsSaturated := index+1 >= c.Capacity
			assert.Equal(t, expectedIsSaturated, cbuffer.IsSaturated())
		}

		count := c.Capacity
		if count > c.CountToPush {
			count = c.CountToPush
		}

		assert.Equal(t, count, cbuffer.GetBufferCount())
		assert.Equal(t, c.CountToPush, cbuffer.GetTotalCount())

		for offset := 0; offset < count; offset += 1 {
			expectedValue := values[len(values)-count+offset]
			assert.NotNil(t, expectedValue)

			actualValueTail, err := cbuffer.GetTail(offset)
			assert.Nil(t, err)
			assert.NotNil(t, actualValueTail)
			assert.Equal(t, expectedValue, actualValueTail)

			actualValueHead, err := cbuffer.GetHead(count - offset - 1)
			assert.Nil(t, err)
			assert.NotNil(t, actualValueHead)
			assert.Equal(t, expectedValue, actualValueHead)
		}

		if c.CountToPush > 0 {
			expectedWindow := make([]float64, c.Capacity)
			for offset := 0; offset < c.Capacity; offset += 1 {
				index := len(values) - c.Capacity + offset
				if index >= 0 && index < len(values) {
					expectedWindow[offset] = values[index]
				}
			}

			assert.ElementsMatch(t, expectedWindow, cbuffer.PeekWindow())
		}
	}
}

func TestCircularBufferShouldPushPAndAccessValues(t *testing.T) {
	cases := []struct {
		Capacity    int
		CountToPush int
	}{
		{1, 1},
		{1, 2},
		{5, 1},
		{5, 2},
		{5, 3},
		{5, 4},
		{5, 5},
		{5, 6},
		{5, 7},
		{5, 8},
		{5, 12},
	}

	type bufferEntry struct {
		Value float64
	}

	for _, c := range cases {
		values := make([]float64, 0, c.CountToPush)

		saturationValues := make([]*bufferEntry, c.Capacity, c.Capacity)
		for index, _ := range saturationValues {
			saturationValues[index] = &bufferEntry{}
		}

		cbuffer := NewSaturatedCircularBuffer[*bufferEntry](saturationValues)

		for index := 0; index < c.CountToPush; index += 1 {
			value := float64(index)
			values = append(values, value)

			cbuffer.PushP().Value = value

			assert.True(t, cbuffer.IsSaturated())
		}

		count := c.Capacity
		if count > c.CountToPush {
			count = c.CountToPush
		}

		assert.Equal(t, c.Capacity, cbuffer.GetBufferCount())
		assert.Equal(t, c.CountToPush, cbuffer.GetTotalCount())

		for offset := 0; offset < count; offset += 1 {
			expectedValue := values[len(values)-count+offset]
			assert.NotNil(t, expectedValue)

			actualValueTail, err := cbuffer.GetTail(offset)
			assert.Nil(t, err)
			assert.NotNil(t, actualValueTail)
			assert.Equal(t, expectedValue, actualValueTail.Value)

			actualValueHead, err := cbuffer.GetHead(count - offset - 1)
			assert.Nil(t, err)
			assert.NotNil(t, actualValueHead)
			assert.Equal(t, expectedValue, actualValueHead.Value)
		}

		if c.CountToPush > 0 {
			expectedWindow := make([]float64, c.Capacity)
			for offset := 0; offset < c.Capacity; offset += 1 {
				index := len(values) - c.Capacity + offset
				if index >= 0 && index < len(values) {
					expectedWindow[offset] = values[index]
				}
			}

			actualWindow := make([]float64, c.Capacity)
			for index, actualWindowValue := range cbuffer.PeekWindow() {
				actualWindow[index] = actualWindowValue.Value
			}

			assert.ElementsMatch(t, expectedWindow, actualWindow)
		}
	}
}

func TestCircularBufferShouldNotGetForInvalidIndex(t *testing.T) {
	cbuffer := NewCircularBuffer[*string](4)
	assert.NotNil(t, cbuffer)

	frame, err := cbuffer.GetTail(-1)
	assert.Nil(t, frame)
	assert.NotNil(t, err)

	frame, err = cbuffer.GetTail(1)
	assert.Nil(t, frame)
	assert.NotNil(t, err)

	frame, err = cbuffer.GetHead(-1)
	assert.Nil(t, frame)
	assert.NotNil(t, err)

	frame, err = cbuffer.GetHead(1)
	assert.Nil(t, frame)
	assert.NotNil(t, err)
}
