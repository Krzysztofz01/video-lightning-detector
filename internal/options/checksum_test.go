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

	a, err = CalculateChecksum(GetDefaultDetectorOptions(), []byte{})

	assert.NotEmpty(t, a)
	assert.Nil(t, err)

	b, err = CalculateChecksum(GetDefaultDetectorOptions(), []byte{})

	assert.NotEmpty(t, b)
	assert.Nil(t, err)

	assert.Equal(t, a, b)
}

func TestDifferentOptionsShouldProduceDifferentChecksums(t *testing.T) {
	var (
		aFileFingerprint []byte          = []byte{0x00, 0x01, 0x02, 0x03}
		bFileFingerprint []byte          = []byte{0x04, 0x05, 0x06, 0x07}
		a                DetectorOptions = GetDefaultDetectorOptions()
		b                DetectorOptions = GetDefaultDetectorOptions()
		aChecksum        string
		bChecksum        string
		err              error
	)

	// NOTE: Changing options that do not alter the checksum with the same file
	a.ExportReport = true
	b.ExportReport = false

	aChecksum, err = CalculateChecksum(a, aFileFingerprint)
	assert.Nil(t, err)
	assert.NotEmpty(t, aChecksum)

	bChecksum, err = CalculateChecksum(b, aFileFingerprint)
	assert.Nil(t, err)
	assert.NotEmpty(t, bChecksum)

	assert.Equal(t, aChecksum, bChecksum)

	// NOTE: Changing options that do not alter the checksum with different file
	a.ExportReport = true
	b.ExportReport = false

	aChecksum, err = CalculateChecksum(a, aFileFingerprint)
	assert.Nil(t, err)
	assert.NotEmpty(t, aChecksum)

	bChecksum, err = CalculateChecksum(b, bFileFingerprint)
	assert.Nil(t, err)
	assert.NotEmpty(t, bChecksum)

	assert.NotEqual(t, aChecksum, bChecksum)

	// NOTE: Changing options that alter the checksum with the same file
	a.AutoThresholds = true
	b.AutoThresholds = false

	aChecksum, err = CalculateChecksum(a, aFileFingerprint)
	assert.Nil(t, err)
	assert.NotEmpty(t, aChecksum)

	bChecksum, err = CalculateChecksum(b, aFileFingerprint)
	assert.Nil(t, err)
	assert.NotEmpty(t, bChecksum)

	assert.NotEqual(t, aChecksum, bChecksum)

	// NOTE: Changing options that alter the checksum with different file
	a.AutoThresholds = true
	b.AutoThresholds = false

	aChecksum, err = CalculateChecksum(a, aFileFingerprint)
	assert.Nil(t, err)
	assert.NotEmpty(t, aChecksum)

	bChecksum, err = CalculateChecksum(b, bFileFingerprint)
	assert.Nil(t, err)
	assert.NotEmpty(t, bChecksum)

	assert.NotEqual(t, aChecksum, bChecksum)
}

func TestInvalidOptionsShouldNotCalculateChecksum(t *testing.T) {
	// NOTE: Invalid options
	options := GetDefaultDetectorOptions()
	options.Denoise = -1

	valid, _ := options.AreValid()
	assert.False(t, valid)

	checksum, err := CalculateChecksum(options, []byte{})
	assert.NotNil(t, err)
	assert.Empty(t, checksum)

	// NOTE: Invalid input file checksum
	options = GetDefaultDetectorOptions()

	valid, _ = options.AreValid()
	assert.True(t, valid)

	checksum, err = CalculateChecksum(options, nil)
	assert.NotNil(t, err)
	assert.Empty(t, checksum)
}
