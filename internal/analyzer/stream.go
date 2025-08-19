package analyzer

import (
	"errors"
	"fmt"
	"image"
	"io"
	"time"

	"github.com/Krzysztofz01/video-lightning-detector/internal/denoise"
	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/printer"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
	"github.com/Krzysztofz01/video-lightning-detector/internal/video"
)

type timedFrame struct {
	Frame     *frame.Frame
	Timestamp time.Time
}

type StreamAnalyzer interface {
	// Initialize the internal video stream and read the next frames. Each call will return the next
	// frame. If the stream is finished it will return io.EOF.
	Next() error

	// Access a latest frame and its timestamp specified by the FIFO index
	PeekFrame(index int) (*frame.Frame, time.Time, error)

	// Access a latest frame image specified by the FIFO index
	PeekFrameImage(index int) (*image.RGBA, error)

	// Close the internal video stream and cleanup internals
	Close() error
}

type streamAnalyzer struct {
	StreamUrl         string
	Options           options.StreamDetectorOptions
	Printer           printer.Printer
	FrameBuffer       utils.CircularBuffer[*timedFrame]
	FrameImageBuffer  utils.CircularBuffer[*image.RGBA]
	FrameImageCurrent *image.RGBA
	VideoStream       video.VideoStream
	FrameNumber       int
	IsInitialized     bool
}

func (analyzer *streamAnalyzer) PeekFrame(index int) (*frame.Frame, time.Time, error) {
	if f, err := analyzer.FrameBuffer.GetHead(index); err != nil {
		return nil, time.Time{}, err
	} else {
		return f.Frame, f.Timestamp, nil
	}
}

func (analyzer *streamAnalyzer) PeekFrameImage(index int) (*image.RGBA, error) {
	return analyzer.FrameImageBuffer.GetHead(index)
}

func (analyzer *streamAnalyzer) Initialize() error {
	video, err := video.NewVideoStream(analyzer.StreamUrl)
	if err != nil {
		return fmt.Errorf("analyzer: failed to open the video stream for the analysis stage: %w", err)
	}

	if err = video.SetScale(analyzer.Options.FrameScalingFactor); err != nil {
		return fmt.Errorf("analyzer: failed to set the video scaling to the given frame scaling factor: %w", err)
	}

	if err = video.SetScaleAlgorithm(analyzer.Options.ScaleAlgorithm); err != nil {
		return fmt.Errorf("analyzer: failed to set the video scaling algorithm for the video: %w", err)
	}

	if len(analyzer.Options.DetectionBoundsExpression) != 0 {
		x, y, w, h, err := utils.ParseBoundsExpression(analyzer.Options.DetectionBoundsExpression)
		if err != nil {
			return fmt.Errorf("analyzer: failed to parse the detection bounds expression: %w", err)
		}

		if err = video.SetBbox(x, y, w, h); err != nil {
			return fmt.Errorf("analyzer: failed to apply the detection bounds to the video: %w", err)
		}
	}

	targetWidth, targetHeight := video.GetOutputDimensions()
	frameCurrent := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	if err = video.SetFrameBuffer(frameCurrent.Pix); err != nil {
		return fmt.Errorf("analyzer: failed to apply the given buffer as the video frame buffer: %w", err)
	}

	capacity := utils.MaxInt(4, int(analyzer.Options.MovingMeanResolution))
	frameImageBufferAlloc := make([]*image.RGBA, capacity, capacity)
	for index := range frameImageBufferAlloc {
		frameImageBufferAlloc[index] = image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	}

	analyzer.FrameBuffer = utils.NewCircularBuffer[*timedFrame](capacity)
	analyzer.FrameImageBuffer = utils.NewSaturatedCircularBuffer[*image.RGBA](frameImageBufferAlloc)
	analyzer.FrameImageCurrent = frameCurrent
	analyzer.VideoStream = video
	analyzer.FrameNumber = 1
	analyzer.IsInitialized = true

	return nil
}

func (analyzer *streamAnalyzer) Next() error {
	if !analyzer.IsInitialized {
		if err := analyzer.Initialize(); err != nil {
			return fmt.Errorf("analyzer: failed to initialize the analyzer video stream: %w", err)
		}
	}

	// FIXME: The timestamp is dependent on the frame 'receive' and not 'creation' time which makes the process latency sensitive
	timestamp := time.Now().UTC()
	if err := analyzer.VideoStream.Read(); err != nil {
		if errors.Is(err, io.EOF) {
			return io.EOF
		} else {
			return fmt.Errorf("analyzer: failed to access the the next frame for analysis: %w", err)
		}
	}

	if analyzer.Options.Denoise != options.NoDenoise {
		if err := denoise.Denoise(analyzer.FrameImageCurrent, analyzer.FrameImageCurrent, analyzer.Options.Denoise); err != nil {
			return fmt.Errorf("analyzer: failed to apply denoise to the current frame image on the analyze stage: %w", err)
		}
	}

	var (
		frameImagePrevious *image.RGBA = nil
		err                error
	)

	if analyzer.FrameNumber > 1 {
		if frameImagePrevious, err = analyzer.FrameImageBuffer.GetHead(0); err != nil {
			return fmt.Errorf("analyzer: failed to access the previous frame image pointer from the buffer: %w", err)
		}
	}

	f := &timedFrame{
		Frame:     frame.CreateNewFrame(analyzer.FrameImageCurrent, frameImagePrevious, analyzer.FrameNumber, frame.BinaryThresholdParam),
		Timestamp: timestamp,
	}

	analyzer.FrameNumber += 1

	analyzer.FrameBuffer.Push(f)
	copy(analyzer.FrameImageBuffer.PushP().Pix, analyzer.FrameImageCurrent.Pix)

	return nil
}

func (analyzer *streamAnalyzer) Close() error {
	if analyzer.VideoStream != nil {
		analyzer.VideoStream.Close()
		analyzer.VideoStream = nil
	}

	analyzer.FrameBuffer = nil
	analyzer.FrameImageBuffer = nil
	analyzer.FrameImageCurrent = nil
	analyzer.FrameNumber = 1
	analyzer.IsInitialized = false

	return nil
}

func NewStreamAnalyzer(inputVideoStream string, o options.StreamDetectorOptions, p printer.Printer) StreamAnalyzer {
	return &streamAnalyzer{
		StreamUrl:         inputVideoStream,
		Options:           o,
		Printer:           p,
		FrameBuffer:       nil,
		FrameImageBuffer:  nil,
		FrameImageCurrent: nil,
		VideoStream:       nil,
		FrameNumber:       1,
		IsInitialized:     false,
	}
}
