package utils

import "fmt"

type CircularBuffer[T any] interface {
	Push(value T)
	PushP() T
	GetHead(index int) (T, error)
	GetTail(index int) (T, error)
	GetTotalCount() int
	GetBufferCount() int
	IsSaturated() bool
	PeekWindow() []T
}

type circularBuffer[T any] struct {
	Data            []T
	Index           int
	Capacity        int
	SaturatedOnInit bool
}

func (cb *circularBuffer[T]) Push(value T) {
	cb.Data[cb.Index%cb.Capacity] = value

	cb.Index += 1
}

func (cb *circularBuffer[T]) PushP() T {
	value := cb.Data[cb.Index%cb.Capacity]
	cb.Index += 1

	return value
}

func (cb *circularBuffer[T]) GetHead(index int) (T, error) {
	length := cb.Capacity
	if cb.Index < cb.Capacity {
		length = cb.Index
	}

	return cb.Get(length - index - 1)
}

func (cb *circularBuffer[T]) GetTail(index int) (T, error) {
	return cb.Get(index)
}

func (cb *circularBuffer[T]) Get(index int) (T, error) {
	if index < 0 || index >= cb.Index {
		return *new(T), fmt.Errorf("utils: the index for the head access is out of range")
	}

	if cb.Index >= cb.Capacity {
		index += cb.Index % cb.Capacity
	}

	return cb.Data[index%cb.Capacity], nil
}

func (cb *circularBuffer[T]) GetTotalCount() int {
	return cb.Index
}

func (cb *circularBuffer[T]) GetBufferCount() int {
	if cb.SaturatedOnInit {
		return cb.Capacity
	}

	return MinInt(cb.Index, cb.Capacity)
}

func (cb *circularBuffer[T]) IsSaturated() bool {
	if cb.SaturatedOnInit {
		return true
	}

	return cb.Index >= cb.Capacity
}

func (cb *circularBuffer[T]) PeekWindow() []T {
	return cb.Data
}

func NewCircularBuffer[T any](capacity int) CircularBuffer[T] {
	if capacity <= 0 {
		panic("utils: the circular buffer must have a capacity greater than zero")
	}

	return &circularBuffer[T]{
		Data:            make([]T, capacity, capacity),
		Index:           0,
		Capacity:        capacity,
		SaturatedOnInit: false,
	}
}

func NewSaturatedCircularBuffer[T any](values []T) CircularBuffer[T] {
	if len(values) <= 0 {
		panic("utils: the values slice to saturate the circular buffer can not be empty")
	}

	cbuffer := &circularBuffer[T]{
		Data:            make([]T, len(values), len(values)),
		Index:           0,
		Capacity:        len(values),
		SaturatedOnInit: true,
	}

	copy(cbuffer.Data, values)
	return cbuffer
}
