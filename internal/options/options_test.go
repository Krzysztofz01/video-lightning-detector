package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldValidateDefaultOptions(t *testing.T) {
	options := GetDefaultDetectorOptions()

	valid, msg := options.AreValid()
	assert.True(t, valid)
	assert.Empty(t, msg)
}

func TestShouldNotValidateInvalidBrightnessDetectionThreshold(t *testing.T) {
	cases := []float64{-0.1, 1.1}

	for _, value := range cases {
		options := GetDefaultDetectorOptions()
		options.BrightnessDetectionThreshold = value

		valid, msg := options.AreValid()
		assert.False(t, valid)
		assert.NotEmpty(t, msg)
	}
}

func TestShouldNotValidateInvalidColorDifferenceDetectionThreshold(t *testing.T) {
	cases := []float64{-0.1, 1.1}

	for _, value := range cases {
		options := GetDefaultDetectorOptions()
		options.ColorDifferenceDetectionThreshold = value

		valid, msg := options.AreValid()
		assert.False(t, valid)
		assert.NotEmpty(t, msg)
	}
}

func TestShouldNotValidateInvalidBinaryThresholdDetectionThreshold(t *testing.T) {
	cases := []float64{-0.1, 1.1}

	for _, value := range cases {
		options := GetDefaultDetectorOptions()
		options.BinaryThresholdDifferenceDetectionThreshold = value

		valid, msg := options.AreValid()
		assert.False(t, valid)
		assert.NotEmpty(t, msg)
	}
}

func TestShouldNotValidateInvalidFrameScalingFactor(t *testing.T) {
	cases := []float64{-0.1, 1.1}

	for _, value := range cases {
		options := GetDefaultDetectorOptions()
		options.FrameScalingFactor = value

		valid, msg := options.AreValid()
		assert.False(t, valid)
		assert.NotEmpty(t, msg)
	}
}
