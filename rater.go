package openskill

// All models implement the Rater interface.
type Rater interface {
	Rate(teams [][]Rating, ranks, scores []int, weights [][]float64) (updatedRatings [][]Rating, err error)
}

// Rating represents a player's skill level as a Gaussian distribution with a mean (Mu) and standard deviation (Sigma).
type Rating struct {
	Mu    float64
	Sigma float64
}

// Ordinal returns a single scalar value that represents a player's rating where their true rating is 99.7% likely to be higher.
func (r Rating) Ordinal() float64 {
	return r.Mu - 3*r.Sigma
}

// checkRateParameters validates the input parameters for the Rate method.
func checkRateParameters(teams [][]Rating, ranks, scores []int, weights [][]float64) error {
	if len(teams) < 2 {
		return ErrLessThanTwoTeams
	}
	for _, team := range teams {
		if len(team) < 1 {
			return ErrEmptyTeam
		}
	}

	if ranks != nil && scores != nil {
		return ErrRanksAndScores
	}
	if ranks == nil && scores == nil {
		return ErrNoRanksOrScores
	}

	if ranks != nil && len(teams) != len(ranks) {
		return ErrRanksAndTeamsMismatch
	}
	if scores != nil && len(teams) != len(scores) {
		return ErrScoresAndTeamsMismatch
	}

	if weights != nil {
		if len(teams) != len(weights) {
			return ErrWeightsAndTeamsMismatch
		}
		for i, teamWeights := range weights {
			if len(teams[i]) != len(teamWeights) {
				return ErrWeightsAndTeamsMismatch
			}
		}
	}

	return nil
}
