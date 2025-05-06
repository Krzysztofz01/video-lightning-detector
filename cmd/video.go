package cmd

import (
	"fmt"

	"github.com/Krzysztofz01/video-lightning-detector/internal/detector"
	"github.com/Krzysztofz01/video-lightning-detector/internal/printer"
	"github.com/spf13/cobra"
)

var (
	InputVideoPath      string
	OutputDirectoryPath string
)

func init() {
	videoCmd.Flags().StringVarP(&InputVideoPath, "input-video-path", "i", "", "Input video to perform the lightning detection.")
	videoCmd.MarkPersistentFlagRequired("input-video-path")

	videoCmd.PersistentFlags().StringVarP(&OutputDirectoryPath, "output-directory-path", "o", "", "Output directory path for export artifacts such as frames and reports in selected formats.")
	videoCmd.MarkPersistentFlagRequired("output-directory-path")

	rootCmd.AddCommand(videoCmd)
}

var videoCmd = &cobra.Command{
	Use:   "video",
	Short: "Perform the analysis, detection and export stage on single video.",
	Long:  "Perform the analysis, detection and export stage on single video.",
	RunE: func(cmd *cobra.Command, args []string) error {
		detectorInstance, err := detector.CreateDetector(printer.Instance(), DetectorOptions)
		if err != nil {
			return fmt.Errorf("cmd: failed to create the detector instance: %w", err)
		}

		if err := detectorInstance.Run(InputVideoPath, OutputDirectoryPath); err != nil {
			return fmt.Errorf("cmd: detector run failed: %w", err)
		}

		return nil
	},
}
