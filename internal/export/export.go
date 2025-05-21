package export

import (
	"fmt"
	"image"
	"io"
	"path"
	"slices"
	"time"

	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/printer"
	"github.com/Krzysztofz01/video-lightning-detector/internal/statistics"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
	"github.com/Krzysztofz01/video-lightning-detector/internal/video"
)

type Exporter interface {
	Export(fc frame.FrameCollection, ds statistics.DescriptiveStatistics, detections []int) error
}

type exporter struct {
	InputVideoPath string
	OutputDirPath  string
	Options        options.DetectorOptions
	Printer        printer.Printer
}

func (exporter *exporter) Export(fc frame.FrameCollection, ds statistics.DescriptiveStatistics, detections []int) error {
	exportTime := time.Now()

	if err := tableDescriptiveStatistics(exporter.Printer, ds, options.Verbose); err != nil {
		return fmt.Errorf("export: failed to export descriptive statistics: %w", err)
	}

	if !exporter.Options.SkipFramesExport {
		if err := exporter.ExportPngFrameImages(detections); err != nil {
			return fmt.Errorf("export: failed to perform the detected frames images export: %w", err)
		}
	}

	var confusionMatrix statistics.ConfusionMatrix
	if exporter.Options.ExportConfusionMatrix {
		actualClassification, err := utils.ParseRangeExpression(exporter.Options.ConfusionMatrixActualDetectionsExpression)
		if err != nil {
			return fmt.Errorf("export: failed to parse the confusion matrix actual detections range expression: %w", err)
		}

		exporter.Printer.Debug("Frames used as actual detection classification: %v", actualClassification)

		confusionMatrix = statistics.CreateConfusionMatrix(actualClassification, detections, fc.Count())

		if err := tableConfusionMatrix(exporter.Printer, confusionMatrix, options.Verbose); err != nil {
			return fmt.Errorf("export: failed to export the confusion matrix: %w", err)
		}
	}

	if exporter.Options.ExportCsvReport {
		csvProgressFinalize := exporter.Printer.Progress("Exporting reports in CSV format")
		defer csvProgressFinalize()

		if path, err := exportCsvFrames(exporter.OutputDirPath, fc); err != nil {
			return fmt.Errorf("export: failed to export csv frames report: %w", err)
		} else {
			exporter.Printer.Info("Frames report in CSV format exported to: %s", path)
		}

		if path, err := exportCsvDescriptiveStatistics(exporter.OutputDirPath, ds); err != nil {
			return fmt.Errorf("export: failed to export csv descriptive statistics report: %w", err)
		} else {
			exporter.Printer.Info("Descriptive statistics in CSV format exported to %s", path)
		}

		if exporter.Options.ExportConfusionMatrix {
			if path, err := exportCsvConfusionMatrix(exporter.OutputDirPath, confusionMatrix); err != nil {
				return fmt.Errorf("export: failed to export csv confusion matrix report: %w", err)
			} else {
				exporter.Printer.Info("Confusion matrix in CSV format exported to %s", path)
			}
		}

		if path, err := exportCsvDetectionThresholds(exporter.OutputDirPath, exporter.Options); err != nil {
			return fmt.Errorf("export: failed to export csv detection thresholds report: %w", err)
		} else {
			exporter.Printer.Info("Detections thresholds in CSV format exported to %s", path)
		}

		csvProgressFinalize()
	}

	if exporter.Options.ExportJsonReport {
		jsonProgressFinalize := exporter.Printer.Progress("Exporting reports in JSON format")
		defer jsonProgressFinalize()

		if path, err := exportJsonFrames(exporter.OutputDirPath, fc); err != nil {
			return fmt.Errorf("export: failed to export json frames report: %w", err)
		} else {
			exporter.Printer.Info("Frames report in JSON format exported to: %s", path)
		}

		if path, err := exportJsonDescriptiveStatistics(exporter.OutputDirPath, ds); err != nil {
			return fmt.Errorf("export: failed to export json descriptive statistics report: %w", err)
		} else {
			exporter.Printer.Info("Descriptive statistics in JSON format exported to %s", path)
		}

		if exporter.Options.ExportConfusionMatrix {
			if path, err := exportJsonConfusionMatrix(exporter.OutputDirPath, confusionMatrix); err != nil {
				return fmt.Errorf("export: failed to export json confusion matrix report: %w", err)
			} else {
				exporter.Printer.Info("Confusion matrix in JSON format exported to %s", path)
			}
		}

		if path, err := exportJsonDetectionThresholds(exporter.OutputDirPath, exporter.Options); err != nil {
			return fmt.Errorf("export: failed to export json detection thresholds report: %w", err)
		} else {
			exporter.Printer.Info("Detection thresholds in JSON format exported to %s", path)
		}

		jsonProgressFinalize()
	}

	if exporter.Options.ExportChartReport {
		chartProgressFinalize := exporter.Printer.Progress("Exporting chart report")
		defer chartProgressFinalize()

		path, err := exportFramesChart(
			exporter.OutputDirPath,
			fc,
			ds,
			detections,
			exporter.Options.BrightnessDetectionThreshold,
			exporter.Options.ColorDifferenceDetectionThreshold,
			exporter.Options.BinaryThresholdDifferenceDetectionThreshold)

		if err != nil {
			return fmt.Errorf("export: failed to export the frames chart: %w", err)
		} else {
			exporter.Printer.Info("Frames chart exported to: %s", path)
		}

		chartProgressFinalize()
	}

	exporter.Printer.Info("Export finished. Stage took: %s", time.Since(exportTime))
	return nil
}

func (exporter *exporter) ExportPngFrameImages(detections []int) error {
	framesExportTime := time.Now()
	exporter.Printer.Debug("Starting the frames export stage.")
	exporter.Printer.Info("About to export %d frames.", len(detections))

	slices.Sort(detections)

	video, err := video.NewVideo(exporter.InputVideoPath)
	if err != nil {
		return fmt.Errorf("export: failed to open the video file for the frame export stage: %w", err)
	}

	defer video.Close()

	targetWidth, targetHeight := video.GetOutputDimensions()

	frame := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	if err := video.SetFrameBuffer(frame.Pix); err != nil {
		return fmt.Errorf("export: failed to apply the given buffer as the video frame buffer: %w", err)
	}

	if err := video.SetTargetFrames(detections...); err != nil {
		return fmt.Errorf("export: failed to set the detection frames as the video target frames: %w", err)
	}

	progressStep, progressFinalize := exporter.Printer.ProgressSteps("Video frames export stage.", len(detections))

	for _, frameIndex := range detections {
		if err := video.Read(); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("export: failed to read the video export frame: %w", err)
		}

		frameImageName := fmt.Sprintf("frame-%d.png", frameIndex+1)
		frameImagePath := path.Join(exporter.OutputDirPath, frameImageName)
		if err := utils.ExportImageAsPng(frameImagePath, frame); err != nil {
			return fmt.Errorf("export: failed to export the frame image: %w", err)
		}

		progressStep()
		exporter.Printer.Info("Frame: [%d/%d]. Frame image exported at: %s", frameIndex+1, video.FramesCountApprox(), frameImagePath)
	}

	progressFinalize()
	exporter.Printer.Debug("Frames export stage finished. Stage took: %s", time.Since(framesExportTime))
	return nil
}

func NewExporter(inputVideo, outputDir string, o options.DetectorOptions, p printer.Printer) Exporter {
	return &exporter{
		InputVideoPath: inputVideo,
		OutputDirPath:  outputDir,
		Options:        o,
		Printer:        p,
	}
}
