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

// Calculate the moving mean and standard deviation in a incremental way. Provide the push and pop values, current dataset mean
// stddev and length. Other functions MovingMeanStd functions here are taking the position of the calculation center point and
// taking half of the bias left and right. This function takes a bias amount of left values.
func MovingMeanStdDevInc(value, discardValue, mean, stdDev float64, length, bias int) (float64, float64) {
	if length < 0 {
		panic("utils: can not calculate the incremental moving mean of an empty set")
	}

	if length == 0 {
		return value, 0
	}

	var (
		lengthf    float64
		meanNext   float64
		stdDevNext float64
	)

	if length >= bias {
		lengthf = float64(bias)

		meanDelta := (value - discardValue) / lengthf
		meanNext = mean + meanDelta

		varSqrtDelta := (value - discardValue) * (value - meanNext + discardValue - mean)
		stdDevNext = math.Sqrt((stdDev*stdDev*lengthf + varSqrtDelta) / float64(bias))
	} else {
		lengthf = float64(length)

		meanNext = (mean*lengthf + value) / (lengthf + 1)

		stdDevDelta := (value - mean) * (value - meanNext)
		stdDevNext = math.Sqrt(math.Max(0, (lengthf*stdDev*stdDev+stdDevDelta)/(lengthf+1)))
	}

	return meanNext, stdDevNext
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

// Calculate the mean and standard deviation values in an incremental way using the previous mean, standard
// deviation, set length and the incoming new value. Panic if the length is negative
func MeanStdDevInc(value, mean, stdDev float64, length int) (float64, float64) {
	if length < 0 {
		panic("utils: can not calculate the incremental mean and standard deviation with an negative length")
	}

	if length == 0 {
		return value, 0
	}

	lf := float64(length)

	meanInc := (mean*lf + value) / (lf + 1)

	stdDevInc := math.Sqrt(math.Max(0, ((lf)*stdDev*stdDev+(value-meanInc)*(value-mean))/(lf+1)))

	return meanInc, stdDevInc
}

// Calcualte the min and max value of the provided set. Panic if the value set is empty.
func MinMax(x []float64) (float64, float64) {
	if len(x) == 0 {
		panic("utils: can not calculate the max value of an empty set")
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

// Calculate the min and max value in a incremental way using the three provided values.
func MinMaxInc(value, min, max float64) (float64, float64) {
	if value < min {
		min = value
	}

	if value > max {
		max = value
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

// Return the bigger value of x or y. This function does not support the edge cases like math.Max
func MaxInt(x, y int) int {
	if x > y {
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
