package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Krzysztofz01/video-lightning-detector/internal/detector"
	"github.com/Krzysztofz01/video-lightning-detector/internal/render"
)

var rootCmd = &cobra.Command{
	Use:   "video-ligtning-detector",
	Short: "",
	Long:  "",
	RunE:  run,
}

var (
	InputVideoPath      string
	OutputDirectoryPath string
	VerboseMode         bool
	DetectorOptions     detector.DetectorOptions = detector.GetDefaultDetectorOptions()
)

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().StringVarP(&InputVideoPath, "input-video-path", "i", "", "Input video to perform the lightning detection.")
	rootCmd.MarkPersistentFlagRequired("input-video-path")

	rootCmd.PersistentFlags().StringVarP(&OutputDirectoryPath, "output-directory-path", "o", "", "Output directory to store detected frames.")
	rootCmd.MarkPersistentFlagRequired("output-directory-path")

	rootCmd.PersistentFlags().BoolVarP(&VerboseMode, "verbose", "v", false, "Enable verbose logging.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.AutoThresholds,
		"auto-thresholds", "a",
		DetectorOptions.AutoThresholds,
		"Automatically select thresholds for all parameters based on calculated frame values. Values that are explicitly provided will not be overwritten.")

	rootCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.ColorDifferenceDetectionThreshold,
		"color-difference-threshold", "c",
		DetectorOptions.ColorDifferenceDetectionThreshold,
		"The threshold used to determine the difference between two neighbouring frames on the color basis. Detection is credited when the value for a given frame is greater than the sum of the threshold of tripping and the moving average.")

	rootCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.BinaryThresholdDifferenceDetectionThreshold,
		"binary-threshold-difference-threshold", "t",
		DetectorOptions.BinaryThresholdDifferenceDetectionThreshold,
		"The threshold used to determine the difference between two neighbouring frames after the binary thresholding process. Detection is credited when the value for a given frame is greater than the sum of the threshold of tripping and the moving average")

	rootCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.BrightnessDetectionThreshold,
		"brightness-threshold", "b",
		DetectorOptions.BrightnessDetectionThreshold,
		"The threshold used to determine the brightness of the frame. Detection is credited when the value for a given frame is greater than the sum of the threshold of tripping and the moving average")

	rootCmd.PersistentFlags().Int32VarP(
		&DetectorOptions.MovingMeanResolution,
		"moving-mean-resolution", "m",
		DetectorOptions.MovingMeanResolution,
		"The number of elements of the subset on which the moving mean will be calculated, for each parameter.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.SkipFramesExport,
		"skip-frames-export", "f",
		DetectorOptions.SkipFramesExport,
		"Value indicating if the detected frames should not be exported.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.ExportCsvReport,
		"export-csv-report", "e",
		DetectorOptions.ExportCsvReport,
		"Value indicating if the frames statistics report in CSV format should be exported.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.ExportJsonReport,
		"export-json-report", "j",
		DetectorOptions.ExportJsonReport,
		"Value indicating if the frames statistics report in JSON format should be exported.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.ExportChartReport,
		"export-chart-report", "r",
		DetectorOptions.ExportChartReport,
		"Value indicating if the frames statistics chart in HTML format should be exported.")

	rootCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.FrameScalingFactor,
		"scaling-factor", "s",
		DetectorOptions.FrameScalingFactor,
		"The frame scaling factor used to downscale frames for better performance.")

	rootCmd.PersistentFlags().VarPF(
		&DetectorOptions.Denoise,
		"denoise", "n",
		"Apply de-noising to the frames. This may have a positivie effect on the frames statistics precision.")

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
		"Value indicating the the the pre-analyzed frames should be imported from the previously exported JSON report.")

	rootCmd.PersistentFlags().BoolVar(
		&DetectorOptions.StrictExplicitThreshold,
		"strict-explicit-threshold",
		DetectorOptions.StrictExplicitThreshold,
		"Value indicating if explicit thresholds range should be validated.")
}

func Execute(args []string) {
	// getImg := func(p string) image.Image {
	// 	f, _ := os.Open(p)
	// 	defer f.Close()

	// 	i, _ := png.Decode(f)
	// 	return i
	// }

	// i1 := getImg("/home/krzysztof/Desktop/Repos/video-lightning-detector/bin/inz-test-out/test-day-2/frames/a/frame0210.png")
	// i2 := getImg("/home/krzysztof/Desktop/Repos/video-lightning-detector/bin/inz-test-out/test-day-2/frames/a/frame0219.png")

	// f1 := frame.CreateNewFrame(i1, nil, 1, frame.BinaryThresholdParam)
	// f2 := frame.CreateNewFrame(i2, i1, 2, frame.BinaryThresholdParam)

	// _, _ = f1, f2

	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func run(cmd *cobra.Command, args []string) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	renderer := render.CreateRenderer(VerboseMode)
	detectorInstance, err := detector.CreateDetector(renderer, DetectorOptions)
	if err != nil {
		return fmt.Errorf("cmd: failed to create the detector instance: %w", err)
	}

	if err := detectorInstance.Run(InputVideoPath, OutputDirectoryPath); err != nil {
		return fmt.Errorf("cmd: detector run failed: %w", err)
	}

	return nil
}
