package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"

	"github.com/Krzysztofz01/video-lightning-detector/internal/detector"
	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/printer"
)

var (
	InputVideoStreamUrl   string
	StreamDetectorOptions options.StreamDetectorOptions = options.GetDefaultStreamDetectorOptions()
)

func init() {
	streamCmd.Flags().StringVarP(&InputVideoStreamUrl, "input-video-stream-url", "i", "", "Input video stream url to perform the lightning detection.")
	streamCmd.MarkPersistentFlagRequired("input-video-stream-url")

	streamCmd.PersistentFlags().Float64VarP(
		&StreamDetectorOptions.ColorDifferenceDetectionThreshold,
		"color-difference-threshold", "c",
		DetectorOptions.ColorDifferenceDetectionThreshold,
		"The threshold used to determine the difference between two neighbouring frames on the color basis. See the documentation for more information on detection threshold values.")

	streamCmd.PersistentFlags().Float64VarP(
		&StreamDetectorOptions.BinaryThresholdDifferenceDetectionThreshold,
		"binary-threshold-difference-threshold", "t",
		DetectorOptions.BinaryThresholdDifferenceDetectionThreshold,
		"The threshold used to determine the difference between two neighbouring frames after the binary thresholding segmentation process. See the documentation for more information on detection threshold values.")

	streamCmd.PersistentFlags().Float64VarP(
		&StreamDetectorOptions.BrightnessDetectionThreshold,
		"brightness-threshold", "b",
		DetectorOptions.BrightnessDetectionThreshold,
		"The threshold used to determine the brightness of the frame. See the documentation for more information on detection threshold values.")

	streamCmd.PersistentFlags().Int32VarP(
		&StreamDetectorOptions.MovingMeanResolution,
		"moving-mean-resolution", "m",
		DetectorOptions.MovingMeanResolution,
		"Resolution of the moving mean used when determining the statistics of the analysed frames. Has a direct impact on the accuracy of detection.")

	streamCmd.PersistentFlags().Float64VarP(
		&StreamDetectorOptions.FrameScalingFactor,
		"scaling-factor", "s",
		DetectorOptions.FrameScalingFactor,
		"Scaling factor for the frame size of the recording. Has a direct impact on the performance, quality and processing time of recordings.")

	denoiseValues := strings.Join(options.GetDenoiseAlgorithmValues(), ", ")
	streamCmd.PersistentFlags().VarP(
		&StreamDetectorOptions.Denoise,
		"denoise", "n",
		fmt.Sprintf("The use of de-noising in the form of low-pass filters. Impact on the quality of weighting determination. Values: [ %s ]", denoiseValues))

	streamCmd.PersistentFlags().BoolVar(
		&StreamDetectorOptions.StrictExplicitThreshold,
		"strict-explicit-threshold",
		DetectorOptions.StrictExplicitThreshold,
		"Omit strict validation of detection threshold ranges.")

	streamCmd.PersistentFlags().StringVar(
		&StreamDetectorOptions.DetectionBoundsExpression,
		"detection-bounds-expression",
		DetectorOptions.DetectionBoundsExpression,
		"An expression indicating consecutively the coordinates of the upper left point, width and height of the cutout (bounding box) of the recording to be processed.  Example: 0:0:100:200")

	scalingValues := strings.Join(options.GetScaleAlgorithmValues(), ", ")
	streamCmd.PersistentFlags().Var(
		&StreamDetectorOptions.ScaleAlgorithm,
		"scaling-algorithm",
		fmt.Sprintf("Sampling interpolation algorithm to be used when scaling the video during analysis. Values: [ %s ]", scalingValues))

	streamCmd.PersistentFlags().Float64Var(
		&StreamDetectorOptions.FrameDetectionPlotThreshold,
		"frame-detection-plot-threshold",
		StreamDetectorOptions.FrameDetectionPlotThreshold,
		"Binary threshold segmentation argument for the frame image processing during detection plot calculation.")

	streamCmd.PersistentFlags().IntVar(
		&StreamDetectorOptions.FrameDetectionPlotResolution,
		"frame-detection-plot-resolution",
		StreamDetectorOptions.FrameDetectionPlotResolution,
		"The resolution of the x and y axis of the frame image detection plot.")

	rootCmd.AddCommand(streamCmd)
}

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Perform the analysis and detection stages on continuous video stream.",
	Long:  "Perform the analysis and detection stages on continuous video stream.",
	RunE: func(cmd *cobra.Command, args []string) error {
		printer.Configure(printer.PrinterConfig{
			UseColor:     true,
			LogLevel:     LogLevel,
			OutStream:    os.Stdout,
			ParsableMode: true,
		})

		detectorInstance, err := detector.CreateStreamDetector(printer.Instance(), StreamDetectorOptions)
		if err != nil {
			return fmt.Errorf("cmd: failed to create the stream detector instance: %w", err)
		}

		if err := detectorInstance.Run(InputVideoStreamUrl); err != nil {
			return fmt.Errorf("cmd: stream detector run failed: %w", err)
		}

		return nil
	},
}
