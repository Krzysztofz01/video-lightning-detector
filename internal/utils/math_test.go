package utils

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeanStdDevShouldPanicForEmptyValueSet(t *testing.T) {
	assert.Panics(t, func() {
		MeanStdDev([]float64{})
	})
}

func TestMeanStdDevShouldCalculateMeanAndStdDevForValueSet(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6}
	meanExpected := 3.5
	stdDevExpected := 1.70782

	const delta float64 = 1e-5

	meanActual, stdDevActual := MeanStdDev(values)

	assert.InDelta(t, meanExpected, meanActual, delta)
	assert.InDelta(t, stdDevExpected, stdDevActual, delta)
}

func TestMeanStdDevIncShouldPanicForNegativeLength(t *testing.T) {
	assert.Panics(t, func() {
		MeanStdDevInc(1, 1, 0, -1)
	})
}

func TestMeanStdDevIncShouldCalculateMeanAndStdDevForValueSet(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6}
	meanExpected := 3.5
	stdDevExpected := 1.70782

	const delta float64 = 1e-5

	var (
		meanActual   float64 = 0.0
		stdDevActual float64 = 0.0
	)

	for index, value := range values {
		meanActual, stdDevActual = MeanStdDevInc(value, meanActual, stdDevActual, index)
	}

	assert.InDelta(t, meanExpected, meanActual, delta)
	assert.InDelta(t, stdDevExpected, stdDevActual, delta)
}

func TestMovingMeanStdDevShouldPanicForEmptyValueSet(t *testing.T) {
	assert.Panics(t, func() {
		MovingMeanStdDev([]float64{}, 0, 2)
	})
}

func TestMovingMeanStdDevShouldPanicForPositionOutOfBounds(t *testing.T) {
	assert.Panics(t, func() {
		MovingMeanStdDev([]float64{1, 2}, 2, 2)
	})
}

func TestMovingMeanShouldCalculateMeanForValueSet(t *testing.T) {
	cases := []struct {
		set            []float64
		position       int
		bias           int
		meanExpected   float64
		stdDevExpected float64
	}{
		{[]float64{3, 2, 1, 2, 3, 4}, 0, 1, 2.5, 0.50000},
		{[]float64{3, 2, 1, 2, 3, 4}, 1, 1, 2.0, 0.81649},
		{[]float64{3, 2, 1, 2, 3, 4}, 2, 1, 1.666667, 0.471404},
		{[]float64{3, 2, 1, 2, 3, 4}, 3, 1, 2.0, 0.816496},
		{[]float64{3, 2, 1, 2, 3, 4}, 4, 1, 3.0, 0.816496},
		{[]float64{3, 2, 1, 2, 3, 4}, 5, 1, 3.5, 0.5},

		{[]float64{3, 2, 1, 2, 3, 4}, 0, 0, 3, 0},
		{[]float64{3, 2, 1, 2, 3, 4}, 1, 0, 2, 0},
		{[]float64{3, 2, 1, 2, 3, 4}, 2, 0, 1, 0},
		{[]float64{3, 2, 1, 2, 3, 4}, 3, 0, 2, 0},
		{[]float64{3, 2, 1, 2, 3, 4}, 4, 0, 3, 0},
		{[]float64{3, 2, 1, 2, 3, 4}, 5, 0, 4, 0},
	}

	const delta float64 = 1e-5
	for _, c := range cases {
		meanActual, stdDevActual := MovingMeanStdDev(c.set, c.position, c.bias)

		assert.InDelta(t, c.meanExpected, meanActual, delta)
		assert.InDelta(t, c.stdDevExpected, stdDevActual, delta)
	}
}

func TestMinMaxShouldPanicForEmptyValueSet(t *testing.T) {
	assert.Panics(t, func() {
		MinMax([]float64{})
	})
}

func TestMinMaxShouldCalculateMinMaxForValueSet(t *testing.T) {
	values := []float64{2, 3, 4, 5, 6, 1}
	minExpected := 1.0
	maxExpected := 6.0

	const delta float64 = 1e-5

	minActual, maxActual := MinMax(values)

	assert.InDelta(t, minExpected, minActual, delta)
	assert.InDelta(t, maxExpected, maxActual, delta)
}

func TestMinMaxIncShouldCalculateMinMaxForValueSet(t *testing.T) {
	values := []float64{2, 3, 4, 5, 6, 1}
	minExpected := 1.0
	maxExpected := 6.0

	const delta float64 = 1e-5

	var (
		minActual float64 = math.Inf(+1)
		maxActual float64 = math.Inf(-1)
	)

	for _, value := range values {
		minActual, maxActual = MinMaxInc(value, minActual, maxActual)
	}

	assert.InDelta(t, minExpected, minActual, delta)
	assert.InDelta(t, maxExpected, maxActual, delta)
}

func TestMinIntShouldReturnTheSmallerValues(t *testing.T) {
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
func TestDivShouldDivAndUseFallback(t *testing.T) {
	cases := []struct {
		a        float64
		b        float64
		fallback float64
		expected float64
	}{
		{1, 1, 0, 1 / 1},
		{1, -1, 0, 1 / -1},
		{1, 0, 0, 0},
		{1, 0, 1, 1},
	}

	for _, c := range cases {
		actual := Div(c.a, c.b, c.fallback)

		assert.Equal(t, c.expected, actual)
	}
}
