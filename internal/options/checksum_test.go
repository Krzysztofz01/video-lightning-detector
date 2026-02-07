package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestDifferentOptionsShouldProduceDifferentChecksums(t *testing.T) {
	var (
		a         DetectorOptions = GetDefaultDetectorOptions()
		b         DetectorOptions = GetDefaultDetectorOptions()
		aChecksum string
		bChecksum string
		err       error
	)

	// NOTE: Changing options that do not alter the checksum
	a.ExportReport = true
	b.ExportReport = false

	aChecksum, err = CalculateChecksum(a)
	assert.Nil(t, err)
	assert.NotEmpty(t, aChecksum)

	bChecksum, err = CalculateChecksum(b)
	assert.Nil(t, err)
	assert.NotEmpty(t, bChecksum)

	assert.Equal(t, aChecksum, bChecksum)

	// NOTE: Changing options that alter the checksum
	a.AutoThresholds = true
	b.AutoThresholds = false

	aChecksum, err = CalculateChecksum(a)
	assert.Nil(t, err)
	assert.NotEmpty(t, aChecksum)

	bChecksum, err = CalculateChecksum(b)
	assert.Nil(t, err)
	assert.NotEmpty(t, bChecksum)

	assert.NotEqual(t, aChecksum, bChecksum)

	// NOTE: Restoring the options that alter the checksum
	a.AutoThresholds = false
	b.AutoThresholds = false

	aChecksum, err = CalculateChecksum(a)
	assert.Nil(t, err)
	assert.NotEmpty(t, aChecksum)

	bChecksum, err = CalculateChecksum(b)
	assert.Nil(t, err)
	assert.NotEmpty(t, bChecksum)

	assert.Equal(t, aChecksum, bChecksum)
}

func TestInvalidOptionsShouldNotCalculateChecksum(t *testing.T) {
	options := GetDefaultDetectorOptions()
	options.Denoise = -1

	valid, _ := options.AreValid()
	assert.False(t, valid)

	checksum, err := CalculateChecksum(options)
	assert.NotNil(t, err)
	assert.Empty(t, checksum)
}
