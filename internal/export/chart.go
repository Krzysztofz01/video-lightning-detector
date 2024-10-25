package export

import (
	"fmt"
	"path"

	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/statistics"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

type ChartData struct {
	FrameCollection                             frame.FrameCollection
	DescriptiveStatistics                       statistics.DescriptiveStatistics
	Detections                                  []int
	BrightnessDetectionThreshold                float64
	ColorDifferenceDetectionThreshold           float64
	BinaryThresholdDifferenceDetectionThreshold float64
}

func ExportFramesChart(outputDirectoryPath string, fc frame.FrameCollection, ds statistics.DescriptiveStatistics, det []int, brightnessT float64, colorDiffT float64, btDiffT float64) (string, error) {
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
		PageTitle: fmt.Sprintf("Video Lightning Detector [%s]", outputDirectoryPath),
		Width:     "100vw",
		Height:    "90vh",
		Theme:     types.ThemeWesteros,
	})

	titleOpts := charts.WithTitleOpts(opts.Title{
		Title: "Video-Lightning-Detector",
	})

	zoomOpts := charts.WithDataZoomOpts(opts.DataZoom{
		Type: "inside",
	})

	animationOpts := charts.WithAnimation()

	chart := charts.NewScatter()
	chart.SetGlobalOptions(initializationOpts, titleOpts, zoomOpts, animationOpts)

	lineChart := charts.NewLine()
	lineChart.SetGlobalOptions(initializationOpts, titleOpts, zoomOpts, animationOpts)

	frames := fc.GetAll()

	detectionsMap := make(map[int]int, len(det))
	for _, detectionIndex := range det {
		detectionsMap[detectionIndex] = detectionIndex
	}

	var (
		xAxis                     []int              = make([]int, 0, len(frames))
		brightness                []opts.ScatterData = make([]opts.ScatterData, 0, len(frames))
		brightnessMovingMean      []opts.ScatterData = make([]opts.ScatterData, 0, len(frames))
		brightnessThreshold       []opts.LineData    = make([]opts.LineData, 0, len(frames))
		colorDiff                 []opts.ScatterData = make([]opts.ScatterData, 0, len(frames))
		colorDiffMovingMean       []opts.ScatterData = make([]opts.ScatterData, 0, len(frames))
		colorDiffThreshold        []opts.LineData    = make([]opts.LineData, 0, len(frames))
		binaryThreshold           []opts.ScatterData = make([]opts.ScatterData, 0, len(frames))
		binaryThresholdMovingMean []opts.ScatterData = make([]opts.ScatterData, 0, len(frames))
		binaryThresholdThreshold  []opts.LineData    = make([]opts.LineData, 0, len(frames))
	)

	for frameIndex, frame := range frames {
		xAxis = append(xAxis, frameIndex+1)

		var symbol string
		if _, ok := detectionsMap[frameIndex]; ok {
			symbol = "arrow"
		} else {
			symbol = "circle"
		}

		brightness = append(brightness, opts.ScatterData{
			Value:  frame.Brightness,
			Symbol: symbol,
		})

		brightnessMovingMeanValue := ds.BrightnessMovingMean[frameIndex]
		brightnessMovingMean = append(brightnessMovingMean, opts.ScatterData{
			Value: brightnessMovingMeanValue,
		})

		brightnessThreshold = append(brightnessThreshold, opts.LineData{
			Value: brightnessT + brightnessMovingMeanValue,
		})

		colorDiff = append(colorDiff, opts.ScatterData{
			Value:  frame.ColorDifference,
			Symbol: symbol,
		})

		colorDiffMovingMeanValue := ds.ColorDifferenceMovingMean[frameIndex]
		colorDiffMovingMean = append(colorDiffMovingMean, opts.ScatterData{
			Value: colorDiffMovingMeanValue,
		})

		colorDiffThreshold = append(colorDiffThreshold, opts.LineData{
			Value: colorDiffT + colorDiffMovingMeanValue,
		})

		binaryThreshold = append(binaryThreshold, opts.ScatterData{
			Value:  frame.BinaryThresholdDifference,
			Symbol: symbol,
		})

		binaryThresholdMovingMeanValue := ds.BinaryThresholdDifferenceMovingMean[frameIndex]
		binaryThresholdMovingMean = append(binaryThresholdMovingMean, opts.ScatterData{
			Value: binaryThresholdMovingMeanValue,
		})

		binaryThresholdThreshold = append(binaryThresholdThreshold, opts.LineData{
			Value: btDiffT + binaryThresholdMovingMeanValue,
		})
	}

	chart.SetXAxis(xAxis)

	chart.AddSeries("Brightness", brightness, getSeriesOptions("#D4BEE4")...)

	chart.AddSeries("Brightness moving mean", brightnessMovingMean, getSeriesOptions("#9B7EBD")...)

	lineChart.AddSeries("Brightness threshold", brightnessThreshold, getSeriesOptions("#3B1E54")...)

	chart.AddSeries("Color difference", colorDiff, getSeriesOptions("#C4DAD2")...)

	chart.AddSeries("Color difference moving mean", colorDiffMovingMean, getSeriesOptions("#6A9C89")...)

	lineChart.AddSeries("Color difference threshold", colorDiffThreshold, getSeriesOptions("#16423C")...)

	chart.AddSeries("Binary threshold", binaryThreshold, getSeriesOptions("#37B7C3")...)

	chart.AddSeries("Binary threshold moving mean", binaryThresholdMovingMean, getSeriesOptions("#088395")...)

	lineChart.AddSeries("Binary threshold threshold", binaryThresholdThreshold, getSeriesOptions("#071952")...)

	chart.Overlap(lineChart)

	if err := chart.Render(framesChartFile); err != nil {
		return "", fmt.Errorf("export: failed to render the frames chart to the file: %w", err)
	}

	return framesChartPath, nil
}

func getSeriesOptions(color string) []charts.SeriesOpts {
	options := make([]charts.SeriesOpts, 0, 2)

	options = append(options, charts.WithItemStyleOpts(opts.ItemStyle{
		Color: color,
	}))

	options = append(options, charts.WithSeriesAnimation(false))

	return options
}
