package statistics

import (
	"fmt"
	"math"

	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

type ConfusionMatrix struct {
	// TP - True positive
	Tp float64

	// TN - True negative
	Tn float64

	// FP - False positive
	Fp float64

	// FN - False negative
	Fn float64

	// P - Positive
	P float64

	// N - Negative
	N float64

	// TPR - Sensitivity/Recall
	Tpr float64

	// TNR - Specificity SPC
	Tnr float64

	// ACC - Accuracy
	Acc float64

	// PPV - Precision
	Ppv float64

	// NPV - Negative predictive value
	Npv float64

	// FPR - False positive rate
	Fpr float64

	// FNR - False negative rate
	Fnr float64

	// MCC - Matthews correlation coefficient
	Mcc float64

	// F-Score
	Fs float64
}

// Get the confusion matrix of binarty classified frame lightning detection. The confusion matrix is calculated via the
// provided actual classified frame indices, prediction classified frame indices and total count of frames in the video
func CreateConfusionMatrix(actualClassifcationFrames, predictedClassificationFrames []int, totalCount int) ConfusionMatrix {
	actual := getTotalClassification(actualClassifcationFrames, totalCount)
	predicted := getTotalClassification(predictedClassificationFrames, totalCount)

	// TP - True positive
	tp := evalMeasure(actual, predicted, func(a, p bool) bool {
		return a && p
	})

	// TN - True negative
	tn := evalMeasure(actual, predicted, func(a, p bool) bool {
		return !a && !p
	})

	// FP - False positive
	fp := evalMeasure(actual, predicted, func(a, p bool) bool {
		return !a && p
	})

	// FN - False negative
	fn := evalMeasure(actual, predicted, func(a, p bool) bool {
		return a && !p
	})

	tpr := utils.Div(tp, (tp + fn), 0) // tp / (tp + fn)
	ppv := utils.Div(tp, (tp + fp), 0) // tp / (tp + fp)

	return ConfusionMatrix{
		Tp:  tp,
		Tn:  tn,
		Fp:  fp,
		Fn:  fn,
		P:   tp + fn,
		N:   tn + fp,
		Tpr: tpr,
		Tnr: utils.Div(tn, (fp + tn), 0),                  // tn / (fp + tn),
		Acc: utils.Div((tp + tn), (tp + fn + tn + fp), 0), // (tp + tn) / (tp + fn + tn + fp),
		Ppv: ppv,
		Npv: utils.Div(tn, (tn + fn), 0),                                               // tn / (tn + fn),
		Fpr: utils.Div(fp, (fp + tn), 0),                                               // fp / (fp + tn),
		Fnr: utils.Div(fp, (tp + fn), 0),                                               // fp / (tp + fn),
		Mcc: utils.Div((tp*tn - fp*fn), math.Sqrt((tp+fp)*(tp+fn)*(tn+fp)*(tn+fn)), 0), // (tp*tn - fp*fn) / math.Sqrt((tp+fp)*(tp+fn)*(tn+fp)*(tn+fn)),
		Fs:  2 * utils.Div((ppv*tpr), (ppv+tpr), 0),                                    // 2 * ((ppv * tpr) / (ppv + tpr))
	}
}

func getTotalClassification(classifiedIndices []int, totalCount int) []bool {
	classification := make([]bool, totalCount)

	for _, classifiedFrame := range classifiedIndices {
		classification[classifiedFrame-1] = true
	}

	return classification
}

func evalMeasure(actual, predicted []bool, f func(a, p bool) bool) float64 {
	if len(actual) != len(predicted) {
		panic(fmt.Errorf("statistics: attempt to evaluate confustion matrix measure for uneven sets"))
	}

	count := 0
	for index := 0; index < len(actual); index += 1 {
		if f(actual[index], predicted[index]) {
			count += 1
		}
	}

	return float64(count)
}
