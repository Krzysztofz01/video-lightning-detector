package detector

import (
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"time"

	"github.com/Krzysztofz01/video-lightning-detector/internal/analyzer"
	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/printer"
	"github.com/Krzysztofz01/video-lightning-detector/internal/statistics"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

// Detector instance that is able to perform a search after ligntning strikes on a continuous video stream.
type StreamDetector interface {
	Run(inputVideoStreamUrl string, ctx context.Context) error
}

type streamDetector struct {
	Options options.StreamDetectorOptions
	Printer printer.Printer
}

// Create a new stream video lightning detector instance with the specified options.
func CreateStreamDetector(printer printer.Printer, options options.StreamDetectorOptions) (StreamDetector, error) {
	if printer == nil {
		return nil, errors.New("detector: invalid nil reference renderer provided")
	}

	if ok, msg := options.AreValid(); !ok {
		return nil, fmt.Errorf("detector: invalid options %s", msg)
	}

	printer.Debug("Continuous detector create with options %+v", options)

	return &streamDetector{
		Options: options,
		Printer: printer,
	}, nil
}

func (detector *streamDetector) Run(inputVideoStreamUrl string, ctx context.Context) error {
	runTime := time.Now()
	detector.Printer.InfoA("starting the lightning hunt.")

	var (
		movingMeanResolution int                                         = int(detector.Options.MovingMeanResolution)
		analyzer             analyzer.StreamAnalyzer                     = analyzer.NewStreamAnalyzer(inputVideoStreamUrl, detector.Options, detector.Printer)
		stats                statistics.IncrementalDescriptiveStatistics = statistics.NewIncrementalDescriptiveStatistics(movingMeanResolution)
		detectionBuffer      ContinuousDetectionBuffer                   = NewContinuousDetectionBuffer(detector.Options, AboveMovingMeanAllWeights)
		detectionIndexes     utils.DecayingHashSet[int]                  = utils.NewDecayingHashSet[int](4)
	)

	var (
		currentFrame            *frame.Frame
		currentFrameTimestamp   time.Time
		detectionFrame          *frame.Frame
		detectionFrameTimestamp time.Time
		detectionFrameImage     *image.RGBA
		windowStatistics        statistics.DescriptiveStatisticsEntry
		err                     error
		frameStrikeDetector     FrameStrikeDetector
	)

readStream:
	for {
		select {
		case <-ctx.Done():
			detector.Printer.InfoA("Stopping the lightning hunt.")
			break readStream
		default:
		}

		if err = analyzer.Next(); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return fmt.Errorf("detector: video analysis frame access failed: %w", err)
		}

		// NOTE: Lazy initialization of frame strike detector due to the requirement of analyzer initialization running first
		if frameStrikeDetector == nil {
			fullFrameWidth, fullFrameHeight, err := analyzer.PeekFrameImageDimensions(true)
			if err != nil {
				return fmt.Errorf("detector: failed to access the full frame dimensions via analyzer: %w", err)
			}

			if frameStrikeDetector, err = CreateFrameStrikeDetector(fullFrameWidth, fullFrameHeight, detector.Options); err != nil {
				return fmt.Errorf("detector: failed to create the frame strike detector: %w", err)
			}
		}

		if currentFrame, currentFrameTimestamp, err = analyzer.PeekFrame(0); err != nil {
			return fmt.Errorf("detector: failed to peek the current frame: %w", err)
		}

		stats.Push(currentFrame)
		windowStatistics = stats.Peek()

		if detector.Printer.IsLogLevel(options.Verbose) {
			detector.Printer.Debug("Frame: [%d - %s]. Brightness: %1.6f (%1.4f) ColorDiff: %1.6f (%1.4f) BTDiff: %1.6f (%1.4f)",
				currentFrame.OrdinalNumber,
				currentFrameTimestamp.Format("15:04:05.000"),
				currentFrame.Brightness,
				windowStatistics.BrightnessMovingMeanAtPoint,
				currentFrame.ColorDifference,
				windowStatistics.ColorDifferenceMovingMeanAtPoint,
				currentFrame.BinaryThresholdDifference,
				windowStatistics.BinaryThresholdDifferenceMovingMeanAtPoint)
		}

		detections, err := detectionBuffer.PushAndResolveIndexes(currentFrame, windowStatistics)
		if err != nil {
			return fmt.Errorf("detector: failed to push and resolve detections via the detection buffer: %w", err)
		}

		for _, frameIndex := range detections {
			if detectionIndexes.Contains(frameIndex) {
				continue
			}

			detectionPeekIndex := currentFrame.OrdinalNumber - 1 - frameIndex

			if detectionFrame, detectionFrameTimestamp, err = analyzer.PeekFrame(detectionPeekIndex); err != nil {
				return fmt.Errorf("detector: failed to access the detection frame: %w", err)
			}

			if detectionFrameImage, err = analyzer.PeekFrameImage(detectionPeekIndex); err != nil {
				return fmt.Errorf("detector: failed to access the detection frame image: %w", err)
			}

			detectionPlot, err := frameStrikeDetector.GetDetectionPlot(detectionFrameImage)
			if err != nil {
				return fmt.Errorf("detector: failed to process the frame strike detection plot: %w", err)
			}

			detector.Printer.Debug("Frame with ordinal number %d has been classified as a detection", detectionFrame.OrdinalNumber)

			detector.Printer.WriteParsable(struct {
				Timestamp     time.Time    `json:"timestamp"`
				DetectionPlot [2][]float64 `json:"plot"`
			}{
				Timestamp:     detectionFrameTimestamp,
				DetectionPlot: detectionPlot,
			})

			detectionIndexes.Add(frameIndex)
		}
	}

	if detector.Options.DiagnosticMode {
		detector.Printer.WriteParsable(struct {
			Statistics statistics.DescriptiveStatisticsEntry `json:"statistics"`
			FrameCount int                                   `json:"frame-count"`
		}{
			Statistics: stats.Peek(),
			FrameCount: analyzer.FrameCount(),
		})
	}

	detector.Printer.InfoA("Lightning hunt was running for: %s", time.Since(runTime))
	return nil
}
