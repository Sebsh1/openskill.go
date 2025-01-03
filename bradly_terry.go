package openskill

import (
	"math"
	"sort"
)

type BradlyTerryFullModel struct {
	mu         float64
	sigma      float64
	beta       float64
	kappa      float64
	tau        float64
	limitSigma bool
	balance    bool
}

func DefaultBradlyTerryFullModel() Rater {
	return BradlyTerryFullModel{
		mu:         25.0,
		sigma:      25.0 / 3.0,
		beta:       25.0 / 6.0,
		kappa:      0.0001,
		tau:        25.0 / 300.0,
		limitSigma: false,
		balance:    false,
	}
}

func NewBradlyTerryFullModel(mu, sigma, beta, kappa, tau float64, limitSigma, balance bool) Rater {
	return BradlyTerryFullModel{
		mu:         mu,
		sigma:      sigma,
		beta:       beta,
		kappa:      kappa,
		tau:        tau,
		limitSigma: limitSigma,
		balance:    balance,
	}
}

func (b BradlyTerryFullModel) Rate(teams [][]Rating, ranks, scores []int, weights [][]float64) ([][]Rating, error) {
	if err := checkRateParameters(teams, ranks, scores, weights); err != nil {
		return nil, err
	}

	originalTeams := make([][]Rating, len(teams))
	for i := range teams {
		originalTeams[i] = make([]Rating, len(teams[i]))
		for j := range teams[i] {
			originalTeams[i][j] = Rating{Mu: teams[i][j].Mu, Sigma: teams[i][j].Sigma}
		}
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
			teamsCopy[teamIndex][playerIndex].Sigma = player.Sigma + b.tau*b.tau
		}
	}

	if scores != nil {
		ranks = make([]int, len(scores))
		for i, score := range scores {
			ranks[i] = -score
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
		result = b.compute(teamsCopy, ranks, weights)
		result, _ = unwind(tenet, result)
	} else {
		result = b.compute(teamsCopy, nil, weights)
	}

	finalResult := make([][]Rating, len(result))
	for i := range result {
		finalResult[i] = make([]Rating, len(result[i]))
		for j := range result[i] {
			finalResult[i][j] = Rating{Mu: result[i][j].Mu, Sigma: result[i][j].Sigma}
		}
	}

	if b.limitSigma {
		for teamIndex, team := range finalResult {
			for playerIndex, player := range team {
				finalResult[teamIndex][playerIndex].Sigma = math.Min(player.Sigma, originalTeams[teamIndex][playerIndex].Sigma)
			}
		}
	}
	return finalResult, nil
}

func (b BradlyTerryFullModel) compute(teams [][]Rating, ranks []int, weights [][]float64) [][]Rating {
	originalTeams := make([][]Rating, len(teams))
	for i := range teams {
		originalTeams[i] = make([]Rating, len(teams[i]))
		copy(originalTeams[i], teams[i])
	}
	teamRatings := calculateTeamRatings(teams, ranks, b.balance, b.kappa)

	result := make([][]Rating, len(teamRatings))
	for i, t1 := range teamRatings {
		omega := 0.0
		delta := 0.0

		for q, t2 := range teamRatings {
			if q == i {
				continue
			}

			cIq := math.Sqrt(t1.SigmaSquared + t2.SigmaSquared + (2 * math.Pow(b.beta, 2)))
			piq := 1.0 / (1.0 + math.Exp((t2.Mu-t1.Mu)/cIq))
			sigmaSquaredToCiq := t1.SigmaSquared / cIq

			s := 0.0
			if t2.Rank > t1.Rank {
				s = 1.0
			} else if t2.Rank == t1.Rank {
				s = 0.5
			}

			omega += sigmaSquaredToCiq * (s - piq)
			gammaValue := math.Sqrt(t1.SigmaSquared / cIq)

			delta += ((gammaValue * sigmaSquaredToCiq) / cIq) * piq * (1 - piq)
		}

		intermediateResultPerTeam := make([]Rating, len(t1.Team))
		for j, r := range t1.Team {
			weight := 1.0
			if weights != nil && len(weights) > i && len(weights[i]) > j {
				weight = weights[i][j]
			}

			mu := r.Mu
			sigmaSq := r.Sigma * r.Sigma

			if omega > 0 {
				mu += (sigmaSq / t1.SigmaSquared) * omega * weight
				sigmaSq *= math.Sqrt(math.Max(1-(sigmaSq/t1.SigmaSquared)*delta*weight, b.kappa))
			} else {
				mu += (sigmaSq / t1.SigmaSquared) * omega / weight
				sigmaSq *= math.Sqrt(math.Max(1-(sigmaSq/t1.SigmaSquared)*delta/weight, b.kappa))
			}

			intermediateResultPerTeam[j] = Rating{Mu: mu, Sigma: r.Sigma}
		}
		result[i] = intermediateResultPerTeam
	}
	return result
}

type BradlyTerryPartialModel struct {
	mu         float64
	sigma      float64
	beta       float64
	kappa      float64
	tau        float64
	limitSigma bool
	balance    bool
}

func DefaultBradlyTerryPartialModel() Rater {
	return BradlyTerryPartialModel{
		mu:         25.0,
		sigma:      25.0 / 3.0,
		beta:       25.0 / 6.0,
		kappa:      0.0001,
		tau:        25.0 / 300.0,
		limitSigma: false,
		balance:    false,
	}
}

func NewBradlyTerryPartialModell(mu, sigma, beta, kappa, tau float64, limitSigma, balance bool) Rater {
	return BradlyTerryPartialModel{
		mu:         mu,
		sigma:      sigma,
		beta:       beta,
		kappa:      kappa,
		tau:        tau,
		limitSigma: limitSigma,
		balance:    balance,
	}
}

func (b BradlyTerryPartialModel) Rate(teams [][]Rating, ranks, scores []int, weights [][]float64) ([][]Rating, error) {
	if err := checkRateParameters(teams, ranks, scores, weights); err != nil {
		return nil, err
	}

	originalTeams := make([][]Rating, len(teams))
	for i := range teams {
		originalTeams[i] = make([]Rating, len(teams[i]))
		for j := range teams[i] {
			originalTeams[i][j] = Rating{Mu: teams[i][j].Mu, Sigma: teams[i][j].Sigma}
		}
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
			teamsCopy[teamIndex][playerIndex].Sigma = math.Sqrt(player.Sigma*player.Sigma + math.Pow(b.tau, 2))
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

	tenet := make([]int, len(ranks))
	for i := range ranks {
		tenet[i] = i
	}
	if ranks != nil {
		sort.Ints(ranks)
	}

	processedResult := make([][]Rating, len(teams))
	var result [][]Rating

	if ranks != nil {
		result = b.compute(teamsCopy, ranks, weights)
		unwoundResult, _ := unwind(tenet, result)
		for _, item := range unwoundResult {
			team := make([]Rating, len(item))
			copy(team, item)
			processedResult = append(processedResult, team)
		}
	} else {
		result = b.compute(teamsCopy, nil, weights)
		for _, item := range result {
			team := make([]Rating, len(item))
			copy(team, item)
			processedResult = append(processedResult, team)
		}
	}

	finalResult := processedResult

	if b.limitSigma {
		finalResult = make([][]Rating, len(processedResult))
		for teamIndex, team := range processedResult {
			finalTeam := make([]Rating, len(team))
			for playerIndex, player := range team {
				player.Sigma = math.Min(player.Sigma, originalTeams[teamIndex][playerIndex].Sigma)
				finalTeam[playerIndex] = player
			}
			finalResult[teamIndex] = finalTeam
		}
	}

	return finalResult, nil
}

func (b BradlyTerryPartialModel) compute(teams [][]Rating, ranks []int, weights [][]float64) [][]Rating {
	originalTeams := make([][]Rating, len(teams))
	for i := range teams {
		originalTeams[i] = make([]Rating, len(teams[i]))
		copy(originalTeams[i], teams[i])
	}
	teamRatings := calculateTeamRatings(teams, ranks, b.balance, b.kappa)
	adjacentTeams := ladderPairs(teamRatings)

	result := make([][]Rating, len(teamRatings))
	for i, t1 := range teamRatings {
		omega := 0.0
		delta := 0.0

		for q, t2 := range adjacentTeams[i] {
			if q == i {
				continue
			}

			cIq := math.Sqrt(t1.SigmaSquared + t2.SigmaSquared + (2 * math.Pow(b.beta, 2)))
			pIq := 1.0 / (1.0 + math.Exp((t2.Mu-t1.Mu)/cIq))
			sigmaSquaredToCiq := t1.SigmaSquared / cIq

			s := 0.0
			if t2.Rank > t1.Rank {
				s = 1.0
			} else if t2.Rank == t1.Rank {
				s = 0.5
			}

			omega += sigmaSquaredToCiq * (s - pIq)
			gammaValue := math.Sqrt(t1.SigmaSquared / cIq)

			delta += ((gammaValue * sigmaSquaredToCiq) / cIq) * pIq * (1 - pIq)
		}

		intermediateResultPerTeam := make([]Rating, len(t1.Team))
		for j, r := range t1.Team {
			weight := 1.0
			if weights != nil && len(weights) > i && len(weights[i]) > j {
				weight = weights[i][j]
			}

			mu := r.Mu
			sigma := r.Sigma

			if omega > 0 {
				mu += (math.Pow(sigma, 2) / t1.SigmaSquared) * omega * weight
				sigma *= math.Sqrt(math.Max(1-(math.Pow(sigma, 2)/t1.SigmaSquared)*delta*weight, b.kappa))
			} else {
				mu += (math.Pow(sigma, 2) / t1.SigmaSquared) * omega / weight
				sigma *= math.Sqrt(math.Max(1-(math.Pow(sigma, 2)/t1.SigmaSquared)*delta/weight, b.kappa))
			}

			intermediateResultPerTeam[j] = Rating{Mu: mu, Sigma: sigma}
		}
		result[i] = intermediateResultPerTeam
	}
	return result
}
