package frame

import (
	"image"
	"runtime"
	"sync"

	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

type kernelResult struct {
	BrightnessSum                float64
	ColorDifferenceSum           float64
	BinaryThresholdDifferenceSum uint64
}

type aggregatedKernelResult struct {
	Brightness                float64
	ColorDifference           float64
	BinaryThresholdDifference float64
}

type frame aggregatedKernelResult

func processFrame(currentFrame, previousFrame *image.RGBA, ordinal int, bThreshold float64) frame {
	var (
		workers                int               = runtime.NumCPU()
		pixelCount             int               = currentFrame.Bounds().Dx() * currentFrame.Bounds().Dy()
		countPerWorker         int               = pixelCount / workers
		countPerWorkerReminder int               = pixelCount % workers
		kernelResultChannel    chan kernelResult = make(chan kernelResult, workers)
		wg                     sync.WaitGroup    = sync.WaitGroup{}
		currentFrameBuffer     []uint8           = currentFrame.Pix
		previousFrameBuffer    []uint8           = make([]uint8, 0)
	)

	if previousFrame != nil {
		previousFrameBuffer = previousFrame.Pix
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

	aggregatedResult := aggregateKernels(kernelResultChannel, pixelCount)
	return frame(aggregatedResult)
}

func processKernel(current, previous []uint8, offset, count, ordinal int, bthreshold float64, kernelChannel chan<- kernelResult, wg *sync.WaitGroup) {
	defer wg.Done()

	result := kernelResult{
		BrightnessSum:                0,
		ColorDifferenceSum:           0,
		BinaryThresholdDifferenceSum: 0,
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
		cR = current[indexOffset+0]
		cG = current[indexOffset+1]
		cB = current[indexOffset+2]

		result.BrightnessSum += utils.GetColorBrightness(cR, cG, cB)

		if ordinal == 1 {
			continue
		}

		pR = previous[indexOffset+0]
		pG = previous[indexOffset+1]
		pB = previous[indexOffset+2]

		result.ColorDifferenceSum += utils.GetColorDifference(cR, cG, cB, pR, pG, pB)

		cBt := utils.BinaryThreshold(cR, cG, cB, bthreshold)
		pBt := utils.BinaryThreshold(pR, pG, pB, bthreshold)
		if cBt != pBt {
			result.BinaryThresholdDifferenceSum += 1
		}
	}

	kernelChannel <- result
}

func aggregateKernels(kernelChannel <-chan kernelResult, pixelCount int) aggregatedKernelResult {
	result := aggregatedKernelResult{
		Brightness:                0,
		ColorDifference:           0,
		BinaryThresholdDifference: 0,
	}

	for kernel := range kernelChannel {
		result.Brightness += kernel.BrightnessSum
		result.ColorDifference += kernel.ColorDifferenceSum
		result.BinaryThresholdDifference += float64(kernel.BinaryThresholdDifferenceSum)
	}

	count := float64(pixelCount)
	result.Brightness /= count
	result.ColorDifference /= count
	result.BinaryThresholdDifference /= count

	return result
}
