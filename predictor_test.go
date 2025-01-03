package openskill

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultPredictor(t *testing.T) {
	t.Parallel()

	p := DefaultPredictor()

	assert.Equal(t, 25.0/6.0, p.(predictor).beta)
	assert.Equal(t, 0.0001, p.(predictor).kappa)
	assert.False(t, p.(predictor).balance)
}

func TestNewPredictor(t *testing.T) {
	t.Parallel()

	p := NewPredictor(1, 2, true)

	assert.Equal(t, 1.0, p.(predictor).beta)
	assert.Equal(t, 2.0, p.(predictor).kappa)
	assert.True(t, p.(predictor).balance)
}

func TestCheckTeams(t *testing.T) {
	t.Parallel()

	t.Run("nil teams", func(t *testing.T) {
		err := checkTeams(nil)

		assert.ErrorIs(t, err, ErrLessThanTwoTeams)
	})

	t.Run("one team", func(t *testing.T) {
		err := checkTeams([][]Rating{{}})

		assert.ErrorIs(t, err, ErrLessThanTwoTeams)
	})

	t.Run("empty team", func(t *testing.T) {
		err := checkTeams([][]Rating{{{1, 2}}, {}})

		assert.ErrorIs(t, err, ErrEmptyTeam)
	})

	t.Run("valid", func(t *testing.T) {
		err := checkTeams([][]Rating{{{1, 2}}, {{1, 2}}})

		assert.NoError(t, err)
	})
}
