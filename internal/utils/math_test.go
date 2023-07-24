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
