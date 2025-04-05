package denoise

import (
	"fmt"
	"strings"
)

type Algorithm int

// TODO: Research and implement bilateral filter
const (
	NoDenoise Algorithm = iota
	StackBlur8
	StackBlur16
	StackBlur32
)

var algorithmNames = map[string]Algorithm{
	"none":        NoDenoise,
	"stackblur8":  StackBlur8,
	"stackblur16": StackBlur16,
	"stackblur32": StackBlur32,
}

func (a *Algorithm) String() string {
	for name, algorith := range algorithmNames {
		if algorith == *a {
			return name
		}
	}

	panic("denoise: invalid unknown algorithm")
}

func (a *Algorithm) Set(s string) error {
	if algorithm, ok := algorithmNames[strings.ToLower(s)]; !ok {
		return fmt.Errorf("denoise: invalid unknown algorithm name")
	} else {
		*a = algorithm
	}

	return nil
}

func (a *Algorithm) Type() string {
	return "algorithm"
}
