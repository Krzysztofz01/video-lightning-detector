package detector

import (
	"fmt"
	"image"
	"path"
	"time"

	vidio "github.com/AlexEidt/Vidio"
	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/draw"
)

// Detector instance that is able to perform a search after ligntning strikes on a video file.
type Detector interface {
	Run(inputVideoPath, outputDirectoryPath string) error
}

type detector struct {
	options DetectorOptions
}

// Create a new video lightning detector instance with the specified options.
func CreateDetector(options DetectorOptions) (Detector, error) {
	if ok, msg := options.AreValid(); !ok {
		return nil, fmt.Errorf("detector: invalid options %s", msg)
	}

	logrus.Debugf("Detector created with options: %+v", options)

	return &detector{
		options: options,
	}, nil
}

// Perform a lightning detection on the provided video specified by the file path and store the results at the specified directory path.
func (detector *detector) Run(inputVideoPath, outputDirectoryPath string) error {
	runTime := time.Now()
	logrus.Debugln("Starting the lightning hunt.")

	frames, err := detector.performVideoAnalysis(inputVideoPath)
	if err != nil {
		return fmt.Errorf("detector: analysis run stage failed: %w", err)
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

	logrus.Debugf("Lightning hunting took: %s", time.Since(runTime))
	return nil
}

// Helper function used to iterate over the video frames in order to generate a collection of frames instances containing
// processed values about given frames and neighbouring frames relations.
func (detector *detector) performVideoAnalysis(inputVideoPath string) (*frame.FramesCollection, error) {
	videoAnalysisTime := time.Now()
	logrus.Debugln("Starting the video analysis stage.")

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

	for video.Read() {
		if detector.options.FrameScalingFactor == 1.0 {
			copy(frameCurrent.Pix, frameCurrentBuffer.Pix)
		} else {
			draw.NearestNeighbor.Scale(frameCurrent, frameCurrent.Rect, frameCurrentBuffer, frameCurrentBuffer.Bounds(), draw.Over, nil)
		}

		if detector.options.Denoise {
			if err := utils.BlurImage(frameCurrent, frameCurrent, 8); err != nil {
				return nil, fmt.Errorf("detector: failed to blur the current frame image on the analyze stage: %w", err)
			}
		}

		frame := frame.CreateNewFrame(frameCurrent, framePrevious, frameNumber)
		frames.Append(frame)

		logrus.Infof("Frame: [%d/%d]. Brightness: %f ColorDiff: %f BTDiff: %f", frameNumber, frameCount, frame.Brightness, frame.ColorDifference, frame.BinaryThresholdDifference)

		frameNumber += 1
		copy(framePrevious.Pix, frameCurrent.Pix)
	}

	logrus.Debugf("Video analysis stage finished. Stage took: %s", time.Since(videoAnalysisTime))
	return frames, nil
}

// Helper function used to auto-calculate the detection thresholds based on the frames and apply the threshold to the detector options
// TODO: More versatile results could be achived using "moving standard deviation"???
func (detector *detector) applyAutoThresholds(framesCollection *frame.FramesCollection) {
	autoThresholdTime := time.Now()
	logrus.Debugln("Starting the auto thresholds calculation stage.")

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
		logrus.Warnf("The brightness detection threshold (%f) value was explicitly specified and would not be replace by the auto-calculated one (%f)",
			detector.options.BrightnessDetectionThreshold,
			gDiffBrightnessValue)
	}

	if defaultOptions.ColorDifferenceDetectionThreshold == defaultOptions.ColorDifferenceDetectionThreshold {
		detector.options.ColorDifferenceDetectionThreshold = gDiffColorDiffValue
	} else {
		logrus.Warnf("The color difference detection threshold (%f) value was explicitly specified and would not be replace by the auto-calculated one (%f)",
			detector.options.ColorDifferenceDetectionThreshold,
			gDiffColorDiffValue)
	}

	if defaultOptions.BinaryThresholdDifferenceDetectionThreshold == defaultOptions.BinaryThresholdDifferenceDetectionThreshold {
		detector.options.BinaryThresholdDifferenceDetectionThreshold = gDiffBTDiffValue
	} else {
		logrus.Warnf("The binary threshold detection threshold (%f) value was explicitly specified and would not be replace by the auto-calculated one (%f)",
			detector.options.BinaryThresholdDifferenceDetectionThreshold,
			gDiffBTDiffValue)
	}

	logrus.Debugf("Auto thresholds calculation stage finished. Stage took: %s", time.Since(autoThresholdTime))
}

// Helper function used to filter out indecies representing frames wihich meet the requirement thresholds.
func (detector *detector) performVideoDetection(framesCollection *frame.FramesCollection) []int {
	videoDetectionTime := time.Now()
	logrus.Debugf("Starting the video detection stage.")

	detections := CreateDetectionBuffer()

	frames := framesCollection.GetAll()
	statistics := framesCollection.CalculateStatistics(int(detector.options.MovingMeanResolution))

	for frameIndex, frame := range frames {
		logPrefix := fmt.Sprintf("Frame: [%d/%d].", frameIndex+1, len(frames))
		logrus.Infof("%s Checking frame thresholds.", logPrefix)

		if frame.Brightness < detector.options.BrightnessDetectionThreshold+statistics.BrightnessMovingMean[frameIndex] {
			logrus.Debugf("%s Frame brightenss requirements not met. (%f < %f + %f)",
				logPrefix,
				frame.Brightness,
				detector.options.BrightnessDetectionThreshold,
				statistics.BrightnessMovingMean[frameIndex])

			detections.Append(frameIndex, false)
			continue
		}

		if frame.ColorDifference < detector.options.ColorDifferenceDetectionThreshold+statistics.ColorDifferenceMovingMean[frameIndex] {
			logrus.Debugf("%s Frame color difference requirements not met. (%f < %f + %f)",
				logPrefix,
				frame.ColorDifference,
				detector.options.ColorDifferenceDetectionThreshold,
				statistics.ColorDifferenceMovingMean[frameIndex])

			detections.Append(frameIndex, false)
			continue
		}

		if frame.BinaryThresholdDifference < detector.options.BinaryThresholdDifferenceDetectionThreshold+statistics.BinaryThresholdDifferenceMovingMean[frameIndex] {
			logrus.Debugf("%s Frame binary threshold difference requirements not met. (%f < %f + %f)",
				logPrefix,
				frame.BinaryThresholdDifference,
				detector.options.BinaryThresholdDifferenceDetectionThreshold,
				statistics.BinaryThresholdDifferenceMovingMean[frameIndex])

			detections.Append(frameIndex, false)
			continue
		}

		logrus.Infof("%s Frame meets the threshold requirements.", logPrefix)
		detections.Append(frameIndex, true)
	}

	logrus.Debugf("Video detection stage finished. Stage took: %s", time.Since(videoDetectionTime))
	return detections.Resolve()
}

// Helper function used to export frames which meet the requirement thresholds to png files.
func (detector *detector) performFramesExport(inputVideoPath, outputDirectoryPath string, detections []int) error {
	framesExportTime := time.Now()
	logrus.Debugf("Starting the frames export stage.")
	logrus.Debugf("About to export %d frames.", len(detections))

	video, err := vidio.NewVideo(inputVideoPath)
	if err != nil {
		return fmt.Errorf("detector: failed to open the video file for the frames export stage: %w", err)
	}

	defer video.Close()

	frameCurrentBuffer := image.NewRGBA(image.Rect(0, 0, video.Width(), video.Height()))
	video.SetFrameBuffer(frameCurrentBuffer.Pix)

	indexVideo := 0
	indexDetections := 0
	videoFramesCount := video.Frames()

	for video.Read() && indexDetections < len(detections) {
		if indexVideo == detections[indexDetections] {
			frameImageName := fmt.Sprintf("frame-%d.png", indexVideo+1)
			frameImagePath := path.Join(outputDirectoryPath, frameImageName)
			if err := utils.ExportImageAsPng(frameImagePath, frameCurrentBuffer); err != nil {
				return fmt.Errorf("detector: failed to export the frame image: %w", err)
			}

			logrus.Infof("Frame: [%d/%d]. Frame image exported at: %s", indexVideo+1, videoFramesCount, frameImagePath)
			indexDetections += 1
		}

		indexVideo += 1
	}

	logrus.Debugf("Frames export stage finished. Stage took: %s", time.Since(framesExportTime))
	return nil
}

// Helper function used to print out descriptive statistics aboout the frames collection
func (detector *detector) performStatisticsLogging(framesCollection *frame.FramesCollection) {
	statistics := framesCollection.CalculateStatistics(int(detector.options.MovingMeanResolution))

	logrus.Infof("Frame brightness mean: %f", statistics.BrightnessMean)
	logrus.Infof("Frame brightness standard deviation: %f", statistics.BrightnessStandardDeviation)
	logrus.Infof("Frame brightness max: %f", statistics.BrightnessMax)
	logrus.Infof("Frame color difference mean: %f", statistics.ColorDifferenceMean)
	logrus.Infof("Frame color difference standard deviation: %f", statistics.ColorDifferenceStandardDeviation)
	logrus.Infof("Frame color difference max: %f", statistics.ColorDifferenceMax)
	logrus.Infof("Frame color binary threshold mean: %f", statistics.BinaryThresholdDifferenceMean)
	logrus.Infof("Frame color binary threshold standard deviation: %f", statistics.BinaryThresholdDifferenceStandardDeviation)
	logrus.Infof("Frame color binary threshold max: %f", statistics.BinaryThresholdDifferenceMax)
}

// Helper function used to export the frames collection report in the CSV format.
func (detector *detector) handleCsvReportExport(outputDirectoryPath string, frames *frame.FramesCollection) error {
	logrus.Info("Exporting reports in CSV format.")
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
		logrus.Infof("Frames report in CSV format exported to: %s", csvFramesReportPath)
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
		logrus.Infof("Statistics report in CSV format exported to: %s", csvStatisticsReportPath)
	}

	return nil
}

// Helper function used to export the frames collection report in the JSON format.
func (detector *detector) handleJsonReportExport(outputDirectoryPath string, frames *frame.FramesCollection) error {
	logrus.Info("Exporting the frames report in JSON format.")
	jsonReportPath := path.Join(outputDirectoryPath, "report.json")
	file, err := utils.CreateFileWithTree(jsonReportPath)
	if err != nil {
		return fmt.Errorf("detector: failed to create the json report file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	if err := frames.ExportJsonReport(file); err != nil {
		return fmt.Errorf("detector: failed to export the json report: %w", err)
	}

	logrus.Infof("Frames report in JSON format exported to: %s", jsonReportPath)
	return nil
}
