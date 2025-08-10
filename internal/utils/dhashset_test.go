package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecayingHashSetShouldCreate(t *testing.T) {
	set := NewDecayingHashSet[int](2)
	assert.NotNil(t, set)

	assert.Panics(t, func() {
		NewDecayingHashSet[int](0)
	})

	assert.Panics(t, func() {
		NewDecayingHashSet[int](-1)
	})
}

func TestDecayingHashSetShouldAddAndTellIfContains(t *testing.T) {
	values := []int{
		1, 2,
		3, 4,
		5, 6,

		7, 8,

		9, 9, 1,
		2, 3,

		4, 5,

		6, 7,
		8, 9, 9,

		1, 1, 1, 1, 2, 2,

		3, 3, 3, 3,
	}

	expectedValues := []int{
		1, 2, 3,
	}

	var (
		decayCount int                  = 2
		set        DecayingHashSet[int] = NewDecayingHashSet[int](decayCount)
	)

	setImpl, ok := set.(*decayingHashSet[int])
	assert.True(t, ok)

	for index, value := range values {
		set.Add(value)

		assert.True(t, set.Contains(value), "%d", index)
		assert.False(t, set.Contains(-value), "%d", index)

		assert.Equal(t, decayCount, setImpl.DecayCount)
		assert.LessOrEqual(t, len(setImpl.DataA), decayCount*4)
		assert.LessOrEqual(t, len(setImpl.DataB), decayCount*4)
	}

	assert.ElementsMatch(t, expectedValues, set.Values())
}
