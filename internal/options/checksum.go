package options

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

var byteOrder binary.ByteOrder = binary.LittleEndian

// TODO: This implementation can be replaced with a safer hashing algorithm (even though it is not used for security-critical purposes).
// Generate a SHA1, hex-encoded checksum of the non data dependent detector options
func CalculateChecksum(options DetectorOptions, inputFileFingerprint []byte) (string, error) {
	if ok, _ := options.AreValid(); !ok {
		return "", fmt.Errorf("options: failed to calculate the checksum for invalid options")
	}

	if inputFileFingerprint == nil {
		return "", fmt.Errorf("options: failed to calculate the checksum for invalid input file fingerprint")
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

	if err := binary.Write(buffer, byteOrder, inputFileFingerprint); err != nil {
		return "", fmt.Errorf("options: failed to binary encode the input file fingerprint: %w", err)
	}

	hash := sha1.Sum(buffer.Bytes())
	hashHex := hex.EncodeToString(hash[:])

	return hashHex, nil
}
