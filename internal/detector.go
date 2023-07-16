package internal

import (
	"encoding/csv"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"time"

	vidio "github.com/AlexEidt/Vidio"
	"github.com/Krzysztofz01/pimit"
	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
)

type Detector interface {
	Run(inputVideoPath, outputDirectoryPath string) ([]*Frame, error)
}

type detector struct {
	options DetectorOptions
}

func CreateDetector(options DetectorOptions) (Detector, error) {
	if ok, msg := options.AreValid(); !ok {
		return nil, fmt.Errorf("detector: invalid options %s", msg)
	}

	return &detector{
		options: options,
	}, nil
}

func (detector *detector) Run(inputVideoPath, outputDirectoryPath string) ([]*Frame, error) {
	runTime := time.Now()
	logrus.Debugln("Starting the lightning hunt.")

	video, err := vidio.NewVideo(inputVideoPath)
	if err != nil {
		return nil, fmt.Errorf("detector: failed to open the video file for the detection iteration: %w", err)
	}

	framePrevious := image.NewRGBA(image.Rect(0, 0, video.Width(), video.Height()))
	frameCurrent := image.NewRGBA(image.Rect(0, 0, video.Width(), video.Height()))

	frames := make([]*Frame, 0, video.Frames())

	video.SetFrameBuffer(frameCurrent.Pix)

	frameNumber := 1
	frameCount := video.Frames()
	frameSize := float64(video.Width() * video.Height())

	for video.Read() {
		wg := sync.WaitGroup{}
		wg.Add(2)

		frame := NewFrame(frameNumber)

		go func() {
			defer wg.Done()

			brightness := atomic.NewFloat64(0)
			pimit.ParallelColumnColorRead(frameCurrent, func(color color.Color) {
				brightness.Add(GetGrayscaleBasedBrightness(color))
			})

			frame.SetBrightness(brightness.Load() / frameSize)
		}()

		go func() {
			defer wg.Done()
			if frameNumber == 1 {
				frame.SetDifference(0)
				return
			}

			difference := atomic.NewFloat64(0)
			pimit.ParallelColumnRead(frameCurrent, func(xIndex, yIndex int, color color.Color) {
				colorPrevious := framePrevious.At(xIndex, yIndex)

				difference.Add(GetColorDifference(color, colorPrevious))
			})

			frame.SetDifference(difference.Load() / frameSize)
		}()

		wg.Wait()

		frames = append(frames, frame)

		logrus.Infof("Frame: [%d/%d]. Difference: %f  Brightness: %f", frameNumber, frameCount, frame.GetDifference(), frame.GetBrightness())

		frameNumber += 1
		copy(framePrevious.Pix, frameCurrent.Pix)
	}

	video, err = vidio.NewVideo(inputVideoPath)
	if err != nil {
		return nil, fmt.Errorf("detector: failed to open the video file for the export iteration: %w", err)
	}

	video.SetFrameBuffer(frameCurrent.Pix)

	if !detector.options.SkipFramesExport {
		for _, frame := range frames {
			if !video.Read() {
				return nil, errors.New("detector: video and detection frames frame count missmatch")
			}

			if err := detector.handleExportFrame(frame, frameCurrent, outputDirectoryPath); err != nil {
				return nil, fmt.Errorf("detector: failed to handle the frame export process: %w", err)
			}
		}
	}

	if !detector.options.SkipReportExport {
		if err := detector.exportReport(frames, outputDirectoryPath); err != nil {
			return nil, fmt.Errorf("detector: failed to generate frames report: %w", err)
		}
	}

	logrus.Debugf("Lightning hunting took: %s", time.Since(runTime))
	return frames, nil
}

func (detector *detector) handleExportFrame(frame *Frame, frameImage *image.RGBA, outputDirectoryPath string) error {
	if frame.GetBrightness() < detector.options.FrameBrightnessThreshold {
		logrus.Debugln("Skipping frame export. Brightness is not matching the threshold.")
		return nil
	}

	if frame.GetDifference() < detector.options.FrameDifferenceThreshold {
		logrus.Debugln("Skipping frame export. Difference is not matching the threshold.")
		return nil
	}

	logrus.Debugln("Performing frame export.")

	exportFramePath := filepath.Join(outputDirectoryPath, fmt.Sprintf("frame-%d.png", frame.OrdinalNumber))
	imageFile, err := os.Create(exportFramePath)
	if err != nil {
		return fmt.Errorf("detector: failed to create the export frame image file: %w", err)
	}

	if err := png.Encode(imageFile, frameImage); err != nil {
		return fmt.Errorf("detector: failed to encode the export frame image: %w", err)
	}

	if err := imageFile.Close(); err != nil {
		return fmt.Errorf("detector: failed to close the export frame image file: %w", err)
	}

	logrus.Debugf("Frame exported. Location: %s", exportFramePath)
	return nil
}

func (detector *detector) exportReport(frames []*Frame, outputDirectoryPath string) error {
	exportCsvPath := filepath.Join(outputDirectoryPath, "frames.csv")
	logrus.Debugln("Generating the frames report.")

	csvFile, err := CreateFileWithTree(exportCsvPath)
	if err != nil {
		return fmt.Errorf("detector: failed to create the frames report file: %w", err)
	}

	csvWriter := csv.NewWriter(csvFile)
	if err := csvWriter.Write([]string{"Frame", "Brightness", "Difference"}); err != nil {
		return fmt.Errorf("detector: failed to write the header to the frames report file: %w", err)
	}

	for _, frame := range frames {
		if err := csvWriter.Write(frame.ToBuffer()); err != nil {
			return fmt.Errorf("detector: failed to write the frame to the frames report file: %w", err)
		}
	}

	csvWriter.Flush()
	if err := csvFile.Close(); err != nil {
		return fmt.Errorf("detector: failed to close the frames report file: %w", err)
	}

	logrus.Debugf("Frames report generated. Location: %s", exportCsvPath)
	return nil
}

type DetectorOptions struct {
	FrameDifferenceThreshold float64
	FrameBrightnessThreshold float64
	SkipFramesExport         bool
	SkipReportExport         bool
}

func (options *DetectorOptions) AreValid() (bool, string) {
	if options.FrameBrightnessThreshold < 0.0 || options.FrameBrightnessThreshold > 1.0 {
		return false, "the frame brightness threshold must be between zero and one"
	}

	if options.FrameDifferenceThreshold < 0.0 || options.FrameDifferenceThreshold > 1.0 {
		return false, "the frame difference threshold must be between zero and one"
	}

	return true, ""
}

func GetDefaultDetectorOptions() DetectorOptions {
	return DetectorOptions{
		FrameDifferenceThreshold: 0,
		FrameBrightnessThreshold: 0,
		SkipFramesExport:         false,
		SkipReportExport:         false,
	}
}
