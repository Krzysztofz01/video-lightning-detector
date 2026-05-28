package utils

import (
	"fmt"
	"image"
	"image/draw"
	"math"
	"runtime"
	"sync"
)

// Create a RGBA copy of the image represented with a image.Image interface
func CopyAsRgba(i image.Image) (*image.RGBA, error) {
	if i == nil {
		return nil, fmt.Errorf("utils: invalid image reference provided")
	}

	rgba := image.NewRGBA(i.Bounds())
	draw.Draw(rgba, rgba.Bounds(), i, i.Bounds().Min, draw.Src)

	return rgba, nil
}

// Run a convolution filter on the provided image. The image channels will be merged using the specified func
// which should return a (0,1) range normalized value. The kernel must be a square, odd dimension length, not
// normalized matrix. The resulting float64 slice is a 1D representation of the image matching the input bounds.
func MergeConvolveFloat64(i *image.RGBA, merge func(r, g, b, a uint8) float64, kernel []float64) ([]float64, error) {
	if i == nil {
		return nil, fmt.Errorf("utils: invalid image reference provided")
	}

	if merge == nil {
		return nil, fmt.Errorf("utils: invalid merge func reference provided")
	}

	kl, perf := IsPerfectSquare(len(kernel))
	if !perf || kl <= 0 || kl%2 == 0 {
		return nil, fmt.Errorf("utils: invalid kernel dimensions")
	}

	kernel = normalizeConvKernel(kernel)

	var (
		h   = i.Bounds().Dy()
		w   = i.Bounds().Dx()
		klh = kl / 2
		src = i.Pix
		dst = make([]float64, h*w)
	)

	var workers = runtime.NumCPU()
	if workers > h {
		workers = 1
	}

	var (
		rows = h / workers
		wg   = sync.WaitGroup{}
	)

	for i := 0; i < workers; i += 1 {
		start := i * rows
		end := start + rows

		if i == workers-1 {
			end = h
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			var (
				value                     float64
				kOffset, sOffset, dOffset int
				weight                    float64
			)

			for y := start; y < end; y += 1 {
				for x := 0; x < w; x += 1 {
					dOffset = y*w + x

					if x-klh < 0 || x+klh >= w || y-klh < 0 || y+klh >= h {
						sOffset = y*w*4 + x*4

						dst[dOffset] = float64(src[sOffset]) / 255
						continue
					}

					value = 0
					for ky := -klh; ky <= klh; ky += 1 {
						for kx := -klh; kx <= klh; kx += 1 {
							kOffset = (ky+klh)*kl + (kx + klh)
							sOffset = (y+ky)*w*4 + (x+kx)*4

							weight = kernel[kOffset]
							value += merge(src[sOffset], src[sOffset+1], src[sOffset+2], src[sOffset+3]) * weight
						}
					}

					dOffset = y*w + x
					dst[dOffset] = math.Max(0, math.Min(value, 1))
				}
			}
		}(start, end)
	}

	wg.Wait()

	return dst, nil
}

// Run a convolution filter on the provided image. The image channels will be merged using the specified func
// which should return a (0,1) range normalized value. The kernel must be a square, odd dimension length, not
// normalized matrix.
func MergeConvolveRGBA(i *image.RGBA, merge func(r, g, b, a uint8) float64, kernel []float64) (*image.RGBA, error) {
	if i == nil {
		return nil, fmt.Errorf("utils: invalid image reference provided")
	}

	if merge == nil {
		return nil, fmt.Errorf("utils: invalid merge func reference provided")
	}

	kl, perf := IsPerfectSquare(len(kernel))
	if !perf || kl <= 0 || kl%2 == 0 {
		return nil, fmt.Errorf("utils: invalid kernel dimensions")
	}

	kernel = normalizeConvKernel(kernel)

	var (
		h   = i.Bounds().Dy()
		w   = i.Bounds().Dx()
		klh = kl / 2
		src = i.Pix
		dst = image.NewRGBA(i.Bounds())
	)

	var (
		workers = runtime.NumCPU()
		rows    = h / workers
		wg      = sync.WaitGroup{}
	)

	for i := 0; i < workers; i += 1 {
		start := i * rows
		end := start + rows

		if i == workers-1 {
			end = h
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			var (
				valueF                    float64
				value                     uint8
				kOffset, sOffset, dOffset int
				weight                    float64
			)

			for y := start; y < end; y += 1 {
				for x := 0; x < w; x += 1 {
					dOffset = y*w*4 + x*4

					if x-klh < 0 || x+klh >= w || y-klh < 0 || y+klh >= h {
						sOffset = dOffset

						dst.Pix[dOffset+0] = src[sOffset+0]
						dst.Pix[dOffset+1] = src[sOffset+1]
						dst.Pix[dOffset+2] = src[sOffset+2]
						dst.Pix[dOffset+3] = src[sOffset+3]
						continue
					}

					valueF = 0
					for ky := -klh; ky <= klh; ky += 1 {
						for kx := -klh; kx <= klh; kx += 1 {
							kOffset = (ky+klh)*kl + (kx + klh)
							sOffset = (y+ky)*w*4 + (x+kx)*4

							weight = kernel[kOffset]
							valueF += merge(src[sOffset], src[sOffset+1], src[sOffset+2], src[sOffset+3]) * weight
						}
					}

					valueF = math.Max(0, math.Min(valueF*255, 255))
					value = uint8(valueF)

					dst.Pix[dOffset+0] = value
					dst.Pix[dOffset+1] = value
					dst.Pix[dOffset+2] = value
					dst.Pix[dOffset+3] = src[dOffset+3]
				}
			}
		}(start, end)
	}

	wg.Wait()

	return dst, nil
}

func normalizeConvKernel(k []float64) []float64 {
	var (
		sum    float64 = 0
		sumPos float64 = 0
	)

	for _, v := range k {
		sum += v
		if v > 0.0 {
			sumPos += v
		}
	}

	if sum != 0 {
		for i := range k {
			k[i] /= sum
		}
	} else {
		for i := range k {
			k[i] /= sumPos
		}
	}

	return k
}
