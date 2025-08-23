package analyzer

import (
	"context"
	"fmt"
	"image"
	"io"
	"os"
	"path"
	"time"

	"github.com/Krzysztofz01/video-lightning-detector/internal/denoise"
	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/printer"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
	"github.com/Krzysztofz01/video-lightning-detector/internal/video"
)

const frameCollectionCacheFilename string = ".vld-cache"

type Analyzer interface {
	// Perform the analysis of the video frames. Depending on the options, this function will perform
	// the analysis or import the result of the previous analysis with a fallback to a standard analysis.
	// Depending on the options the frames analysis will be exported for future usage.
	GetFrames(ctx context.Context) (frame.FrameCollection, error)
}

type analyzer struct {
	InputVideoPath string
	OutputDirPath  string
	Options        options.DetectorOptions
	Printer        printer.Printer
}

func (analyzer *analyzer) GetFrames(ctx context.Context) (frame.FrameCollection, error) {
	var (
		frames      frame.FrameCollection
		preanalyzed bool
		err         error
	)

	if analyzer.Options.ImportPreanalyzed {
		preanalizedImportTime := time.Now()

		frames, preanalyzed, err = analyzer.ImportPreanalyzedFrames()
		if err != nil {
			return nil, fmt.Errorf("analyzer: failed to import the preanalyzed frames: %w", err)
		}

		if preanalyzed {
			analyzer.Printer.Info("Importing the pre-analyzed frames data. Stage took: %s", time.Since(preanalizedImportTime))
			return frames, nil
		}

		analyzer.Printer.Warning("No exported pre-analzyed frames JSON file found. Fallback to frames analysis.")
	}

	if frames, err = analyzer.PerformFramesAnalysis(ctx); err != nil {
		return nil, fmt.Errorf("analyzer: failed to perform the frames analysis: %w", err)
	}

	if analyzer.Options.ImportPreanalyzed {
		if err := analyzer.ExportPreanalyzedFrames(frames); err != nil {
			return nil, fmt.Errorf("analyzer: preanalyzed frames export stage failed: %w", err)
		}
	}

	return frames, nil
}

// Helper function used to iterate over the video frames in order to generate a collection of frames instances containing
// processed values about given frames and neighbouring frames relations.
func (analyzer *analyzer) PerformFramesAnalysis(ctx context.Context) (frame.FrameCollection, error) {
	videoAnalysisTime := time.Now()
	analyzer.Printer.Debug("Starting the video analysis stage.")

	video, err := video.NewVideo(analyzer.InputVideoPath)
	if err != nil {
		return nil, fmt.Errorf("analyzer: failed to open the video file for the analysis stage: %w", err)
	}

	defer video.Close()

	if err := video.SetScale(analyzer.Options.FrameScalingFactor); err != nil {
		return nil, fmt.Errorf("analyzer: failed to set the video scaling to the given frame scaling factor: %w", err)
	}

	if err := video.SetScaleAlgorithm(analyzer.Options.ScaleAlgorithm); err != nil {
		return nil, fmt.Errorf("analyzer: failed to set the video scaling algorithm for the video: %w", err)
	}

	if len(analyzer.Options.DetectionBoundsExpression) != 0 {
		x, y, w, h, err := utils.ParseBoundsExpression(analyzer.Options.DetectionBoundsExpression)
		if err != nil {
			return nil, fmt.Errorf("analyzer: failed to parse the detection bounds expression: %w", err)
		}

		if err := video.SetBbox(x, y, w, h); err != nil {
			return nil, fmt.Errorf("analyzer: failed to apply the detection bounds to the video: %w", err)
		}
	}

	targetWidth, targetHeight := video.GetOutputDimensions()
	frameCurrent := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	framePrevious := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	if err := video.SetFrameBuffer(frameCurrent.Pix); err != nil {
		return nil, fmt.Errorf("analyzer: failed to apply the given buffer as the video frame buffer: %w", err)
	}

	// NOTE: Due to the fact that the internal video implementation works in such a way, that the frame count is an approximation instead of an
	// exact value, the frameCount is used as an initial capacity value for the frames collection and not the result fixed size/frames count.
	frameNumber := 1
	frameCount := video.FramesCountApprox()
	frames := frame.NewFrameCollection(frameCount)
	defer frames.Lock()

	progressStep, progressFinalize := analyzer.Printer.ProgressSteps("Video analysis stage.", frameCount)

videoRead:
	for {
		select {
		case <-ctx.Done():
			analyzer.Printer.Info("Stopping the video analysis stage")
			break videoRead
		default:
		}

		if err := video.Read(); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("analyzer: failed to read the video frame: %w", err)
		}

		if analyzer.Options.Denoise != options.NoDenoise {
			if err := denoise.Denoise(frameCurrent, frameCurrent, analyzer.Options.Denoise); err != nil {
				return nil, fmt.Errorf("analyzer: failed to apply denoise to the current frame image on the analyze stage: %w", err)
			}
		}

		frame := frame.CreateNewFrame(frameCurrent, framePrevious, frameNumber, frame.BinaryThresholdParam)
		if err := frames.Push(frame); err != nil {
			return nil, fmt.Errorf("analyzer: failed to push the frame to the collection: %w", err)
		}

		analyzer.Printer.Debug("Frame: [%d/%d]. Brightness: %f ColorDiff: %f BTDiff: %f", frameNumber, frameCount, frame.Brightness, frame.ColorDifference, frame.BinaryThresholdDifference)

		frameNumber += 1
		progressStep()

		// TODO: This can be run concurrently together with CreateNewFrame on separeted goroutines but will require a double-buffered framePrevious.
		copy(framePrevious.Pix, frameCurrent.Pix)
	}

	progressFinalize()
	analyzer.Printer.Debug("Video analysis stage finished. Stage took: %s", time.Since(videoAnalysisTime))
	return frames, nil
}

// Helper function used to import the pre-analyzed frames collection from the JSON export file.
func (analyzer *analyzer) ImportPreanalyzedFrames() (frame.FrameCollection, bool, error) {
	frameCollectionCachePath := path.Join(analyzer.OutputDirPath, frameCollectionCacheFilename)
	if !utils.FileExists(frameCollectionCachePath) {
		return nil, false, nil
	}

	frameCollectionCacheFile, err := os.Open(frameCollectionCachePath)
	if err != nil {
		return nil, true, fmt.Errorf("analyzer: failed to open the frame collection cache with preanalyzed frames: %w", err)
	}

	defer func() {
		if err := frameCollectionCacheFile.Close(); err != nil {
			panic(err)
		}
	}()

	optionsChecksum, err := options.CalculateChecksum(analyzer.Options)
	if err != nil {
		return nil, true, fmt.Errorf("analyzer: failed to access the detector options checksum: %w", err)
	}

	if equal, err := frame.ChecksumEqualPeek(frameCollectionCacheFile, optionsChecksum); err != nil {
		return nil, false, fmt.Errorf("analyzer: failed to peek and compare preanalzyed frames checksum: %w", err)
	} else if !equal {
		return nil, false, nil
	}

	if _, err := frameCollectionCacheFile.Seek(0, io.SeekStart); err != nil {
		return nil, false, fmt.Errorf("analyzer: failed to reset the frame collection cache file reading offset: %w", err)
	}

	if frames, _, err := frame.ImportCachedFrameCollection(frameCollectionCacheFile); err != nil {
		return nil, true, fmt.Errorf("analyzer: failed to import the json frames report with preanalyzed frames: %w", err)
	} else {
		return frames, true, nil
	}
}

func (analyzer *analyzer) ExportPreanalyzedFrames(fc frame.FrameCollection) error {
	frameCollectionCachePath := path.Join(analyzer.OutputDirPath, frameCollectionCacheFilename)

	var (
		frameCollectionCacheFile *os.File
		optionsChecksum          string
		err                      error
	)

	if optionsChecksum, err = options.CalculateChecksum(analyzer.Options); err != nil {
		return fmt.Errorf("analyzer: failed to access the options checksum: %w", err)
	}

	defer func() {
		if frameCollectionCacheFile == nil {
			return
		}

		if err := frameCollectionCacheFile.Close(); err != nil {
			panic(err)
		}
	}()

	if utils.FileExists(frameCollectionCachePath) {
		frameCollectionCacheFile, err = os.Open(frameCollectionCachePath)
		if err != nil {
			return fmt.Errorf("analyzer: failed to open the frame collection cache with preanalyzed frames: %w", err)
		}

		if equal, err := frame.ChecksumEqualPeek(frameCollectionCacheFile, optionsChecksum); err != nil {
			return fmt.Errorf("analyzer: failed to peek and compare preanalzyed frames checksum: %w", err)
		} else {
			if equal {
				return nil
			}
		}
	}

	frameCollectionCacheFile, err = utils.CreateFileWithTree(frameCollectionCachePath)
	if err != nil {
		return fmt.Errorf("analyzer: failed to creatae the frame collection cache with preanalyzed frames: %w", err)
	}

	if err := frame.ExportCachedFrameCollection(frameCollectionCacheFile, fc, optionsChecksum); err != nil {
		return fmt.Errorf("analyzer: failed to export the preanalyzed frames cache: %w", err)
	}

	return nil
}

func NewAnalyzer(inputVideo, outputDir string, o options.DetectorOptions, p printer.Printer) Analyzer {
	return &analyzer{
		InputVideoPath: inputVideo,
		OutputDirPath:  outputDir,
		Options:        o,
		Printer:        p,
	}
}
