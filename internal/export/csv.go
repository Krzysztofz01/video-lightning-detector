package export

import (
	"encoding/csv"
	"fmt"
	"path"
	"strconv"

	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/statistics"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

func ExportCsvFrames(outputDirectoryPath string, fc frame.FrameCollection) (string, error) {
	csvFramesReportPath := path.Join(outputDirectoryPath, CsvFramesReportFilename)
	framesReportFile, err := utils.CreateFileWithTree(csvFramesReportPath)
	if err != nil {
		return "", fmt.Errorf("export: failed to create the csv frames report file: %w", err)
	}

	defer func() {
		if err := framesReportFile.Close(); err != nil {
			panic(err)
		}
	}()

	writer := csv.NewWriter(framesReportFile)

	if err := writer.Write([]string{"Frame", "Brightness", "ColorDifference", "BinaryThresholdDifference"}); err != nil {
		return "", fmt.Errorf("export: failed to write the header to the frames report file: %w", err)
	}

	var (
		frames    []*frame.Frame = fc.GetAll()
		rowBuffer []string       = make([]string, 4)
	)

	for _, frame := range frames {
		rowBuffer = rowBuffer[:0]
		rowBuffer = append(rowBuffer, strconv.Itoa(frame.OrdinalNumber))
		rowBuffer = append(rowBuffer, strconv.FormatFloat(frame.Brightness, 'f', -1, 64))
		rowBuffer = append(rowBuffer, strconv.FormatFloat(frame.ColorDifference, 'f', -1, 64))
		rowBuffer = append(rowBuffer, strconv.FormatFloat(frame.BinaryThresholdDifference, 'f', -1, 64))

		if err := writer.Write(rowBuffer); err != nil {
			return "", fmt.Errorf("export: failed to write the frame row to the frames report file: %w", err)
		}
	}

	writer.Flush()

	return csvFramesReportPath, nil
}

func ExportCsvDescriptiveStatistics(outputDirectoryPath string, ds statistics.DescriptiveStatistics) (string, error) {
	csvDescriptiveStatisticsReportPath := path.Join(outputDirectoryPath, CsvDescriptiveStatisticsReportFilename)
	statisticsReportFile, err := utils.CreateFileWithTree(csvDescriptiveStatisticsReportPath)
	if err != nil {
		return "", fmt.Errorf("export: failed to create the csv descriptive statistics report file: %w", err)
	}

	defer func() {
		if err := statisticsReportFile.Close(); err != nil {
			panic(err)
		}
	}()

	writer := csv.NewWriter(statisticsReportFile)

	rows := [][]string{
		{"", "Brightness mean", "Brightness standard deviation", "Brightness max"},
		valuesToCsvRow(1, ds.BrightnessMean, ds.BrightnessStandardDeviation, ds.BrightnessMax),
		{},
		{"", "Color difference mean", "Color difference standard deviation", "Color difference max"},
		valuesToCsvRow(1, ds.ColorDifferenceMean, ds.ColorDifferenceStandardDeviation, ds.ColorDifferenceMax),
		{},
		{"", "Binary threshold difference mean", "Binary threshold difference standard deviation", "Binary threshold difference max"},
		valuesToCsvRow(1, ds.BinaryThresholdDifferenceMean, ds.BinaryThresholdDifferenceStandardDeviation, ds.BinaryThresholdDifferenceMax),
		{},
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("export: failed to write descriptive statistics rows to the report file: %w", err)
		}
	}

	if err := writer.Write([]string{"Frame (Moving mean center point)", "Brightness moving mean", "ColorDifference moving mean", "BinaryThresholdDifference moving mean"}); err != nil {
		return "", fmt.Errorf("export: failed to write the moving mean header to the descriptive statistics report file: %w", err)
	}

	for index := 0; index < len(ds.BrightnessMovingMean); index += 1 {
		values := valuesToCsvRow(0, ds.BrightnessMovingMean[index], ds.ColorDifferenceMovingMean[index], ds.BinaryThresholdDifferenceMovingMean[index])
		values = append([]string{strconv.Itoa(index + 1)}, values...)

		if err := writer.Write(values); err != nil {
			return "", fmt.Errorf("export: failed to write moving mean row to the descriptive statistics report file: %w", err)
		}
	}

	writer.Flush()

	return csvDescriptiveStatisticsReportPath, nil
}

func ExportCsvConfusionMatrix(outputDirectoryPath string, cm statistics.ConfusionMatrix) (string, error) {
	csvConfusionMatrixReportPath := path.Join(outputDirectoryPath, CsvConfusionMatrixReportFilename)
	confusionMatrixReportFile, err := utils.CreateFileWithTree(csvConfusionMatrixReportPath)
	if err != nil {
		return "", fmt.Errorf("export: failed to create the csv confusion matrix report file: %w", err)
	}

	defer func() {
		if err := confusionMatrixReportFile.Close(); err != nil {
			panic(err)
		}
	}()

	writer := csv.NewWriter(confusionMatrixReportFile)

	rows := [][]string{
		{"Tp", strconv.FormatFloat(cm.Tp, 'f', -1, 64)},
		{"Tn", strconv.FormatFloat(cm.Tn, 'f', -1, 64)},
		{"Fp", strconv.FormatFloat(cm.Fp, 'f', -1, 64)},
		{"Fn", strconv.FormatFloat(cm.Fn, 'f', -1, 64)},
		{"P:", strconv.FormatFloat(cm.P, 'f', -1, 64)},
		{"N:", strconv.FormatFloat(cm.N, 'f', -1, 64)},
		{"Tpr", strconv.FormatFloat(cm.Tpr, 'f', -1, 64)},
		{"Tnr", strconv.FormatFloat(cm.Tnr, 'f', -1, 64)},
		{"Acc", strconv.FormatFloat(cm.Acc, 'f', -1, 64)},
		{"Ppv", strconv.FormatFloat(cm.Ppv, 'f', -1, 64)},
		{"Npv", strconv.FormatFloat(cm.Npv, 'f', -1, 64)},
		{"Fpr", strconv.FormatFloat(cm.Fpr, 'f', -1, 64)},
		{"Fnr", strconv.FormatFloat(cm.Fnr, 'f', -1, 64)},
		{"Mcc", strconv.FormatFloat(cm.Mcc, 'f', -1, 64)},
		{"Fs", strconv.FormatFloat(cm.Fs, 'f', -1, 64)},
	}

	if err := writer.WriteAll(rows); err != nil {
		return "", fmt.Errorf("export: failed to write the confusion matrix rows to the csv file: %w", err)
	}

	return csvConfusionMatrixReportPath, nil
}

func valuesToCsvRow(leftPadding int, values ...float64) []string {
	buffer := make([]string, 0, len(values)+leftPadding)
	for index := 0; index < leftPadding; index += 1 {
		buffer = append(buffer, "")
	}

	for _, value := range values {
		buffer = append(buffer, strconv.FormatFloat(value, 'f', -1, 64))
	}

	return buffer
}
