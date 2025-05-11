package frame

import (
	"fmt"
)

// Structure representing the collection of video frames.
type FrameCollection interface {
	Push(frame *Frame) error
	GetAll() []*Frame
	Count() int
}

type frameCollection struct {
	Frames   []*Frame
	Index    int
	Capacity int
	Strict   bool
}

func (fc *frameCollection) Push(frame *Frame) error {
	if frame == nil {
		return fmt.Errorf("frame: invalid uninitialized frame reference")
	}

	if fc.Index >= fc.Capacity {
		return fmt.Errorf("frame: frame collection capacity exceeded")
	}

	if fc.Index != frame.OrdinalNumber-1 {
		return fmt.Errorf("frame: collection indexing and provided frame order missmatch")
	}

	fc.Frames[fc.Index] = frame
	fc.Index += 1

	return nil
}

func (fc *frameCollection) GetAll() []*Frame {
	if fc.Strict && fc.Index != fc.Capacity {
		panic("frame: can not access a unsaturated frame collection")
	}

	return fc.Frames
}

func (fc *frameCollection) Count() int {
	if fc.Strict && fc.Index != fc.Capacity {
		panic("frame: can not access a unsaturated frame collection")
	}

	return len(fc.Frames)
}

func NewFrameCollection(cap int) FrameCollection {
	if cap <= 0 {
		panic("frame: frame collection capacity must be greater than zero")
	}

	return &frameCollection{
		Frames:   make([]*Frame, cap),
		Index:    0,
		Capacity: cap,
		Strict:   true,
	}
}
