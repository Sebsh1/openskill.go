package openskill

import (
	"math"
	"sort"
)

type Predictor interface {
	ChanceOfWinning(teams [][]Rating) ([]float64, error)
	ChanceOfDraw(teams [][]Rating) (float64, error)
	ChanceOfRanks(teams [][]Rating) ([]int, []float64, error)
}

type predictor struct {
	beta    float64
	kappa   float64
	balance bool
}

// DefaultPredictor returns a new Predictor with sensible default parameter values.
func DefaultPredictor() Predictor {
	return predictor{
		beta:    25.0 / 6.0,
		kappa:   0.0001,
		balance: false,
	}
}

// NewPredictor returns a new Predictor with custom parameter values.
func NewPredictor(beta, kappa float64, balance bool) Predictor {
	return predictor{
		beta:    beta,
		kappa:   kappa,
		balance: balance,
	}
}

// ChanceOfWinning returns the probability of each team winning a match as a number between 0 and 1.
func (p predictor) ChanceOfWinning(teams [][]Rating) ([]float64, error) {
	if err := checkTeams(teams); err != nil {
		return nil, err
	}

	n := len(teams)

	if n == 2 {
		teamsRatings := calculateTeamRatings(teams, nil, p.balance, p.kappa)
		a := teamsRatings[0]
		b := teamsRatings[1]
		result := []float64{
			phiMajor((a.Mu - b.Mu) / math.Sqrt(2*p.beta*p.beta+a.SigmaSquared+b.SigmaSquared)),
			1 - phiMajor((a.Mu-b.Mu)/math.Sqrt(2*p.beta*p.beta+a.SigmaSquared+b.SigmaSquared)),
		}

		return result, nil
	}

	pairwiseProbabilities := make([]float64, 0)
	for i := range teams {
		for j := range teams {
			if i != j {
				teamA := teams[i]
				teamB := teams[j]

				teamASubset := calculateTeamRatings([][]Rating{teamA}, nil, p.balance, p.kappa)
				teamBSubset := calculateTeamRatings([][]Rating{teamB}, nil, p.balance, p.kappa)

				muA := teamASubset[0].Mu
				muB := teamBSubset[0].Mu
				sigmaA := teamASubset[0].SigmaSquared
				sigmaB := teamBSubset[0].SigmaSquared

				pairwiseProbabilities = append(pairwiseProbabilities, phiMajor((muA-muB)/math.Sqrt(2*p.beta*p.beta+sigmaA+sigmaB)))
			}
		}
	}

	winProbabilities := make([]float64, n)
	for i := range teams {
		sum := 0.0
		for j := i * (n - 1); j < (i+1)*(n-1); j++ {
			sum += pairwiseProbabilities[j]
		}
		winProbabilities[i] = sum / float64(n-1)
	}

	totalProbability := 0.0
	for _, prob := range winProbabilities {
		totalProbability += prob
	}

	normalizedProbabilities := make([]float64, n)
	for i, prob := range winProbabilities {
		normalizedProbabilities[i] = prob / totalProbability
	}

	return normalizedProbabilities, nil
}

// ChanceOfDraw returns the probability of a match ending in a draw as a number between 0 and 1.
func (p predictor) ChanceOfDraw(teams [][]Rating) (float64, error) {
	if err := checkTeams(teams); err != nil {
		return 0, err
	}

	totalPlayerCount := 0
	for _, team := range teams {
		totalPlayerCount += len(team)
	}

	drawProbability := 1 / float64(totalPlayerCount)
	drawMargin := math.Sqrt(float64(totalPlayerCount)) * p.beta * phiMajorInv((1+drawProbability)/2)

	pairwiseProbabilities := make([]float64, 0)

	for i := 0; i < len(teams); i++ {
		for j := i + 1; j < len(teams); j++ {
			pairA := teams[i]
			pairB := teams[j]

			pairASubset := calculateTeamRatings([][]Rating{pairA}, nil, p.balance, p.kappa)
			pairBSubset := calculateTeamRatings([][]Rating{pairB}, nil, p.balance, p.kappa)

			muA := pairASubset[0].Mu
			muB := pairBSubset[0].Mu
			sigmaA := pairASubset[0].SigmaSquared
			sigmaB := pairBSubset[0].SigmaSquared

			prob := phiMajor((drawMargin-muA+muB)/math.Sqrt(2*p.beta*p.beta+sigmaA+sigmaB)) -
				phiMajor((muB-muA-drawMargin)/math.Sqrt(2*p.beta*p.beta+sigmaA+sigmaB))

			pairwiseProbabilities = append(pairwiseProbabilities, prob)
		}
	}

	sum := 0.0
	for _, prob := range pairwiseProbabilities {
		sum += prob
	}
	return sum / float64(len(pairwiseProbabilities)), nil

}

// ChanceOfRanks returns the most likely ranking of the teams and the probability of these.
func (p predictor) ChanceOfRanks(teams [][]Rating) ([]int, []float64, error) {
	if err := checkTeams(teams); err != nil {
		return nil, nil, err
	}

	n := len(teams)
	teamRatings := calculateTeamRatings(teams, nil, p.balance, p.kappa)

	winProbabilities := make([]float64, n)
	for i, t1 := range teamRatings {
		for j, t2 := range teamRatings {
			if i != j {
				diff := t1.Mu - t2.Mu
				denominator := math.Sqrt(2*p.beta*p.beta + t1.SigmaSquared + t2.SigmaSquared)
				winProbabilities[i] += phiMajor(diff / denominator)
			}
		}
		winProbabilities[i] /= (float64(n) - 1)
	}

	var totalProbability = 0.0
	for _, p := range winProbabilities {
		totalProbability += p
	}

	normalizedProbabilities := make([]float64, n)
	for i, p := range winProbabilities {
		normalizedProbabilities[i] = p / totalProbability
	}

	sortedTeams := make([][2]int, n)
	for i, prob := range normalizedProbabilities {
		sortedTeams[i] = [2]int{i, int(math.Round(prob * 1000000))}
	}
	sort.SliceStable(sortedTeams, func(i, j int) bool {
		return sortedTeams[i][1] > sortedTeams[j][1]
	})

	ranks := make([]int, n)
	currentRank := 1
	for i, team := range sortedTeams {
		if i > 0 && sortedTeams[i][1] < sortedTeams[i-1][1] {
			currentRank = i + 1
		}
		ranks[team[0]] = currentRank
	}

	return ranks, normalizedProbabilities, nil
}

// checkTeams validates teams input for the predictor methods.
func checkTeams(teams [][]Rating) error {
	if len(teams) < 2 {
		return ErrLessThanTwoTeams
	}

	for _, team := range teams {
		if len(team) < 1 {
			return ErrEmptyTeam
		}
	}

	return nil
}
