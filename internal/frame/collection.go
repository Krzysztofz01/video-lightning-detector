package frame

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
)

// Structure representing the collection of video frames.
type FramesCollection struct {
	Frames                     map[int]*Frame
	cachedStatisticsValue      *FramesStatistics
	cachedStatisticsResolution int
	mu                         sync.RWMutex
}

// Create a new frames collection with a given capacity of frames.
func CreateNewFramesCollection(frames int) *FramesCollection {
	return &FramesCollection{
		Frames:                     make(map[int]*Frame, frames),
		cachedStatisticsValue:      nil,
		cachedStatisticsResolution: 0,
		mu:                         sync.RWMutex{},
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
	frames.cachedStatisticsValue = nil
	frames.cachedStatisticsResolution = 0
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

// Get all frames sorted by the frame ordinal number.
// TODO: Add tests
func (frames *FramesCollection) GetAll() []*Frame {
	frames.mu.RLock()
	defer frames.mu.RUnlock()

	return frames.mapFramesToSlice()
}

// Get all frames sorted by the frame ordinal nubmer. This function does not lock and should only be used intrnaly by the FramesCollection
func (frames *FramesCollection) mapFramesToSlice() []*Frame {
	values := make([]*Frame, len(frames.Frames))
	for index := 0; index < len(frames.Frames); index += 1 {
		frameNumber := index + 1
		frame, ok := frames.Frames[frameNumber]
		if !ok {
			panic("frame: missing frame spotted during frames iteration")
		}

		values[index] = frame
	}

	return values
}

// Calculate the descriptive statistics values for the given frames collection.
func (frames *FramesCollection) CalculateStatistics(movingMeanResolution int) FramesStatistics {
	frames.mu.RLock()
	defer frames.mu.RUnlock()

	if frames.cachedStatisticsValue == nil || frames.cachedStatisticsResolution != movingMeanResolution {
		frames.cachedStatisticsValue = CreateNewFramesStatistics(frames.mapFramesToSlice(), movingMeanResolution)
		frames.cachedStatisticsResolution = movingMeanResolution
	}

	return *frames.cachedStatisticsValue
}

// Write the JSON format frames report to the provided writer which can be a file reference.
func (frames *FramesCollection) ExportJsonReport(file io.Writer) error {
	framesSlice := frames.GetAll()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(framesSlice); err != nil {
		return fmt.Errorf("frame: failed to encode the frames collection to json report file: %w", err)
	}

	return nil
}

// Write the CSV format frames report to the provided writer which can be a file reference.
func (frames *FramesCollection) ExportCsvReport(file io.Writer) error {
	framesSlice := frames.GetAll()

	csvWriter := csv.NewWriter(file)
	if err := csvWriter.Write([]string{"Frame", "Brightness", "ColorDifference", "BinaryThresholdDifference"}); err != nil {
		return fmt.Errorf("frame: failed to write the header to the frames report file: %w", err)
	}

	for _, frame := range framesSlice {
		if err := csvWriter.Write(frame.ToBuffer()); err != nil {
			return fmt.Errorf("frame: failed to write the frame to the frames report file: %w", err)
		}
	}

	csvWriter.Flush()
	return nil
}

// Get the count of frames in the frame collection.
func (frames *FramesCollection) Count() int {
	return len(frames.Frames)
}
