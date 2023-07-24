package detector

import (
	"fmt"
	"image"
	"image/png"
	"path"
	"time"

	vidio "github.com/AlexEidt/Vidio"
	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/draw"
)

// Detector instance that is able to perform a search after ligntning strikes on a video file.
type Detector interface {
	Run(inputVideoPath, outputDirectoryPath string) error
}

type detector struct {
	options DetectorOptions
}

// Create a new video lightning detector instance with the specified options.
func CreateDetector(options DetectorOptions) (Detector, error) {
	if ok, msg := options.AreValid(); !ok {
		return nil, fmt.Errorf("detector: invalid options %s", msg)
	}

	logrus.Debugf("Detector created with options: %+v", options)

	return &detector{
		options: options,
	}, nil
}

// Perform a lightning detection on the provided video specified by the file path and store the results at the specified directory path.
func (detector *detector) Run(inputVideoPath, outputDirectoryPath string) error {
	runTime := time.Now()
	logrus.Debugln("Starting the lightning hunt.")

	video, err := vidio.NewVideo(inputVideoPath)
	if err != nil {
		return fmt.Errorf("detector: failed to open the video file for the detection iteration: %w", err)
	}

	targetWidth := int(float64(video.Width()) * detector.options.FrameScalingFactor)
	targetHeight := int(float64(video.Height()) * detector.options.FrameScalingFactor)

	frameCurrentBuffer := image.NewRGBA(image.Rect(0, 0, video.Width(), video.Height()))
	video.SetFrameBuffer(frameCurrentBuffer.Pix)

	frameCurrent := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	framePrevious := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	frameNumber := 1
	frameCount := video.Frames()
	frames := frame.CreateNewFramesCollection(frameCount)

	for video.Read() {
		if detector.options.FrameScalingFactor == 1.0 {
			copy(frameCurrent.Pix, frameCurrentBuffer.Pix)
		} else {
			draw.NearestNeighbor.Scale(frameCurrent, frameCurrent.Rect, frameCurrentBuffer, frameCurrentBuffer.Bounds(), draw.Over, nil)
		}

		frame := frame.CreateNewFrame(frameCurrent, framePrevious, frameNumber)
		frames.Append(frame)

		logrus.Infof("Frame: [%d/%d]. Brightness: %f ColorDiff: %f BTDiff: %f", frameNumber, frameCount, frame.Brightness, frame.ColorDifference, frame.BinaryThresholdDifference)

		frameNumber += 1
		copy(framePrevious.Pix, frameCurrent.Pix)
	}

	// statistics := frames.CalculateStatistics()

	video.Close()
	video, err = vidio.NewVideo(inputVideoPath)
	if err != nil {
		return fmt.Errorf("detector: failed to open the video file for the export iteration: %w", err)
	}

	video.SetFrameBuffer(frameCurrent.Pix)

	if !detector.options.SkipFramesExport {
		frameNumber = 1
		for video.Read() {
			frame, err := frames.Get(frameNumber)
			if err != nil {
				return fmt.Errorf("detector: failed to retrieve a frame from the frames collection: %w", err)
			}

			logrus.Infof("Frame: [%d/%d]. Checking frame thresholds.", frameNumber, frameCount)
			if !frameMeetThresholds(frame, &detector.options) {
				logrus.Infof("Frame: [%d/%d]. Thresholds not meet.", frameNumber, frameCount)
				frameNumber += 1
				continue
			}

			logrus.Infof("Frame: [%d/%d]. Thresholds meet, exporting frame.", frameNumber, frameCount)
			if err := handleFrameExport(outputDirectoryPath, frame, frameCurrentBuffer); err != nil {
				return fmt.Errorf("detector: frame iamge export failed: %w", err)
			}

			frameNumber += 1
		}
	}

	video.Close()

	if detector.options.ExportCsvReport {
		if err := handleCsvReportExport(outputDirectoryPath, frames); err != nil {
			return fmt.Errorf("detector: csv report export failed: %w", err)
		}
	}

	if detector.options.ExportJsonReport {
		if err := handleJsonReportExport(outputDirectoryPath, frames); err != nil {
			return fmt.Errorf("detector: json report export failed: %w", err)
		}
	}

	logrus.Debugf("Lightning hunting took: %s", time.Since(runTime))
	return nil
}

func handleCsvReportExport(outputDirectoryPath string, frames *frame.FramesCollection) error {
	logrus.Info("Exporting the frames report in CSV format.")
	csvReportPath := path.Join(outputDirectoryPath, "report.csv")
	file, err := utils.CreateFileWithTree(csvReportPath)
	if err != nil {
		return fmt.Errorf("detector: failed to create the csv report file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	if err := frames.ExportCsvReport(file); err != nil {
		return fmt.Errorf("detector: failed to export the csv report: %w", err)
	}

	logrus.Infof("Frames report in CSV format exported to: %s", csvReportPath)
	return nil
}

func handleJsonReportExport(outputDirectoryPath string, frames *frame.FramesCollection) error {
	logrus.Info("Exporting the frames report in JSON format.")
	jsonReportPath := path.Join(outputDirectoryPath, "report.json")
	file, err := utils.CreateFileWithTree(jsonReportPath)
	if err != nil {
		return fmt.Errorf("detector: failed to create the json report file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	if err := frames.ExportJsonReport(file); err != nil {
		return fmt.Errorf("detector: failed to export the json report: %w", err)
	}

	logrus.Infof("Frames report in JSON format exported to: %s", jsonReportPath)
	return nil
}

func handleFrameExport(outputDirectoryPath string, frame *frame.Frame, image image.Image) error {
	frameImageName := fmt.Sprintf("frame-%d.png", frame.OrdinalNumber)
	frameImagePath := path.Join(outputDirectoryPath, frameImageName)
	file, err := utils.CreateFileWithTree(frameImagePath)
	if err != nil {
		return fmt.Errorf("detector: failed to create the frame image file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	if err := png.Encode(file, image); err != nil {
		return fmt.Errorf("decoder: failed to export the frame image: %w", err)
	}

	return nil
}

func frameMeetThresholds(frame *frame.Frame, options *DetectorOptions) bool {
	if frame.Brightness < options.BrightnessDetectionThreshold {
		logrus.Debugf("Brightness detection threshold not met. (%f < %f)",
			frame.Brightness,
			options.BrightnessDetectionThreshold)

		return false
	}

	if frame.ColorDifference < options.ColorDifferenceDetectionThreshold {
		logrus.Debugf("Color difference detection threshold not met. (%f < %f)",
			frame.ColorDifference,
			options.ColorDifferenceDetectionThreshold)

		return false
	}

	if frame.BinaryThresholdDifference < options.BinaryThresholdDifferenceDetectionThreshold {
		logrus.Debugf("Binary threshold difference detection threshold not met. (%f < %f)",
			frame.BinaryThresholdDifference,
			options.BinaryThresholdDifferenceDetectionThreshold)

		return false
	}

	return true
}
