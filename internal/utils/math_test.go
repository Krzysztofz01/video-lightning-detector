package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeanShouldPanicForEmptyValueSet(t *testing.T) {
	assert.Panics(t, func() {
		Mean([]float64{})
	})
}

func TestMeanShouldCalculateMeanForValueSet(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6}
	expected := 3.5

	const delta float64 = 1e-7

	actual := Mean(values)

	assert.InDelta(t, expected, actual, delta)
}

func TestMovingMeanShouldPanicForEmptyValueSet(t *testing.T) {
	assert.Panics(t, func() {
		MovingMean([]float64{}, 0, 2)
	})
}

func TestMovingMeanShouldPanicForPositionOutOfBounds(t *testing.T) {
	assert.Panics(t, func() {
		MovingMean([]float64{1, 2}, 2, 2)
	})
}

func TestMovingMeanShouldCalculateMeanForValueSet(t *testing.T) {
	cases := []struct {
		set      []float64
		position int
		bias     int
		expected float64
	}{
		{[]float64{3, 2, 1, 2, 3, 4}, 0, 1, 2.5},
		{[]float64{3, 2, 1, 2, 3, 4}, 1, 1, 2.0},
		{[]float64{3, 2, 1, 2, 3, 4}, 2, 1, 1.666667},
		{[]float64{3, 2, 1, 2, 3, 4}, 3, 1, 2.0},
		{[]float64{3, 2, 1, 2, 3, 4}, 4, 1, 3.0},
		{[]float64{3, 2, 1, 2, 3, 4}, 5, 1, 3.5},
	}

	const delta float64 = 1e-5
	for _, c := range cases {
		actual := MovingMean(c.set, c.position, c.bias)

		assert.InDelta(t, c.expected, actual, delta)
	}
}

func TestStandardDeviationShouldPanicForEmptyValueSet(t *testing.T) {
	assert.Panics(t, func() {
		StandardDeviation([]float64{})
	})
}

func TestStandardDeviationShouldCalculateStandardDeviationForValueSet(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6}
	expected := 1.70782

	const delta float64 = 1e-5

	actual := StandardDeviation(values)

	assert.InDelta(t, expected, actual, delta)
}

func TestMaxShouldPanicForEmptyValueSet(t *testing.T) {
	assert.Panics(t, func() {
		Max([]float64{})
	})
}

func TestMaxShouldCalculateMaxForValueSet(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6}
	expected := 6.0

	const delta float64 = 1e-5

	actual := Max(values)

	assert.InDelta(t, expected, actual, delta)
}
