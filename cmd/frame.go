package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/Krzysztofz01/video-lightning-detector/internal/detector"
	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/printer"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

var (
	InputFramePath             string
	FrameStreamDetectorOptions options.StreamDetectorOptions = options.GetDefaultStreamDetectorOptions()
)

func init() {
	frameCmd.Flags().StringVarP(&InputFramePath, "input-frame-path", "i", "", "Input frame to perform the lightning detection.")
	frameCmd.MarkPersistentFlagRequired("input-frame-path")

	frameCmd.PersistentFlags().Float64VarP(
		&FrameStreamDetectorOptions.FrameDetectionPlotThreshold,
		"frame-detection-plot-threshold", "t",
		FrameStreamDetectorOptions.FrameDetectionPlotThreshold,
		"Binary threshold segmentation argument for the frame image processing during detection plot calculation.")

	frameCmd.PersistentFlags().IntVarP(
		&FrameStreamDetectorOptions.FrameDetectionPlotResolution,
		"frame-detection-plot-resolution", "r",
		FrameStreamDetectorOptions.FrameDetectionPlotResolution,
		"The resolution of the x and y axis of the frame image detection plot.")

	if DeveloperMode == "true" {
		rootCmd.AddCommand(frameCmd)
	}
}

var frameCmd = &cobra.Command{
	Use:   "frame",
	Short: "[Developer Mode Only] Perform the frame strike detection on a provided image.",
	Long:  "[Developer Mode Only] Perform the frame strike detection on a provided image.",
	RunE: func(cmd *cobra.Command, args []string) error {
		printer.Configure(printer.PrinterConfig{
			UseColor:     true,
			LogLevel:     LogLevel,
			OutStream:    os.Stdout,
			ParsableMode: false,
		})

		img, err := utils.ImportImageRgba(InputFramePath)
		if err != nil {
			return fmt.Errorf("cmd: failed to import the image: %w", err)
		}

		detector, err := detector.CreateFrameStrikeDetector(img.Bounds().Dx(), img.Bounds().Dy(), FrameStreamDetectorOptions)
		if err != nil {
			return fmt.Errorf("cmd: failed to create frame strike detector: %w", err)
		}

		plot, err := detector.GetDetectionPlot(img)
		if err != nil {
			return fmt.Errorf("cmd: failed to perform the frame strike detection: %w", err)
		}

		plotMarshal, err := json.Marshal(plot)
		if err != nil {
			return fmt.Errorf("cmd: failed to marshal the frame strike detection plot: %w", err)
		}

		printer.Instance().WriteRaw("%s\n", string(plotMarshal))
		return nil
	},
}
