package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/Krzysztofz01/video-lightning-detector/internal/frame"
	"github.com/Krzysztofz01/video-lightning-detector/internal/statistics"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

func ExportJsonFrames(outputDirectoryPath string, fc frame.FrameCollection) (string, error) {
	jsonFramesReportPath := path.Join(outputDirectoryPath, JsonFramesReportFilename)
	framesReportFile, err := utils.CreateFileWithTree(jsonFramesReportPath)
	if err != nil {
		return "", fmt.Errorf("export: failed to create the json frames report file: %w", err)
	}

	defer func() {
		if err := framesReportFile.Close(); err != nil {
			panic(err)
		}
	}()

	encoder := createEncoder(framesReportFile)

	frames := fc.GetAll()

	if err := encoder.Encode(frames); err != nil {
		return "", fmt.Errorf("export: failed to encode the frames to json report file: %w", err)
	}

	return jsonFramesReportPath, nil
}

func ExportJsonDescriptiveStatistics(outputDirectoryPath string, ds statistics.DescriptiveStatistics) (string, error) {
	jsonDescriptiveStatisticsReportPath := path.Join(outputDirectoryPath, JsonDescriptiveStatisticsReportFilename)
	statisticsReportFile, err := utils.CreateFileWithTree(jsonDescriptiveStatisticsReportPath)
	if err != nil {
		return "", fmt.Errorf("export: failed to create the json descriptive statistics report file: %w", err)
	}

	defer func() {
		if err := statisticsReportFile.Close(); err != nil {
			panic(err)
		}
	}()

	encoder := createEncoder(statisticsReportFile)

	if err := encoder.Encode(ds); err != nil {
		return "", fmt.Errorf("export: failed to encode the descriptive statistics: %w", err)
	}

	return jsonDescriptiveStatisticsReportPath, nil
}

func ExportJsonConfusionMatrix(outputDirectoryPath string, cm statistics.ConfusionMatrix) (string, error) {
	jsonConfusionMatrixReportPath := path.Join(outputDirectoryPath, JsonConfusionMatrixReportFilename)
	confusionMatrixReportFile, err := utils.CreateFileWithTree(jsonConfusionMatrixReportPath)
	if err != nil {
		return "", fmt.Errorf("export: failed to create the json confusion matrix report file: %w", err)
	}

	defer func() {
		if err := confusionMatrixReportFile.Close(); err != nil {
			panic(err)
		}
	}()

	encoder := createEncoder(confusionMatrixReportFile)

	if err := encoder.Encode(cm); err != nil {
		return "", fmt.Errorf("export: failed to encode the confusion matrix: %w", err)
	}

	return jsonConfusionMatrixReportPath, nil
}

func createEncoder(file *os.File) *json.Encoder {
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	return encoder
}
