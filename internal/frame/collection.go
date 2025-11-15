package frame

import (
	"fmt"
)

const baseFrameCollectionCapacity = 32

// Structure representing the collection of video frames.
type FrameCollection interface {
	Push(frame *Frame) error
	GetAll() []*Frame
	Count() int
	Lock()
}

type frameCollection struct {
	Frames   []*Frame
	Index    int
	Capacity int
	Locked   bool
}

func (fc *frameCollection) Push(frame *Frame) error {
	if frame == nil {
		return fmt.Errorf("frame: invalid uninitialized frame reference")
	}

	if fc.Locked {
		return fmt.Errorf("frame: frame collection is locked")
	}

	if fc.Index != frame.OrdinalNumber-1 {
		return fmt.Errorf("frame: collection indexing and provided frame order missmatch")
	}

	if fc.Index < fc.Capacity {
		fc.Frames[fc.Index] = frame
	} else {
		fc.Frames = append(fc.Frames, frame)
	}

	fc.Index += 1

	return nil
}

func (fc *frameCollection) GetAll() []*Frame {
	if !fc.Locked {
		panic("frame: can not read from an unlocked frame collection")
	}

	a := fc.Frames[:fc.Index]
	return a
}

func (fc *frameCollection) Count() int {
	if !fc.Locked {
		panic("frame: can not read from an unlocked frame collection")
	}

	return fc.Index
}

func (fc *frameCollection) Lock() {
	fc.Locked = true
}

func NewFrameCollection(cap int) FrameCollection {
	if cap <= 0 {
		panic("frame: frame collection capacity must be greater than zero")
	}

	return &frameCollection{
		Frames:   make([]*Frame, cap, baseFrameCollectionCapacity+cap),
		Index:    0,
		Capacity: cap,
		Locked:   false,
	}
}
