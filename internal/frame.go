package internal

import "sync"

type Frame struct {
	OrdinalNumber int     `json:"ordinal-number"`
	Difference    float64 `json:"difference"`
	Brightness    float64 `json:"brightness"`
	mu            sync.Mutex
}

func NewFrame(ordinalNumber int) *Frame {
	return &Frame{
		OrdinalNumber: ordinalNumber,
		Difference:    0,
		Brightness:    0,
		mu:            sync.Mutex{},
	}
}

func (frame *Frame) SetDifference(difference float64) {
	frame.mu.Lock()
	defer frame.mu.Unlock()

	frame.Difference = difference
}

func (frame *Frame) SetBrightness(brightnes float64) {
	frame.mu.Lock()
	defer frame.mu.Unlock()

	frame.Brightness = brightnes
}
