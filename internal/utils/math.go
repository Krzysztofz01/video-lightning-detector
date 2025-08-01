package utils

import "math"

// Calculate the moving mean and standard deviation values of the provided set. The position paramter is the index of
// the central subset element and the bias is the amount of "left" and "right" neighbours. Elements out of index are
// not taken under account.
func MovingMeanStdDev(x []float64, position, bias int) (float64, float64) {
	if len(x) == 0 {
		panic("utils: can not calcualte the mean of an empty set")
	}

	if position >= len(x) {
		panic("utils: the position is out of bounds of the value set")
	}

	nominator := 0.0
	denominator := 0.0

	for index := position - bias; index <= position+bias; index += 1 {
		if index >= 0 && index < len(x) {
			nominator += x[index]
			denominator += 1
		}
	}

	mean := nominator / denominator

	nominator = 0.0
	denominator = 0.0

	for index := position - bias; index <= position+bias; index += 1 {
		if index >= 0 && index < len(x) {
			nominator += (x[index] - mean) * (x[index] - mean)
			denominator += 1
		}
	}

	stdDev := math.Sqrt(nominator / denominator)

	return mean, stdDev
}

// Calculate the mean and standard deviation values of the provided set. Panic if the value set is empty.
func MeanStdDev(x []float64) (float64, float64) {
	if len(x) == 0 {
		panic("utils: can not calculate the mean and standard deviation of an empty set")
	}

	sum := 0.0
	for _, value := range x {
		sum += value
	}

	mean := sum / float64(len(x))

	meanDiffSum := 0.0
	for _, value := range x {
		meanDiffSum += math.Pow(mean-value, 2)
	}

	stdDev := math.Sqrt(meanDiffSum / float64(len(x)))

	return mean, stdDev
}

// Calcualte the min and max value of the provided set. Panic if the value set is empty.
func MinMax(x []float64) (float64, float64) {
	if len(x) == 0 {
		panic("utils: can not calucalte the max value of an empty set")
	}

	var (
		min float64 = x[0]
		max float64 = x[0]
	)

	for _, value := range x {
		if value < min {
			min = value
		}

		if value > max {
			max = value
		}
	}

	return min, max
}

// Return the smaller value of x or y. This functions does not support the edge cases like math.Min
func MinInt(x, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}

// Divide a by b and return fallback for zero-division
func Div(a, b, fallback float64) float64 {
	if b == 0 {
		return fallback
	}

	return a / b
}
