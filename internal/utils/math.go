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

// Calculate the moving mean value of the provided set. The position paramter is the index of the central subset element
// and the bias is the amount of "left" and "right" neighbours. Elements out of index are not taken under account.
func MovingMean(x []float64, position, bias int) float64 {
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

	return nominator / denominator
}

// TODO: Add docs and tests
func MovingMeanStdDev(x []float64, position, bias int) (float64, float64) {
	if len(x) == 0 {
		panic("utils: can not calcualte the mean of an empty set")
	}

	if position >= len(x) {
		panic("utils: the position is out of bounds of the value set")
	}

	mean := MovingMean(x, position, bias)

	nominator := 0.0
	denominator := 0.0

	for index := position - bias; index <= position+bias; index += 1 {
		if index >= 0 && index < len(x) {
			nominator += (x[index] - mean) * (x[index] - mean)
			denominator += 1
		}
	}

	stdDev := math.Sqrt(nominator / denominator)

	return mean, stdDev
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
