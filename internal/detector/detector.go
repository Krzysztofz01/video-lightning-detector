package detector

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Krzysztofz01/video-lightning-detector/internal/analyzer"
	"github.com/Krzysztofz01/video-lightning-detector/internal/export"
	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/printer"
	"github.com/Krzysztofz01/video-lightning-detector/internal/statistics"
)

// Detector instance that is able to perform a search after ligntning strikes on a video file.
type Detector interface {
	Run(inputVideoPath, outputDirectoryPath string, ctx context.Context) error
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
func (detector *detector) Run(inputVideoPath, outputDirectoryPath string, ctx context.Context) error {
	runTime := time.Now()
	detector.printer.InfoA("Starting the lightning hunt.")

	analyzer := analyzer.NewAnalyzer(inputVideoPath, outputDirectoryPath, detector.options, detector.printer)

	frames, err := analyzer.GetFrames(ctx)
	if err != nil {
		return fmt.Errorf("detector: video analysis stage failed: %w", err)
	}

	descriptiveStatistics := statistics.CreateDescriptiveStatistics(frames, int(detector.options.MovingMeanResolution))

	if detector.options.AutoThresholds {
		threshold := NewAutoThreshold(frames, descriptiveStatistics, detector.printer)

		if options, err := threshold.ApplyToOptions(detector.options, AboveMeanOfDeviations); err != nil {
			return fmt.Errorf("detector: failed to perform the auto-threshold calculation: %w", err)
		} else {
			detector.options = options
		}
	}

	detections, err := detector.PerformVideoDetection(frames, descriptiveStatistics)
	if err != nil {
		return fmt.Errorf("detector: video detection stage failed: %w", err)
	}

	exporter := export.NewExporter(inputVideoPath, outputDirectoryPath, detector.options, detector.printer)
	if err := exporter.Export(frames, descriptiveStatistics, detections); err != nil {
		return fmt.Errorf("detector: export stage failed: %w", err)
	}

	detector.printer.InfoA("Lightning hunting took: %s", time.Since(runTime))
	return nil
}

// Helper function used to filter out indecies representing frames wihich meet the requirement thresholds.
func (detector *detector) PerformVideoDetection(framesCollection frame.FrameCollection, ds statistics.DescriptiveStatistics) ([]int, error) {
	videoDetectionTime := time.Now()
	detector.printer.Debug("Starting the video detection stage.")

	var (
		frames          []*frame.Frame          = framesCollection.GetAll()
		detectionBuffer DiscreteDetectionBuffer = NewDiscreteDetectionBuffer(detector.options, AboveMovingMeanAllWeights)
		statistics      statistics.DescriptiveStatisticsEntry
	)

	progressStep, progressFinalize := detector.printer.ProgressSteps("Video detection stage.", len(frames))
	defer progressFinalize()

	for frameIndex, frame := range frames {
		if err := ds.AtP(frameIndex, &statistics); err != nil {
			return nil, fmt.Errorf("detector: failed to access frame descriptive statistics: %w", err)
		}

		if err := detectionBuffer.Push(frame, statistics); err != nil {
			return nil, fmt.Errorf("detector: failed to push the frame the detection buffer: %w", err)
		}

		progressStep()
	}

	detector.printer.Debug("Video detection stage finished. Stage took: %s", time.Since(videoDetectionTime))

	return detectionBuffer.ResolveIndexes(), nil
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
