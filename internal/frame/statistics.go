package frame

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

const (
	MovingMeanBias = 25
)

// Structure containing frames descriptive statistics values.
type FramesStatistics struct {
	BrightnessMean                             float64
	BrightnessMovingMean                       []float64
	BrightnessStandardDeviation                float64
	BrightnessMax                              float64
	ColorDifferenceMean                        float64
	ColorDifferenceMovingMean                  []float64
	ColorDifferenceStandardDeviation           float64
	ColorDifferenceMax                         float64
	BinaryThresholdDifferenceMean              float64
	BinaryThresholdDifferenceMovingMean        []float64
	BinaryThresholdDifferenceStandardDeviation float64
	BinaryThresholdDifferenceMax               float64
}

func CreateNewFramesStatistics(frames []*Frame) *FramesStatistics {
	var (
		brightness                    []float64 = make([]float64, 0, len(frames))
		colorDiff                     []float64 = make([]float64, 0, len(frames))
		binaryThresholdDiff           []float64 = make([]float64, 0, len(frames))
		brightnessMovingMean          []float64 = make([]float64, 0, len(frames))
		colorDiffMovingMean           []float64 = make([]float64, 0, len(frames))
		binaryThresholdDiffMovingMean []float64 = make([]float64, 0, len(frames))
	)

	for _, frame := range frames {
		brightness = append(brightness, frame.Brightness)
		colorDiff = append(colorDiff, frame.ColorDifference)
		binaryThresholdDiff = append(binaryThresholdDiff, frame.BinaryThresholdDifference)
	}

	for index := range frames {
		brightnessMovingMean = append(brightnessMovingMean, utils.MovingMean(brightness, index, MovingMeanBias))
		colorDiffMovingMean = append(colorDiffMovingMean, utils.MovingMean(colorDiff, index, MovingMeanBias))
		binaryThresholdDiffMovingMean = append(binaryThresholdDiffMovingMean, utils.MovingMean(binaryThresholdDiff, index, MovingMeanBias))
	}

	return &FramesStatistics{
		BrightnessMean:                             utils.Mean(brightness),
		BrightnessMovingMean:                       brightnessMovingMean,
		BrightnessStandardDeviation:                utils.StandardDeviation(brightness),
		BrightnessMax:                              utils.Max(brightness),
		ColorDifferenceMean:                        utils.Mean(colorDiff),
		ColorDifferenceMovingMean:                  colorDiffMovingMean,
		ColorDifferenceStandardDeviation:           utils.StandardDeviation(colorDiff),
		ColorDifferenceMax:                         utils.Max(colorDiff),
		BinaryThresholdDifferenceMean:              utils.Mean(binaryThresholdDiff),
		BinaryThresholdDifferenceMovingMean:        binaryThresholdDiffMovingMean,
		BinaryThresholdDifferenceStandardDeviation: utils.StandardDeviation(binaryThresholdDiff),
		BinaryThresholdDifferenceMax:               utils.Max(binaryThresholdDiff),
	}
}

// Write the CSV format statistics report to the provided writer which can be a file reference.
func (statistics *FramesStatistics) ExportCsvReport(file io.Writer) error {
	csvWriter := csv.NewWriter(file)
	rows := [][]string{
		{"", "Brightness mean", "Brightness standard deviation", "Brightness max"},
		statistics.valuesToBuffer(1, statistics.BrightnessMean, statistics.BrightnessStandardDeviation, statistics.BrightnessMax),
		{},
		{"", "Color difference mean", "Color difference standard deviation", "Color difference max"},
		statistics.valuesToBuffer(1, statistics.ColorDifferenceMean, statistics.ColorDifferenceStandardDeviation, statistics.ColorDifferenceMax),
		{},
		{"", "Binary threshold difference mean", "Binary threshold difference standard deviation", "Binary threshold difference max"},
		statistics.valuesToBuffer(1, statistics.BinaryThresholdDifferenceMean, statistics.BinaryThresholdDifferenceStandardDeviation, statistics.BinaryThresholdDifferenceMax),
		{},
	}

	for _, row := range rows {
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("frame: failed to write descriptive statistics to the report file: %w", err)
		}
	}

	if err := csvWriter.Write([]string{"Frame (Moving mean center point)", "Brightness moving mean", "ColorDifference moving mean", "BinaryThresholdDifference moving mean"}); err != nil {
		return fmt.Errorf("frame: failed to write the moving mean header to the statistics report file: %w", err)
	}

	for index := 0; index < len(statistics.BrightnessMovingMean); index += 1 {
		values := statistics.valuesToBuffer(0, statistics.BrightnessMovingMean[index], statistics.ColorDifferenceMovingMean[index], statistics.BinaryThresholdDifferenceMovingMean[index])
		if err := csvWriter.Write(append([]string{strconv.Itoa(index + 1)}, values...)); err != nil {
			return fmt.Errorf("frame: failed to write moving mean row to the statistics report file: %w", err)
		}
	}

	csvWriter.Flush()
	return nil
}

func (statistics *FramesStatistics) valuesToBuffer(padding int, values ...float64) []string {
	buffer := make([]string, 0, len(values)+padding)
	for index := 0; index < padding; index += 1 {
		buffer = append(buffer, "")
	}

	for _, value := range values {
		buffer = append(buffer, strconv.FormatFloat(value, 'f', -1, 64))
	}

	return buffer
}
