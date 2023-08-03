package cmd

import (
	"fmt"
	"os"

	nestedFormatter "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Krzysztofz01/video-lightning-detector/internal/detector"
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

	rootCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.FrameScalingFactor,
		"scaling-factor", "s",
		DetectorOptions.FrameScalingFactor,
		"The frame scaling factor used to downscale frames for better performance.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.Denoise,
		"denoise", "n",
		DetectorOptions.Denoise,
		"Apply de-noising to the frames. This may have a positivie effect on the frames statistics precision.")

	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&nestedFormatter.Formatter{
		TimestampFormat:  "",
		HideKeys:         true,
		NoColors:         false,
		NoFieldsColors:   false,
		NoFieldsSpace:    false,
		ShowFullLevel:    false,
		NoUppercaseLevel: false,
		TrimMessages:     false,
		CallerFirst:      false,
	})
}

func Execute(args []string) {
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err.Error())
	}
}

func run(cmd *cobra.Command, args []string) error {
	defer func() {
		if err := recover(); err != nil {
			logrus.Fatal(err)
		}
	}()

	if VerboseMode {
		logrus.SetLevel(logrus.DebugLevel)
	}

	detectorInstance, err := detector.CreateDetector(DetectorOptions)
	if err != nil {
		return fmt.Errorf("cmd: failed to create the detector instance: %w", err)
	}

	if err := detectorInstance.Run(InputVideoPath, OutputDirectoryPath); err != nil {
		logrus.Error(err)
		return fmt.Errorf("cmd: detector run failed: %w", err)
	}

	return nil
}
