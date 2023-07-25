package frame

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

// Structure representing the collection of video frames.
type FramesCollection struct {
	Frames map[int]*Frame
	mu     sync.RWMutex
}

// Structure containing frames descriptive statistics values.
type FramesStatistics struct {
	BrightnessMean                             float64
	BrightnessStandardDeviation                float64
	BrightnessMax                              float64
	ColorDifferenceMean                        float64
	ColorDifferenceStandardDeviation           float64
	ColorDifferenceMax                         float64
	BinaryThresholdDifferenceMean              float64
	BinaryThresholdDifferenceStandardDeviation float64
	BinaryThresholdDifferenceMax               float64
}

// Create a new frames collection with a given capacity of frames.
func CreateNewFramesCollection(frames int) *FramesCollection {
	return &FramesCollection{
		Frames: make(map[int]*Frame, frames),
		mu:     sync.RWMutex{},
	}
}

// Add a new frame to the frames collection.
func (frames *FramesCollection) Append(frame *Frame) error {
	if frame == nil {
		return errors.New("frame: can not appenda nil frame to the frames collection")
	}

	frames.mu.Lock()
	defer frames.mu.Unlock()

	if _, exists := frames.Frames[frame.OrdinalNumber]; exists {
		return errors.New("frame: frame with a given ordinal number already exists")
	}

	frames.Frames[frame.OrdinalNumber] = frame
	return nil
}

// Get a frame from the frames collection by the frame ordinal number.
func (frames *FramesCollection) Get(frameNumber int) (*Frame, error) {
	frames.mu.RLock()
	defer frames.mu.RUnlock()

	if frame, exists := frames.Frames[frameNumber]; !exists {
		return nil, errors.New("frame: frame with a given ordinal number does not exist")
	} else {
		return frame, nil
	}
}

// Calculate the descriptive statistics values for the given frames collection.
func (frames *FramesCollection) CalculateStatistics() FramesStatistics {
	frames.mu.RLock()
	defer frames.mu.RUnlock()

	var (
		framesBrightness                []float64 = make([]float64, 0, len(frames.Frames))
		framesColorDifference           []float64 = make([]float64, 0, len(frames.Frames))
		framesBinaryThresholdDifference []float64 = make([]float64, 0, len(frames.Frames))
	)

	for _, frame := range frames.Frames {
		framesBrightness = append(framesBrightness, frame.Brightness)
		framesColorDifference = append(framesColorDifference, frame.ColorDifference)
		framesBinaryThresholdDifference = append(framesBinaryThresholdDifference, frame.BinaryThresholdDifference)
	}

	return FramesStatistics{
		BrightnessMean:                             utils.Mean(framesBrightness),
		BrightnessStandardDeviation:                utils.StandardDeviation(framesBrightness),
		BrightnessMax:                              utils.Max(framesBrightness),
		ColorDifferenceMean:                        utils.Mean(framesColorDifference),
		ColorDifferenceStandardDeviation:           utils.StandardDeviation(framesColorDifference),
		ColorDifferenceMax:                         utils.Max(framesColorDifference),
		BinaryThresholdDifferenceMean:              utils.Mean(framesBinaryThresholdDifference),
		BinaryThresholdDifferenceStandardDeviation: utils.StandardDeviation(framesBinaryThresholdDifference),
		BinaryThresholdDifferenceMax:               utils.Max(framesBinaryThresholdDifference),
	}
}

// Write the JSON format frames report to the provided writer which can be a file reference.
func (frames *FramesCollection) ExportJsonReport(file io.Writer) error {
	frames.mu.RLock()
	defer frames.mu.RUnlock()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(frames.Frames); err != nil {
		return fmt.Errorf("frame: failed to encode the frames collection to json report file: %w", err)
	}

	return nil
}

// Write the CSV format frames report to the provided writer which can be a file reference.
func (frames *FramesCollection) ExportCsvReport(file io.Writer) error {
	frames.mu.RLock()
	defer frames.mu.RUnlock()

	csvWriter := csv.NewWriter(file)
	if err := csvWriter.Write([]string{"Frame", "Brightness", "ColorDifference", "BinaryThresholdDifference"}); err != nil {
		return fmt.Errorf("frame: failed to write the header to the frames report file: %w", err)
	}

	for _, frame := range frames.Frames {
		if err := csvWriter.Write(frame.ToBuffer()); err != nil {
			return fmt.Errorf("frame: failed to write the frame to the frames report file: %w", err)
		}
	}

	csvWriter.Flush()
	return nil
}
