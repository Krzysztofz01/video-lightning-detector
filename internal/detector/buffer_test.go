package detector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectionBufferShouldCreate(t *testing.T) {
	buffer := CreateDetectionBuffer()

	assert.NotNil(t, buffer)
}

func TestDetectionBufferShouldCorrectlyResolveAppendedValues(t *testing.T) {
	cases := []struct {
		detections []bool
		expected   []int
	}{
		{[]bool{false, false, false, false}, []int{}},
		{[]bool{false, true, false, false}, []int{1}},
		{[]bool{true, false, false, false}, []int{0}},
		{[]bool{true, true, false, false}, []int{0, 1}},
		{[]bool{false, false, true, false}, []int{2}},
		{[]bool{false, true, true, false}, []int{1, 2}},
		{[]bool{true, false, true, false}, []int{0, 1, 2}},
		{[]bool{true, true, true, false}, []int{0, 1, 2}},
		{[]bool{false, false, false, true}, []int{3}},
		{[]bool{false, true, false, true}, []int{1, 2, 3}},
		{[]bool{true, false, false, true}, []int{0, 1, 2, 3}},
		{[]bool{true, true, false, true}, []int{0, 1, 2, 3}},
		{[]bool{false, false, true, true}, []int{2, 3}},
		{[]bool{false, true, true, true}, []int{1, 2, 3}},
		{[]bool{true, false, true, true}, []int{0, 1, 2, 3}},
		{[]bool{true, true, true, true}, []int{0, 1, 2, 3}},
		{[]bool{false, false, true, true, false}, []int{2, 3}},
		{[]bool{false, true, true, true, false}, []int{1, 2, 3}},
		{[]bool{true, false, true, true, false}, []int{0, 1, 2, 3}},
		{[]bool{true, true, true, true, false}, []int{0, 1, 2, 3}},
	}

	for _, c := range cases {
		buffer := CreateDetectionBuffer()
		for frameIndex, detected := range c.detections {
			buffer.Append(frameIndex, detected, detected, detected)
		}

		actual := buffer.ResolveClassifiedIndex()

		assert.NotNil(t, actual)
		assert.Equal(t, c.expected, actual)
	}
}
