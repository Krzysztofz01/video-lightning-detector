package statistics

import "fmt"

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

	// LR+ - Positive likehood ratio
	Plr float64

	// LR- - Negative likehood ratio
	Nlr float64

	// DOR - Diagnostic Odds ratio
	Dor float64
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

	return ConfusionMatrix{
		Tp:  tp,
		Tn:  tn,
		Fp:  fp,
		Fn:  fn,
		P:   tp + fn,
		N:   tn + fp,
		Tpr: tp / (tp + fn),
		Tnr: tn / (fp + tn),
		Acc: (tp + tn) / (tp + fn + tn + fp),
		Ppv: tp / (tp + fp),
		Npv: tn / (tn + fn),
		Fpr: fp / (fp + tn),
		Fnr: fp / (tp + fn),
		Plr: (tp * (fp + tn)) / (fp * (tp + fn)),
		Nlr: (fn * (fp + tn)) / (tn * (tp + fn)),
		Dor: (tp * tn) / (fp * fn),
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
