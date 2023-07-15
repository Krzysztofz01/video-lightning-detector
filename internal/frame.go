package internal

import (
	"strconv"
	"sync"
)

type Frame struct {
	OrdinalNumber int     `json:"ordinal-number"`
	Difference    float64 `json:"difference"`
	Brightness    float64 `json:"brightness"`
	mu            sync.RWMutex
}

func NewFrame(ordinalNumber int) *Frame {
	return &Frame{
		OrdinalNumber: ordinalNumber,
		Difference:    0,
		Brightness:    0,
		mu:            sync.RWMutex{},
	}
}

func (frame *Frame) SetDifference(difference float64) {
	frame.mu.Lock()
	defer frame.mu.Unlock()

	frame.Difference = difference
}

func (frame *Frame) GetDifference() float64 {
	frame.mu.RLock()
	defer frame.mu.RUnlock()

	return frame.Difference
}

func (frame *Frame) SetBrightness(brightnes float64) {
	frame.mu.Lock()
	defer frame.mu.Unlock()

	frame.Brightness = brightnes
}

func (frame *Frame) GetBrightness() float64 {
	frame.mu.RLock()
	defer frame.mu.RUnlock()

	return frame.Brightness
}

func (frame *Frame) ToBuffer() []string {
	frame.mu.Lock()
	defer frame.mu.Unlock()

	buffer := make([]string, 0, 3)
	buffer = append(buffer, strconv.Itoa(frame.OrdinalNumber))
	buffer = append(buffer, strconv.FormatFloat(frame.Brightness, 'f', -1, 64))
	buffer = append(buffer, strconv.FormatFloat(frame.Difference, 'f', -1, 64))

	return buffer
}
