package export

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/statistics"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

const (
	JsonReportFileName string = "report.json"
)

// FIXME: Create a separate export wrapper to avoid passing a huge amount of args
func exportJsonReport(outputDirectoryPath string, options options.DetectorOptions, fc frame.FrameCollection, ds statistics.DescriptiveStatistics, cm statistics.ConfusionMatrix) (string, error) {
	frames := make([]jsonFrameReport, 0, fc.Count())
	for index, f := range fc.GetAll() {
		frames = append(frames, jsonFrameReport{
			OrdinalNumber:                         f.OrdinalNumber,
			ColorDifference:                       f.ColorDifference,
			BinaryThresholdDifference:             f.BinaryThresholdDifference,
			Brightness:                            f.Brightness,
			BrightnessMovingMean:                  ds.BrightnessMovingMean[index],
			BrightnessMovingStdDev:                ds.BrightnessMovingStdDev[index],
			ColorDifferenceMovingMean:             ds.ColorDifferenceMovingMean[index],
			ColorDifferenceMovingStdDev:           ds.ColorDifferenceMovingStdDev[index],
			BinaryThresholdDifferenceMovingMean:   ds.BinaryThresholdDifferenceMovingMean[index],
			BinaryThresholdDifferenceMovingStdDev: ds.BinaryThresholdDifferenceMovingStdDev[index],
		})
	}

	var confusionMatrix *jsonConfusionMatrixReport = nil
	if (cm != statistics.ConfusionMatrix{}) {
		confusionMatrix = &jsonConfusionMatrixReport{
			Tp:  cm.Tp,
			Tn:  cm.Tn,
			Fp:  cm.Fp,
			Fn:  cm.Fn,
			P:   cm.P,
			N:   cm.N,
			Tpr: cm.Tpr,
			Tnr: cm.Tnr,
			Acc: cm.Acc,
			Ppv: cm.Ppv,
			Npv: cm.Npv,
			Fpr: cm.Fpr,
			Fnr: cm.Fnr,
			Mcc: cm.Mcc,
			Fs:  cm.Fs,
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
			BrightnessMean:                             ds.BrightnessMean,
			BrightnessStandardDeviation:                ds.BrightnessStandardDeviation,
			BrightnessMin:                              ds.BrightnessMin,
			BrightnessMax:                              ds.BrightnessMax,
			ColorDifferenceMean:                        ds.ColorDifferenceMean,
			ColorDifferenceStandardDeviation:           ds.ColorDifferenceStandardDeviation,
			ColorDifferenceMin:                         ds.ColorDifferenceMin,
			ColorDifferenceMax:                         ds.ColorDifferenceMax,
			BinaryThresholdDifferenceMean:              ds.BinaryThresholdDifferenceMean,
			BinaryThresholdDifferenceStandardDeviation: ds.BinaryThresholdDifferenceStandardDeviation,
			BinaryThresholdDifferenceMin:               ds.BinaryThresholdDifferenceMin,
			BinaryThresholdDifferenceMax:               ds.BinaryThresholdDifferenceMax,
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

// FIXME: Add the expected and actual detections
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
