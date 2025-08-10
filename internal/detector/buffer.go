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
	Classifier                  BufferClassifier
	ClassificationQueue         []detectionBufferElement
	DetectionSet                map[int]bool
	ClassificationIndexPreAlloc []int
}

func (buffer *discreteDetectionBuffer) Push(f *frame.Frame, s statistics.DescriptiveStatisticsEntry) error {
	queueLength := len(buffer.ClassificationQueue)
	if queueLength > 0 && buffer.ClassificationQueue[queueLength-1].Index >= f.OrdinalNumber {
		return fmt.Errorf("detector: frame pushing failed due to invalid ordinal number")
	}

	element := buffer.Classifier.CreateElement(f, s)

	if len(buffer.ClassificationQueue) < ClassificationRingQueueSize {
		buffer.ClassificationQueue = append(buffer.ClassificationQueue, element)
	} else {
		buffer.ClassificationQueue = append(buffer.ClassificationQueue[1:], element)
	}

	buffer.ClassificationIndexPreAlloc = buffer.Classifier.ResolveElement(buffer.ClassificationQueue, buffer.ClassificationIndexPreAlloc)
	for _, frameIndex := range buffer.ClassificationIndexPreAlloc {
		buffer.DetectionSet[frameIndex] = true
	}

	return nil
}

func (buffer *discreteDetectionBuffer) ResolveIndexes() []int {
	indexes := make([]int, 0)
	for index, ok := range buffer.DetectionSet {
		if ok {
			// FIXME: Workaround due to the fact that the classification indexing is based around the frame ordinal number which is 1-indexed, but
			// 		  the consumer of the resolved detections expects the result indexing to be 0-based.
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
		Classifier:                  createBufferClassifier(opt, s),
		ClassificationQueue:         make([]detectionBufferElement, 0, ClassificationRingQueueSize),
		DetectionSet:                make(map[int]bool, 0),
		ClassificationIndexPreAlloc: make([]int, 0, ClassificationRingQueueSize),
	}
}

type ContinuousDetectionBuffer interface {
	PushAndResolveIndexes(f *frame.Frame, s statistics.DescriptiveStatisticsEntry) ([]int, error)
}

type continuousDetectionBuffer struct {
	Classifier                  BufferClassifier
	ClassificationQueue         []detectionBufferElement
	ClassificationIndexPreAlloc []int
}

func (buffer *continuousDetectionBuffer) PushAndResolveIndexes(f *frame.Frame, s statistics.DescriptiveStatisticsEntry) ([]int, error) {
	queueLength := len(buffer.ClassificationQueue)
	if queueLength > 0 && buffer.ClassificationQueue[queueLength-1].Index >= f.OrdinalNumber {
		return nil, fmt.Errorf("detector: frame pushing failed due to invalid ordinal number")
	}

	element := buffer.Classifier.CreateElement(f, s)

	if len(buffer.ClassificationQueue) < ClassificationRingQueueSize {
		buffer.ClassificationQueue = append(buffer.ClassificationQueue, element)
	} else {
		buffer.ClassificationQueue = append(buffer.ClassificationQueue[1:], element)
	}

	buffer.ClassificationIndexPreAlloc = buffer.Classifier.ResolveElement(buffer.ClassificationQueue, buffer.ClassificationIndexPreAlloc)

	// FIXME: Workaround due to the fact that the classification indexing is based around the frame ordinal number which is 1-indexed, but
	// 		  the consumer of the resolved detections expects the result indexing to be 0-based.
	for index := range buffer.ClassificationIndexPreAlloc {
		buffer.ClassificationIndexPreAlloc[index] -= 1
	}

	return buffer.ClassificationIndexPreAlloc, nil
}

func NewContinuousDetectionBuffer(opt options.StreamDetectorOptions, s DetectionStrategy) ContinuousDetectionBuffer {
	return &continuousDetectionBuffer{
		Classifier:                  createBufferClassifier(opt, s),
		ClassificationQueue:         make([]detectionBufferElement, 0, ClassificationRingQueueSize),
		ClassificationIndexPreAlloc: make([]int, 0, ClassificationRingQueueSize),
	}
}

type BufferClassifier interface {
	CreateElement(f *frame.Frame, s statistics.DescriptiveStatisticsEntry) detectionBufferElement
	ResolveElement(queue []detectionBufferElement, resultAlloc []int) []int
}

type bufferClassifier struct {
	Strategy                                    DetectionStrategy
	BrightnessDetectionThreshold                float64
	ColorDifferenceDetectionThreshold           float64
	BinaryThresholdDifferenceDetectionThreshold float64
}

func (classifier *bufferClassifier) CreateElement(f *frame.Frame, s statistics.DescriptiveStatisticsEntry) detectionBufferElement {
	cl := detectionBufferElement{
		Index:                               f.OrdinalNumber,
		ColorDifferenceClassified:           false,
		BinaryThresholdDifferenceClassified: false,
		BrightnessClassified:                false,
	}

	switch classifier.Strategy {
	case AboveMovingMeanAllWeights:
		cl.BrightnessClassified = f.Brightness >= classifier.BrightnessDetectionThreshold+s.BrightnessMovingMeanAtPoint
		cl.ColorDifferenceClassified = f.ColorDifference >= classifier.ColorDifferenceDetectionThreshold+s.ColorDifferenceMovingMeanAtPoint
		cl.BinaryThresholdDifferenceClassified = f.BinaryThresholdDifference >= classifier.BinaryThresholdDifferenceDetectionThreshold+s.BinaryThresholdDifferenceMovingMeanAtPoint
	case AboveGlobalMeanAllWeights:
		cl.BrightnessClassified = f.Brightness >= classifier.BrightnessDetectionThreshold+s.BrightnessMean
		cl.ColorDifferenceClassified = f.ColorDifference >= classifier.ColorDifferenceDetectionThreshold+s.ColorDifferenceMean
		cl.BinaryThresholdDifferenceClassified = f.BinaryThresholdDifference >= classifier.BinaryThresholdDifferenceDetectionThreshold+s.BinaryThresholdDifferenceMean
	case AboveZeroAllWeights:
		cl.BrightnessClassified = f.Brightness >= classifier.BrightnessDetectionThreshold
		cl.ColorDifferenceClassified = f.ColorDifference >= classifier.ColorDifferenceDetectionThreshold
		cl.BinaryThresholdDifferenceClassified = f.BinaryThresholdDifference >= classifier.BinaryThresholdDifferenceDetectionThreshold
	default:
		panic("detector: invalid detection strategy specified")
	}

	return cl
}

func (classifier *bufferClassifier) ResolveElement(queue []detectionBufferElement, classificationResult []int) []int {
	classificationResult = classificationResult[:0]

	if len(queue) < ClassificationRingQueueSize {
		for _, e := range queue {
			if e.BrightnessClassified && e.ColorDifferenceClassified && e.BinaryThresholdDifferenceClassified {
				classificationResult = append(classificationResult, e.Index)
			}
		}

		return classificationResult
	}

	var (
		c0  = queue[0]
		c1  = queue[1]
		c2  = queue[2]
		c3  = queue[3]
		wc0 = c0.BrightnessClassified && c0.ColorDifferenceClassified && c0.BinaryThresholdDifferenceClassified
		wc1 = c1.BrightnessClassified && c1.ColorDifferenceClassified && c1.BinaryThresholdDifferenceClassified
		wc2 = c2.BrightnessClassified && c2.ColorDifferenceClassified && c2.BinaryThresholdDifferenceClassified
		wc3 = c3.BrightnessClassified && c3.ColorDifferenceClassified && c3.BinaryThresholdDifferenceClassified
	)

	if wc0 {
		classificationResult = append(classificationResult, c0.Index)
	}

	if wc1 || (wc0 && wc3) || (wc0 && wc2) {
		classificationResult = append(classificationResult, c1.Index)
	}

	if wc2 || (wc1 && wc3) || (wc0 && wc3) {
		classificationResult = append(classificationResult, c2.Index)
	}

	if wc3 {
		classificationResult = append(classificationResult, c3.Index)
	}

	return classificationResult
}

type detectorOptionsConstraint interface {
	options.DetectorOptions | options.StreamDetectorOptions
}

func createBufferClassifier[TOptions detectorOptionsConstraint](opt TOptions, s DetectionStrategy) BufferClassifier {
	switch s {
	case AboveMovingMeanAllWeights, AboveGlobalMeanAllWeights, AboveZeroAllWeights:
	default:
		panic("detector: invalid detection strategy specified")
	}

	switch o := any(opt).(type) {
	case options.DetectorOptions:
		{
			return &bufferClassifier{
				Strategy:                                    s,
				BrightnessDetectionThreshold:                o.BrightnessDetectionThreshold,
				ColorDifferenceDetectionThreshold:           o.ColorDifferenceDetectionThreshold,
				BinaryThresholdDifferenceDetectionThreshold: o.BinaryThresholdDifferenceDetectionThreshold,
			}
		}
	case options.StreamDetectorOptions:
		{
			return &bufferClassifier{
				Strategy:                                    s,
				BrightnessDetectionThreshold:                o.BrightnessDetectionThreshold,
				ColorDifferenceDetectionThreshold:           o.ColorDifferenceDetectionThreshold,
				BinaryThresholdDifferenceDetectionThreshold: o.BinaryThresholdDifferenceDetectionThreshold,
			}
		}
	default:
		panic("detector: invalid detector options specified")
	}
}

type detectionBufferElement struct {
	Index                               int
	ColorDifferenceClassified           bool
	BinaryThresholdDifferenceClassified bool
	BrightnessClassified                bool
}
