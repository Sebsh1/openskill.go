package openskill

import "fmt"

var (
	ErrLessThanTwoTeams        = fmt.Errorf("less than two teams")
	ErrEmptyTeam               = fmt.Errorf("empty team")
	ErrNoRanksOrScores         = fmt.Errorf("ranks or scores must be provided")
	ErrRanksAndScores          = fmt.Errorf("ranks and scores cannot be provided together")
	ErrRanksAndTeamsMismatch   = fmt.Errorf("ranks must have same shape as teams")
	ErrScoresAndTeamsMismatch  = fmt.Errorf("scores must have same shape as teams")
	ErrWeightsAndTeamsMismatch = fmt.Errorf("weights must have same shape as teams")
)
