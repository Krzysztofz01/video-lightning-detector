package frame

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// Structure representing the collection of video frames.
type FrameCollection interface {
	Append(frame *Frame) error
	Get(frameOrdinalNumber int) (*Frame, error)
	GetAll() []*Frame
	Count() int
	ExportCache(file io.Writer, checksum string) error
}

type frameCollection struct {
	Frames  []*Frame
	Checked bool
	mu      sync.RWMutex
}

type frameCollectionCache struct {
	Checksum string   `json:"checksum"`
	Frames   []*Frame `json:"frames"`
}

// Create a new frames collection with a given capacity of frames.
func CreateNewFrameCollection(frames int) FrameCollection {
	return &frameCollection{
		Frames:  make([]*Frame, frames),
		Checked: false,
		mu:      sync.RWMutex{},
	}
}

func ImportCachedFrameCollection(file io.Reader) (fc FrameCollection, checksum string, err error) {
	defer func() {
		if err := recover(); err != nil {
			fc = nil
			checksum = ""
			err = fmt.Errorf("frame: failed to create the frame collection from the provided frames: %s", err)
		}
	}()

	decoder := json.NewDecoder(file)

	var fcc frameCollectionCache
	if err = decoder.Decode(&fcc); err != nil {
		return nil, "", fmt.Errorf("frame: failed to decode the frame collection cached: %w", err)
	}

	fc = &frameCollection{
		Frames:  fcc.Frames,
		Checked: false,
		mu:      sync.RWMutex{},
	}

	// NOTE: Call GetAll() to perform the internal order validation and mark frame collection as checked
	fc.GetAll()

	return fc, fcc.Checksum, err
}

func (fc *frameCollection) Append(frame *Frame) error {
	if frame == nil {
		return fmt.Errorf("frame: provided frame to append is nil")
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	frameIndex := frame.OrdinalNumber - 1

	if frameIndex < 0 || frameIndex >= len(fc.Frames) {
		return fmt.Errorf("frame: frame ordinal number does not fit the frame collection capacity")
	}

	if storedFrame := fc.Frames[frameIndex]; storedFrame != nil {
		return fmt.Errorf("frame: frame with given ordinal number is already in the collection")
	}

	fc.Frames[frameIndex] = frame
	fc.Checked = false
	return nil
}

func (fc *frameCollection) Get(frameOrdinalNumber int) (*Frame, error) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	frameIndex := frameOrdinalNumber - 1

	if frameIndex < 0 || frameIndex >= len(fc.Frames) {
		return nil, fmt.Errorf("frame: frame ordinal number is out of the frame collection capacity range")
	}

	if frame := fc.Frames[frameIndex]; frame == nil {
		return nil, fmt.Errorf("frame: frame with given ordinal number is not present in the frames collection")
	} else {
		return frame, nil
	}
}

func (fc *frameCollection) GetAll() []*Frame {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if !fc.Checked {
		for index, frame := range fc.Frames {
			if frame == nil {
				panic("frame: missing frame found in the collection")
			}

			if index+1 != frame.OrdinalNumber {
				panic("frame: out of order frame found in the collection")
			}
		}
	}

	return fc.Frames
}

func (fc *frameCollection) Count() int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	return len(fc.Frames)
}

func (fc *frameCollection) ExportCache(file io.Writer, checksum string) (err error) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	defer func() {
		if err := recover(); err != nil {
			err = fmt.Errorf("frame: failed to encode the frame collection due to incorrect data: %s", err)
		}
	}()

	encoder := json.NewEncoder(file)

	fcc := frameCollectionCache{
		Checksum: checksum,
		Frames:   fc.Frames,
	}

	if err := encoder.Encode(fcc); err != nil {
		return fmt.Errorf("frame: failed to encode the frame collection frames to cache: %w", err)
	}

	return nil
}
