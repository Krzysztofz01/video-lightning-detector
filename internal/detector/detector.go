package detector

import (
	"errors"
	"fmt"
	"image"
	"math"
	"os"
	"path"
	"time"

	vidio "github.com/AlexEidt/Vidio"
	"github.com/Krzysztofz01/video-lightning-detector/internal/export"
	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/render"
	"github.com/Krzysztofz01/video-lightning-detector/internal/statistics"
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

	var frames frame.FrameCollection

	frames, err := detector.GetAnalyzedFrames(inputVideoPath, outputDirectoryPath)
	if err != nil {
		return fmt.Errorf("detector: video analysis stage failed: %w", err)
	}

	descriptiveStatistics := statistics.CreateDescriptiveStatistics(frames, int(detector.options.MovingMeanResolution))

	if detector.options.AutoThresholds {
		detector.ApplyAutoThresholds(frames, descriptiveStatistics)
	}

	detections := detector.PerformVideoDetection(frames, descriptiveStatistics)

	if detector.options.ImportPreanalyzed {
		if err := detector.ExportPreanalyzedFrames(frames, outputDirectoryPath); err != nil {
			return fmt.Errorf("detector: preanalyzed frames export stage failed: %w", err)
		}
	}

	if err := detector.PerformExports(inputVideoPath, outputDirectoryPath, frames, descriptiveStatistics, detections); err != nil {
		return fmt.Errorf("detector: export stage failed: %w", err)
	}

	detector.renderer.LogInfo("Lightning hunting took: %s", time.Since(runTime))
	return nil
}

// Helper function used to perform the analysis of the video frames. Depending on the options, this function will perform
// the analysis or import the result of the previous analysis with a fallback to a standard analysis.
func (detector *detector) GetAnalyzedFrames(inputVideoPath, outputDirectoryPath string) (frame.FrameCollection, error) {
	var (
		frames         frame.FrameCollection
		wasPreanalzyed bool
		err            error
	)

	if detector.options.ImportPreanalyzed {
		frames, wasPreanalzyed, err = detector.ImportPreanalyzedFrames(outputDirectoryPath)
		if err != nil {
			return nil, fmt.Errorf("detector: failed to import the preanalyzed frames: %w", err)
		}

		if wasPreanalzyed {
			detector.renderer.LogInfo("Importing the pre-analyzed frames data.")
			return frames, nil
		}

		detector.renderer.LogWarning("No exported pre-analzyed frames JSON file found. Fallback to frames analysis.")
	}

	if frames, err = detector.PerformFramesAnalysis(inputVideoPath); err != nil {
		return nil, fmt.Errorf("detector: failed to perform the frames analysis: %w", err)
	} else {
		return frames, nil
	}
}

// Helper function used to iterate over the video frames in order to generate a collection of frames instances containing
// processed values about given frames and neighbouring frames relations.
func (detector *detector) PerformFramesAnalysis(inputVideoPath string) (frame.FrameCollection, error) {
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
	frames := frame.CreateNewFrameCollection(frameCount)

	progressBarStep, progressBarClose := detector.renderer.Progress("Video analysis stage.", frameCount)

	for video.Read() {
		if err := utils.ScaleImage(frameCurrentBuffer, frameCurrent, detector.options.FrameScalingFactor); err != nil {
			return nil, fmt.Errorf("detector: failed to scale the current frame image on the analyze stage: %w", err)
		}

		if detector.options.Denoise {
			if err := utils.BlurImage(frameCurrent, frameCurrent, 8); err != nil {
				return nil, fmt.Errorf("detector: failed to blur the current frame image on the analyze stage: %w", err)
			}
		}

		frame := frame.CreateNewFrame(frameCurrent, framePrevious, frameNumber, frame.BinaryThresholdParam)
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

// Helper function used to import the pre-analyzed frames collection from the JSON export file.
func (detector *detector) ImportPreanalyzedFrames(outputDirectoryPath string) (frame.FrameCollection, bool, error) {
	frameCollectionCachePath := path.Join(outputDirectoryPath, FrameCollectionCacheFilename)
	if !utils.FileExists(frameCollectionCachePath) {
		return nil, false, nil
	}

	frameCollectionCacheFile, err := os.Open(frameCollectionCachePath)
	if err != nil {
		return nil, true, fmt.Errorf("detector: failed to open the frame collection cache with preanalyzed frames: %w", err)
	}

	defer func() {
		if err := frameCollectionCacheFile.Close(); err != nil {
			panic(err)
		}
	}()

	optionsChecksum, err := detector.options.GetChecksum()
	if err != nil {
		return nil, true, fmt.Errorf("detector: failed to access the detector options checksum: %w", err)
	}

	frames, checksum, err := frame.ImportCachedFrameCollection(frameCollectionCacheFile)
	if err != nil {
		return nil, true, fmt.Errorf("detector: failed to import the json frames report with preanalyzed frames: %w", err)
	}

	if optionsChecksum != checksum {
		return nil, false, nil
	}

	return frames, true, nil
}

func (detector *detector) ExportPreanalyzedFrames(fc frame.FrameCollection, outputDirectoryPath string) error {
	frameCollectionCachePath := path.Join(outputDirectoryPath, FrameCollectionCacheFilename)

	var (
		frameCollectionCacheFile *os.File
		optionsChecksum          string
		err                      error
	)

	if optionsChecksum, err = detector.options.GetChecksum(); err != nil {
		return fmt.Errorf("detector: failed to access the options checksum: %w", err)
	}

	defer func() {
		if frameCollectionCacheFile == nil {
			return
		}

		if err := frameCollectionCacheFile.Close(); err != nil {
			panic(err)
		}
	}()

	if utils.FileExists(frameCollectionCachePath) {
		frameCollectionCacheFile, err = os.Open(frameCollectionCachePath)
		if err != nil {
			return fmt.Errorf("detector: failed to open the frame collection cache with preanalyzed frames: %w", err)
		}

		var importedChecksum string
		if _, importedChecksum, err = frame.ImportCachedFrameCollection(frameCollectionCacheFile); err != nil {
			return fmt.Errorf("detector: failed to access the cached frame collection: %w", err)
		}

		if optionsChecksum == importedChecksum {
			return nil
		}
	}

	frameCollectionCacheFile, err = utils.CreateFileWithTree(frameCollectionCachePath)
	if err != nil {
		return fmt.Errorf("detector: failed to creatae the frame collection cache with preanalyzed frames: %w", err)
	}

	if err := fc.ExportCache(frameCollectionCacheFile, optionsChecksum); err != nil {
		return fmt.Errorf("detector: failed to export the preanalyzed frames cache: %w", err)
	}

	return nil
}

// Helper function used to auto-calculate the detection thresholds based on the frames and apply the threshold to the detector options
// TODO: More versatile results could be achived using "moving standard deviation"???
// TODO: Render thresholds in a table
func (detector *detector) ApplyAutoThresholds(framesCollection frame.FrameCollection, ds statistics.DescriptiveStatistics) {
	autoThresholdTime := time.Now()
	detector.renderer.LogDebug("Starting the auto thresholds calculation stage.")

	frames := framesCollection.GetAll()

	var (
		gDiffBrightnessValue float64 = 0
		gDiffBrightnessCount int     = 0
		gDiffColorDiffValue  float64 = 0
		gDiffColorDiffCount  int     = 0
		gDiffBTDiffValue     float64 = 0
		gDiffBTDiffCount     int     = 0
	)

	for i := 0; i < len(frames); i += 1 {
		diffBrightness := frames[i].Brightness - ds.BrightnessMovingMean[i]
		if diffBrightness > 0 {
			gDiffBrightnessValue += diffBrightness
			gDiffBrightnessCount += 1
		}

		diffColorDiff := frames[i].ColorDifference - ds.ColorDifferenceMovingMean[i]
		if diffColorDiff > 0 {
			gDiffColorDiffValue += diffColorDiff
			gDiffColorDiffCount += 1
		}

		diffBTDiff := frames[i].BinaryThresholdDifference - ds.BinaryThresholdDifferenceMovingMean[i]
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
func (detector *detector) PerformVideoDetection(framesCollection frame.FrameCollection, ds statistics.DescriptiveStatistics) []int {
	videoDetectionTime := time.Now()
	detector.renderer.LogDebug("Starting the video detection stage.")

	detections := CreateDetectionBuffer()

	frames := framesCollection.GetAll()

	progressBarStep, progressBarClose := detector.renderer.Progress("Video detection stage.", len(frames))

	for frameIndex, frame := range frames {
		logPrefix := fmt.Sprintf("Frame: [%d/%d].", frameIndex+1, len(frames))
		detector.renderer.LogDebug("%s Checking frame thresholds.", logPrefix)

		if frame.Brightness < detector.options.BrightnessDetectionThreshold+ds.BrightnessMovingMean[frameIndex] {
			detector.renderer.LogDebug("%s Frame brightenss requirements not met. (%f < %f + %f)",
				logPrefix,
				frame.Brightness,
				detector.options.BrightnessDetectionThreshold,
				ds.BrightnessMovingMean[frameIndex])

			detections.Append(frameIndex, false)
			progressBarStep()
			continue
		}

		if frame.ColorDifference < detector.options.ColorDifferenceDetectionThreshold+ds.ColorDifferenceMovingMean[frameIndex] {
			detector.renderer.LogDebug("%s Frame color difference requirements not met. (%f < %f + %f)",
				logPrefix,
				frame.ColorDifference,
				detector.options.ColorDifferenceDetectionThreshold,
				ds.ColorDifferenceMovingMean[frameIndex])

			detections.Append(frameIndex, false)
			progressBarStep()
			continue
		}

		if frame.BinaryThresholdDifference < detector.options.BinaryThresholdDifferenceDetectionThreshold+ds.BinaryThresholdDifferenceMovingMean[frameIndex] {
			detector.renderer.LogDebug("%s Frame binary threshold difference requirements not met. (%f < %f + %f)",
				logPrefix,
				frame.BinaryThresholdDifference,
				detector.options.BinaryThresholdDifferenceDetectionThreshold,
				ds.BinaryThresholdDifferenceMovingMean[frameIndex])

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

// Helper function used to perform exports to varius formats selected via the options
func (detector *detector) PerformExports(inputVideoPath, outputDirectoryPath string, fc frame.FrameCollection, ds statistics.DescriptiveStatistics, detections []int) error {
	if err := export.RenderDescriptiveStatistics(detector.renderer, ds); err != nil {
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

		detector.renderer.LogDebug("Frames used as actual detection classification: %v", actualClassification)

		confusionMatrix = statistics.CreateConfusionMatrix(actualClassification, detections, fc.Count())
	}

	if detector.options.ExportCsvReport {
		csvSpinnerStop := detector.renderer.Spinner("Exporting reports in CSV format")
		defer csvSpinnerStop()

		if path, err := export.ExportCsvFrames(outputDirectoryPath, fc); err != nil {
			return fmt.Errorf("detector: failed to export csv frames report: %w", err)
		} else {
			detector.renderer.LogInfo("Frames report in CSV format exported to: %s", path)
		}

		if path, err := export.ExportCsvDescriptiveStatistics(outputDirectoryPath, ds); err != nil {
			return fmt.Errorf("detector: failed to export csv descriptive statistics report: %w", err)
		} else {
			detector.renderer.LogInfo("Descriptive statistics in CSV format exported to %s", path)
		}

		if detector.options.ExportConfusionMatrix {
			if path, err := export.ExportCsvConfusionMatrix(outputDirectoryPath, confusionMatrix); err != nil {
				return fmt.Errorf("detector: failed to export csv confusion matrix report: %w", err)
			} else {
				detector.renderer.LogInfo("Confusion matrix in CSV format exported to %s", path)
			}
		}

		csvSpinnerStop()
	}

	if detector.options.ExportJsonReport {
		jsonSpinnerStop := detector.renderer.Spinner("Exporting reports in JSON format")
		defer jsonSpinnerStop()

		if path, err := export.ExportJsonFrames(outputDirectoryPath, fc); err != nil {
			return fmt.Errorf("detector: failed to export json frames report: %w", err)
		} else {
			detector.renderer.LogInfo("Frames report in JSON format exported to: %s", path)
		}

		if path, err := export.ExportJsonDescriptiveStatistics(outputDirectoryPath, ds); err != nil {
			return fmt.Errorf("detector: failed to export json descriptive statistics report: %w", err)
		} else {
			detector.renderer.LogInfo("Descriptive statistics in JSON format exported to %s", path)
		}

		if detector.options.ExportConfusionMatrix {
			if path, err := export.ExportJsonConfusionMatrix(outputDirectoryPath, confusionMatrix); err != nil {
				return fmt.Errorf("detector: failed to export json confusion matrix report: %w", err)
			} else {
				detector.renderer.LogInfo("Confusion matrix in JSON format exported to %s", path)
			}
		}

		jsonSpinnerStop()
	}

	if detector.options.ExportChartReport {
		chartSpinnerStop := detector.renderer.Spinner("Exporting chart report")
		defer chartSpinnerStop()

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
			detector.renderer.LogInfo("Frames chart exported to: %s", path)
		}

		chartSpinnerStop()
	}

	if detector.options.ExportConfusionMatrix {
		confusionMatrixSpinnerStop := detector.renderer.Spinner("Exporting confusion matrix")
		defer confusionMatrixSpinnerStop()

		if err := export.RenderConfusionMatrix(detector.renderer, confusionMatrix); err != nil {
			return fmt.Errorf("detector: failed to export the confusion matrix: %w", err)
		}

		confusionMatrixSpinnerStop()
	}

	return nil
}

// NOTE: Experimental sampling of binary threshold from recording
func (detector *detector) SampleBinaryThreshold(inputVideoPath string) (float64, error) {
	video, err := vidio.NewVideo(inputVideoPath)
	if err != nil {
		return 0, fmt.Errorf("detector: failed to open the video file for the binary threshold sampling stage: %w", err)
	}

	defer video.Close()

	var (
		frames      int     = video.Frames()
		fps         float64 = video.FPS()
		duration    float64 = float64(frames) / fps
		sampleCount int     = int(math.Max(1.43*math.Log(duration/60)-0.64, 1))
		framesStep  int     = frames / sampleCount
	)

	var sampleFrameIndexes []int = make([]int, 0, sampleCount)
	for sampleIndex := 0; sampleIndex < sampleCount; sampleIndex += 1 {
		sampleFrameIndexes = append(sampleFrameIndexes, sampleIndex*framesStep)
	}

	sampleFrames, err := video.ReadFrames(sampleFrameIndexes...)
	if err != nil {
		return 0, fmt.Errorf("detector: failed to read the specified frames from the video: %w", err)
	}

	var thresholdSum float64 = 0
	for _, sampleFrame := range sampleFrames {
		thresholdSum += utils.Otsu(sampleFrame)

	}

	return thresholdSum / float64(sampleCount), nil
}

// Helper function used to export frame images which meet the requirement thresholds to png files.
func (detector *detector) PerformFrameImagesExport(inputVideoPath, outputDirectoryPath string, detections []int) error {
	framesExportTime := time.Now()
	detector.renderer.LogDebug("Starting the frames export stage.")
	detector.renderer.LogInfo("About to export %d frames.", len(detections))

	video, err := vidio.NewVideo(inputVideoPath)
	if err != nil {
		return fmt.Errorf("detector: failed to open the video file for the frames export stage: %w", err)
	}

	defer video.Close()

	progressBarStep, progressBarClose := detector.renderer.Progress("Video frames export stage.", len(detections))

	// TODO: Limit for large detections
	frames, err := video.ReadFrames(detections...)
	if err != nil {
		return fmt.Errorf("detector: failed to read the specified frames from the video: %w", err)
	}

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
