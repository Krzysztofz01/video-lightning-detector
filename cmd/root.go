package cmd

import (
	"fmt"
	"os"

	nestedFormatter "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Krzysztofz01/video-lightning-detector/internal"
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
	DetectorOptions     internal.DetectorOptions = internal.GetDefaultDetectorOptions()
)

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().StringVarP(&InputVideoPath, "input-video-path", "i", "", "Input video to perform the lightning detection.")
	rootCmd.MarkPersistentFlagRequired("input-video-path")

	rootCmd.PersistentFlags().StringVarP(&OutputDirectoryPath, "output-directory-path", "o", "", "Output directory to store detected frames.")
	rootCmd.MarkPersistentFlagRequired("output-directory-path")

	rootCmd.PersistentFlags().BoolVarP(&VerboseMode, "verbose", "v", false, "Enable verbose logging.")

	rootCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.FrameDifferenceThreshold,
		"difference-threshold", "d",
		DetectorOptions.FrameDifferenceThreshold,
		"The threshold used to determine the difference between two neighbouring frames.")

	rootCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.FrameBrightnessThreshold,
		"brightness-threshold", "b",
		DetectorOptions.FrameBrightnessThreshold,
		"The threshold used to determine the brightness of the frame.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.SkipFramesExport,
		"skip-frames-export", "f",
		DetectorOptions.SkipFramesExport,
		"Value indicating if the detected frams should not be exported.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.SkipReportExport,
		"skip-report-export", "r",
		DetectorOptions.SkipReportExport,
		"Value indicating if the frames statistics report should not be exported.")

	rootCmd.PersistentFlags().BoolVarP(
		&DetectorOptions.SkipThresholdSuggestion,
		"skip-threshold-suggestion", "t",
		DetectorOptions.SkipThresholdSuggestion,
		"Value indicating if the thresholds suggestion shoul not be calculated.")

	rootCmd.PersistentFlags().Float64VarP(
		&DetectorOptions.FrameScalingFactor,
		"scaling-factor", "s",
		DetectorOptions.FrameScalingFactor,
		"The frame scaling factor used to downscale frames for better performance.")

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
	if VerboseMode {
		logrus.SetLevel(logrus.DebugLevel)
	}

	detector, err := internal.CreateDetector(DetectorOptions)
	if err != nil {
		return fmt.Errorf("cmd: failed to create the detector instance: %w", err)
	}

	_, err = detector.Run(InputVideoPath, OutputDirectoryPath)
	if err != nil {
		logrus.Error(err)
		return fmt.Errorf("cmd: detector run failed: %w", err)
	}

	return nil
}
