package options

import "github.com/Krzysztofz01/video-lightning-detector/internal/utils"

// Structure representing the options for the stream detector.
type StreamDetectorOptions struct {
	BrightnessDetectionThreshold                float64
	ColorDifferenceDetectionThreshold           float64
	BinaryThresholdDifferenceDetectionThreshold float64
	MovingMeanResolution                        int32
	Denoise                                     DenoiseAlgorithm
	FrameScalingFactor                          float64
	StrictExplicitThreshold                     bool
	DetectionBoundsExpression                   string
	ScaleAlgorithm                              ScaleAlgorithm
	FrameDetectionPlotResolution                int
	FrameDetectionPlotThreshold                 float64
	DiagnosticMode                              bool
}

// Return a boolean value representing if the stream detector options are valid. If any validation errors occured
// a message will be stored in the string return value.
func (options *StreamDetectorOptions) AreValid() (bool, string) {
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

	if options.MovingMeanResolution < 0 {
		return false, "the moving mean/stddev resolution can not be negative"
	}

	if !IsValidDenoiseAlgorithm(options.Denoise) {
		return false, "the specified denoise algorithm is invalid"
	}

	if len(options.DetectionBoundsExpression) != 0 && !utils.IsBoundsExpressionValid(options.DetectionBoundsExpression) {
		return false, "the detection bounds expression has a invalid format"
	}

	if !IsValidScaleAlgorithm(options.ScaleAlgorithm) {
		return false, "the specified scale algorithm is invalid"
	}

	if options.FrameDetectionPlotResolution <= 0 {
		return false, "the specified frame detection plot resolution must be greater than 0"
	}

	if options.FrameDetectionPlotThreshold < 0.0 || options.FrameDetectionPlotThreshold > 1.0 {
		return false, "the frame detection plot threshold must be between zero and one"
	}

	return true, ""
}

// Create a copy of the stream detector options
func (options *StreamDetectorOptions) Clone() StreamDetectorOptions {
	return StreamDetectorOptions{
		BrightnessDetectionThreshold:                options.BrightnessDetectionThreshold,
		ColorDifferenceDetectionThreshold:           options.ColorDifferenceDetectionThreshold,
		BinaryThresholdDifferenceDetectionThreshold: options.BinaryThresholdDifferenceDetectionThreshold,
		MovingMeanResolution:                        options.MovingMeanResolution,
		Denoise:                                     options.Denoise,
		FrameScalingFactor:                          options.FrameScalingFactor,
		StrictExplicitThreshold:                     options.StrictExplicitThreshold,
		DetectionBoundsExpression:                   options.DetectionBoundsExpression,
		ScaleAlgorithm:                              options.ScaleAlgorithm,
		FrameDetectionPlotResolution:                options.FrameDetectionPlotResolution,
		FrameDetectionPlotThreshold:                 options.FrameDetectionPlotThreshold,
		DiagnosticMode:                              options.DiagnosticMode,
	}
}

// Return the default stream detector options.
func GetDefaultStreamDetectorOptions() StreamDetectorOptions {
	return StreamDetectorOptions{
		BrightnessDetectionThreshold:                0.0,
		ColorDifferenceDetectionThreshold:           0.0,
		BinaryThresholdDifferenceDetectionThreshold: 0.0,
		MovingMeanResolution:                        50,
		Denoise:                                     NoDenoise,
		FrameScalingFactor:                          0.5,
		StrictExplicitThreshold:                     true,
		DetectionBoundsExpression:                   "",
		ScaleAlgorithm:                              Default,
		FrameDetectionPlotResolution:                25,
		FrameDetectionPlotThreshold:                 0.95,
		DiagnosticMode:                              false,
	}
}
