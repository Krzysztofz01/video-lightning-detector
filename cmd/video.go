package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Krzysztofz01/video-lightning-detector/internal/detector"
	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/printer"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

var (
	InputVideoPath      string
	OutputDirectoryPath string
	DetectorOptions     options.DetectorOptions = options.GetDefaultDetectorOptions()
)

const (
	InputVideoPathFlagName      string = "input-video-path"
	OutputDirectoryPathFlagName string = "output-directory-path"
)

func init() {
	videoCmd.Flags().StringVarP(&InputVideoPath, InputVideoPathFlagName, "i", "", "Input video to perform the lightning detection.")
	_ = videoCmd.MarkFlagRequired(InputVideoPathFlagName)
	_ = videoCmd.MarkFlagFilename(InputVideoPathFlagName)

	videoCmd.Flags().StringVarP(&OutputDirectoryPath, OutputDirectoryPathFlagName, "o", "", "Output directory path for export artifacts such as frames and reports in selected formats. By default, the input video name is used to derive the output directory name.")
	_ = videoCmd.MarkFlagDirname(OutputDirectoryPathFlagName)

	videoCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.AutoThresholds,
		"auto-thresholds", "a",
		DetectorOptions.AutoThresholds,
		"Automatic determination of thresholds after video analysis. The specified thresholds will overwrite those determined.")

	videoCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.ColorDifferenceDetectionThreshold,
		"color-difference-threshold", "c",
		DetectorOptions.ColorDifferenceDetectionThreshold,
		"The threshold used to determine the difference between two neighbouring frames on the color basis. See the documentation for more information on detection threshold values.")

	videoCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.BinaryThresholdDifferenceDetectionThreshold,
		"binary-threshold-difference-threshold", "t",
		DetectorOptions.BinaryThresholdDifferenceDetectionThreshold,
		"The threshold used to determine the difference between two neighbouring frames after the binary thresholding segmentation process. See the documentation for more information on detection threshold values.")

	videoCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.BrightnessDetectionThreshold,
		"brightness-threshold", "b",
		DetectorOptions.BrightnessDetectionThreshold,
		"The threshold used to determine the brightness of the frame. See the documentation for more information on detection threshold values.")

	videoCmd.PersistentFlags().Int32VarP(
		&DetectorOptions.MovingMeanResolution,
		"moving-mean-resolution", "m",
		DetectorOptions.MovingMeanResolution,
		"Resolution of the moving mean used when determining the statistics of the analysed frames. Has a direct impact on the accuracy of detection.")

	videoCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.SkipFramesExport,
		"skip-frames-export", "f",
		DetectorOptions.SkipFramesExport,
		"Skipping the step in which positively classified frames are exported to image files.")

	videoCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.ExportCsvReport,
		"export-csv-report", "e",
		DetectorOptions.ExportCsvReport,
		"Export of reports in CSV format.")

	_ = videoCmd.PersistentFlags().MarkDeprecated("export-csv-report", "The CSV export is deprecated and will be removed in the future")

	videoCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.ExportJsonReport,
		"export-json-report", "j",
		DetectorOptions.ExportJsonReport,
		"Export of reports in JSON format.")

	_ = videoCmd.PersistentFlags().MarkDeprecated("export-json-report", "This flag is deprecated and will be removed in the future, use the -r flag instead")

	videoCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.FrameScalingFactor,
		"scaling-factor", "s",
		DetectorOptions.FrameScalingFactor,
		"Scaling factor for the frame size of the recording. Has a direct impact on the performance, quality and processing time of recordings.")

	denoiseValues := strings.Join(options.GetDenoiseAlgorithmValues(), ", ")
	videoCmd.PersistentFlags().VarP(
		&DetectorOptions.Denoise,
		"denoise", "n",
		fmt.Sprintf("The use of de-noising in the form of low-pass filters. Impact on the quality of weighting determination. Values: [ %s ]", denoiseValues))

	videoCmd.PersistentFlags().BoolVar(
		&DetectorOptions.ExportConfusionMatrix,
		"export-confusion-matrix",
		DetectorOptions.ExportConfusionMatrix,
		"Value indicating if the frames detection classification confusion matrix should be rendered.")

	_ = videoCmd.PersistentFlags().MarkDeprecated("export-confusion-matrix", "This flag is deprecated and will be removed in the future. The confusion matrix results are exported by default.")

	videoCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.ExportReport,
		"export-report", "r",
		DetectorOptions.ExportReport,
		"Export a full report in JSON format.")

	videoCmd.PersistentFlags().StringVar(
		&DetectorOptions.ConfusionMatrixActualDetectionsExpression,
		"confusion-matrix-actual-detections-expression",
		DetectorOptions.ConfusionMatrixActualDetectionsExpression,
		"Expression indicating the range of frames that should be used as actual classification. Example: 4,5,8-10,12,14")

	videoCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.ImportPreanalyzed,
		"import-preanalyzed", "p",
		DetectorOptions.ImportPreanalyzed,
		"Use the cached data associated with the video analysis or save it in case the video has not already been analysed.")

	videoCmd.PersistentFlags().BoolVar(
		&DetectorOptions.StrictExplicitThreshold,
		"strict-explicit-threshold",
		DetectorOptions.StrictExplicitThreshold,
		"Omit strict validation of detection threshold ranges.")

	videoCmd.PersistentFlags().StringVar(
		&DetectorOptions.DetectionBoundsExpression,
		"detection-bounds-expression",
		DetectorOptions.DetectionBoundsExpression,
		"An expression indicating consecutively the coordinates of the upper left point, width and height of the cutout (bounding box) of the recording to be processed.  Example: 0:0:100:200")

	scalingValues := strings.Join(options.GetScaleAlgorithmValues(), ", ")
	videoCmd.PersistentFlags().Var(
		&DetectorOptions.ScaleAlgorithm,
		"scaling-algorithm",
		fmt.Sprintf("Sampling interpolation algorithm to be used when scaling the video during analysis. Values: [ %s ]", scalingValues))

	videoCmd.PersistentFlags().StringVar(
		&DetectorOptions.ExportFramesPrefix,
		"export-frames-prefix",
		DetectorOptions.ExportFramesPrefix,
		"Name prefix for exported image files.")

	rootCmd.AddCommand(videoCmd)
}

var videoCmd = &cobra.Command{
	Use:   "video",
	Short: "Perform the analysis, detection and export stage on single video.",
	Long:  "Perform the analysis, detection and export stage on single video.",
	RunE: func(cmd *cobra.Command, args []string) error {
		printer.Configure(printer.PrinterConfig{
			UseColor:     true,
			LogLevel:     LogLevel,
			OutStream:    os.Stdout,
			ParsableMode: false,
		})

		detectorInstance, err := detector.CreateDetector(printer.Instance(), DetectorOptions)
		if err != nil {
			return fmt.Errorf("cmd: failed to create the detector instance: %w", err)
		}

		if len(OutputDirectoryPath) == 0 {
			videoName, err := utils.GetFilenameWithoutExtensions(InputVideoPath)
			if err != nil {
				printer.Instance().Error("The output directory name can not be derived and must be specified by the flag.")
				return fmt.Errorf("cmd: failed to derive the output directory name from input video path: %w", err)
			} else {
				OutputDirectoryPath = fmt.Sprintf("%s_vld", videoName)
			}
		}

		if err := detectorInstance.Run(InputVideoPath, OutputDirectoryPath, cmd.Context()); err != nil {
			return fmt.Errorf("cmd: detector run failed: %w", err)
		}

		return nil
	},
}
