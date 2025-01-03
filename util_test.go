package openskill

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var delta = 0.001

func TestNormalize(t *testing.T) {
	t.Parallel()

	assert.Equal(t, normalize([]float64{1, 2, 3}, 0, 1), []float64{0, 0.5, 1})
	assert.Equal(t, normalize([]float64{1, 2, 3}, 0, 100), []float64{0, 50, 100})
	assert.Equal(t, normalize([]float64{1, 2, 3}, 0, 10), []float64{0, 5, 10})
	assert.Equal(t, normalize([]float64{1, 2, 3}, 1, 0), []float64{1, 0.5, 0})
	assert.Equal(t, normalize([]float64{1, 1, 1}, 0, 1), []float64{0, 0, 0})
	assert.Equal(t, normalize([]float64{1}, 0, 10), []float64{10})
}

func TestV(t *testing.T) {
	t.Parallel()

	assert.InDelta(t, v(1, 2), 1.525135276160981, delta)
	assert.InDelta(t, v(0, 2), 2.373215532822843, delta)
	assert.InDelta(t, v(0, -1), 0.287599970939178, delta)
	assert.Equal(t, v(0, 10), 10.0)
}

func TestVt(t *testing.T) {
	t.Parallel()

	assert.Equal(t, vt(-1000, -100), 1100.0)
	assert.Equal(t, vt(1000, -100), -1100.0)
	assert.InDelta(t, vt(-1000, 1000), 0.79788, delta)
	assert.Equal(t, vt(0, 1000), 0.0)
}

func TestW(t *testing.T) {
	t.Parallel()

	assert.InDelta(t, w(1, 2), 0.800902334429651, delta)
	assert.InDelta(t, w(0, 2), 0.885720899585924, delta)
	assert.InDelta(t, w(0, -1), 0.3703137142233946, delta)
	assert.Equal(t, w(0, 10), 0.0)
	assert.Equal(t, w(-1, 10), 1.0)
}

func TestWt(t *testing.T) {
	t.Parallel()

	assert.InDelta(t, wt(1, 2), 0.3838582646421707, delta)
	assert.InDelta(t, wt(0, 2), 0.2262586964500768, delta)
	assert.InDelta(t, wt(0, 10), 0.0, delta)
	assert.Equal(t, wt(0, -1), 1.0)
	assert.Equal(t, wt(0, 0), 1.0)
}

func TestLadderPairs(t *testing.T) {
	t.Parallel()

	assert.Equal(t, ladderPairs([]int{}), [][]int{{}})
	assert.Equal(t, ladderPairs([]int{1}), [][]int{{}})
	assert.Equal(t, ladderPairs([]int{1, 2}), [][]int{{2}, {1}})
	assert.Equal(t, ladderPairs([]int{1, 2, 3, 4}), [][]int{{2}, {1, 3}, {2, 4}, {3}})
}

func TestUnwind(t *testing.T) {
	t.Parallel()

	t.Run("zero items", func(t *testing.T) {
		output, tenet := unwind([]int{}, []any{})

		assert.Equal(t, output, []any{})
		assert.Equal(t, tenet, []int{})
	})

	t.Run("one item", func(t *testing.T) {
		output, tenet := unwind([]int{0}, []string{"a"})

		assert.Equal(t, output, []string{"a"})
		assert.Equal(t, tenet, []int{0})
	})

	t.Run("multiple items", func(t *testing.T) {
		output, tenet := unwind([]int{1, 3, 2, 0}, []string{"b", "d", "c", "a"})

		assert.Equal(t, output, []string{"a", "b", "c", "d"})
		assert.Equal(t, tenet, []int{3, 0, 2, 1})
	})

	t.Run("non-zero-indexed ranks", func(t *testing.T) {
		source := []string{"a", "b", "c", "d", "e", "f"}
		rank := []int{3, 7, 5, 1, 10, 8}

		output, _ := unwind(rank, source)

		assert.Equal(t, output, []string{"d", "a", "c", "b", "f", "e"})
	})

	t.Run("undo the ranking", func(t *testing.T) {
		source := []string{"b", "d", "c", "a"}
		rank := []int{1, 3, 2, 0}

		tempOutput, tempTenet := unwind(rank, source)
		output, tenet := unwind(tempTenet, tempOutput)

		assert.Equal(t, output, source)
		assert.Equal(t, tenet, rank)
	})
}

func TestC(t *testing.T) {
	t.Parallel()

	mu := 25.0
	sigma := 25.0 / 3.0
	beta := 25.0 / 6.0
	kappa := 0.0001

	t.Run("1v2", func(t *testing.T) {
		t1 := []Rating{{mu, sigma}}
		t2 := []Rating{{mu, sigma}, {mu, sigma}}

		teamRatings := calculateTeamRatings([][]Rating{t1, t2}, nil, false, kappa)

		assert.InDelta(t, c(teamRatings, beta), 15.590239, delta)
	})

	t.Run("5v5", func(t *testing.T) {
		t1 := []Rating{{mu, sigma}, {mu, sigma}, {mu, sigma}, {mu, sigma}, {mu, sigma}}
		t2 := []Rating{{mu, sigma}, {mu, sigma}, {mu, sigma}, {mu, sigma}, {mu, sigma}}

		teamRatings := calculateTeamRatings([][]Rating{t1, t2}, nil, false, kappa)

		assert.InDelta(t, c(teamRatings, beta), 27.003, delta)
	})
}

func TestA(t *testing.T) {
	t.Parallel()

	mu := 25.0
	sigma := 25.0 / 3.0
	kappa := 0.0001

	t1 := []Rating{{mu, sigma}}
	t2 := []Rating{{mu, sigma}, {mu, sigma}}
	t3 := []Rating{{mu, sigma}, {mu, sigma}}
	t4 := []Rating{{mu, sigma}}

	t.Run("one team per rank", func(t *testing.T) {
		teamRatings := calculateTeamRatings([][]Rating{t1, t2, t3, t4}, nil, false, kappa)

		assert.Equal(t, a(teamRatings), []int{1, 1, 1, 1})
	})

	t.Run("shared ranks", func(t *testing.T) {
		teamRatings := calculateTeamRatings([][]Rating{t1, t2, t3, t4}, []int{1, 1, 1, 4}, false, kappa)

		assert.Equal(t, a(teamRatings), []int{3, 3, 3, 1})
	})
}

func TestSumQ(t *testing.T) {
	t.Parallel()

	mu := 25.0
	sigma := 25.0 / 3.0
	beta := 25.0 / 6.0
	kappa := 0.0001

	t.Run("1v2", func(t *testing.T) {
		t1 := []Rating{{mu, sigma}}
		t2 := []Rating{{mu, sigma}, {mu, sigma}}

		teamRatings := calculateTeamRatings([][]Rating{t1, t2}, nil, false, kappa)
		c := c(teamRatings, beta)
		sums := sumQ(teamRatings, c)

		assert.InDelta(t, sums[0], 29.67892702634643, delta)
		assert.InDelta(t, sums[1], 24.70819334370875, delta)
	})

	t.Run("5v5", func(t *testing.T) {
		t1 := []Rating{{mu, sigma}, {mu, sigma}, {mu, sigma}, {mu, sigma}, {mu, sigma}}
		t2 := []Rating{{mu, sigma}, {mu, sigma}, {mu, sigma}, {mu, sigma}, {mu, sigma}}

		teamRatings := calculateTeamRatings([][]Rating{t1, t2}, nil, false, kappa)
		c := c(teamRatings, beta)
		sums := sumQ(teamRatings, c)

		assert.InDelta(t, sums[0], 204.843788, delta)
		assert.InDelta(t, sums[1], 102.421894, delta)
	})
}
