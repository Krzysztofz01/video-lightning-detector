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

type exporterArguments struct {
	FrameCollection    frame.FrameCollection
	ExpectedDetections []int
	ActualDetections   []int
	Statistics         statistics.DescriptiveStatistics
	HasConfusionMatrix bool
	ConfusionMatrix    statistics.ConfusionMatrix
}

func (exporter *exporter) buildExporterArguments(fc frame.FrameCollection, ds statistics.DescriptiveStatistics, detections []int) (*exporterArguments, error) {
	var (
		hasConfusionMatrix bool = len(exporter.Options.ConfusionMatrixActualDetectionsExpression) > 0
		confusionMatrix    statistics.ConfusionMatrix
		expectedDetections []int
		err                error
	)

	if hasConfusionMatrix {
		expectedDetections, err = utils.ParseRangeExpression(exporter.Options.ConfusionMatrixActualDetectionsExpression)
		if err != nil {
			return nil, fmt.Errorf("export: failed to parse the confusion matrix actual detections range expression: %w", err)
		}

		confusionMatrix = statistics.CreateConfusionMatrix(expectedDetections, detections, fc.Count())
	}

	return &exporterArguments{
		FrameCollection:    fc,
		ExpectedDetections: expectedDetections,
		ActualDetections:   detections,
		Statistics:         ds,
		HasConfusionMatrix: hasConfusionMatrix,
		ConfusionMatrix:    confusionMatrix,
	}, nil
}

func (exporter *exporter) Export(fc frame.FrameCollection, ds statistics.DescriptiveStatistics, detections []int) error {
	exportTime := time.Now()

	args, err := exporter.buildExporterArguments(fc, ds, detections)
	if err != nil {
		return fmt.Errorf("export: failed to build the export arguments: %w", err)
	}

	if err := tableDescriptiveStatistics(exporter.Printer, args.Statistics, options.Verbose); err != nil {
		return fmt.Errorf("export: failed to export descriptive statistics: %w", err)
	}

	if !exporter.Options.SkipFramesExport {
		if err := exporter.ExportPngFrameImages(args.ActualDetections); err != nil {
			return fmt.Errorf("export: failed to perform the detected frames images export: %w", err)
		}
	}

	if args.HasConfusionMatrix {
		exporter.Printer.Debug("Frames used as expected detection classification: %v", args.ExpectedDetections)

		if err := tableConfusionMatrix(exporter.Printer, args.ConfusionMatrix, options.Verbose); err != nil {
			return fmt.Errorf("export: failed to export the confusion matrix: %w", err)
		}
	}

	if exporter.Options.ExportCsvReport {
		csvProgressFinalize := exporter.Printer.Progress("Exporting reports in CSV format")
		defer csvProgressFinalize()

		if path, err := exportCsvFrames(exporter.OutputDirPath, args.FrameCollection); err != nil {
			return fmt.Errorf("export: failed to export csv frames report: %w", err)
		} else {
			exporter.Printer.Info("Frames report in CSV format exported to: %s", path)
		}

		if path, err := exportCsvDescriptiveStatistics(exporter.OutputDirPath, args.Statistics); err != nil {
			return fmt.Errorf("export: failed to export csv descriptive statistics report: %w", err)
		} else {
			exporter.Printer.Info("Descriptive statistics in CSV format exported to %s", path)
		}

		if args.HasConfusionMatrix {
			if path, err := exportCsvConfusionMatrix(exporter.OutputDirPath, args.ConfusionMatrix); err != nil {
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

	if exporter.Options.ExportJsonReport || exporter.Options.ExportReport {
		jsonProgressFinalize := exporter.Printer.Progress("Exporting report in JSON format")
		defer jsonProgressFinalize()

		reportPath, err := exportJsonReport(exporter.OutputDirPath, exporter.Options, args)
		if err != nil {
			return fmt.Errorf("export: failed to export the json report: %w", err)
		} else {
			exporter.Printer.Info("Report in JSON format exported to: %s", reportPath)
		}

		jsonProgressFinalize()
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
