package detector

import (
	"errors"
	"fmt"
	"image"
	"path"
	"strconv"
	"time"

	vidio "github.com/AlexEidt/Vidio"
	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/render"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

// Detector instance that is able to perform a search after ligntning strikes on a video file.
type Detector interface {
	Run(inputVideoPath, outputDirectoryPath string) error
}

type detector struct {
	options  DetectorOptions
	renderer render.Renderer
}

// Create a new video lightning detector instance with the specified options.
func CreateDetector(renderer render.Renderer, options DetectorOptions) (Detector, error) {
	if renderer == nil {
		return nil, errors.New("detector: invalid nil reference renderer provided")
	}

	if ok, msg := options.AreValid(); !ok {
		return nil, fmt.Errorf("detector: invalid options %s", msg)
	}

	renderer.LogDebug("Detector create with options %+v", options)

	return &detector{
		options:  options,
		renderer: renderer,
	}, nil
}

// Perform a lightning detection on the provided video specified by the file path and store the results at the specified directory path.
func (detector *detector) Run(inputVideoPath, outputDirectoryPath string) error {
	runTime := time.Now()
	detector.renderer.LogInfo("Starting the lightning hunt.")

	frames, err := detector.performVideoAnalysis(inputVideoPath)
	if err != nil {
		return fmt.Errorf("detector: video analysis stage failed: %w", err)
	}

	if detector.options.AutoThresholds {
		detector.applyAutoThresholds(frames)
	}

	detector.performStatisticsLogging(frames)
	detections := detector.performVideoDetection(frames)

	if !detector.options.SkipFramesExport {
		if err := detector.performFramesExport(inputVideoPath, outputDirectoryPath, detections); err != nil {
			return fmt.Errorf("detector: failed to perform the detected frames images export: %w", err)
		}
	}

	if detector.options.ExportCsvReport {
		if err := detector.handleCsvReportExport(outputDirectoryPath, frames); err != nil {
			return fmt.Errorf("detector: csv report export failed: %w", err)
		}
	}

	if detector.options.ExportJsonReport {
		if err := detector.handleJsonReportExport(outputDirectoryPath, frames); err != nil {
			return fmt.Errorf("detector: json report export failed: %w", err)
		}
	}

	detector.renderer.LogInfo("Lightning hunting took: %s", time.Since(runTime))
	return nil
}

// Helper function used to iterate over the video frames in order to generate a collection of frames instances containing
// processed values about given frames and neighbouring frames relations.
func (detector *detector) performVideoAnalysis(inputVideoPath string) (*frame.FramesCollection, error) {
	videoAnalysisTime := time.Now()
	detector.renderer.LogDebug("Starting the video analysis stage.")

	video, err := vidio.NewVideo(inputVideoPath)
	if err != nil {
		return nil, fmt.Errorf("detector: failed to open the video file for the analysis stage: %w", err)
	}

	defer video.Close()

	targetWidth := int(float64(video.Width()) * detector.options.FrameScalingFactor)
	targetHeight := int(float64(video.Height()) * detector.options.FrameScalingFactor)

	frameCurrentBuffer := image.NewRGBA(image.Rect(0, 0, video.Width(), video.Height()))
	video.SetFrameBuffer(frameCurrentBuffer.Pix)

	frameCurrent := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	framePrevious := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	frameNumber := 1
	frameCount := video.Frames()
	frames := frame.CreateNewFramesCollection(frameCount)

	progressBarStep, progressBarClose := detector.renderer.Progress("Video analysis stage.", frameCount)

	for video.Read() {
		if utils.ScaleImage(frameCurrentBuffer, frameCurrent, detector.options.FrameScalingFactor); err != nil {
			return nil, fmt.Errorf("detector: failed to scale the current frame image on the analyze stage: %w", err)
		}

		if detector.options.Denoise {
			if err := utils.BlurImage(frameCurrent, frameCurrent, 8); err != nil {
				return nil, fmt.Errorf("detector: failed to blur the current frame image on the analyze stage: %w", err)
			}
		}

		frame := frame.CreateNewFrame(frameCurrent, framePrevious, frameNumber)
		frames.Append(frame)

		detector.renderer.LogDebug("Frame: [%d/%d]. Brightness: %f ColorDiff: %f BTDiff: %f", frameNumber, frameCount, frame.Brightness, frame.ColorDifference, frame.BinaryThresholdDifference)

		frameNumber += 1
		progressBarStep()
		copy(framePrevious.Pix, frameCurrent.Pix)
	}

	progressBarClose()
	detector.renderer.LogDebug("Video analysis stage finished. Stage took: %s", time.Since(videoAnalysisTime))
	return frames, nil
}

// Helper function used to auto-calculate the detection thresholds based on the frames and apply the threshold to the detector options
// TODO: More versatile results could be achived using "moving standard deviation"???
// TODO: Render thresholds in a table
func (detector *detector) applyAutoThresholds(framesCollection *frame.FramesCollection) {
	autoThresholdTime := time.Now()
	detector.renderer.LogDebug("Starting the auto thresholds calculation stage.")

	frames := framesCollection.GetAll()
	statistics := framesCollection.CalculateStatistics(int(detector.options.MovingMeanResolution))

	var (
		gDiffBrightnessValue float64 = 0
		gDiffBrightnessCount int     = 0
		gDiffColorDiffValue  float64 = 0
		gDiffColorDiffCount  int     = 0
		gDiffBTDiffValue     float64 = 0
		gDiffBTDiffCount     int     = 0
	)

	for i := 0; i < len(frames); i += 1 {
		diffBrightness := frames[i].Brightness - statistics.BrightnessMovingMean[i]
		if diffBrightness > 0 {
			gDiffBrightnessValue += diffBrightness
			gDiffBrightnessCount += 1
		}

		diffColorDiff := frames[i].ColorDifference - statistics.ColorDifferenceMovingMean[i]
		if diffColorDiff > 0 {
			gDiffColorDiffValue += diffColorDiff
			gDiffColorDiffCount += 1
		}

		diffBTDiff := frames[i].BinaryThresholdDifference - statistics.BinaryThresholdDifferenceMovingMean[i]
		if diffBTDiff > 0 {
			gDiffBTDiffValue += diffBTDiff
			gDiffBTDiffCount += 1
		}
	}

	gDiffBrightnessValue /= float64(gDiffBrightnessCount)
	gDiffColorDiffValue /= float64(gDiffColorDiffCount)
	gDiffBTDiffValue /= float64(gDiffBTDiffCount)

	defaultOptions := GetDefaultDetectorOptions()

	if defaultOptions.BrightnessDetectionThreshold == defaultOptions.BrightnessDetectionThreshold {
		detector.options.BrightnessDetectionThreshold = gDiffBrightnessValue
	} else {
		detector.renderer.LogWarning("The brightness detection threshold (%f) value was explicitly specified and would not be replace by the auto-calculated one (%f)",
			detector.options.BrightnessDetectionThreshold,
			gDiffBrightnessValue)
	}

	if defaultOptions.ColorDifferenceDetectionThreshold == defaultOptions.ColorDifferenceDetectionThreshold {
		detector.options.ColorDifferenceDetectionThreshold = gDiffColorDiffValue
	} else {
		detector.renderer.LogWarning("The color difference detection threshold (%f) value was explicitly specified and would not be replace by the auto-calculated one (%f)",
			detector.options.ColorDifferenceDetectionThreshold,
			gDiffColorDiffValue)
	}

	if defaultOptions.BinaryThresholdDifferenceDetectionThreshold == defaultOptions.BinaryThresholdDifferenceDetectionThreshold {
		detector.options.BinaryThresholdDifferenceDetectionThreshold = gDiffBTDiffValue
	} else {
		detector.renderer.LogWarning("The binary threshold detection threshold (%f) value was explicitly specified and would not be replace by the auto-calculated one (%f)",
			detector.options.BinaryThresholdDifferenceDetectionThreshold,
			gDiffBTDiffValue)
	}

	detector.renderer.LogDebug("Auto thresholds calculation stage finished. Stage took: %s", time.Since(autoThresholdTime))
}

// Helper function used to filter out indecies representing frames wihich meet the requirement thresholds.
func (detector *detector) performVideoDetection(framesCollection *frame.FramesCollection) []int {
	videoDetectionTime := time.Now()
	detector.renderer.LogDebug("Starting the video detection stage.")

	detections := CreateDetectionBuffer()

	frames := framesCollection.GetAll()
	statistics := framesCollection.CalculateStatistics(int(detector.options.MovingMeanResolution))

	progressBarStep, progressBarClose := detector.renderer.Progress("Video detection stage.", len(frames))

	for frameIndex, frame := range frames {
		logPrefix := fmt.Sprintf("Frame: [%d/%d].", frameIndex+1, len(frames))
		detector.renderer.LogDebug("%s Checking frame thresholds.", logPrefix)

		if frame.Brightness < detector.options.BrightnessDetectionThreshold+statistics.BrightnessMovingMean[frameIndex] {
			detector.renderer.LogDebug("%s Frame brightenss requirements not met. (%f < %f + %f)",
				logPrefix,
				frame.Brightness,
				detector.options.BrightnessDetectionThreshold,
				statistics.BrightnessMovingMean[frameIndex])

			detections.Append(frameIndex, false)
			progressBarStep()
			continue
		}

		if frame.ColorDifference < detector.options.ColorDifferenceDetectionThreshold+statistics.ColorDifferenceMovingMean[frameIndex] {
			detector.renderer.LogDebug("%s Frame color difference requirements not met. (%f < %f + %f)",
				logPrefix,
				frame.ColorDifference,
				detector.options.ColorDifferenceDetectionThreshold,
				statistics.ColorDifferenceMovingMean[frameIndex])

			detections.Append(frameIndex, false)
			progressBarStep()
			continue
		}

		if frame.BinaryThresholdDifference < detector.options.BinaryThresholdDifferenceDetectionThreshold+statistics.BinaryThresholdDifferenceMovingMean[frameIndex] {
			detector.renderer.LogDebug("%s Frame binary threshold difference requirements not met. (%f < %f + %f)",
				logPrefix,
				frame.BinaryThresholdDifference,
				detector.options.BinaryThresholdDifferenceDetectionThreshold,
				statistics.BinaryThresholdDifferenceMovingMean[frameIndex])

			detections.Append(frameIndex, false)
			progressBarStep()
			continue
		}

		detector.renderer.LogInfo("%s Frame meets the threshold requirements.", logPrefix)
		detections.Append(frameIndex, true)

		progressBarStep()
	}

	progressBarClose()
	detector.renderer.LogDebug("Video detection stage finished. Stage took: %s", time.Since(videoDetectionTime))
	return detections.Resolve()
}

// Helper function used to export frames which meet the requirement thresholds to png files.
func (detector *detector) performFramesExport(inputVideoPath, outputDirectoryPath string, detections []int) error {
	framesExportTime := time.Now()
	detector.renderer.LogDebug("Starting the frames export stage.")
	detector.renderer.LogInfo("About to export %d frames.", len(detections))

	video, err := vidio.NewVideo(inputVideoPath)
	if err != nil {
		return fmt.Errorf("detector: failed to open the video file for the frames export stage: %w", err)
	}

	defer video.Close()

	// TODO: Limit for large detections
	frames, err := video.ReadFrames(detections...)
	if err != nil {
		return fmt.Errorf("detector: failed to read the specified frames from the video: %w", err)
	}

	progressBarStep, progressBarClose := detector.renderer.Progress("Video frames export stage.", len(detections))

	for index, frame := range frames {
		frameIndex := detections[index]

		frameImageName := fmt.Sprintf("frame-%d.png", frameIndex+1)
		frameImagePath := path.Join(outputDirectoryPath, frameImageName)
		if err := utils.ExportImageAsPng(frameImagePath, frame); err != nil {
			return fmt.Errorf("detector: failed to export the frame image: %w", err)
		}

		progressBarStep()
		detector.renderer.LogInfo("Frame: [%d/%d]. Frame image exported at: %s", frameIndex+1, video.Frames(), frameImagePath)
	}

	progressBarClose()
	detector.renderer.LogDebug("Frames export stage finished. Stage took: %s", time.Since(framesExportTime))
	return nil
}

// Helper function used to print out descriptive statistics aboout the frames collection
func (detector *detector) performStatisticsLogging(framesCollection *frame.FramesCollection) {
	statistics := framesCollection.CalculateStatistics(int(detector.options.MovingMeanResolution))

	values := [][]string{
		{"Frame brightness mean", strconv.FormatFloat(statistics.BrightnessMean, 'f', -1, 64)},
		{"Frame brightness standard deviation", strconv.FormatFloat(statistics.BrightnessStandardDeviation, 'f', -1, 64)},
		{"Frame brightness max", strconv.FormatFloat(statistics.BrightnessMax, 'f', -1, 64)},
		{"Frame color difference mean", strconv.FormatFloat(statistics.ColorDifferenceMean, 'f', -1, 64)},
		{"Frame color difference standard deviation", strconv.FormatFloat(statistics.ColorDifferenceStandardDeviation, 'f', -1, 64)},
		{"Frame color difference max", strconv.FormatFloat(statistics.ColorDifferenceMax, 'f', -1, 64)},
		{"Frame color binary threshold mean", strconv.FormatFloat(statistics.BinaryThresholdDifferenceMean, 'f', -1, 64)},
		{"Frame color binary threshold standard deviation", strconv.FormatFloat(statistics.BinaryThresholdDifferenceStandardDeviation, 'f', -1, 64)},
		{"Frame color binary threshold max", strconv.FormatFloat(statistics.BinaryThresholdDifferenceMax, 'f', -1, 64)},
	}

	detector.renderer.Table(values)
}

// Helper function used to export the frames collection report in the CSV format.
func (detector *detector) handleCsvReportExport(outputDirectoryPath string, frames *frame.FramesCollection) error {
	csvSpinnerStop := detector.renderer.Spinner("Exporting report in CSV format")
	defer csvSpinnerStop()

	csvFramesReportPath := path.Join(outputDirectoryPath, "frames-report.csv")
	framesReportFile, err := utils.CreateFileWithTree(csvFramesReportPath)
	if err != nil {
		return fmt.Errorf("detector: failed to create the csv frames report file: %w", err)
	}

	defer func() {
		if err := framesReportFile.Close(); err != nil {
			panic(err)
		}
	}()

	if err := frames.ExportCsvReport(framesReportFile); err != nil {
		return fmt.Errorf("detector: failed to export the csv frames report: %w", err)
	} else {
		detector.renderer.LogInfo("Frames report in CSV format exported to: %s", csvFramesReportPath)
	}

	csvStatisticsReportPath := path.Join(outputDirectoryPath, "statistics-report.csv")
	statisticsReportFile, err := utils.CreateFileWithTree(csvStatisticsReportPath)
	if err != nil {
		return fmt.Errorf("detector: failed to create the csv statistics report file: %w", err)
	}

	defer func() {
		if err := statisticsReportFile.Close(); err != nil {
			panic(err)
		}
	}()

	statistics := frames.CalculateStatistics(int(detector.options.MovingMeanResolution))
	if err := statistics.ExportCsvReport(statisticsReportFile); err != nil {
		return fmt.Errorf("detector: failed to export the csv statistics report: %w", err)
	} else {
		detector.renderer.LogInfo("Statistics report in CSV format exported to: %s", csvStatisticsReportPath)
	}

	return nil
}

// Helper function used to export the frames collection report in the JSON format.
func (detector *detector) handleJsonReportExport(outputDirectoryPath string, frames *frame.FramesCollection) error {
	jsonSpinnerClose := detector.renderer.Spinner("Exporting the frames report in JSON format.")
	defer jsonSpinnerClose()

	jsonFramesReportPath := path.Join(outputDirectoryPath, "frames-report.json")
	framesReportFile, err := utils.CreateFileWithTree(jsonFramesReportPath)
	if err != nil {
		return fmt.Errorf("detector: failed to create the json frames report file: %w", err)
	}

	defer func() {
		if err := framesReportFile.Close(); err != nil {
			panic(err)
		}
	}()

	if err := frames.ExportJsonReport(framesReportFile); err != nil {
		return fmt.Errorf("detector: failed to export the json frames report: %w", err)
	} else {
		detector.renderer.LogInfo("Frames report in JSON format exported to: %s", jsonFramesReportPath)
	}

	jsonStatisticsReportPath := path.Join(outputDirectoryPath, "statistics-report.json")
	statisticsReportFile, err := utils.CreateFileWithTree(jsonStatisticsReportPath)
	if err != nil {
		return fmt.Errorf("detector: failed to create the json statistics report file: %w", err)
	}

	defer func() {
		if err := statisticsReportFile.Close(); err != nil {
			panic(err)
		}
	}()

	statistics := frames.CalculateStatistics(int(detector.options.MovingMeanResolution))
	if err := statistics.ExportJsonReport(statisticsReportFile); err != nil {
		return fmt.Errorf("detector: failed to export the json statistics report: %w", err)
	} else {
		detector.renderer.LogInfo("Statistics report in JSON format exported to %s", jsonStatisticsReportPath)
	}

	return nil
}
