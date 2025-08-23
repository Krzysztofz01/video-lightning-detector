package detector

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/statistics"
)

func TestDiscreteDetectionBufferShouldCreate(t *testing.T) {
	cases := map[DetectionStrategy]bool{
		AboveMovingMeanAllWeights: true,
		AboveGlobalMeanAllWeights: true,
		AboveZeroAllWeights:       true,
		-1:                        false,
	}

	for strategy, validStrategy := range cases {
		if validStrategy {
			buffer := NewDiscreteDetectionBuffer(options.GetDefaultDetectorOptions(), strategy)

			assert.NotNil(t, buffer)
		} else {
			assert.Panics(t, func() {
				NewDiscreteDetectionBuffer(options.GetDefaultDetectorOptions(), strategy)
			})
		}
	}
}

func TestDiscreteDetectionBufferShouldFailOnAppendingInvalidValues(t *testing.T) {
	buffer := NewDiscreteDetectionBuffer(options.GetDefaultDetectorOptions(), AboveMovingMeanAllWeights)
	assert.NotNil(t, buffer)

	stats := statistics.DescriptiveStatisticsEntry{}

	frame1 := &frame.Frame{
		OrdinalNumber:             10,
		ColorDifference:           0,
		BinaryThresholdDifference: 0,
		Brightness:                0,
	}

	frame2 := &frame.Frame{
		OrdinalNumber:             frame1.OrdinalNumber - 1,
		ColorDifference:           0,
		BinaryThresholdDifference: 0,
		Brightness:                0,
	}

	err := buffer.Push(frame1, stats)
	assert.Nil(t, err)

	err = buffer.Push(frame2, stats)
	assert.NotNil(t, err)
}

func TestDiscreteDetectionBufferCorrectlyResolveAppendedValues(t *testing.T) {
	const (
		brightnessThreshold                float64 = 0.5
		colorDifferenceThreshold           float64 = 0.5
		binaryThresholdDifferenceThreshold float64 = 0.5
		delta                              float64 = 0.25
	)

	var (
		options    options.DetectorOptions = options.GetDefaultDetectorOptions()
		statistics statistics.DescriptiveStatisticsEntry
	)

	options.BrightnessDetectionThreshold = brightnessThreshold
	options.ColorDifferenceDetectionThreshold = colorDifferenceThreshold
	options.BinaryThresholdDifferenceDetectionThreshold = binaryThresholdDifferenceThreshold

	times := func(v bool, n int) []bool {
		values := make([]bool, 0, n)
		for i := 0; i < n; i += 1 {
			values = append(values, v)
		}

		return values
	}

	cases := []struct {
		Frames   []bool
		Expected []int
	}{
		{[]bool{false, false, false, false}, []int{}},
		{[]bool{false, true, false, false}, []int{1}},
		{[]bool{true, false, false, false}, []int{0}},
		{[]bool{true, true, false, false}, []int{0, 1}},
		{[]bool{false, false, true, false}, []int{2}},
		{[]bool{false, true, true, false}, []int{1, 2}},
		{[]bool{true, false, true, false}, []int{0, 1, 2}},
		{[]bool{true, true, true, false}, []int{0, 1, 2}},
		{[]bool{false, false, false, true}, []int{3}},
		{[]bool{false, true, false, true}, []int{1, 2, 3}},
		{[]bool{true, false, false, true}, []int{0, 1, 2, 3}},
		{[]bool{true, true, false, true}, []int{0, 1, 2, 3}},
		{[]bool{false, false, true, true}, []int{2, 3}},
		{[]bool{false, true, true, true}, []int{1, 2, 3}},
		{[]bool{true, false, true, true}, []int{0, 1, 2, 3}},
		{[]bool{true, true, true, true}, []int{0, 1, 2, 3}},
		{[]bool{false, false, true, true, false}, []int{2, 3}},
		{[]bool{false, true, true, true, false}, []int{1, 2, 3}},
		{[]bool{true, false, true, true, false}, []int{0, 1, 2, 3}},
		{[]bool{true, true, true, true, false}, []int{0, 1, 2, 3}},
		{append(times(false, 30), false, false, true, true, false), []int{30 + 2, 30 + 3}},
		{append(times(false, 30), false, true, true, true, false), []int{30 + 1, 30 + 2, 30 + 3}},
		{append(times(false, 30), true, false, true, true, false), []int{30 + 0, 30 + 1, 30 + 2, 30 + 3}},
		{append(times(false, 30), true, true, true, true, false), []int{30 + 0, 30 + 1, 30 + 2, 30 + 3}},
	}

	for _, c := range cases {
		frames := make([]*frame.Frame, 0, len(c.Frames))
		for index, detection := range c.Frames {
			var sign float64
			if detection {
				sign = 1
			} else {
				sign = -1
			}

			frames = append(frames, &frame.Frame{
				OrdinalNumber:             index + 1,
				ColorDifference:           colorDifferenceThreshold + (delta * sign),
				BinaryThresholdDifference: binaryThresholdDifferenceThreshold + (delta * sign),
				Brightness:                brightnessThreshold + (delta * sign),
			})
		}

		detectionBuffer := NewDiscreteDetectionBuffer(options, AboveZeroAllWeights)
		for _, frame := range frames {
			err := detectionBuffer.Push(frame, statistics)
			assert.Nil(t, err)
		}

		actual := detectionBuffer.ResolveIndexes()
		assert.NotNil(t, actual)
		assert.Equal(t, c.Expected, actual)
	}
}

func TestContinuousDetectionBufferShouldCreate(t *testing.T) {
	cases := map[DetectionStrategy]bool{
		AboveMovingMeanAllWeights: true,
		AboveGlobalMeanAllWeights: true,
		AboveZeroAllWeights:       true,
		-1:                        false,
	}

	for strategy, validStrategy := range cases {
		if validStrategy {
			buffer := NewContinuousDetectionBuffer(options.GetDefaultStreamDetectorOptions(), strategy)

			assert.NotNil(t, buffer)
		} else {
			assert.Panics(t, func() {
				NewContinuousDetectionBuffer(options.GetDefaultStreamDetectorOptions(), strategy)
			})
		}
	}
}

func TestContinuousDetectionBufferShouldFailOnAppendingInvalidValues(t *testing.T) {
	buffer := NewContinuousDetectionBuffer(options.GetDefaultStreamDetectorOptions(), AboveMovingMeanAllWeights)
	assert.NotNil(t, buffer)

	stats := statistics.DescriptiveStatisticsEntry{}

	frame1 := &frame.Frame{
		OrdinalNumber:             10,
		ColorDifference:           0,
		BinaryThresholdDifference: 0,
		Brightness:                0,
	}

	frame2 := &frame.Frame{
		OrdinalNumber:             frame1.OrdinalNumber - 1,
		ColorDifference:           0,
		BinaryThresholdDifference: 0,
		Brightness:                0,
	}

	result, err := buffer.PushAndResolveIndexes(frame1, stats)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	result, err = buffer.PushAndResolveIndexes(frame2, stats)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestContinuousDetectionBufferCorrectlyResolveAppendedValues(t *testing.T) {
	const (
		brightnessThreshold                float64 = 0.5
		colorDifferenceThreshold           float64 = 0.5
		binaryThresholdDifferenceThreshold float64 = 0.5
		delta                              float64 = 0.25
	)

	var (
		options    options.StreamDetectorOptions = options.GetDefaultStreamDetectorOptions()
		statistics statistics.DescriptiveStatisticsEntry
	)

	options.BrightnessDetectionThreshold = brightnessThreshold
	options.ColorDifferenceDetectionThreshold = colorDifferenceThreshold
	options.BinaryThresholdDifferenceDetectionThreshold = binaryThresholdDifferenceThreshold

	times := func(v bool, n int) []bool {
		values := make([]bool, 0, n)
		for i := 0; i < n; i += 1 {
			values = append(values, v)
		}

		return values
	}

	cases := []struct {
		Frames   []bool
		Expected []int
	}{
		{[]bool{false, false, false, false}, []int{}},
		{[]bool{false, true, false, false}, []int{1}},
		{[]bool{true, false, false, false}, []int{0}},
		{[]bool{true, true, false, false}, []int{0, 1}},
		{[]bool{false, false, true, false}, []int{2}},
		{[]bool{false, true, true, false}, []int{1, 2}},
		{[]bool{true, false, true, false}, []int{0, 1, 2}},
		{[]bool{true, true, true, false}, []int{0, 1, 2}},
		{[]bool{false, false, false, true}, []int{3}},
		{[]bool{false, true, false, true}, []int{1, 2, 3}},
		{[]bool{true, false, false, true}, []int{0, 1, 2, 3}},
		{[]bool{true, true, false, true}, []int{0, 1, 2, 3}},
		{[]bool{false, false, true, true}, []int{2, 3}},
		{[]bool{false, true, true, true}, []int{1, 2, 3}},
		{[]bool{true, false, true, true}, []int{0, 1, 2, 3}},
		{[]bool{true, true, true, true}, []int{0, 1, 2, 3}},
		{[]bool{false, false, true, true, false}, []int{2, 3}},
		{[]bool{false, true, true, true, false}, []int{1, 2, 3}},
		{[]bool{true, false, true, true, false}, []int{0, 1, 2, 3}},
		{[]bool{true, true, true, true, false}, []int{0, 1, 2, 3}},
		{append(times(false, 30), false, false, true, true, false), []int{30 + 2, 30 + 3}},
		{append(times(false, 30), false, true, true, true, false), []int{30 + 1, 30 + 2, 30 + 3}},
		{append(times(false, 30), true, false, true, true, false), []int{30 + 0, 30 + 1, 30 + 2, 30 + 3}},
		{append(times(false, 30), true, true, true, true, false), []int{30 + 0, 30 + 1, 30 + 2, 30 + 3}},
	}

	for _, c := range cases {
		frames := make([]*frame.Frame, 0, len(c.Frames))
		for index, detection := range c.Frames {
			var sign float64
			if detection {
				sign = 1
			} else {
				sign = -1
			}

			frames = append(frames, &frame.Frame{
				OrdinalNumber:             index + 1,
				ColorDifference:           colorDifferenceThreshold + (delta * sign),
				BinaryThresholdDifference: binaryThresholdDifferenceThreshold + (delta * sign),
				Brightness:                brightnessThreshold + (delta * sign),
			})
		}

		detectionBuffer := NewContinuousDetectionBuffer(options, AboveZeroAllWeights)

		allClassificationIndexes := make([]int, 0)
		for _, frame := range frames {
			classificationIndexes, err := detectionBuffer.PushAndResolveIndexes(frame, statistics)
			assert.NotNil(t, classificationIndexes)
			assert.Nil(t, err)

			allClassificationIndexes = append(allClassificationIndexes, classificationIndexes...)
		}

		slices.Sort(allClassificationIndexes)
		actual := slices.Compact(allClassificationIndexes)
		assert.NotNil(t, actual)
		assert.Equal(t, c.Expected, actual)
	}
}
