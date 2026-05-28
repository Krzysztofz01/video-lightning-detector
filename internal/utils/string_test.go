package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestParseRangeExpressionShouldFailForInvalidExpression(t *testing.T) {
	cases := []string{
		"-1,1,2",
		"--1",
		"3-2",
		"0,-1",
		",",
	}

	for _, c := range cases {
		r, err := ParseRangeExpression(c)
		assert.NotNil(t, err)
		assert.Nil(t, r)
	}
}

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

func TestParseBoundsExpressionShouldFailForInvalidExpression(t *testing.T) {
	cases := []string{
		"0:0:0:0:",
		":0:0:0:0:",
		"::::",
		"-1:-1:-1:-1",
	}

	for _, c := range cases {
		_, _, _, _, err := ParseBoundsExpression(c)
		assert.NotNil(t, err)
	}
}
