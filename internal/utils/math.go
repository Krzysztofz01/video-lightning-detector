package utils

import "math"

// Calculate the mean value of the provided set. Panic if the value set is empty.
func Mean(x []float64) float64 {
	if len(x) == 0 {
		panic("utils: can not calculate the mean of an empty set")
	}

	sum := 0.0
	for _, value := range x {
		sum += value
	}

	return sum / float64(len(x))
}

// Calculate the population standard deviation value of the provided set. Panic if the value set is empty.
func StandardDeviation(x []float64) float64 {
	if len(x) == 0 {
		panic("utils: can not calculate the standard deviation of an empty set")
	}

	mean := Mean(x)
	meanDiffSum := 0.0
	for _, value := range x {
		meanDiffSum += math.Pow(mean-value, 2)
	}

	return math.Sqrt(meanDiffSum / float64(len(x)))

}

// Calcualte the max value of the provided set. Panic if the value set is empty.
func Max(x []float64) float64 {
	if len(x) == 0 {
		panic("utils: can not calucalte the max value of an empty set")
	}

	max := x[0]
	for _, value := range x {
		if value > max {
			max = value
		}
	}

	return max
}
