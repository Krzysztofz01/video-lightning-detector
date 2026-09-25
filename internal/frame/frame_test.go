package frame

import (
	"image"
	"image/color"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

func TestFactoryShouldCreateNewFirstFrame(t *testing.T) {
	defer goleak.VerifyNone(t)

	var (
		a *image.RGBA = mockImage(color.White)
		b *image.RGBA = nil
	)

	factory := CreateFrameFactory(BinaryThresholdParam)

	frame, err := factory.CreateNewFrame(a, b)

	assert.Nil(t, err)
	assert.NotNil(t, frame)
	assert.Equal(t, 1.0, frame.Brightness)
	assert.Equal(t, 0.0, frame.ColorDifference)
	assert.Equal(t, 0.0, frame.BinaryThresholdDifference)
}

func TestFactoryShouldNotCreateNewFrameForInvalidImageReferences(t *testing.T) {
	defer goleak.VerifyNone(t)

	factory := CreateFrameFactory(BinaryThresholdParam)
	img := mockImage(color.White)

	_, err := factory.CreateNewFrame(nil, nil)
	assert.NotNil(t, err)

	_, err = factory.CreateNewFrame(nil, img)
	assert.NotNil(t, err)

	_, err = factory.CreateNewFrame(img, nil)
	assert.Nil(t, err)

	_, err = factory.CreateNewFrame(nil, nil)
	assert.NotNil(t, err)

	_, err = factory.CreateNewFrame(img, nil)
	assert.NotNil(t, err)

	_, err = factory.CreateNewFrame(nil, img)
	assert.NotNil(t, err)
}

func TestShouldCreateNewFrameWithDifferentNeighbour(t *testing.T) {
	defer goleak.VerifyNone(t)

	var (
		a *image.RGBA = mockImage(color.White)
		b *image.RGBA = mockImage(color.Black)
	)

	factory := CreateFrameFactory(BinaryThresholdParam)

	frame, err := factory.CreateNewFrame(b, nil)
	assert.Nil(t, err)
	assert.NotNil(t, frame)

	frame, err = factory.CreateNewFrame(a, b)

	assert.Nil(t, err)
	assert.NotNil(t, frame)
	assert.Equal(t, 1.0, frame.Brightness)
	assert.Equal(t, 1.0, frame.ColorDifference)
	assert.Equal(t, 1.0, frame.BinaryThresholdDifference)
}

func TestShouldCreateNewFrameWithIdenticalNeighbour(t *testing.T) {
	defer goleak.VerifyNone(t)

	var (
		a *image.RGBA = mockImage(color.White)
		b *image.RGBA = mockImage(color.White)
	)

	factory := CreateFrameFactory(BinaryThresholdParam)

	frame, err := factory.CreateNewFrame(b, nil)
	assert.Nil(t, err)
	assert.NotNil(t, frame)

	frame, err = factory.CreateNewFrame(a, b)

	assert.Nil(t, err)
	assert.NotNil(t, frame)
	assert.Equal(t, 1.0, frame.Brightness)
	assert.Equal(t, 0.0, frame.ColorDifference)
	assert.Equal(t, 0.0, frame.BinaryThresholdDifference)
}

func TestShouldCreateAndCalculateCorrectValuesForWeightsForFirstAndNthFrame(t *testing.T) {
	defer goleak.VerifyNone(t)

	cases := []struct {
		h int
		w int
	}{
		{1, 1},
		{1, 2},
		{2, 1},
		{2, 2},
		{3, 3},
		{3, 2},
		{10, 10},
		{11, 11},
		{11, 20},
		{20, 20},
		{25, 25},
		{26, 26},
		{1, 25},
		{1, 26},
		{26, 1},
		{33, 33},
	}

	rng := rand.New(rand.NewSource(seed))

	for _, c := range cases {
		var (
			img1                *image.RGBA = image.NewRGBA(image.Rect(0, 0, c.w, c.h))
			expectedBrightness1 float64     = 0
			expectedColorDiff1  float64     = 0
			expectedBtDiff1     float64     = 0
		)

		var (
			img2                *image.RGBA = image.NewRGBA(image.Rect(0, 0, c.w, c.h))
			expectedBrightness2 float64     = 0
			expectedColorDiff2  float64     = 0
			expectedBtDiff2     float64     = 0
		)

		for y := 0; y < c.h; y += 1 {
			for x := 0; x < c.w; x += 1 {
				c1 := color.RGBA{
					R: uint8(rng.Intn(256)),
					G: uint8(rng.Intn(256)),
					B: uint8(rng.Intn(256)),
					A: 0xff,
				}

				c2 := color.RGBA{
					R: uint8(rng.Intn(256)),
					G: uint8(rng.Intn(256)),
					B: uint8(rng.Intn(256)),
					A: 0xff,
				}

				expectedBrightness1 += utils.GetColorBrightness(c1.R, c1.G, c1.B)

				expectedBrightness2 += utils.GetColorBrightness(c2.R, c2.G, c2.B)
				expectedColorDiff2 += utils.GetColorDifference(c1.R, c1.G, c1.B, c2.R, c2.G, c2.B)

				btDiff1 := utils.BinaryThreshold(c1.R, c1.G, c1.B, BinaryThresholdParam)
				btDiff2 := utils.BinaryThreshold(c2.R, c2.G, c2.B, BinaryThresholdParam)
				if btDiff1 != btDiff2 {
					expectedBtDiff2 += 1
				}

				img1.SetRGBA(x, y, c1)
				img2.SetRGBA(x, y, c2)
			}
		}

		count := float64(c.w * c.h)

		expectedBrightness1 /= count
		expectedColorDiff1 /= count
		expectedBtDiff1 /= count

		expectedBrightness2 /= count
		expectedColorDiff2 /= count
		expectedBtDiff2 /= count

		factory := CreateFrameFactory(BinaryThresholdParam)

		frame1, err1 := factory.CreateNewFrame(img1, nil)
		frame2, err2 := factory.CreateNewFrame(img2, img1)

		assert.Nil(t, err1)
		assert.NotNil(t, frame1)
		assert.Equal(t, 1, frame1.OrdinalNumber)
		assert.InDelta(t, expectedBrightness1, frame1.Brightness, delta)
		assert.InDelta(t, expectedColorDiff1, frame1.ColorDifference, delta)
		assert.InDelta(t, expectedBtDiff1, frame1.BinaryThresholdDifference, delta)

		assert.Nil(t, err2)
		assert.NotNil(t, frame2)
		assert.Equal(t, 2, frame2.OrdinalNumber)
		assert.InDelta(t, expectedBrightness2, frame2.Brightness, delta)
		assert.InDelta(t, expectedColorDiff2, frame2.ColorDifference, delta)
		assert.InDelta(t, expectedBtDiff2, frame2.BinaryThresholdDifference, delta)
	}
}

const (
	seed  int64   = 0xbeef
	delta float64 = 1e-14
)

func mockImage(c color.Color) *image.RGBA {
	width := 4
	height := 4

	image := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x += 1 {
		for y := 0; y < height; y += 1 {
			image.Set(x, y, c)
		}
	}

	return image
}
