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

	for expresion, expected := range cases {
		actual, err := ParseRangeExpression(expresion)

		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	}
}
