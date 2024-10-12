package detector

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

const FrameCollectionCacheFilename string = "video-lightning-detector.cache"

// Structure representing the options for the detector.
type DetectorOptions struct {
	AutoThresholds                              bool
	BrightnessDetectionThreshold                float64
	ColorDifferenceDetectionThreshold           float64
	BinaryThresholdDifferenceDetectionThreshold float64
	MovingMeanResolution                        int32
	ExportCsvReport                             bool
	ExportJsonReport                            bool
	ExportChartReport                           bool
	ExportConfusionMatrix                       bool
	ConfusionMatrixActualDetectionsExpression   string
	SkipFramesExport                            bool
	Denoise                                     bool
	FrameScalingFactor                          float64
	ImportPreanalyzed                           bool
}

// Return a boolean value representing if the detector options are valid. If any validation errors occured
// a message will be stored in the string return value.
// TODO: MovingMeanResolution validation >1
func (options *DetectorOptions) AreValid() (bool, string) {
	if options.BrightnessDetectionThreshold < 0.0 || options.BrightnessDetectionThreshold > 1.0 {
		return false, "the frame brightness detection threshold must be between zero and one"
	}

	if options.ColorDifferenceDetectionThreshold < 0.0 || options.ColorDifferenceDetectionThreshold > 1.0 {
		return false, "the frame color difference detection threshold must be between zero and one"
	}

	if options.BinaryThresholdDifferenceDetectionThreshold < 0.0 || options.BinaryThresholdDifferenceDetectionThreshold > 1.0 {
		return false, "the frame binary threshold difference detection threshold must be between zero and one"
	}

	if options.FrameScalingFactor < 0.0 || options.FrameScalingFactor > 1.0 {
		return false, "the scaling factor must be between zero and one"
	}

	if options.ExportConfusionMatrix && len(options.ConfusionMatrixActualDetectionsExpression) == 0 {
		return false, "the confusion matrix actual detections expressions must be specified to export the confusion matrix"
	}

	if len(options.ConfusionMatrixActualDetectionsExpression) != 0 && !utils.IsRangeExpressionValid(options.ConfusionMatrixActualDetectionsExpression) {
		return false, "the confusion matrix actual detections expression has a invalid format"
	}

	return true, ""
}

func (options *DetectorOptions) GetChecksum() (string, error) {
	optionsCopy := *options

	if optionsCopy.AutoThresholds {
		defaultOptions := GetDefaultDetectorOptions()
		optionsCopy.BinaryThresholdDifferenceDetectionThreshold = defaultOptions.BinaryThresholdDifferenceDetectionThreshold
		optionsCopy.BrightnessDetectionThreshold = defaultOptions.BrightnessDetectionThreshold
		optionsCopy.ColorDifferenceDetectionThreshold = defaultOptions.ColorDifferenceDetectionThreshold
	}

	encodingBuffer := new(bytes.Buffer)
	encoder := json.NewEncoder(encodingBuffer)

	if err := encoder.Encode(optionsCopy); err != nil {
		return "", fmt.Errorf("detector: failed to encode the options to json: %w", err)
	}

	hash := sha1.New()
	if _, err := hash.Write(encodingBuffer.Bytes()); err != nil {
		return "", fmt.Errorf("detector: failed to hash the ddetector options: %w", err)
	}

	str := hex.EncodeToString(hash.Sum(nil))

	fmt.Printf("Hash: %s | %s\n", str, encodingBuffer.String())

	return str, nil
}

// Return the default detector options.
func GetDefaultDetectorOptions() DetectorOptions {
	return DetectorOptions{
		AutoThresholds:                              false,
		BrightnessDetectionThreshold:                0.0,
		ColorDifferenceDetectionThreshold:           0.0,
		BinaryThresholdDifferenceDetectionThreshold: 0.0,
		MovingMeanResolution:                        50,
		ExportCsvReport:                             false,
		ExportJsonReport:                            false,
		ExportChartReport:                           false,
		ExportConfusionMatrix:                       false,
		ConfusionMatrixActualDetectionsExpression:   "",
		SkipFramesExport:                            false,
		Denoise:                                     false,
		FrameScalingFactor:                          0.5,
	}
}
