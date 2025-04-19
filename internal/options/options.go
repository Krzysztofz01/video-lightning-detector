package options

import (
	"github.com/Krzysztofz01/video-lightning-detector/internal/denoise"
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
	Denoise                                     denoise.Algorithm
	FrameScalingFactor                          float64
	ImportPreanalyzed                           bool
	StrictExplicitThreshold                     bool
	DetectionBoundsExpression                   string
	UseInternalFrameScaling                     bool
}

// Return a boolean value representing if the detector options are valid. If any validation errors occured
// a message will be stored in the string return value.
func (options *DetectorOptions) AreValid() (bool, string) {
	if options.StrictExplicitThreshold && (options.BrightnessDetectionThreshold < 0.0 || options.BrightnessDetectionThreshold > 1.0) {
		return false, "the frame brightness detection threshold must be between zero and one"
	}

	if options.StrictExplicitThreshold && (options.ColorDifferenceDetectionThreshold < 0.0 || options.ColorDifferenceDetectionThreshold > 1.0) {
		return false, "the frame color difference detection threshold must be between zero and one"
	}

	if options.StrictExplicitThreshold && (options.BinaryThresholdDifferenceDetectionThreshold < 0.0 || options.BinaryThresholdDifferenceDetectionThreshold > 1.0) {
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

	if options.MovingMeanResolution < 0 {
		return false, "the moving mean/stddev resolution can not be negative"
	}

	if !denoise.IsValidAlgorithm(options.Denoise) {
		return false, "the specified denoise algorithm is invalid"
	}

	if len(options.DetectionBoundsExpression) != 0 && !utils.IsBoundsExpressionValid(options.DetectionBoundsExpression) {
		return false, "the detection bounds expression has a invalid format"
	}

	return true, ""
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
		Denoise:                                     denoise.NoDenoise,
		FrameScalingFactor:                          0.5,
		ImportPreanalyzed:                           false,
		StrictExplicitThreshold:                     true,
		DetectionBoundsExpression:                   "",
		UseInternalFrameScaling:                     false,
	}
}
