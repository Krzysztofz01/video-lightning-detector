package export

import (
	"fmt"
	"strconv"

	"github.com/Krzysztofz01/video-lightning-detector/internal/printer"
	"github.com/Krzysztofz01/video-lightning-detector/internal/statistics"
)

func RenderDescriptiveStatistics(p printer.Printer, ds statistics.DescriptiveStatistics) error {
	p.Table([][]string{
		{"Frame brightness mean", strconv.FormatFloat(ds.BrightnessMean, 'f', -1, 64)},
		{"Frame brightness standard deviation", strconv.FormatFloat(ds.BrightnessStandardDeviation, 'f', -1, 64)},
		{"Frame brightness max", strconv.FormatFloat(ds.BrightnessMax, 'f', -1, 64)},
		{"Frame color difference mean", strconv.FormatFloat(ds.ColorDifferenceMean, 'f', -1, 64)},
		{"Frame color difference standard deviation", strconv.FormatFloat(ds.ColorDifferenceStandardDeviation, 'f', -1, 64)},
		{"Frame color difference max", strconv.FormatFloat(ds.ColorDifferenceMax, 'f', -1, 64)},
		{"Frame color binary threshold mean", strconv.FormatFloat(ds.BinaryThresholdDifferenceMean, 'f', -1, 64)},
		{"Frame color binary threshold standard deviation", strconv.FormatFloat(ds.BinaryThresholdDifferenceStandardDeviation, 'f', -1, 64)},
		{"Frame color binary threshold max", strconv.FormatFloat(ds.BinaryThresholdDifferenceMax, 'f', -1, 64)},
	})

	return nil
}

func RenderConfusionMatrix(p printer.Printer, cm statistics.ConfusionMatrix) error {
	p.Table([][]string{
		{"TP", "[True positive]", fmt.Sprintf("%f", cm.Tp)},
		{"TN", "[True negative]", fmt.Sprintf("%f", cm.Tn)},
		{"FP", "[False positive]", fmt.Sprintf("%f", cm.Fp)},
		{"FN", "[False negative]", fmt.Sprintf("%f", cm.Fn)},
		{"P", "[Positive]", fmt.Sprintf("%f", cm.P)},
		{"N", "[Negative]", fmt.Sprintf("%f", cm.N)},
		{"TPR", "[Sensitivity / Recall]", fmt.Sprintf("%f", cm.Tpr)},
		{"TNR", "[Specificity / SPC]", fmt.Sprintf("%f", cm.Tnr)},
		{"ACC", "[Accuracy]", fmt.Sprintf("%f", cm.Acc)},
		{"PPV", "[Precision]", fmt.Sprintf("%f", cm.Ppv)},
		{"NPV", "[Negative predictive value]", fmt.Sprintf("%f", cm.Npv)},
		{"FPR", "[False positive rate]", fmt.Sprintf("%f", cm.Fpr)},
		{"FNR", "[False negative rate]", fmt.Sprintf("%f", cm.Fnr)},
		{"MCC", "[Matthews correlation coefficient]", fmt.Sprintf("%f", cm.Mcc)},
		{"FS", "[F-Score]", fmt.Sprintf("%f", cm.Fs)},
	})

	return nil
}
