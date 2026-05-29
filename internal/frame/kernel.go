package frame

import (
	"image"
	"math"
	"runtime"
	"sync"

	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

type kernelResult struct {
	Count                         float64
	BrightnessMean                float64
	BrightnessM2                  float64
	BrightnessMin                 float64
	BrightnessMax                 float64
	ColorDifferenceMean           float64
	ColorDifferenceM2             float64
	ColorDifferenceMin            float64
	ColorDifferenceMax            float64
	BinaryThresholdDifferenceMean float64
	BinaryThresholdDifferenceM2   float64
	BinaryThresholdCount          float64
	LuminanceMean                 float64
	LuminanceM2                   float64
	LuminanceMin                  float64
	LuminanceMax                  float64
	SaturationMean                float64
	SaturationM2                  float64
	SaturationMin                 float64
	SaturationMax                 float64
}

type aggregatedKernelResult struct {
	Brightness                 float64
	ColorDifference            float64
	BinaryThresholdDifference  float64
	BrightnessStdDev           float64
	BrightnessMin              float64
	BrightnessMax              float64
	BrightnessFirstDerivative  float64
	BrightnessSecondDerivative float64
	SaturationMean             float64
	SaturationStdDev           float64
	ColorDifferenceVariance    float64
	LuminanceMean              float64
	LuminanceMin               float64
	LuminanceMax               float64
	BinaryThresholdRatio       float64
}

type frame aggregatedKernelResult

func processFrame(frameImage0, frameImage1 *image.RGBA, frame1, frame2 *Frame, ordinal int, bThreshold float64) frame {
	var (
		workers    int = runtime.NumCPU()
		pixelCount int = frameImage0.Bounds().Dx() * frameImage0.Bounds().Dy()
	)

	if workers > pixelCount {
		workers = 1
	}

	var (
		countPerWorker         int               = pixelCount / workers
		countPerWorkerReminder int               = pixelCount % workers
		kernelResultChannel    chan kernelResult = make(chan kernelResult, workers)
		wg                     sync.WaitGroup    = sync.WaitGroup{}
		currentFrameBuffer     []uint8           = frameImage0.Pix
		previousFrameBuffer    []uint8           = make([]uint8, 0)
	)

	if ordinal > 1 {
		previousFrameBuffer = frameImage1.Pix
	}

	for index := 0; index < workers; index += 1 {
		offset := index * countPerWorker

		count := countPerWorker
		if index+1 == workers {
			count += countPerWorkerReminder
		}

		wg.Add(1)
		go processKernel(currentFrameBuffer, previousFrameBuffer, offset, count, ordinal, bThreshold, kernelResultChannel, &wg)
	}

	wg.Wait()
	close(kernelResultChannel)

	aggregatedResult := aggregateKernels(kernelResultChannel, frame1, frame2, ordinal)
	return frame(aggregatedResult)
}

func processKernel(current, previous []uint8, offset, count, ordinal int, bthreshold float64, kernelChannel chan<- kernelResult, wg *sync.WaitGroup) {
	defer wg.Done()

	kr := kernelResult{
		Count:                         0,
		BrightnessMean:                0,
		BrightnessM2:                  0,
		BrightnessMin:                 1,
		BrightnessMax:                 0,
		ColorDifferenceMean:           0,
		ColorDifferenceM2:             0,
		ColorDifferenceMin:            1,
		ColorDifferenceMax:            0,
		BinaryThresholdDifferenceMean: 0,
		BinaryThresholdDifferenceM2:   0,
		BinaryThresholdCount:          0,
		SaturationMean:                0,
		SaturationM2:                  0,
		SaturationMin:                 1,
		SaturationMax:                 0,
		LuminanceMean:                 0,
		LuminanceM2:                   0,
		LuminanceMin:                  1,
		LuminanceMax:                  0,
	}

	const (
		step int = 4
	)

	var (
		indexOffset int   = step * offset
		indexCount  int   = step * (offset + count)
		cR          uint8 = 0
		cG          uint8 = 0
		cB          uint8 = 0
		pR          uint8 = 0
		pG          uint8 = 0
		pB          uint8 = 0
	)

	for index := indexOffset; index < indexCount; index += 4 {
		cR = current[index+0]
		cG = current[index+1]
		cB = current[index+2]

		kr.Count += 1

		brightness := utils.GetColorBrightness(cR, cG, cB)

		brightnessDelta := brightness - kr.BrightnessMean
		kr.BrightnessMean += brightnessDelta / kr.Count
		kr.BrightnessM2 += brightnessDelta * (brightness - kr.BrightnessMean)

		if kr.BrightnessMin > brightness {
			kr.BrightnessMin = brightness
		}

		if kr.BrightnessMax < brightness {
			kr.BrightnessMax = brightness
		}

		saturation, luminance := utils.GetSaturationLuminance(cR, cG, cB)

		saturationDelta := saturation - kr.BrightnessMean
		kr.SaturationMean += saturationDelta / kr.Count
		kr.SaturationM2 += saturationDelta * (saturation - kr.SaturationMean)

		luminanceDelta := luminance - kr.LuminanceMean
		kr.LuminanceMean += luminanceDelta / kr.Count
		kr.LuminanceM2 += luminanceDelta * (luminance - kr.LuminanceMean)

		if kr.SaturationMin > saturation {
			kr.SaturationMin = saturation
		}

		if kr.SaturationMax < saturation {
			kr.SaturationMax = saturation
		}

		if kr.LuminanceMin > luminance {
			kr.LuminanceMin = luminance
		}

		if kr.LuminanceMax < luminance {
			kr.LuminanceMax = luminance
		}

		if ordinal == 1 {
			continue
		}

		pR = previous[index+0]
		pG = previous[index+1]
		pB = previous[index+2]

		colorDifference := utils.GetColorDifference(cR, cG, cB, pR, pG, pB)

		colorDifferenceDelta := colorDifference - kr.ColorDifferenceMean
		kr.ColorDifferenceMean += colorDifferenceDelta / kr.Count
		kr.ColorDifferenceM2 += colorDifferenceDelta * (colorDifference - kr.ColorDifferenceMean)

		if kr.ColorDifferenceMin > colorDifference {
			kr.ColorDifferenceMin = colorDifference
		}

		if kr.ColorDifferenceMax < colorDifference {
			kr.ColorDifferenceMax = colorDifference
		}

		cBt := utils.BinaryThreshold(cR, cG, cB, bthreshold)
		pBt := utils.BinaryThreshold(pR, pG, pB, bthreshold)

		if cBt == 0xFF {
			kr.BinaryThresholdCount += 1
		}

		binaryThresholdDifference := 0.0
		if cBt != pBt {
			binaryThresholdDifference = 1.0
		}

		binaryThresholdDifferenceDelta := binaryThresholdDifference - kr.BinaryThresholdDifferenceMean
		kr.BinaryThresholdDifferenceMean += binaryThresholdDifferenceDelta / kr.Count
		kr.BinaryThresholdDifferenceM2 += binaryThresholdDifferenceDelta * (colorDifference - kr.ColorDifferenceMean)

		if kr.ColorDifferenceMin > colorDifference {
			kr.ColorDifferenceMin = colorDifference
		}

		if kr.ColorDifferenceMax < colorDifference {
			kr.ColorDifferenceMax = colorDifference
		}
	}

	kernelChannel <- kr
}

func mergeTwoKernels(a, b kernelResult) kernelResult {
	if a.Count == 0 {
		return b
	}

	if b.Count == 0 {
		return a
	}

	var kr kernelResult

	kr.Count = a.Count + b.Count

	bCc := b.Count / kr.Count
	aCbCc := a.Count * b.Count / kr.Count

	brightnessDelta := b.BrightnessMean - a.BrightnessMean
	kr.BrightnessMean = a.BrightnessMean + brightnessDelta*bCc
	kr.BrightnessM2 = a.BrightnessM2 + b.BrightnessM2 + brightnessDelta*brightnessDelta*aCbCc

	if a.BrightnessMin < b.BrightnessMin {
		kr.BrightnessMin = a.BrightnessMin
	} else {
		kr.BrightnessMin = b.BrightnessMin
	}

	if a.BrightnessMax > b.BrightnessMax {
		kr.BrightnessMax = a.BrightnessMax
	} else {
		kr.BrightnessMax = b.BrightnessMax
	}

	saturationDelta := b.SaturationMean - a.SaturationMean
	kr.SaturationMean = a.SaturationMean + saturationDelta*bCc
	kr.SaturationM2 = a.SaturationM2 + b.SaturationM2 + saturationDelta*saturationDelta*aCbCc

	if a.SaturationMin < b.SaturationMin {
		kr.SaturationMin = a.SaturationMin
	} else {
		kr.SaturationMin = b.SaturationMin
	}

	if a.SaturationMax > b.SaturationMax {
		kr.SaturationMax = a.SaturationMax
	} else {
		kr.SaturationMax = b.SaturationMax
	}

	luminanceDelta := b.LuminanceMean - a.LuminanceMean
	kr.LuminanceMean = a.LuminanceMean + luminanceDelta*bCc
	kr.LuminanceM2 = a.LuminanceM2 + b.LuminanceM2 + luminanceDelta*luminanceDelta*aCbCc

	if a.LuminanceMin < b.LuminanceMin {
		kr.LuminanceMin = a.LuminanceMin
	} else {
		kr.LuminanceMin = b.LuminanceMin
	}

	if a.LuminanceMax > b.LuminanceMax {
		kr.LuminanceMax = a.LuminanceMax
	} else {
		kr.LuminanceMax = b.LuminanceMax
	}

	colorDifferenceDelta := b.ColorDifferenceMean - a.ColorDifferenceMean
	kr.ColorDifferenceMean = a.ColorDifferenceMean + colorDifferenceDelta*bCc
	kr.ColorDifferenceM2 = a.ColorDifferenceM2 + b.ColorDifferenceM2 + colorDifferenceDelta*colorDifferenceDelta*aCbCc

	if a.ColorDifferenceMin < b.ColorDifferenceMin {
		kr.ColorDifferenceMin = a.ColorDifferenceMin
	} else {
		kr.ColorDifferenceMin = b.ColorDifferenceMin
	}

	if a.ColorDifferenceMax > b.ColorDifferenceMax {
		kr.ColorDifferenceMax = a.ColorDifferenceMax
	} else {
		kr.ColorDifferenceMax = b.ColorDifferenceMax
	}

	binaryThresholdDifferenceDelta := b.BinaryThresholdDifferenceMean - a.BinaryThresholdDifferenceMean
	kr.BinaryThresholdDifferenceMean = a.BinaryThresholdDifferenceMean + binaryThresholdDifferenceDelta*bCc
	kr.BinaryThresholdDifferenceM2 = a.BinaryThresholdDifferenceM2 + b.BinaryThresholdDifferenceM2 + binaryThresholdDifferenceDelta*binaryThresholdDifferenceDelta*aCbCc

	kr.BinaryThresholdCount = a.BinaryThresholdCount + b.BinaryThresholdCount

	return kr
}

func aggregateKernels(kernelChannel <-chan kernelResult, frame1, frame2 *Frame, ordinal int) aggregatedKernelResult {
	var kr kernelResult
	for kernel := range kernelChannel {
		kr = mergeTwoKernels(kr, kernel)
	}

	var (
		previousFrameBrightness    float64 = 0
		penultimateFrameBrightness float64 = 0
	)

	if ordinal > 1 {
		previousFrameBrightness = frame1.Brightness
	}

	if ordinal > 2 {
		penultimateFrameBrightness = frame2.Brightness
	}

	return aggregatedKernelResult{
		Brightness:                 kr.BrightnessMean,
		ColorDifference:            kr.ColorDifferenceMean,
		BinaryThresholdDifference:  kr.BinaryThresholdDifferenceMean,
		BrightnessStdDev:           math.Sqrt(kr.BrightnessM2 / kr.Count),
		BrightnessMin:              kr.BrightnessMin,
		BrightnessMax:              kr.BrightnessMax,
		BrightnessFirstDerivative:  kr.BrightnessMean - previousFrameBrightness,
		BrightnessSecondDerivative: kr.BrightnessMean - 2*previousFrameBrightness + penultimateFrameBrightness,
		SaturationMean:             kr.SaturationMean,
		SaturationStdDev:           math.Sqrt(kr.SaturationM2 / kr.Count),
		ColorDifferenceVariance:    kr.ColorDifferenceM2 / kr.Count,
		LuminanceMean:              kr.LuminanceMean,
		LuminanceMin:               kr.LuminanceMin,
		LuminanceMax:               kr.LuminanceMax,
		BinaryThresholdRatio:       kr.BinaryThresholdCount / kr.Count,
	}
}
