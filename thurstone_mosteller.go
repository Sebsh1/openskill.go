package openskill

import (
	"math"
	"sort"
)

type ThurstoneMostellerFullModel struct {
	mu         float64
	sigma      float64
	beta       float64
	kappa      float64
	tau        float64
	epsilon    float64
	limitSigma bool
	balance    bool
}

func DefaultThurstoneMostellerFullModel() Rater {
	return ThurstoneMostellerFullModel{
		mu:         25.0,
		sigma:      25.0 / 3.0,
		beta:       25.0 / 6.0,
		kappa:      0.0001,
		tau:        25.0 / 300.0,
		epsilon:    0.1,
		limitSigma: false,
		balance:    false,
	}
}

func NewThurstoneMostellerFullModel(mu, sigma, beta, kappa, tau, epsilon float64, limitSigma, balance bool) Rater {
	return ThurstoneMostellerFullModel{
		mu:         mu,
		sigma:      sigma,
		beta:       beta,
		kappa:      kappa,
		tau:        tau,
		epsilon:    epsilon,
		limitSigma: limitSigma,
		balance:    balance,
	}
}

func (t ThurstoneMostellerFullModel) Rate(teams [][]Rating, ranks, scores []int, weights [][]float64) ([][]Rating, error) {
	if err := checkRateParameters(teams, ranks, scores, weights); err != nil {
		return nil, err
	}

	if ranks == nil && scores != nil {
		ranks = make([]int, len(scores))
		for i, score := range scores {
			ranks[i] = -score
		}
	}

	for i := range weights {
		weights[i] = normalize(weights[i], 1, 2)
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
			teamsCopy[teamIndex][playerIndex].Sigma = math.Sqrt(player.Sigma*player.Sigma + math.Pow(t.tau, 2))
		}
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
		result = t.compute(teamsCopy, ranks, weights)
		unwoundResult, _ := unwind(tenet, result)
		for index, item := range unwoundResult {
			team := make([]Rating, len(item))
			copy(team, item)
			processedResult[index] = team
		}
	} else {
		result = t.compute(teamsCopy, nil, weights)
		for index, item := range result {
			team := make([]Rating, len(item))
			copy(team, item)
			processedResult[index] = team
		}

	}

	finalResult := processedResult

	if t.limitSigma {
		finalResult = make([][]Rating, len(processedResult))
		for teamIndex, team := range processedResult {
			finalTeam := make([]Rating, len(team))
			for playerIndex, player := range team {
				player.Sigma = math.Min(player.Sigma, teamsCopy[teamIndex][playerIndex].Sigma)
				finalTeam[playerIndex] = player
			}
			finalResult[teamIndex] = finalTeam
		}
	}
	return finalResult, nil
}

func (t ThurstoneMostellerFullModel) compute(teams [][]Rating, ranks []int, weights [][]float64) [][]Rating {
	originalTeams := make([][]Rating, len(teams))
	for i := range teams {
		originalTeams[i] = make([]Rating, len(teams[i]))
		copy(originalTeams[i], teams[i])
	}

	teamRatings := calculateTeamRatings(teams, ranks, t.balance, t.kappa)

	result := make([][]Rating, len(teamRatings))
	for i, teamIRating := range teamRatings {
		omega := 0.0
		delta := 0.0

		for q, teamQRating := range teamRatings {
			if q == i {
				continue
			}

			cIq := math.Sqrt(teamIRating.SigmaSquared + teamQRating.SigmaSquared + (2 * math.Pow(t.beta, 2)))
			deltaMu := (teamIRating.Mu - teamQRating.Mu) / cIq
			sigmaSquaredToCiq := teamIRating.SigmaSquared / cIq
			gamma := math.Sqrt(teamIRating.SigmaSquared) / cIq

			if teamQRating.Rank > teamIRating.Rank {
				omega += sigmaSquaredToCiq * v(deltaMu, t.epsilon)
				delta += (gamma * sigmaSquaredToCiq / cIq) * w(deltaMu, t.epsilon)
			} else if teamQRating.Rank < teamIRating.Rank {
				omega += -sigmaSquaredToCiq * v(-deltaMu, t.epsilon)
				delta += (gamma * sigmaSquaredToCiq / cIq) * w(-deltaMu, t.epsilon)
			} else {
				omega += sigmaSquaredToCiq * vt(deltaMu, t.epsilon)
				delta += (gamma * sigmaSquaredToCiq / cIq) * wt(deltaMu, t.epsilon)
			}
		}

		intermediateResultPerTeam := make([]Rating, len(teamIRating.Team))
		for j, jPlayers := range teamIRating.Team {
			weight := 1.0
			if weights != nil && len(weights) > i && len(weights[i]) > j {
				weight = weights[i][j]
			}

			mu := jPlayers.Mu
			sigma := jPlayers.Sigma

			if omega > 0 {
				mu += (sigma * sigma / teamIRating.SigmaSquared) * omega * weight
				sigma *= math.Sqrt(math.Max(1-(sigma*sigma/teamIRating.SigmaSquared)*delta*weight, t.kappa))
			} else {
				mu += (sigma * sigma / teamIRating.SigmaSquared) * omega / weight
				sigma *= math.Sqrt(math.Max(1-(sigma*sigma/teamIRating.SigmaSquared)*delta/weight, t.kappa))
			}
			intermediateResultPerTeam[j] = Rating{Mu: mu, Sigma: sigma}
		}
		result[i] = intermediateResultPerTeam
	}
	return result
}

type ThurstoneMostellerPartialModel struct {
	mu         float64
	sigma      float64
	beta       float64
	kappa      float64
	tau        float64
	epsilon    float64
	limitSigma bool
	balance    bool
}

func DefaultThurstoneMostellerPartialModel() Rater {
	return ThurstoneMostellerPartialModel{
		mu:         25.0,
		sigma:      25.0 / 3.0,
		beta:       25.0 / 6.0,
		kappa:      0.0001,
		tau:        25.0 / 300.0,
		epsilon:    0.1,
		limitSigma: false,
		balance:    false,
	}
}

func NewThurstoneMostellerPartialModel(mu, sigma, beta, kappa, epsilon, tau float64, limitSigma, balance bool) Rater {
	return ThurstoneMostellerPartialModel{
		mu:         mu,
		sigma:      sigma,
		beta:       beta,
		kappa:      kappa,
		tau:        tau,
		epsilon:    epsilon,
		limitSigma: limitSigma,
		balance:    balance,
	}
}

func (t ThurstoneMostellerPartialModel) Rate(teams [][]Rating, ranks, scores []int, weights [][]float64) ([][]Rating, error) {
	if err := checkRateParameters(teams, ranks, scores, weights); err != nil {
		return nil, err
	}

	if ranks == nil && scores != nil {
		ranks = make([]int, len(scores))
		for i, score := range scores {
			ranks[i] = -score
		}
	}

	for i := range weights {
		weights[i] = normalize(weights[i], 1, 2)
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
			teamsCopy[teamIndex][playerIndex].Sigma = math.Sqrt(player.Sigma*player.Sigma + math.Pow(t.tau, 2))
		}
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
		result = t.compute(teamsCopy, ranks, weights)
		unwoundResult, _ := unwind(tenet, result)
		for index, item := range unwoundResult {
			team := make([]Rating, len(item))
			copy(team, item)
			processedResult[index] = team
		}
	} else {
		result = t.compute(teamsCopy, nil, weights)
		for index, item := range result {
			team := make([]Rating, len(item))
			copy(team, item)
			processedResult[index] = team
		}

	}

	finalResult := processedResult

	if t.limitSigma {
		finalResult = make([][]Rating, len(processedResult))
		for teamIndex, team := range processedResult {
			finalTeam := make([]Rating, len(team))
			for playerIndex, player := range team {
				player.Sigma = math.Min(player.Sigma, teamsCopy[teamIndex][playerIndex].Sigma)
				finalTeam[playerIndex] = player
			}
			finalResult[teamIndex] = finalTeam
		}
	}
	return finalResult, nil
}

func (t ThurstoneMostellerPartialModel) compute(teams [][]Rating, ranks []int, weights [][]float64) [][]Rating {
	originalTeams := make([][]Rating, len(teams))
	for i := range teams {
		originalTeams[i] = make([]Rating, len(teams[i]))
		copy(originalTeams[i], teams[i])
	}
	teamRatings := calculateTeamRatings(teams, ranks, t.balance, t.kappa)
	adjacentTeams := ladderPairs(teamRatings)

	result := make([][]Rating, len(teamRatings))
	for i, t1 := range teamRatings {
		omega := 0.0
		delta := 0.0

		for j, t2 := range adjacentTeams[i] {
			if j == i {
				continue
			}

			c := 2 * math.Sqrt(t1.SigmaSquared+t2.SigmaSquared+(2*math.Pow(t.beta, 2)))
			deltaMu := (t1.Mu - t2.Mu) / c
			sigmaSquaredToC := t1.SigmaSquared / c
			gamma := math.Sqrt(t1.SigmaSquared / c)

			if t2.Rank > t1.Rank {
				omega += sigmaSquaredToC * v(deltaMu, t.epsilon/c)
				delta += (gamma * sigmaSquaredToC / c) * w(deltaMu, t.epsilon/c)
			} else if t2.Rank < t1.Rank {
				omega += -sigmaSquaredToC * v(-deltaMu, t.epsilon/c)
				delta += (gamma * sigmaSquaredToC / c) * w(-deltaMu, t.epsilon/c)
			} else {
				omega += sigmaSquaredToC * vt(deltaMu, t.epsilon/c)
				delta += (gamma * sigmaSquaredToC / c) * wt(deltaMu, t.epsilon/c)
			}
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
				mu += (sigma * sigma / t1.SigmaSquared) * omega * weight
				sigma *= math.Sqrt(math.Max(1-(sigma*sigma/t1.SigmaSquared)*delta*weight, t.kappa))
			} else {
				mu += (sigma * sigma / t1.SigmaSquared) * omega / weight
				sigma *= math.Sqrt(math.Max(1-(sigma*sigma/t1.SigmaSquared)*delta/weight, t.kappa))
			}

			intermediateResultPerTeam[j] = Rating{Mu: mu, Sigma: sigma}
		}
		result[i] = intermediateResultPerTeam
	}
	return result
}
