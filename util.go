package openskill

import (
	"math"
	"sort"
)

// teamRating is an intermediate struct used to calculate the ratings of a team of players.
type teamRating struct {
	Mu           float64
	SigmaSquared float64
	Team         []Rating
	Rank         int
}

func phiMajor(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt2))
}

func phiMajorInv(p float64) float64 {
	q := p - 0.5
	if math.Abs(q) <= 0.425 {
		r := 0.180625 - q*q
		num := (((((((2.5090809287301227e+3*r+
			3.3430575583588128e+4)*r+
			6.7265770927008709e+4)*r+
			4.5921953931549872e+4)*r+
			1.3731693765509461e+4)*r+
			1.9715909503065514e+3)*r+
			1.3314166789178437e+2)*r +
			3.3871328727963666e+0) * q
		den := (((((((5.2264952788528546e+3*r+
			2.8729085735721943e+4)*r+
			3.9307895800092710e+4)*r+
			2.1213794301586599e+4)*r+
			5.3941960214247511e+3)*r+
			6.8718700749205790e+2)*r+
			4.2313330701600911e+1)*r +
			1.0)
		return num / den
	}

	r := p
	if q > 0.0 {
		r = 1.0 - p
	}
	r = math.Sqrt(-math.Log(r))

	var num, den float64
	if r <= 5.0 {
		r = r - 1.6
		num = (((((((7.7454501427834141e-4*r+
			2.2723844989269185e-2)*r+
			2.4178072517745061e-1)*r+
			1.2704582524523684e+0)*r+
			3.6478483247632046e+0)*r+
			5.7694972214606914e+0)*r+
			4.6303378461565453e+0)*r +
			1.4234371107496838e+0)
		den = (((((((1.0507500716444168e-9*r+
			5.4759380849953445e-4)*r+
			1.5198666563616457e-2)*r+
			1.4810397642748007e-1)*r+
			6.8976733498510000e-1)*r+
			1.6763848301838038e+0)*r+
			2.0531916266377588e+0)*r +
			1.0)
	} else {
		r = r - 5.0
		num = (((((((2.0103343992922881e-7*r+
			2.7115555687434876e-5)*r+
			1.2426609473880784e-3)*r+
			2.6532189526576123e-2)*r+
			2.9656057182850489e-1)*r+
			1.7848265399172911e+0)*r+
			5.4637849111641144e+0)*r +
			6.6579046435011038e+0)
		den = (((((((2.0442631033899398e-15*r+
			1.4215117583164459e-7)*r+
			1.8463183175100547e-5)*r+
			7.8686913114561326e-4)*r+
			1.4875361290850615e-2)*r+
			1.3692988092273581e-1)*r+
			5.9983220655588794e-1)*r +
			1.0)
	}
	x := num / den
	if q < 0.0 {
		x = -x
	}
	return x
}

func phiMinor(x float64) float64 {
	return math.Exp(-0.5*x*x) / math.Sqrt(2*math.Pi)
}

func v(x, t float64) float64 {
	xt := x - t
	denominator := phiMajor(xt)

	if denominator < math.SmallestNonzeroFloat64 {
		return -xt
	}
	return phiMinor(xt) / denominator
}

func vt(x, t float64) float64 {
	xx := math.Abs(x)
	b := phiMajor(t-xx) - phiMajor(-t-xx)

	if b < 1e-5 {
		if x < 0 {
			return -x - t
		}
		return -x + t
	}

	a := phiMinor(-t-xx) - phiMinor(t-xx)
	if x < 0 {
		return -a / b
	}
	return a / b
}

func w(x, t float64) float64 {
	xt := x - t
	denominator := phiMajor(xt)

	if denominator < math.SmallestNonzeroFloat64 {
		if x < 0 {
			return 1
		}
		return 0
	}

	return v(x, t) * (v(x, t) + xt)
}

func wt(x, t float64) float64 {
	xx := math.Abs(x)
	b := phiMajor(t-xx) - phiMajor(-t-xx)

	if b < math.SmallestNonzeroFloat64 {
		return 1.0
	}

	numerator := (t-xx)*phiMinor(t-xx) + (t+xx)*phiMinor(-t-xx)
	return numerator/b + vt(x, t)*vt(x, t)
}

// unwind sorts the objects based on the tenet values and returns the sorted objects and their original indices.
func unwind[T any](tenet []int, objects []T) ([]T, []int) {
	if len(objects) == 0 {
		return []T{}, []int{}
	}

	type pair struct {
		Index  int
		Tenet  int
		Object T
	}

	pairs := make([]pair, len(objects))
	for i, obj := range objects {
		pairs[i] = pair{Index: i, Tenet: tenet[i], Object: obj}
	}

	sort.SliceStable(pairs, func(i, j int) bool {
		if pairs[i].Tenet != pairs[j].Tenet {
			return pairs[i].Tenet < pairs[j].Tenet
		}
		return pairs[i].Index < pairs[j].Index
	})

	sortedObjects := make([]T, len(objects))
	indices := make([]int, len(objects))
	for i, p := range pairs {
		sortedObjects[i] = p.Object
		indices[i] = p.Index
	}

	return sortedObjects, indices
}

// ladderPairs returns a list neighbours for every item in the list.
func ladderPairs[T any](list []T) [][]T {
	n := len(list)
	if n <= 1 {
		return [][]T{{}}
	}

	result := make([][]T, n)

	result[0] = []T{list[1]}
	for i := 1; i < n-1; i++ {
		result[i] = []T{list[i-1], list[i+1]}
	}
	result[n-1] = []T{list[n-2]}

	return result
}

// normalize scales the values in the vector to the range [targetMin, targetMax].
func normalize(vector []float64, targetMin, targetMax float64) []float64 {
	if len(vector) == 1 {
		return []float64{targetMax}
	}

	sourceMin := vector[0]
	sourceMax := vector[0]
	for _, v := range vector {
		if v < sourceMin {
			sourceMin = v
		}
		if v > sourceMax {
			sourceMax = v
		}
	}

	sourceRange := sourceMax - sourceMin
	if sourceRange == 0 {
		sourceRange = 0.0001
	}

	targetRange := targetMax - targetMin
	result := make([]float64, len(vector))
	for i, value := range vector {
		result[i] = (((value-sourceMin)/sourceRange)*targetRange + targetMin)
	}

	return result
}

// a returns the number of teams with the same rank for each team.
func a(teamRatings []teamRating) []int {
	rankCount := make(map[int]int, len(teamRatings))
	for _, t := range teamRatings {
		rankCount[t.Rank]++
	}

	result := make([]int, len(teamRatings))
	for i, t := range teamRatings {
		result[i] = rankCount[t.Rank]
	}

	return result
}

// c calculates the sum of the squares of the standard deviations of the teams.
func c(teamRatings []teamRating, beta float64) float64 {
	teamSigma := 0.0
	for _, team := range teamRatings {
		teamSigma += team.SigmaSquared + beta*beta
	}

	return math.Sqrt(teamSigma)
}

func sumQ(teamRatings []teamRating, c float64) []float64 {
	sumQ := make([]float64, len(teamRatings))
	for _, t1 := range teamRatings {
		summed := math.Exp(t1.Mu / c)
		for j, t2 := range teamRatings {
			if t1.Rank >= t2.Rank {
				sumQ[j] += summed
			}
		}
	}
	return sumQ
}

// calculateRankings finds the order of the teams based on a ranking
func calculateRankings(teams [][]Rating, ranks []int) []int {
	teamScores := make([]int, len(teams))
	rankOutput := make([]int, len(teams))
	s := 0

	if ranks != nil {
		for i := range teams {
			teamScores[i] = ranks[i]
		}
	} else {
		for i := range teams {
			teamScores[i] = i
		}
	}

	for i, score := range teamScores {
		if i > 0 {
			if teamScores[i-1] < score {
				s = i
			}
		}
		rankOutput[i] = s
	}

	return rankOutput
}

// calculateTeamRatings calculates the ratings of a team of players used for further computations.
func calculateTeamRatings(teams [][]Rating, ranks []int, balance bool, kappa float64) []teamRating {
	result := make([]teamRating, len(teams))
	rank := calculateRankings(teams, ranks)

	for i, team := range teams {
		sortedTeam := make([]Rating, len(team))
		copy(sortedTeam, team)
		sort.SliceStable(sortedTeam, func(i, j int) bool {
			return sortedTeam[i].Ordinal() > sortedTeam[j].Ordinal()
		})

		maxOrdinal := sortedTeam[0].Ordinal()
		muSummed := 0.0
		sigmaSqSummed := 0.0

		for _, player := range sortedTeam {
			balanceWeight := 1.0
			if balance {
				balanceWeight = 1 + ((maxOrdinal - player.Ordinal()) / (maxOrdinal + kappa))
			}
			muSummed += player.Mu * balanceWeight
			sigmaSqSummed += (player.Sigma * balanceWeight) * (player.Sigma * balanceWeight)
		}

		result[i] = teamRating{
			Mu:           muSummed,
			SigmaSquared: sigmaSqSummed,
			Team:         team,
			Rank:         rank[i],
		}
	}

	return result
}
