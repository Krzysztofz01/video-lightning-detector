package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/printer"
)

var (
	LogLevel        options.LogLevel
	DetectorOptions options.DetectorOptions = options.GetDefaultDetectorOptions()
)

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().VarP(&LogLevel, "log-level", "l", "The verbosity of the log messages printed to the standard output.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.AutoThresholds,
		"auto-thresholds", "a",
		DetectorOptions.AutoThresholds,
		"Automatic determination of thresholds after video analysis. The specified thresholds will overwrite those determined.")

	rootCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.ColorDifferenceDetectionThreshold,
		"color-difference-threshold", "c",
		DetectorOptions.ColorDifferenceDetectionThreshold,
		"The threshold used to determine the difference between two neighbouring frames on the color basis. See the documentation for more information on detection threshold values.")

	rootCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.BinaryThresholdDifferenceDetectionThreshold,
		"binary-threshold-difference-threshold", "t",
		DetectorOptions.BinaryThresholdDifferenceDetectionThreshold,
		"The threshold used to determine the difference between two neighbouring frames after the binary thresholding segmentation process. See the documentation for more information on detection threshold values.")

	rootCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.BrightnessDetectionThreshold,
		"brightness-threshold", "b",
		DetectorOptions.BrightnessDetectionThreshold,
		"The threshold used to determine the brightness of the frame. See the documentation for more information on detection threshold values.")

	rootCmd.PersistentFlags().Int32VarP(
		&DetectorOptions.MovingMeanResolution,
		"moving-mean-resolution", "m",
		DetectorOptions.MovingMeanResolution,
		"Resolution of the moving mean used when determining the statistics of the analysed frames. Has a direct impact on the accuracy of detection.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.SkipFramesExport,
		"skip-frames-export", "f",
		DetectorOptions.SkipFramesExport,
		"Skipping the step in which positively classified frames are exported to image files.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.ExportCsvReport,
		"export-csv-report", "e",
		DetectorOptions.ExportCsvReport,
		"Export of reports in CSV format.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.ExportJsonReport,
		"export-json-report", "j",
		DetectorOptions.ExportJsonReport,
		"Export of reports in JSON format.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.ExportChartReport,
		"export-chart-report", "r",
		DetectorOptions.ExportChartReport,
		"Export of frame statistics as a chart in HTML format.")

	rootCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.FrameScalingFactor,
		"scaling-factor", "s",
		DetectorOptions.FrameScalingFactor,
		"Scaling factor for the frame size of the recording. Has a direct impact on the performance, quality and processing time of recordings.")

	denoiseValues := strings.Join(options.GetDenoiseAlgorithmValues(), ", ")
	rootCmd.PersistentFlags().VarP(
		&DetectorOptions.Denoise,
		"denoise", "n",
		fmt.Sprintf("The use of de-noising in the form of low-pass filters. Impact on the quality of weighting determination. Values: [ %s ]", denoiseValues))

	rootCmd.PersistentFlags().BoolVar(
		&DetectorOptions.ExportConfusionMatrix,
		"export-confusion-matrix",
		DetectorOptions.ExportConfusionMatrix,
		"Value indicating if the frames detection classification confusion matrix should be rendered.")

	rootCmd.PersistentFlags().StringVar(
		&DetectorOptions.ConfusionMatrixActualDetectionsExpression,
		"confusion-matrix-actual-detections-expression",
		DetectorOptions.ConfusionMatrixActualDetectionsExpression,
		"Expression indicating the range of frames that should be used as actual classification. Example: 4,5,8-10,12,14")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.ImportPreanalyzed,
		"import-preanalyzed", "p",
		DetectorOptions.ImportPreanalyzed,
		"Use the cached data associated with the video analysis or save it in case the video has not already been analysed.")

	rootCmd.PersistentFlags().BoolVar(
		&DetectorOptions.StrictExplicitThreshold,
		"strict-explicit-threshold",
		DetectorOptions.StrictExplicitThreshold,
		"Omit strict validation of detection threshold ranges.")

	rootCmd.PersistentFlags().StringVar(
		&DetectorOptions.DetectionBoundsExpression,
		"detection-bounds-expression",
		DetectorOptions.DetectionBoundsExpression,
		"An expression indicating consecutively the coordinates of the upper left point, width and height of the cutout (bounding box) of the recording to be processed.  Example: 0:0:100:200")

	scalingValues := strings.Join(options.GetScaleAlgorithmValues(), ", ")
	rootCmd.PersistentFlags().Var(
		&DetectorOptions.ScaleAlgorithm,
		"scaling-algorithm",
		fmt.Sprintf("Sampling interpolation algorithm to be used when scaling the video during analysis. Values: [ %s ]", scalingValues))
}

func Execute(args []string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stdout, "Unexpected failure: %s\n", err)
			os.Exit(1)
		}
	}()

	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stdout, "Failure: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

var rootCmd = &cobra.Command{
	Use:  "vld",
	Long: "A video analysis tool that allows to detect and export frames that have captured lightning strikes.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		printer.Configure(printer.PrinterConfig{
			UseColor:  true,
			LogLevel:  LogLevel,
			OutStream: os.Stdout,
		})
	},
}
