package export

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

const (
	JsonReportFileName string = "report.json"
)

func exportJsonReport(outputDirectoryPath string, options options.DetectorOptions, args *exporterArguments) (string, error) {
	// NOTE: The lookups are 1-indexed like the frame ordinal numbers, the provided detection values are 0-indexed
	var (
		expectedDetectionsLookup = make(map[int]bool)
		actualDetectionsLookup   = make(map[int]bool)
	)

	for _, detectionIndex := range args.ExpectedDetections {
		expectedDetectionsLookup[detectionIndex+1] = true
	}

	for _, detectionIndex := range args.ActualDetections {
		actualDetectionsLookup[detectionIndex+1] = true
	}

	frames := make([]jsonFrameReport, 0, args.FrameCollection.Count())
	for index, f := range args.FrameCollection.GetAll() {
		_, expectedDetection := expectedDetectionsLookup[f.OrdinalNumber]
		_, actualDetection := actualDetectionsLookup[f.OrdinalNumber]

		frames = append(frames, jsonFrameReport{
			OrdinalNumber:                         f.OrdinalNumber,
			ColorDifference:                       f.ColorDifference,
			BinaryThresholdDifference:             f.BinaryThresholdDifference,
			Brightness:                            f.Brightness,
			BrightnessMovingMean:                  args.Statistics.BrightnessMovingMean[index],
			BrightnessMovingStdDev:                args.Statistics.BrightnessMovingStdDev[index],
			ColorDifferenceMovingMean:             args.Statistics.ColorDifferenceMovingMean[index],
			ColorDifferenceMovingStdDev:           args.Statistics.ColorDifferenceMovingStdDev[index],
			BinaryThresholdDifferenceMovingMean:   args.Statistics.BinaryThresholdDifferenceMovingMean[index],
			BinaryThresholdDifferenceMovingStdDev: args.Statistics.BinaryThresholdDifferenceMovingStdDev[index],
			ExpectedDetection:                     expectedDetection,
			ActualDetection:                       actualDetection,
		})
	}

	var confusionMatrix *jsonConfusionMatrixReport = nil
	if args.HasConfusionMatrix {
		confusionMatrix = &jsonConfusionMatrixReport{
			Tp:  args.ConfusionMatrix.Tp,
			Tn:  args.ConfusionMatrix.Tn,
			Fp:  args.ConfusionMatrix.Fp,
			Fn:  args.ConfusionMatrix.Fn,
			P:   args.ConfusionMatrix.P,
			N:   args.ConfusionMatrix.N,
			Tpr: args.ConfusionMatrix.Tpr,
			Tnr: args.ConfusionMatrix.Tnr,
			Acc: args.ConfusionMatrix.Acc,
			Ppv: args.ConfusionMatrix.Ppv,
			Npv: args.ConfusionMatrix.Npv,
			Fpr: args.ConfusionMatrix.Fpr,
			Fnr: args.ConfusionMatrix.Fnr,
			Mcc: args.ConfusionMatrix.Mcc,
			Fs:  args.ConfusionMatrix.Fs,
		}
	}

	report := jsonReport{
		Options: jsonOptionsReport{
			AutoThresholds:                              options.AutoThresholds,
			BrightnessDetectionThreshold:                options.BrightnessDetectionThreshold,
			ColorDifferenceDetectionThreshold:           options.ColorDifferenceDetectionThreshold,
			BinaryThresholdDifferenceDetectionThreshold: options.BinaryThresholdDifferenceDetectionThreshold,
			MovingMeanResolution:                        options.MovingMeanResolution,
			Denoise:                                     options.Denoise.String(),
			FrameScalingFactor:                          options.FrameScalingFactor,
			ScaleAlgorithm:                              options.ScaleAlgorithm.String(),
		},
		Frames: frames,
		DescriptiveStatistics: jsonDescriptiveStatisticsReport{
			BrightnessMean:                             args.Statistics.BrightnessMean,
			BrightnessStandardDeviation:                args.Statistics.BrightnessStandardDeviation,
			BrightnessMin:                              args.Statistics.BrightnessMin,
			BrightnessMax:                              args.Statistics.BrightnessMax,
			ColorDifferenceMean:                        args.Statistics.ColorDifferenceMean,
			ColorDifferenceStandardDeviation:           args.Statistics.ColorDifferenceStandardDeviation,
			ColorDifferenceMin:                         args.Statistics.ColorDifferenceMin,
			ColorDifferenceMax:                         args.Statistics.ColorDifferenceMax,
			BinaryThresholdDifferenceMean:              args.Statistics.BinaryThresholdDifferenceMean,
			BinaryThresholdDifferenceStandardDeviation: args.Statistics.BinaryThresholdDifferenceStandardDeviation,
			BinaryThresholdDifferenceMin:               args.Statistics.BinaryThresholdDifferenceMin,
			BinaryThresholdDifferenceMax:               args.Statistics.BinaryThresholdDifferenceMax,
		},
		ConfusionMatrix: confusionMatrix,
	}

	reportFilePath := path.Join(outputDirectoryPath, JsonReportFileName)
	reportFile, err := utils.CreateFileWithTree(reportFilePath)
	if err != nil {
		return "", fmt.Errorf("export: failed to create the json report file: %w", err)
	}

	defer func() {
		if err := reportFile.Close(); err != nil {
			panic(err)
		}
	}()

	encoder := json.NewEncoder(reportFile)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(report); err != nil {
		return "", fmt.Errorf("export: failed to encode the report data: %w", err)
	}

	return reportFilePath, nil
}

type jsonReport struct {
	Options               jsonOptionsReport               `json:"options"`
	Frames                []jsonFrameReport               `json:"frames"`
	DescriptiveStatistics jsonDescriptiveStatisticsReport `json:"descriptive-statistics"`
	ConfusionMatrix       *jsonConfusionMatrixReport      `json:"confusion-matrix"`
}

type jsonOptionsReport struct {
	AutoThresholds                              bool    `json:"auto-thresholds"`
	BrightnessDetectionThreshold                float64 `json:"brightness-detection-threshold"`
	ColorDifferenceDetectionThreshold           float64 `json:"color-difference-detection-threshold"`
	BinaryThresholdDifferenceDetectionThreshold float64 `json:"binary-threshold-difference-detection-threshold"`
	MovingMeanResolution                        int32   `json:"moving-mean-resolution"`
	Denoise                                     string  `json:"denoise"`
	FrameScalingFactor                          float64 `json:"frame-scaling-factor"`
	ScaleAlgorithm                              string  `json:"scale-algorithm"`
}

type jsonFrameReport struct {
	OrdinalNumber                         int     `json:"ordinal-number"`
	ColorDifference                       float64 `json:"color-difference"`
	BinaryThresholdDifference             float64 `json:"binary-threshold-difference"`
	Brightness                            float64 `json:"brightness"`
	BrightnessMovingMean                  float64 `json:"brightness-moving-mean"`
	BrightnessMovingStdDev                float64 `json:"brightness-moving-standard-deviation"`
	ColorDifferenceMovingMean             float64 `json:"color-difference-moving-mean"`
	ColorDifferenceMovingStdDev           float64 `json:"color-difference-moving-standard-deviation"`
	BinaryThresholdDifferenceMovingMean   float64 `json:"binary-threshold-difference-moving-mean"`
	BinaryThresholdDifferenceMovingStdDev float64 `json:"binary-threshold-difference-moving-standard-deviation"`
	ExpectedDetection                     bool    `json:"expected-detection"`
	ActualDetection                       bool    `json:"actual-detection"`
}

type jsonDescriptiveStatisticsReport struct {
	BrightnessMean                             float64 `json:"brightness-mean"`
	BrightnessStandardDeviation                float64 `json:"brightness-standard-deviation"`
	BrightnessMin                              float64 `json:"brightness-min"`
	BrightnessMax                              float64 `json:"brightness-max"`
	ColorDifferenceMean                        float64 `json:"color-difference-mean"`
	ColorDifferenceStandardDeviation           float64 `json:"color-difference-standard-deviation"`
	ColorDifferenceMin                         float64 `json:"color-difference-min"`
	ColorDifferenceMax                         float64 `json:"color-difference-max"`
	BinaryThresholdDifferenceMean              float64 `json:"binary-threshold-difference-mean"`
	BinaryThresholdDifferenceStandardDeviation float64 `json:"binary-threshold-difference-standard-deviation"`
	BinaryThresholdDifferenceMin               float64 `json:"binary-threshold-difference-min"`
	BinaryThresholdDifferenceMax               float64 `json:"binary-threshold-difference-max"`
}

type jsonConfusionMatrixReport struct {
	Tp  float64 `json:"tp"`
	Tn  float64 `json:"tn"`
	Fp  float64 `json:"fp"`
	Fn  float64 `json:"fn"`
	P   float64 `json:"p"`
	N   float64 `json:"n"`
	Tpr float64 `json:"tpr"`
	Tnr float64 `json:"tnr"`
	Acc float64 `json:"acc"`
	Ppv float64 `json:"ppv"`
	Npv float64 `json:"npv"`
	Fpr float64 `json:"fpr"`
	Fnr float64 `json:"fnr"`
	Mcc float64 `json:"mcc"`
	Fs  float64 `json:"fs"`
}
