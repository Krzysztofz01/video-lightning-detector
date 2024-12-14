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

func TestMinMaxShouldPanicForEmptyValueSet(t *testing.T) {
	assert.Panics(t, func() {
		MinMax([]float64{})
	})
}

func TestMinMaxShouldCalculateMinMaxForValueSet(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6}
	minExpected := 1.0
	maxExpected := 6.0

	const delta float64 = 1e-5

	minActual, maxActual := MinMax(values)

	assert.InDelta(t, minExpected, minActual, delta)
	assert.InDelta(t, maxExpected, maxActual, delta)
}

func TestMinIntShoudlReturnTheSmallerValues(t *testing.T) {
	cases := map[struct {
		x int
		y int
	}]int{
		{0, 1}:   0,
		{0, -1}:  -1,
		{1, 1}:   1,
		{1, 2}:   1,
		{-1, -2}: -2,
	}

	for c, expected := range cases {
		actual := MinInt(c.x, c.y)

		assert.Equal(t, expected, actual)
	}
}
