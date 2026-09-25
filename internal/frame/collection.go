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
}

func (fc *frameCollection) Push(frame *Frame) error {
	if frame == nil {
		return fmt.Errorf("frame: invalid uninitialized frame reference")
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
	return fc.Frames[:fc.Index]
}

func (fc *frameCollection) Count() int {
	return fc.Index
}

func NewFrameCollection(cap int) FrameCollection {
	if cap <= 0 {
		panic("frame: frame collection capacity must be greater than zero")
	}

	return &frameCollection{
		Frames:   make([]*Frame, cap, cap+64),
		Index:    0,
		Capacity: cap,
	}
}
