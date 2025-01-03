package openskill

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckRateParameters(t *testing.T) {
	t.Parallel()

	t.Run("nil teams", func(t *testing.T) {
		err := checkRateParameters(nil, []int{}, nil, nil)

		assert.ErrorIs(t, err, ErrLessThanTwoTeams)
	})

	t.Run("empty teams", func(t *testing.T) {
		err := checkRateParameters([][]Rating{}, []int{}, nil, nil)

		assert.ErrorIs(t, err, ErrLessThanTwoTeams)
	})

	t.Run("one team", func(t *testing.T) {
		t1 := []Rating{{1, 2}}

		err := checkRateParameters([][]Rating{t1}, []int{1}, nil, nil)

		assert.ErrorIs(t, err, ErrLessThanTwoTeams)
	})

	t.Run("empty team", func(t *testing.T) {
		t1 := []Rating{{1, 2}}
		t2 := []Rating{}

		err := checkRateParameters([][]Rating{t1, t2}, []int{1, 2}, nil, nil)

		assert.ErrorIs(t, err, ErrEmptyTeam)
	})

	t.Run("ranks and scores", func(t *testing.T) {
		t1 := []Rating{{1, 2}}
		t2 := []Rating{{1, 2}}

		err := checkRateParameters([][]Rating{t1, t2}, []int{1, 2}, []int{1, 2}, nil)

		assert.ErrorIs(t, err, ErrRanksAndScores)
	})

	t.Run("no ranks or scores", func(t *testing.T) {
		t1 := []Rating{{1, 2}}
		t2 := []Rating{{1, 2}}

		err := checkRateParameters([][]Rating{t1, t2}, nil, nil, nil)

		assert.ErrorIs(t, err, ErrNoRanksOrScores)
	})

	t.Run("ranks and teams length mismatch", func(t *testing.T) {
		t1 := []Rating{{1, 2}}
		t2 := []Rating{{1, 2}}

		err := checkRateParameters([][]Rating{t1, t2}, []int{1}, nil, nil)

		assert.ErrorIs(t, err, ErrRanksAndTeamsMismatch)
	})

	t.Run("scores and teams length mismatch", func(t *testing.T) {
		t1 := []Rating{{1, 2}}
		t2 := []Rating{{1, 2}}

		err := checkRateParameters([][]Rating{t1, t2}, nil, []int{1}, nil)

		assert.ErrorIs(t, err, ErrScoresAndTeamsMismatch)
	})

	t.Run("team missing weights", func(t *testing.T) {
		t1 := []Rating{{1, 2}}
		t2 := []Rating{{1, 2}}

		err := checkRateParameters([][]Rating{t1, t2}, []int{1, 2}, nil, [][]float64{{1}})

		assert.ErrorIs(t, err, ErrWeightsAndTeamsMismatch)
	})

	t.Run("player missing weight", func(t *testing.T) {
		t1 := []Rating{{1, 2}}
		t2 := []Rating{{1, 2}}

		err := checkRateParameters([][]Rating{t1, t2}, []int{1, 2}, nil, [][]float64{{1}, {}})

		assert.ErrorIs(t, err, ErrWeightsAndTeamsMismatch)
	})

	t.Run("valid", func(t *testing.T) {
		t1 := []Rating{{1, 2}}
		t2 := []Rating{{1, 2}}

		err := checkRateParameters([][]Rating{t1, t2}, []int{1, 2}, nil, [][]float64{{1}, {1}})

		assert.NoError(t, err)
	})
}
