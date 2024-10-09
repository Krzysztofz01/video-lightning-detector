package frame

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
)

// Structure representing the collection of video frames.
type FramesCollection struct {
	Frames map[int]*Frame
	mu     sync.RWMutex
}

// Create a new frames collection with a given capacity of frames.
func CreateNewFramesCollection(frames int) *FramesCollection {
	return &FramesCollection{
		Frames: make(map[int]*Frame, frames),
		mu:     sync.RWMutex{},
	}
}

// Create a new frames collection from a json with pre-analized frames data.
func ImportFramesCollectionFromJson(file io.Reader) (*FramesCollection, error) {
	var decodedFrames []*Frame

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&decodedFrames); err != nil {
		return nil, fmt.Errorf("frame: failed to decode the frames collection json: %w", err)
	}

	frames := make(map[int]*Frame, len(decodedFrames))
	for index, frame := range decodedFrames {
		frames[index+1] = frame
	}

	return &FramesCollection{
		Frames: frames,
		mu:     sync.RWMutex{},
	}, nil
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

// Get the count of frames in the frame collection.
func (frames *FramesCollection) Count() int {
	return len(frames.Frames)
}
