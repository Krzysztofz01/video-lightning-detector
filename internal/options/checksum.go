package options

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

var byteOrder binary.ByteOrder = binary.LittleEndian

// Generate a SHA1, hex-encoded checksum of the non data dependent detector options
func CalculateChecksum(options DetectorOptions) (string, error) {
	if ok, _ := options.AreValid(); !ok {
		return "", fmt.Errorf("options: failed to calculate the checksum for invalid options")
	}

	buffer := &bytes.Buffer{}

	if !options.AutoThresholds {
		if err := binary.Write(buffer, byteOrder, options.BinaryThresholdDifferenceDetectionThreshold); err != nil {
			return "", fmt.Errorf("options: failed to binary encode the BinaryThresholdDifferenceDetectionThreshold: %w", err)
		}

		if err := binary.Write(buffer, byteOrder, options.BrightnessDetectionThreshold); err != nil {
			return "", fmt.Errorf("options: failed to binary encode the BrightnessDetectionThreshold: %w", err)
		}

		if err := binary.Write(buffer, byteOrder, options.ColorDifferenceDetectionThreshold); err != nil {
			return "", fmt.Errorf("options: failed to binary encode the ColorDifferenceDetectionThreshold: %w", err)
		}
	}

	if err := binary.Write(buffer, byteOrder, int64(options.Denoise)); err != nil {
		return "", fmt.Errorf("options: failed to binary encode the Denoise: %w", err)
	}

	if err := binary.Write(buffer, byteOrder, options.FrameScalingFactor); err != nil {
		return "", fmt.Errorf("options: failed to binary encode the FrameScalingFactor: %w", err)
	}

	hash := sha1.Sum(buffer.Bytes())
	hashHex := hex.EncodeToString(hash[:])

	return hashHex, nil
}
