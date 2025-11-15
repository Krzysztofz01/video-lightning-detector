package utils

type DecayingHashSet[T comparable] interface {
	Add(value T)
	Contains(value T) bool
	Values() []T
}

type decayingHashSet[T comparable] struct {
	DataA        map[T]bool
	DataB        map[T]bool
	State        bool
	DecayCount   int
	DecayCounter int
}

func (set *decayingHashSet[T]) Add(value T) {
	if set.DecayCounter == 4*set.DecayCount {
		if set.State {
			clear(set.DataB)
			set.DecayCounter = len(set.DataA)
		} else {
			clear(set.DataA)
			set.DecayCounter = len(set.DataB)
		}

		set.State = !set.State
	}

	if set.DecayCounter < 3*set.DecayCount {
		if set.State {
			if _, ok := set.DataB[value]; !ok {
				set.DataB[value] = true
				set.DecayCounter += 1
			}
		} else {
			if _, ok := set.DataA[value]; !ok {
				set.DataA[value] = true
				set.DecayCounter += 1
			}
		}

		return
	}

	if set.DecayCounter < 4*set.DecayCount {
		if set.State {
			if _, ok := set.DataB[value]; !ok {
				set.DataB[value] = true
				set.DataA[value] = true
				set.DecayCounter += 1
			}
		} else {
			if _, ok := set.DataA[value]; !ok {
				set.DataA[value] = true
				set.DataB[value] = true
				set.DecayCounter += 1
			}
		}

		return
	}
}

func (set *decayingHashSet[T]) Contains(value T) bool {
	var contains bool
	if set.State {
		_, contains = set.DataB[value]
	} else {
		_, contains = set.DataA[value]
	}

	return contains
}

func (set *decayingHashSet[T]) Values() []T {
	valuesMap := make(map[T]bool, set.DecayCount*5)

	for value, _ := range set.DataA {
		valuesMap[value] = true
	}

	for value, _ := range set.DataB {
		valuesMap[value] = true
	}

	values := make([]T, 0, len(valuesMap))
	for value := range valuesMap {
		values = append(values, value)
	}

	return values
}

func NewDecayingHashSet[T comparable](decay int) DecayingHashSet[T] {
	if decay <= 0 {
		panic("utils: decay count must be greater than zero")
	}

	return &decayingHashSet[T]{
		DataA:        make(map[T]bool, decay*4),
		DataB:        make(map[T]bool, decay*4),
		State:        false,
		DecayCount:   decay,
		DecayCounter: 0,
	}
}
