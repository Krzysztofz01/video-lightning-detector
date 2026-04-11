package statistics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateConfusionMatrixShouldCreateAndCalculateMetrics(t *testing.T) {
	// NOTE: Expected using ordinal number and actual using index
	var (
		expected []int = []int{1, 2, 3, 4}
		actual   []int = []int{3 - 1, 4 - 1, 5 - 1, 6 - 1}
		total    int   = 10
	)

	const delta float64 = 1e-5

	var (
		Tp  float64 = 2
		Tn  float64 = 4
		Fp  float64 = 2
		Fn  float64 = 2
		P   float64 = 4
		N   float64 = 6
		Tpr float64 = 0.5
		Tnr float64 = 2.0 / 3.0
		Acc float64 = 0.6
		Ppv float64 = 0.5
		Npv float64 = 2.0 / 3.0
		Fpr float64 = 1.0 / 3.0
		Fnr float64 = 0.5
		Mcc float64 = 1.0 / 6.0
		Fs  float64 = 0.5
	)

	conmat := CreateConfusionMatrix(expected, actual, total)

	assert.InDelta(t, Tp, conmat.Tp, delta)
	assert.InDelta(t, Tn, conmat.Tn, delta)
	assert.InDelta(t, Fp, conmat.Fp, delta)
	assert.InDelta(t, Fn, conmat.Fn, delta)
	assert.InDelta(t, P, conmat.P, delta)
	assert.InDelta(t, N, conmat.N, delta)
	assert.InDelta(t, Tpr, conmat.Tpr, delta)
	assert.InDelta(t, Tnr, conmat.Tnr, delta)
	assert.InDelta(t, Acc, conmat.Acc, delta)
	assert.InDelta(t, Ppv, conmat.Ppv, delta)
	assert.InDelta(t, Npv, conmat.Npv, delta)
	assert.InDelta(t, Fpr, conmat.Fpr, delta)
	assert.InDelta(t, Fnr, conmat.Fnr, delta)
	assert.InDelta(t, Mcc, conmat.Mcc, delta)
	assert.InDelta(t, Fs, conmat.Fs, delta)
}
