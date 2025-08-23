package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: Implement more test cases
func TestParseRangeExpressionShouldCorrectlyParseExpression(t *testing.T) {
	cases := map[string][]int{
		"1,2,3,4-8,10": {1, 2, 3, 4, 5, 6, 7, 8, 10},
		"30-37":        {30, 31, 32, 33, 34, 35, 36, 37},
	}

	for expression, expected := range cases {
		assert.True(t, IsRangeExpressionValid(expression))

		actual, err := ParseRangeExpression(expression)

		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	}
}

// TODO: Implement more test cases
func TestParseBoundsExpressionShouldCorrectlyParseExpression(t *testing.T) {
	cases := map[string]struct {
		X, Y, W, H int
	}{
		"0:0:100:100": {0, 0, 100, 100},
		"3:6:9:12":    {3, 6, 9, 12},
	}

	for expression, expected := range cases {
		assert.True(t, IsBoundsExpressionValid(expression))

		x, y, w, h, err := ParseBoundsExpression(expression)

		assert.Nil(t, err)
		assert.Equal(t, x, expected.X)
		assert.Equal(t, y, expected.Y)
		assert.Equal(t, w, expected.W)
		assert.Equal(t, h, expected.H)
	}
}
