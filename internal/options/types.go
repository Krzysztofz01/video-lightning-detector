package options

import (
	"fmt"
	"strings"
)

type DenoiseAlgorithm int

const (
	NoDenoise DenoiseAlgorithm = iota
	StackBlur8
	StackBlur16
	StackBlur32
)

func IsValidDenoiseAlgorithm(a DenoiseAlgorithm) bool {
	switch a {
	case NoDenoise, StackBlur8, StackBlur16, StackBlur32:
		return true
	default:
		return false
	}
}

var denoiseAlgorithmNames = map[string]DenoiseAlgorithm{
	"none":        NoDenoise,
	"stackblur8":  StackBlur8,
	"stackblur16": StackBlur16,
	"stackblur32": StackBlur32,
}

func (a *DenoiseAlgorithm) String() string {
	for name, algorith := range denoiseAlgorithmNames {
		if algorith == *a {
			return name
		}
	}

	panic("options: invalid unknown denoise algorithm")
}

func (a *DenoiseAlgorithm) Set(s string) error {
	if algorithm, ok := denoiseAlgorithmNames[strings.ToLower(s)]; !ok {
		return fmt.Errorf("options: invalid unknown denoise algorithm name")
	} else {
		*a = algorithm
	}

	return nil
}

func (a *DenoiseAlgorithm) Type() string {
	return "denoisealgorithm"
}

type ScaleAlgorithm int

const (
	Default ScaleAlgorithm = iota
	Bilinear
	Bicubic
	NearestNeighbour
	Lanczos
	Area
)

func IsValidScaleAlgorithm(a ScaleAlgorithm) bool {
	switch a {
	case Default, Bilinear, Bicubic, NearestNeighbour, Lanczos, Area:
		return true
	default:
		return false
	}
}

var scaleAlgorithmNames = map[string]ScaleAlgorithm{
	"default":  Default,
	"bilinear": Bilinear,
	"bicubic":  Bicubic,
	"nearest":  NearestNeighbour,
	"lanczos":  Lanczos,
	"area":     Area,
}

func (a *ScaleAlgorithm) String() string {
	for name, algorith := range scaleAlgorithmNames {
		if algorith == *a {
			return name
		}
	}

	panic("options: invalid unknown scale algorithm")
}

func (a *ScaleAlgorithm) Set(s string) error {
	if algorithm, ok := scaleAlgorithmNames[strings.ToLower(s)]; !ok {
		return fmt.Errorf("options: invalid unknown scale algorithm name")
	} else {
		*a = algorithm
	}

	return nil
}

func (a *ScaleAlgorithm) Type() string {
	return "scalealgorithm"
}
