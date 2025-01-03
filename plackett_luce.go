package openskill

import (
	"math"
	"sort"
)

type PlackettLuceModel struct {
	mu         float64
	sigma      float64
	beta       float64
	kappa      float64
	limitSigma bool
	balance    bool
}

func DefaultPlackettLuceModel() Rater {
	return PlackettLuceModel{
		mu:         25.0,
		sigma:      25.0 / 3.0,
		beta:       25.0 / 6.0,
		kappa:      0.0001,
		limitSigma: false,
		balance:    false,
	}
}

func NewPlackettLuceModel(mu, sigma, beta, kappa float64, limitSigma, balance bool) Rater {
	return PlackettLuceModel{
		mu:         mu,
		sigma:      sigma,
		beta:       beta,
		kappa:      kappa,
		limitSigma: limitSigma,
		balance:    balance,
	}
}

func (p PlackettLuceModel) Rate(teams [][]Rating, ranks, scores []int, weights [][]float64) ([][]Rating, error) {
	if err := checkRateParameters(teams, ranks, scores, weights); err != nil {
		return nil, err
	}

	originalTeams := make([][]Rating, len(teams))
	for i := range teams {
		originalTeams[i] = make([]Rating, len(teams[i]))
		copy(originalTeams[i], teams[i])
	}

	teamsCopy := make([][]Rating, len(teams))
	for i := range teams {
		teamsCopy[i] = make([]Rating, len(teams[i]))
		for j := range teams[i] {
			teamsCopy[i][j] = Rating{Mu: teams[i][j].Mu, Sigma: teams[i][j].Sigma}
		}
	}

	for teamIndex, team := range teamsCopy {
		for playerIndex, player := range team {
			teamsCopy[teamIndex][playerIndex].Sigma = player.Sigma
		}
	}

	if scores != nil {
		ranks = make([]int, len(scores))
		for i, score := range scores {
			ranks[i] = int(-score)
		}
	}

	for i := range weights {
		weights[i] = normalize(weights[i], 1, 2)
	}

	var tenet []int
	var orderedTeams [][]Rating

	if ranks != nil {
		orderedTeams, tenet = unwind(ranks, teamsCopy)
		teamsCopy = orderedTeams
		sort.Ints(ranks)
	}

	var result [][]Rating
	if ranks != nil && tenet != nil {
		result = p.compute(teamsCopy, ranks, weights)
		result, _ = unwind(tenet, result)
	} else {
		result = p.compute(teamsCopy, nil, weights)
	}

	finalResult := make([][]Rating, len(result))
	for i := range result {
		finalResult[i] = make([]Rating, len(result[i]))
		for j := range result[i] {
			finalResult[i][j] = Rating{Mu: result[i][j].Mu, Sigma: result[i][j].Sigma}
		}
	}

	if p.limitSigma {
		for teamIndex, team := range finalResult {
			for playerIndex, player := range team {
				finalResult[teamIndex][playerIndex].Sigma = math.Min(player.Sigma, originalTeams[teamIndex][playerIndex].Sigma)
			}
		}
	}

	return finalResult, nil
}

func (p PlackettLuceModel) compute(teams [][]Rating, ranks []int, weights [][]float64) [][]Rating {
	originalTeams := make([][]Rating, len(teams))
	for i := range teams {
		originalTeams[i] = make([]Rating, len(teams[i]))
		copy(originalTeams[i], teams[i])
	}

	teamRatings := calculateTeamRatings(teams, ranks, p.balance, p.kappa)
	a := a(teamRatings)
	c := c(teamRatings, p.beta)
	sumQ := sumQ(teamRatings, c)

	result := make([][]Rating, len(teams))
	for i, t1 := range teamRatings {
		omega := 0.0
		delta := 0.0
		iMuOverC := math.Exp(t1.Mu / c)

		for j, t2 := range teamRatings {
			iMuOverCOverSumQ := iMuOverC / sumQ[j]
			if t2.Rank <= t1.Rank {
				delta += (iMuOverCOverSumQ * (1 - iMuOverCOverSumQ) / float64(a[j]))
			}
			if j == i {
				omega += (1 - iMuOverCOverSumQ) / float64(a[j])
			} else {
				omega -= iMuOverCOverSumQ / float64(a[j])
			}
		}

		omega *= t1.SigmaSquared / c
		delta *= t1.SigmaSquared / math.Pow(c, 2)
		gammaValue := math.Sqrt(p.sigma*p.sigma) / c
		delta *= gammaValue

		intermediateResultPerTeam := make([]Rating, len(t1.Team))
		for j, player := range t1.Team {
			weight := 1.0
			if weights != nil {
				weight = weights[i][j]
			}

			mu := player.Mu
			sigma := player.Sigma
			if omega > 0 {
				mu += (sigma / t1.SigmaSquared) * omega * weight
				sigma *= math.Max(1-(sigma/t1.SigmaSquared)*delta*weight, p.kappa)
			} else {
				mu += (sigma / t1.SigmaSquared) * omega / weight
				sigma *= math.Max(1-(sigma/t1.SigmaSquared)*delta/weight, p.kappa)
			}

			intermediateResultPerTeam[j] = Rating{Mu: mu, Sigma: sigma}
		}
		result[i] = intermediateResultPerTeam
	}

	return result
}
