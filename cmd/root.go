package cmd

import (
	"fmt"
	"os"

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
	detectorOptions     internal.DetectorOptions
)

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	defaultOptions := internal.GetDefaultDetectorOptions()

	rootCmd.PersistentFlags().StringVarP(&InputVideoPath, "input-video-path", "i", "", "Input video to perform the lightning detection.")
	rootCmd.MarkPersistentFlagRequired("input-video-path")

	rootCmd.PersistentFlags().StringVarP(&OutputDirectoryPath, "output-directory-path", "o", "", "Output directory to store detected frames.")
	rootCmd.MarkPersistentFlagRequired("output-directory-path")

	rootCmd.PersistentFlags().Float64VarP(&detectorOptions.FrameDifferenceThreshold, "difference-threshold", "d", defaultOptions.FrameDifferenceThreshold, "The threshold used to determine the difference between two neighbouring frames.")

	rootCmd.PersistentFlags().Float64VarP(&defaultOptions.FrameBrightnessThreshold, "brightness-threshold", "b", defaultOptions.FrameBrightnessThreshold, "The threshold used to determine the brightness of the frame.")
}

func Execute(args []string) {
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		// TODO: Logger
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// TODO: Validate paths
	detector, err := internal.CreateDetector(detectorOptions)
	if err != nil {
		return fmt.Errorf("cmd: failed to create the detector instance: %w", err)
	}

	_, err = detector.Run(InputVideoPath, OutputDirectoryPath)
	if err != nil {
		return fmt.Errorf("cmd: detector run failed: %w", err)
	}

	return nil
}
