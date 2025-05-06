package detector

import (
	"fmt"
	"slices"

	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/statistics"
)

// TODO: Diagnostic classification logging via printer
// TODO: Implement more tests

const (
	ClassificationRingQueueSize int = 4
)

type DetectionStrategy int

const (
	AboveMovingMeanAllWeights DetectionStrategy = iota
	AboveGlobalMeanAllWeights
	AboveZeroAllWeights
)

type DiscreteDetectionBuffer interface {
	Push(f *frame.Frame, s statistics.DescriptiveStatisticsEntry) error
	ResolveIndexes() []int
}

type discreteDetectionBuffer struct {
	Options             options.DetectorOptions
	ClassificationQueue []detectionBufferElement
	DetectionSet        map[int]bool
	Strategy            DetectionStrategy
}

func (buffer *discreteDetectionBuffer) Push(f *frame.Frame, s statistics.DescriptiveStatisticsEntry) error {
	queueLength := len(buffer.ClassificationQueue)
	if queueLength > 0 && buffer.ClassificationQueue[queueLength-1].Index >= f.OrdinalNumber {
		return fmt.Errorf("detector: frame pushing failed due to invalid ordinal number")
	}

	cl := detectionBufferElement{
		Index:                               f.OrdinalNumber,
		ColorDifferenceClassified:           false,
		BinaryThresholdDifferenceClassified: false,
		BrightnessClassified:                false,
	}

	switch buffer.Strategy {
	case AboveMovingMeanAllWeights:
		cl.BrightnessClassified = f.Brightness >= buffer.Options.BrightnessDetectionThreshold+s.BrightnessMovingMeanAtPoint
		cl.ColorDifferenceClassified = f.ColorDifference >= buffer.Options.ColorDifferenceDetectionThreshold+s.ColorDifferenceMovingMeanAtPoint
		cl.BinaryThresholdDifferenceClassified = f.BinaryThresholdDifference >= buffer.Options.BinaryThresholdDifferenceDetectionThreshold+s.BinaryThresholdDifferenceMovingMeanAtPoint
	case AboveGlobalMeanAllWeights:
		cl.BrightnessClassified = f.Brightness >= buffer.Options.BrightnessDetectionThreshold+s.BrightnessMean
		cl.ColorDifferenceClassified = f.ColorDifference >= buffer.Options.ColorDifferenceDetectionThreshold+s.ColorDifferenceMean
		cl.BinaryThresholdDifferenceClassified = f.BinaryThresholdDifference >= buffer.Options.BinaryThresholdDifferenceDetectionThreshold+s.BinaryThresholdDifferenceMean
	case AboveZeroAllWeights:
		cl.BrightnessClassified = f.Brightness >= buffer.Options.BrightnessDetectionThreshold
		cl.ColorDifferenceClassified = f.ColorDifference >= buffer.Options.ColorDifferenceDetectionThreshold
		cl.BinaryThresholdDifferenceClassified = f.BinaryThresholdDifference >= buffer.Options.BinaryThresholdDifferenceDetectionThreshold
	default:
		panic("detector: invalid detection strategy specified")
	}

	if len(buffer.ClassificationQueue) < ClassificationRingQueueSize {
		buffer.ClassificationQueue = append(buffer.ClassificationQueue, cl)
	} else {
		buffer.ClassificationQueue = append(buffer.ClassificationQueue[1:], cl)
	}

	buffer.ResolveClassificationQueue()
	return nil
}

func (buffer *discreteDetectionBuffer) ResolveClassificationQueue() {
	if len(buffer.ClassificationQueue) < ClassificationRingQueueSize {
		for _, e := range buffer.ClassificationQueue {
			if e.BrightnessClassified && e.ColorDifferenceClassified && e.BinaryThresholdDifferenceClassified {
				buffer.DetectionSet[e.Index] = true
			}
		}

		return
	}

	var (
		c0  = buffer.ClassificationQueue[0]
		c1  = buffer.ClassificationQueue[1]
		c2  = buffer.ClassificationQueue[2]
		c3  = buffer.ClassificationQueue[3]
		wc0 = c0.BrightnessClassified && c0.ColorDifferenceClassified && c0.BinaryThresholdDifferenceClassified
		wc1 = c1.BrightnessClassified && c1.ColorDifferenceClassified && c1.BinaryThresholdDifferenceClassified
		wc2 = c2.BrightnessClassified && c2.ColorDifferenceClassified && c2.BinaryThresholdDifferenceClassified
		wc3 = c3.BrightnessClassified && c3.ColorDifferenceClassified && c3.BinaryThresholdDifferenceClassified
	)

	if wc0 {
		buffer.DetectionSet[c0.Index] = true
	}

	if wc1 || (wc0 && wc3) || (wc0 && wc2) {
		buffer.DetectionSet[c1.Index] = true
	}

	if wc2 || (wc1 && wc3) || (wc0 && wc3) {
		buffer.DetectionSet[c2.Index] = true
	}

	if wc3 {
		buffer.DetectionSet[c3.Index] = true
	}
}

func (buffer *discreteDetectionBuffer) ResolveIndexes() []int {
	indexes := make([]int, 0)
	for index, ok := range buffer.DetectionSet {
		if ok {
			// NOTE: The detection set is holding frame 1-indexed ordinal numbers, and the detector expectes to return 0-indexed frames indexes
			indexes = append(indexes, index-1)
		}
	}

	slices.Sort(indexes)
	return indexes
}

func NewDiscreteDetectionBuffer(opt options.DetectorOptions, s DetectionStrategy) DiscreteDetectionBuffer {
	switch s {
	case AboveMovingMeanAllWeights, AboveGlobalMeanAllWeights, AboveZeroAllWeights:
	default:
		panic("detector: invalid detection strategy specified")
	}

	return &discreteDetectionBuffer{
		Options:             opt,
		ClassificationQueue: make([]detectionBufferElement, 0, ClassificationRingQueueSize),
		DetectionSet:        make(map[int]bool, 0),
		Strategy:            s,
	}
}

type detectionBufferElement struct {
	Index                               int
	ColorDifferenceClassified           bool
	BinaryThresholdDifferenceClassified bool
	BrightnessClassified                bool
}
