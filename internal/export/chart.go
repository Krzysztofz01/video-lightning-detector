package export

import (
	"fmt"
	"path"

	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

func ExportFramesChart(outputDirectoryPath string, fc *frame.FramesCollection) (string, error) {
	framesChartPath := path.Join(outputDirectoryPath, FramesChartFilename)
	framesChartFile, err := utils.CreateFileWithTree(framesChartPath)
	if err != nil {
		return "", fmt.Errorf("export: failed to create the html frames chart file: %w", err)
	}

	defer func() {
		if err := framesChartFile.Close(); err != nil {
			panic(err)
		}
	}()

	initializationOpts := charts.WithInitializationOpts(opts.Initialization{
		Theme: types.ThemeWesteros,
	})

	titleOpts := charts.WithTitleOpts(opts.Title{
		Title: "Video-Lightning-Detector",
	})

	chart := charts.NewScatter()
	chart.SetGlobalOptions(initializationOpts, titleOpts)

	frames := fc.GetAll()

	var (
		xAxis           []int              = make([]int, 0, len(frames))
		brightness      []opts.ScatterData = make([]opts.ScatterData, 0, len(frames))
		colorDiff       []opts.ScatterData = make([]opts.ScatterData, 0, len(frames))
		binaryThreshold []opts.ScatterData = make([]opts.ScatterData, 0, len(frames))
	)

	for frameIndex, frame := range frames {
		xAxis = append(xAxis, frameIndex+1)

		brightness = append(brightness, opts.ScatterData{
			Value: frame.Brightness,
		})

		colorDiff = append(colorDiff, opts.ScatterData{
			Value: frame.ColorDifference,
		})

		binaryThreshold = append(binaryThreshold, opts.ScatterData{
			Value: frame.BinaryThresholdDifference,
		})
	}

	chart.SetXAxis(xAxis)
	chart.AddSeries("Brightness", brightness)
	chart.AddSeries("Color difference", colorDiff)
	chart.AddSeries("Binary threshold", binaryThreshold)

	if err := chart.Render(framesChartFile); err != nil {
		return "", fmt.Errorf("export: failed to render the frames chart to the file: %w", err)
	}

	return framesChartPath, nil
}
