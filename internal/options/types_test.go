package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidDenoiseAlgorithmShouldReturnCorrectBoolean(t *testing.T) {
	cases := map[DenoiseAlgorithm]bool{
		NoDenoise:   true,
		StackBlur8:  true,
		StackBlur16: true,
		StackBlur32: true,
		-1:          false,
	}

	for algorithm, expected := range cases {
		actual := IsValidDenoiseAlgorithm(algorithm)

		assert.Equal(t, expected, actual)
	}
}

func TestIsValidScaleAlgorithmShouldReturnCorrectBoolean(t *testing.T) {
	cases := map[ScaleAlgorithm]bool{
		Default:          true,
		Bilinear:         true,
		Bicubic:          true,
		NearestNeighbour: true,
		Lanczos:          true,
		Area:             true,
		-1:               false,
	}

	for algorithm, expected := range cases {
		actual := IsValidScaleAlgorithm(algorithm)

		assert.Equal(t, expected, actual)
	}
}
