# openskill

An implementation of the Open Skill rating system in Go.

Key Features:
- Any number of teams (>1)
- Uneven team sizes
- Predict outcomes before playing
- Per player contribution
- Accuracy on par with TrueSkill, but faster and open source
- No 3rd party dependencies

## Example
```go
	player1 := openskill.Rating{Mu: 25, Sigma: 3}
	player2 := openskill.Rating{Mu: 10, Sigma: 2.5}
	player3 := openskill.Rating{Mu: 5, Sigma: 2}
	player4 := openskill.Rating{Mu: 17, Sigma: 2}
	teams := [][]openskill.Rating{{player1}, {player2, player3}, {player4}}

	m := openskill.DefaultPlackettLuceModel()
	p := openskill.DefaultPredictor()

	// Predict the outcome of a match
	drawChance, _ := p.ChanceOfDraw(teams)                     // 0.18...
	winChance, _ := p.ChanceOfWinning(teams)                   // [0.59... 0.15... 0.24...]
	predictedRanks, probabilities, _ := p.ChanceOfRanks(teams) // [1 3 2] [0.59... 0.15... 0.24...]

	// Update the ratings based on a match where the first team won
	ranks := []int{1, 2, 3}
	teams, _ = m.Rate(teams, ranks, nil, nil)

	// Update the ratings based on a match where the first team won 10-5-3
	scores := []int{10, 5, 3}
	teams, _ = m.Rate(teams, nil, scores, nil)

	// Update the ratings based on a match where the second team won, but player 2 should get more credit
	ranks = []int{2, 1, 3}
	weights := [][]float64{{1}, {0.8, 0.2}, {1}}
	teams, _ = m.Rate(teams, ranks, nil, weights)
```


## Description
The package provides a simple interface which all of the models implement allowing for easy mocking or extension:
```go
type Rater interface {
	Rate(teams [][]Rating, ranks, scores []int, weights [][]float64) (updatedRatings [][]Rating, err error)
}
```

Use the `Default...Model()` methods to get started quickly or `New...Model(...)` if you want to tune the parameters yourself.

If you do not (want to) understand how the models work, `DefaultPlackettLuceModel()` is the recommended model, but feel free to experiment with what type of model or parameters works best for your type of matches. 

The package also provides a way to predict the outcome of matches between teams using the `Predictor` interface:
```go
type Predictor interface {
	ChanceOfWinning(teams [][]Rating) ([]float64, error)
	ChanceOfDraw(teams [][]Rating) (float64, error)
	ChanceOfRanks(teams [][]Rating) ([]int, []float64, error)
}
```

Use `DefaultPredictor()` to get started quickly or `NewPredictor(...)` if you want to tune the parameters yourself.
If you are using a custom model with `New...Model(...)`, your Predictor should be initialized with the same parameter values. 


## Implementations in other languages

- [Elixir](https://github.com/philihp/openskill.ex)
- [Kotlin](https://github.com/brezinajn/openskill.kt)
- [Lua](https://github.com/bstummer/openskill.lua)
- [Javascript](https://github.com/philihp/openskill.js)
- [Java](https://github.com/pocketcombats/openskill-java)
- [Python](https://github.com/vivekjoshy/openskill.py)

## References

This project is originally based on the [openskill.py](https://github.com/vivekjoshy/openskill.py) package. All of the models are based on the work in this [paper](https://jmlr.org/papers/v12/weng11a.html) or are the derivatives of algorithms found in it.

- Julia Ibstedt, Elsa Rådahl, Erik Turesson, and Magdalena vande Voorde. Application and further development of trueskill™ ranking in sports. 2019. URL: https://www.diva-portal.org/smash/get/diva2:1322103/FULLTEXT01.pdf.
- Ruby C. Weng and Chih-Jen Lin. A bayesian approximation method for online ranking. Journal of Machine Learning Research, 12(9):267–300, 2011. URL: http://jmlr.org/papers/v12/weng11a.html.