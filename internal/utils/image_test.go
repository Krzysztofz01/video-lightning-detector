package utils

import (
	"github.com/stretchr/testify/assert"
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestCopyAsRgbaShouldNotCopyNilImage(t *testing.T) {
	i, err := CopyAsRgba(nil)
	assert.NotNil(t, err)
	assert.Nil(t, i)
}

func TestCopyAsRgbaShouldCopy(t *testing.T) {
	var (
		size   int          = 4
		rect                = image.Rect(0, 0, size, size)
		images []draw.Image = []draw.Image{
			image.NewRGBA(rect),
			image.NewNRGBA(rect),
			image.NewGray(rect),
		}
	)

	for _, i := range images {
		for y := 0; y < i.Bounds().Dy(); y += 1 {
			for x := 0; x < i.Bounds().Dx(); x += 1 {
				i.Set(x, y, color.White)
			}
		}
	}

	for _, i := range images {
		imageRgba, err := CopyAsRgba(i)
		assert.Nil(t, err)
		assert.NotNil(t, imageRgba)

		for y := 0; y < i.Bounds().Dy(); y += 1 {
			for x := 0; x < i.Bounds().Dx(); x += 1 {
				eR, eG, eB, eA := i.At(x, y).RGBA()
				aR, aG, aB, aA := imageRgba.At(x, y).RGBA()

				assert.Equal(t, eR, aR)
				assert.Equal(t, eG, aG)
				assert.Equal(t, eB, aB)
				assert.Equal(t, eA, aA)
			}
		}

		imageRgba.Set(0, 0, color.Black)
		eR, eG, eB, _ := i.At(0, 0).RGBA()
		aR, aG, aB, _ := imageRgba.At(0, 0).RGBA()

		assert.NotEqual(t, eR, aR)
		assert.NotEqual(t, eG, aG)
		assert.NotEqual(t, eB, aB)
	}
}

func TestMergeConvolveFloat64ShouldConvolveImage(t *testing.T) {
	cases := []struct {
		kernel   []float64
		expected []float64
	}{
		{
			kernel: []float64{
				1, 1, 1,
				1, 1, 1,
				1, 1, 1,
			},
			expected: []float64{
				0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
				0.0, 0.0, 2.0 / 9, 4.0 / 9, 4.0 / 9, 0.0,
				0.0, 0.0, 3.0 / 9, 6.0 / 9, 6.0 / 9, 0.0,
				0.0, 0.0, 3.0 / 9, 6.0 / 9, 6.0 / 9, 0.0,
				0.0, 0.0, 2.0 / 9, 4.0 / 9, 4.0 / 9, 0.0,
				0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
			},
		},
		{
			kernel: []float64{
				0, 0, 0,
				-1, 0, 1,
				0, 0, 0,
			},
			expected: []float64{
				0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
				0.0, 0.0, 1.0, 1.0, 0.0, 0.0,
				0.0, 0.0, 1.0, 1.0, 0.0, 0.0,
				0.0, 0.0, 1.0, 1.0, 0.0, 0.0,
				0.0, 0.0, 1.0, 1.0, 0.0, 0.0,
				0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
			},
		},
	}

	const (
		width          = 6
		height         = 6
		delta  float64 = 1e-10
	)

	var (
		img        = image.NewRGBA(image.Rect(0, 0, width, height))
		imgFloat64 = []float64{
			0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
			0.0, 0.0, 0.0, 1.0, 1.0, 0.0,
			0.0, 0.0, 0.0, 1.0, 1.0, 0.0,
			0.0, 0.0, 0.0, 1.0, 1.0, 0.0,
			0.0, 0.0, 0.0, 1.0, 1.0, 0.0,
			0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
		}
	)

	for y := 0; y < height; y += 1 {
		for x := 0; x < width; x += 1 {
			v := uint8(imgFloat64[y*height+x] * 255)
			c := color.RGBA{R: v, G: v, B: v, A: 0xff}
			img.SetRGBA(x, y, c)
		}
	}

	merge := func(r, g, b, _ uint8) float64 {
		return (float64(r) + float64(g) + float64(b)) / (255 * 3)
	}

	for _, c := range cases {
		actual, err := MergeConvolveFloat64(img, merge, c.kernel)
		assert.Nil(t, err)
		assert.NotNil(t, actual)

		for i := 0; i < width*height; i += 1 {
			assert.InDelta(t, c.expected[i], actual[i], delta, "i: %d", i)
		}
	}
}

func TestMergeConvolveFloat64ShouldReturnErrorForInvalidArguments(t *testing.T) {
	cases := []struct {
		img    *image.RGBA
		merge  func(r, g, b, a uint8) float64
		kernel []float64
	}{
		{
			img: nil,
			merge: func(r, g, b, a uint8) float64 {
				return 0
			},
			kernel: make([]float64, 3*3),
		},
		{
			img:    image.NewRGBA(image.Rect(0, 0, 1, 1)),
			merge:  nil,
			kernel: make([]float64, 3*3),
		},
		{
			img: image.NewRGBA(image.Rect(0, 0, 1, 1)),
			merge: func(r, g, b, a uint8) float64 {
				return 0
			},
			kernel: nil,
		},
		{
			img: image.NewRGBA(image.Rect(0, 0, 1, 1)),
			merge: func(r, g, b, a uint8) float64 {
				return 0
			},
			kernel: make([]float64, 3*(3+1)),
		},
		{
			img: image.NewRGBA(image.Rect(0, 0, 1, 1)),
			merge: func(r, g, b, a uint8) float64 {
				return 0
			},
			kernel: make([]float64, 2*2),
		},
		{
			img: image.NewRGBA(image.Rect(0, 0, 1, 1)),
			merge: func(r, g, b, a uint8) float64 {
				return 0
			},
			kernel: make([]float64, 0),
		},
	}

	for _, c := range cases {
		img, err := MergeConvolveFloat64(c.img, c.merge, c.kernel)
		assert.NotNil(t, err)
		assert.Nil(t, img)
	}
}

func TestMergeConvolveRGBAShouldConvolveImage(t *testing.T) {
	cases := []struct {
		kernel   []float64
		expected []uint8
	}{
		{
			kernel: []float64{
				1, 1, 1,
				1, 1, 1,
				1, 1, 1,
			},
			expected: []uint8{
				0, 0, 0, 0, 0, 0,
				0, 0, 56, 113, 113, 0,
				0, 0, 85, 170, 170, 0,
				0, 0, 85, 170, 170, 0,
				0, 0, 57, 113, 113, 0,
				0, 0, 0, 0, 0, 0,
			},
		},
		{
			kernel: []float64{
				0, 0, 0,
				-1, 0, 1,
				0, 0, 0,
			},
			expected: []uint8{
				0, 0, 0, 0, 0, 0,
				0, 0, 255, 255, 0, 0,
				0, 0, 255, 255, 0, 0,
				0, 0, 255, 255, 0, 0,
				0, 0, 255, 255, 0, 0,
				0, 0, 0, 0, 0, 0,
			},
		},
	}

	const (
		width  = 6
		height = 6
	)

	var (
		img        = image.NewRGBA(image.Rect(0, 0, width, height))
		imgFloat64 = []float64{
			0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
			0.0, 0.0, 0.0, 1.0, 1.0, 0.0,
			0.0, 0.0, 0.0, 1.0, 1.0, 0.0,
			0.0, 0.0, 0.0, 1.0, 1.0, 0.0,
			0.0, 0.0, 0.0, 1.0, 1.0, 0.0,
			0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
		}
	)

	for y := 0; y < height; y += 1 {
		for x := 0; x < width; x += 1 {
			v := uint8(imgFloat64[y*height+x] * 255)
			c := color.RGBA{R: v, G: v, B: v, A: 0xff}
			img.SetRGBA(x, y, c)
		}
	}

	merge := func(r, g, b, _ uint8) float64 {
		return (float64(r) + float64(g) + float64(b)) / (255 * 3)
	}

	for _, c := range cases {
		actual, err := MergeConvolveRGBA(img, merge, c.kernel)
		assert.Nil(t, err)
		assert.NotNil(t, actual)

		for i := 0; i < width*height; i += 4 {
			assert.Equal(t, c.expected[i/4], actual.Pix[i], "i: %d", i)
		}
	}
}

func TestMergeConvolveRGBAShouldReturnErrorForInvalidArguments(t *testing.T) {
	cases := []struct {
		img    *image.RGBA
		merge  func(r, g, b, a uint8) float64
		kernel []float64
	}{
		{
			img: nil,
			merge: func(r, g, b, a uint8) float64 {
				return 0
			},
			kernel: make([]float64, 3*3),
		},
		{
			img:    image.NewRGBA(image.Rect(0, 0, 1, 1)),
			merge:  nil,
			kernel: make([]float64, 3*3),
		},
		{
			img: image.NewRGBA(image.Rect(0, 0, 1, 1)),
			merge: func(r, g, b, a uint8) float64 {
				return 0
			},
			kernel: nil,
		},
		{
			img: image.NewRGBA(image.Rect(0, 0, 1, 1)),
			merge: func(r, g, b, a uint8) float64 {
				return 0
			},
			kernel: make([]float64, 3*(3+1)),
		},
		{
			img: image.NewRGBA(image.Rect(0, 0, 1, 1)),
			merge: func(r, g, b, a uint8) float64 {
				return 0
			},
			kernel: make([]float64, 2*2),
		},
		{
			img: image.NewRGBA(image.Rect(0, 0, 1, 1)),
			merge: func(r, g, b, a uint8) float64 {
				return 0
			},
			kernel: make([]float64, 0),
		},
	}

	for _, c := range cases {
		img, err := MergeConvolveRGBA(c.img, c.merge, c.kernel)
		assert.NotNil(t, err)
		assert.Nil(t, img)
	}
}
