package detector

import (
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
	Run(inputVideoStreamUrl string) error
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

func (detector *streamDetector) Run(inputVideoStreamUrl string) error {
	runTime := time.Now()
	detector.Printer.InfoA("starting the lightning hunt.")

	var (
		movingMeanResolution int                                         = int(detector.Options.MovingMeanResolution)
		plotResolution       int                                         = detector.Options.FrameDetectionPlotResolution
		plotThreshold        float64                                     = detector.Options.FrameDetectionPlotThreshold
		analyzer             analyzer.StreamAnalyzer                     = analyzer.NewStreamAnalyzer(inputVideoStreamUrl, detector.Options, detector.Printer)
		stats                statistics.IncrementalDescriptiveStatistics = statistics.NewIncrementalDescriptiveStatistics(movingMeanResolution)
		detectionBuffer      ContinuousDetectionBuffer                   = NewContinuousDetectionBuffer(detector.Options, AboveMovingMeanAllWeights)
		detectionIndexes     utils.DecayingHashSet[int]                  = utils.NewDecayingHashSet[int](4)
	)

	var (
		currentFrame            *frame.Frame
		detectionFrameTimestamp time.Time
		detectionFrameImage     *image.RGBA
		windowStatistics        statistics.DescriptiveStatisticsEntry
		err                     error
	)

	for {
		if err = analyzer.Next(); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return fmt.Errorf("detector: video analysis frame access failed: %w", err)
		}

		if currentFrame, _, err = analyzer.PeekFrame(0); err != nil {
			return fmt.Errorf("detector: failed to peek the current frame: %w", err)
		}

		stats.Push(currentFrame)
		windowStatistics = stats.Peek()

		detections, err := detectionBuffer.PushAndResolveIndexes(currentFrame, windowStatistics)
		if err != nil {
			return fmt.Errorf("detector: failed to push and resolve detections via the detection buffer: %w", err)
		}

		for _, frameIndex := range detections {
			if detectionIndexes.Contains(frameIndex) {
				continue
			}

			detectionPeekIndex := currentFrame.OrdinalNumber - 1 - frameIndex

			if _, detectionFrameTimestamp, err = analyzer.PeekFrame(detectionPeekIndex); err != nil {
				return fmt.Errorf("detector: failed to access the detection frame: %w", err)
			}

			if detectionFrameImage, err = analyzer.PeekFrameImage(detectionPeekIndex); err != nil {
				return fmt.Errorf("detector: failed to access the detection frame image: %w", err)
			}

			detectionPlot, err := getFrameStrikePlot(detectionFrameImage, plotResolution, plotThreshold)
			if err != nil {
				return fmt.Errorf("detector: failed to process the frame strike detection plot: %w", err)
			}

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

	detector.Printer.InfoA("Lightning hunt was running for: %s", time.Since(runTime))
	return nil
}
