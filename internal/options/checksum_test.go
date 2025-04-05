package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: Implement more tests

func TestDefaultOptionsShouldProduceTheSameChecksum(t *testing.T) {
	var (
		a, b string
		err  error
	)

	a, err = CalculateChecksum(GetDefaultDetectorOptions())

	assert.NotEmpty(t, a)
	assert.Nil(t, err)

	b, err = CalculateChecksum(GetDefaultDetectorOptions())

	assert.NotEmpty(t, b)
	assert.Nil(t, err)

	assert.Equal(t, a, b)
}
