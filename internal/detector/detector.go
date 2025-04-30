package detector

import (
	"errors"
	"fmt"
	"image"
	"io"
	"path"
	"slices"
	"time"

	"github.com/Krzysztofz01/video-lightning-detector/internal/analyzer"
	"github.com/Krzysztofz01/video-lightning-detector/internal/export"
	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/printer"
	"github.com/Krzysztofz01/video-lightning-detector/internal/statistics"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
	"github.com/Krzysztofz01/video-lightning-detector/internal/video"
)

// Detector instance that is able to perform a search after ligntning strikes on a video file.
type Detector interface {
	Run(inputVideoPath, outputDirectoryPath string) error
}

type detector struct {
	options options.DetectorOptions
	printer printer.Printer
}

// Create a new video lightning detector instance with the specified options.
func CreateDetector(printer printer.Printer, options options.DetectorOptions) (Detector, error) {
	if printer == nil {
		return nil, errors.New("detector: invalid nil reference renderer provided")
	}

	if ok, msg := options.AreValid(); !ok {
		return nil, fmt.Errorf("detector: invalid options %s", msg)
	}

	printer.Debug("Detector create with options %+v", options)

	return &detector{
		options: options,
		printer: printer,
	}, nil
}

// Perform a lightning detection on the provided video specified by the file path and store the results at the specified directory path.
func (detector *detector) Run(inputVideoPath, outputDirectoryPath string) error {
	runTime := time.Now()
	detector.printer.InfoA("Starting the lightning hunt.")

	analyzer := analyzer.NewAnalyzer(inputVideoPath, outputDirectoryPath, detector.options, detector.printer)

	frames, err := analyzer.GetFrames()
	if err != nil {
		return fmt.Errorf("detector: video analysis stage failed: %w", err)
	}

	descriptiveStatistics := statistics.CreateDescriptiveStatistics(frames, int(detector.options.MovingMeanResolution))

	if detector.options.AutoThresholds {
		detector.ApplyAutoThresholds(frames, descriptiveStatistics)
	}

	detections := detector.PerformVideoDetection(frames, descriptiveStatistics)

	if err := detector.PerformExports(inputVideoPath, outputDirectoryPath, frames, descriptiveStatistics, detections); err != nil {
		return fmt.Errorf("detector: export stage failed: %w", err)
	}

	detector.printer.InfoA("Lightning hunting took: %s", time.Since(runTime))
	return nil
}

// Helper function used to auto-calculate the detection thresholds based on the frames and apply the threshold to the detector options
func (detector *detector) ApplyAutoThresholds(fc frame.FrameCollection, ds statistics.DescriptiveStatistics) {
	autoThresholdTime := time.Now()
	detector.printer.Debug("Starting the auto thresholds calculation stage.")

	frames := fc.GetAll()

	const (
		brightnessDiffCoefficient   float64 = 1.0
		brightnessStdDevCoefficient float64 = 0.0
		colorDiffDiffCoefficient    float64 = 1.0
		colorDiffStdDevCoefficient  float64 = 0.0
		btDiffDiffCoefficient       float64 = 0.25
		btDiffStdDevCoefficient     float64 = 0.15
	)

	var (
		brightnessMeanDiffSum float64 = 0.0
		brightnessStdDevSum   float64 = 0.0
		brightnessCount       int     = 0
		colorDiffMeanDiffSum  float64 = 0.0
		colorDiffStdDevSum    float64 = 0.0
		colorDiffCount        int     = 0
		btDiffMeanDiffSum     float64 = 0.0
		btDiffStdDevSum       float64 = 0.0
		btDiffCount           int     = 0
	)

	for frameIndex, frame := range frames {
		if brightnessDiff := frame.Brightness - ds.BrightnessMovingMean[frameIndex]; brightnessDiff > 0 {
			brightnessMeanDiffSum += brightnessDiff
			brightnessStdDevSum += ds.BrightnessMovingStdDev[frameIndex]
			brightnessCount += 1
		}

		if colorDiff := frame.ColorDifference - ds.ColorDifferenceMovingMean[frameIndex]; colorDiff > 0 {
			colorDiffMeanDiffSum += colorDiff
			colorDiffStdDevSum += ds.ColorDifferenceMovingStdDev[frameIndex]
			colorDiffCount += 1
		}

		if btDiff := frame.BinaryThresholdDifference - ds.BinaryThresholdDifferenceMovingMean[frameIndex]; btDiff > 0 {
			btDiffMeanDiffSum += btDiff
			btDiffStdDevSum += ds.BinaryThresholdDifferenceMovingStdDev[frameIndex]
			btDiffCount += 1
		}
	}

	var brightnessThreshold float64
	if brightnessCount == 0 {
		brightnessThreshold = 0
	} else {
		countf := float64(brightnessCount)
		brightnessThreshold = brightnessDiffCoefficient*brightnessMeanDiffSum/countf + brightnessStdDevCoefficient*brightnessStdDevSum/countf
	}

	var colorDifferenceThreshold float64
	if colorDiffCount == 0 {
		colorDifferenceThreshold = 0
	} else {
		countf := float64(colorDiffCount)
		colorDifferenceThreshold = colorDiffDiffCoefficient*colorDiffMeanDiffSum/countf + colorDiffStdDevCoefficient*colorDiffStdDevSum/countf
	}

	var binaryThresholdDifferenceThreshold float64
	if btDiffCount == 0 {
		binaryThresholdDifferenceThreshold = 0
	} else {
		countf := float64(btDiffCount)
		binaryThresholdDifferenceThreshold = btDiffDiffCoefficient*btDiffMeanDiffSum/countf + btDiffStdDevCoefficient*btDiffStdDevSum/countf
	}

	defaultOptions := options.GetDefaultDetectorOptions()

	if detector.options.BrightnessDetectionThreshold == defaultOptions.BrightnessDetectionThreshold {
		detector.options.BrightnessDetectionThreshold = brightnessThreshold
		detector.printer.Debug("Auto calculated brightness detection threshold: %g", brightnessThreshold)
	} else {
		detector.printer.Warning("The brightness detection threshold (%f) value was explicitly specified and would not be replace by the auto-calculated one (%f)",
			detector.options.BrightnessDetectionThreshold,
			brightnessThreshold)
	}

	if detector.options.ColorDifferenceDetectionThreshold == defaultOptions.ColorDifferenceDetectionThreshold {
		detector.options.ColorDifferenceDetectionThreshold = colorDifferenceThreshold
		detector.printer.Debug("Auth calculated color difference detection threshold: %g", colorDifferenceThreshold)
	} else {
		detector.printer.Warning("The color difference detection threshold (%f) value was explicitly specified and would not be replace by the auto-calculated one (%f)",
			detector.options.ColorDifferenceDetectionThreshold,
			colorDifferenceThreshold)
	}

	if detector.options.BinaryThresholdDifferenceDetectionThreshold == defaultOptions.BinaryThresholdDifferenceDetectionThreshold {
		detector.options.BinaryThresholdDifferenceDetectionThreshold = binaryThresholdDifferenceThreshold
		detector.printer.Debug("Auto calculated binary threshold difference threshold: %g", binaryThresholdDifferenceThreshold)
	} else {
		detector.printer.Warning("The binary threshold detection threshold (%f) value was explicitly specified and would not be replace by the auto-calculated one (%f)",
			detector.options.BinaryThresholdDifferenceDetectionThreshold,
			binaryThresholdDifferenceThreshold)
	}

	detector.printer.Debug("Auto thresholds calculation stage finished. Stage took: %s", time.Since(autoThresholdTime))
}

// Helper function used to filter out indecies representing frames wihich meet the requirement thresholds.
func (detector *detector) PerformVideoDetection(framesCollection frame.FrameCollection, ds statistics.DescriptiveStatistics) []int {
	videoDetectionTime := time.Now()
	detector.printer.Debug("Starting the video detection stage.")

	detectionBuffer := CreateDetectionBuffer()

	frames := framesCollection.GetAll()

	progressStep, progressFinalize := detector.printer.ProgressSteps("Video detection stage.", len(frames))
	defer progressFinalize()

	for frameIndex, frame := range frames {
		var (
			brightnessClassified bool = frame.Brightness >= detector.options.BrightnessDetectionThreshold+ds.BrightnessMovingMean[frameIndex]
			colorDiffClassified  bool = frame.ColorDifference >= detector.options.ColorDifferenceDetectionThreshold+ds.ColorDifferenceMovingMean[frameIndex]
			btDiffClassified     bool = frame.BinaryThresholdDifference >= detector.options.BinaryThresholdDifferenceDetectionThreshold+ds.BinaryThresholdDifferenceMovingMean[frameIndex]
		)

		// TODO: Verbose logging

		detectionBuffer.Append(frameIndex, brightnessClassified, colorDiffClassified, btDiffClassified)

		progressStep()
	}

	detector.printer.Debug("Video detection stage finished. Stage took: %s", time.Since(videoDetectionTime))

	return detectionBuffer.ResolveClassifiedIndex()
}

// Helper function used to perform exports to varius formats selected via the options
func (detector *detector) PerformExports(inputVideoPath, outputDirectoryPath string, fc frame.FrameCollection, ds statistics.DescriptiveStatistics, detections []int) error {
	exportTime := time.Now()

	if err := export.PrintDescriptiveStatistics(detector.printer, ds, options.Verbose); err != nil {
		return fmt.Errorf("detector: failed to export descriptive statistics: %w", err)
	}

	if !detector.options.SkipFramesExport {
		if err := detector.PerformFrameImagesExport(inputVideoPath, outputDirectoryPath, detections); err != nil {
			return fmt.Errorf("detector: failed to perform the detected frames images export: %w", err)
		}
	}

	var confusionMatrix statistics.ConfusionMatrix
	if detector.options.ExportConfusionMatrix {
		actualClassification, err := utils.ParseRangeExpression(detector.options.ConfusionMatrixActualDetectionsExpression)
		if err != nil {
			return fmt.Errorf("detector: failed to parse the confusion matrix actual detections range expression: %w", err)
		}

		detector.printer.Debug("Frames used as actual detection classification: %v", actualClassification)

		confusionMatrix = statistics.CreateConfusionMatrix(actualClassification, detections, fc.Count())

		if err := export.PrintConfusionMatrix(detector.printer, confusionMatrix, options.Verbose); err != nil {
			return fmt.Errorf("detector: failed to export the confusion matrix: %w", err)
		}
	}

	if detector.options.ExportCsvReport {
		csvProgressFinalize := detector.printer.Progress("Exporting reports in CSV format")
		defer csvProgressFinalize()

		if path, err := export.ExportCsvFrames(outputDirectoryPath, fc); err != nil {
			return fmt.Errorf("detector: failed to export csv frames report: %w", err)
		} else {
			detector.printer.Info("Frames report in CSV format exported to: %s", path)
		}

		if path, err := export.ExportCsvDescriptiveStatistics(outputDirectoryPath, ds); err != nil {
			return fmt.Errorf("detector: failed to export csv descriptive statistics report: %w", err)
		} else {
			detector.printer.Info("Descriptive statistics in CSV format exported to %s", path)
		}

		if detector.options.ExportConfusionMatrix {
			if path, err := export.ExportCsvConfusionMatrix(outputDirectoryPath, confusionMatrix); err != nil {
				return fmt.Errorf("detector: failed to export csv confusion matrix report: %w", err)
			} else {
				detector.printer.Info("Confusion matrix in CSV format exported to %s", path)
			}
		}

		csvProgressFinalize()
	}

	if detector.options.ExportJsonReport {
		jsonProgressFinalize := detector.printer.Progress("Exporting reports in JSON format")
		defer jsonProgressFinalize()

		if path, err := export.ExportJsonFrames(outputDirectoryPath, fc); err != nil {
			return fmt.Errorf("detector: failed to export json frames report: %w", err)
		} else {
			detector.printer.Info("Frames report in JSON format exported to: %s", path)
		}

		if path, err := export.ExportJsonDescriptiveStatistics(outputDirectoryPath, ds); err != nil {
			return fmt.Errorf("detector: failed to export json descriptive statistics report: %w", err)
		} else {
			detector.printer.Info("Descriptive statistics in JSON format exported to %s", path)
		}

		if detector.options.ExportConfusionMatrix {
			if path, err := export.ExportJsonConfusionMatrix(outputDirectoryPath, confusionMatrix); err != nil {
				return fmt.Errorf("detector: failed to export json confusion matrix report: %w", err)
			} else {
				detector.printer.Info("Confusion matrix in JSON format exported to %s", path)
			}
		}

		jsonProgressFinalize()
	}

	if detector.options.ExportChartReport {
		chartProgressFinalize := detector.printer.Progress("Exporting chart report")
		defer chartProgressFinalize()

		path, err := export.ExportFramesChart(
			outputDirectoryPath,
			fc,
			ds,
			detections,
			detector.options.BrightnessDetectionThreshold,
			detector.options.ColorDifferenceDetectionThreshold,
			detector.options.BinaryThresholdDifferenceDetectionThreshold)

		if err != nil {
			return fmt.Errorf("detector: failed to export the frames chart: %w", err)
		} else {
			detector.printer.Info("Frames chart exported to: %s", path)
		}

		chartProgressFinalize()
	}

	detector.printer.Info("Export finished. Stage took: %s", time.Since(exportTime))
	return nil
}

// NOTE: Experimental sampling of binary threshold from recording
// func (detector *detector) SampleBinaryThreshold(inputVideoPath string) (float64, error) {
// 	video, err := vidio.NewVideo(inputVideoPath)
// 	if err != nil {
// 		return 0, fmt.Errorf("detector: failed to open the video file for the binary threshold sampling stage: %w", err)
// 	}

// 	defer video.Close()

// 	var (
// 		frames      int     = video.Frames()
// 		fps         float64 = video.FPS()
// 		duration    float64 = float64(frames) / fps
// 		sampleCount int     = int(math.Max(1.43*math.Log(duration/60)-0.64, 1))
// 		framesStep  int     = frames / sampleCount
// 	)

// 	var sampleFrameIndexes []int = make([]int, 0, sampleCount)
// 	for sampleIndex := 0; sampleIndex < sampleCount; sampleIndex += 1 {
// 		sampleFrameIndexes = append(sampleFrameIndexes, sampleIndex*framesStep)
// 	}

// 	sampleFrames, err := video.ReadFrames(sampleFrameIndexes...)
// 	if err != nil {
// 		return 0, fmt.Errorf("detector: failed to read the specified frames from the video: %w", err)
// 	}

// 	var thresholdSum float64 = 0
// 	for _, sampleFrame := range sampleFrames {
// 		thresholdSum += utils.Otsu(sampleFrame)

// 	}

// 	return thresholdSum / float64(sampleCount), nil
// }

// Helper function used to export frame images which meet the requirement thresholds to png files.
func (detector *detector) PerformFrameImagesExport(inputVideoPath, outputDirectoryPath string, detections []int) error {
	framesExportTime := time.Now()
	detector.printer.Debug("Starting the frames export stage.")
	detector.printer.Info("About to export %d frames.", len(detections))

	slices.Sort(detections)

	video, err := video.NewVideo(inputVideoPath)
	if err != nil {
		return fmt.Errorf("detector: failed to open the video file for the frame export stage: %w", err)
	}

	defer video.Close()

	targetWidth, targetHeight := video.GetOutputDimensions()

	frame := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	if err := video.SetFrameBuffer(frame.Pix); err != nil {
		return fmt.Errorf("detector: failed to apply the given buffer as the video frame buffer: %w", err)
	}

	if err := video.SetTargetFrames(detections...); err != nil {
		return fmt.Errorf("detector: failed to set the detection frames as the video target frames: %w", err)
	}

	progressStep, progressFinalize := detector.printer.ProgressSteps("Video frames export stage.", len(detections))

	for _, frameIndex := range detections {
		if err := video.Read(); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("detector: failed to read the video export frame: %w", err)
		}

		frameImageName := fmt.Sprintf("frame-%d.png", frameIndex+1)
		frameImagePath := path.Join(outputDirectoryPath, frameImageName)
		if err := utils.ExportImageAsPng(frameImagePath, frame); err != nil {
			return fmt.Errorf("detector: failed to export the frame image: %w", err)
		}

		progressStep()
		detector.printer.Info("Frame: [%d/%d]. Frame image exported at: %s", frameIndex+1, video.Frames(), frameImagePath)
	}

	progressFinalize()
	detector.printer.Debug("Frames export stage finished. Stage took: %s", time.Since(framesExportTime))
	return nil
}
