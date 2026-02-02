package mmr

import (
	"math"
)

// Calculator handles MMR/ELO rating calculations
type Calculator struct {
	// K-factor: determines how much ratings change per match
	// Higher K = more volatile ratings
	KFactor float64

	// Default MMR for new players
	DefaultMMR int
}

// NewCalculator creates a new MMR calculator with defaults
func NewCalculator() *Calculator {
	return &Calculator{
		KFactor:    32.0, // Standard chess K-factor
		DefaultMMR: 1000,
	}
}

// MatchResult represents the outcome of a match for a player
type MatchResult struct {
	PlayerMMR   int
	OpponentMMR int
	Won         bool
	TeamSize    int // Number of players on team (for team adjustment)
}

// Calculate calculates new MMR based on match result
func (c *Calculator) Calculate(result MatchResult) int {
	// Expected score (probability of winning) using Elo formula
	expected := c.expectedScore(float64(result.PlayerMMR), float64(result.OpponentMMR))

	// Actual score (1 for win, 0 for loss)
	actual := 0.0
	if result.Won {
		actual = 1.0
	}

	// Calculate rating change
	change := c.KFactor * (actual - expected)

	// Apply team size adjustment (larger teams = smaller individual impact)
	if result.TeamSize > 1 {
		teamFactor := 1.0 / math.Sqrt(float64(result.TeamSize))
		change *= teamFactor
	}

	// Calculate new MMR
	newMMR := float64(result.PlayerMMR) + change

	// Floor at 0 (can't go negative)
	if newMMR < 0 {
		newMMR = 0
	}

	return int(math.Round(newMMR))
}

// expectedScore calculates the expected probability of winning
// using the Elo formula: 1 / (1 + 10^((opponentRating - playerRating) / 400))
func (c *Calculator) expectedScore(playerRating, opponentRating float64) float64 {
	return 1.0 / (1.0 + math.Pow(10, (opponentRating-playerRating)/400.0))
}

// CalculateTeamMatch calculates MMR changes for a team-based match
func (c *Calculator) CalculateTeamMatch(winningTeam, losingTeam []int) (winnerChanges, loserChanges []int) {
	// Calculate average MMR for each team
	avgWinner := average(winningTeam)
	avgLoser := average(losingTeam)

	winnerChanges = make([]int, len(winningTeam))
	loserChanges = make([]int, len(losingTeam))

	// Calculate changes for winning team
	for i, playerMMR := range winningTeam {
		result := MatchResult{
			PlayerMMR:   playerMMR,
			OpponentMMR: int(avgLoser),
			Won:         true,
			TeamSize:    len(winningTeam),
		}
		newMMR := c.Calculate(result)
		winnerChanges[i] = newMMR - playerMMR
	}

	// Calculate changes for losing team
	for i, playerMMR := range losingTeam {
		result := MatchResult{
			PlayerMMR:   playerMMR,
			OpponentMMR: int(avgWinner),
			Won:         false,
			TeamSize:    len(losingTeam),
		}
		newMMR := c.Calculate(result)
		loserChanges[i] = newMMR - playerMMR
	}

	return winnerChanges, loserChanges
}

// average calculates the average of integers
func average(nums []int) float64 {
	if len(nums) == 0 {
		return 0
	}
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return float64(sum) / float64(len(nums))
}

// GetKFactorForExperience returns adjusted K-factor based on matches played
// New players have higher K-factor for faster calibration
func (c *Calculator) GetKFactorForExperience(matchesPlayed int) float64 {
	switch {
	case matchesPlayed < 10:
		return 50.0 // High volatility for new players
	case matchesPlayed < 30:
		return 40.0 // Medium-high for calibration
	case matchesPlayed < 100:
		return 32.0 // Standard
	default:
		return 24.0 // Lower for experienced players (more stable)
	}
}

// CalculateWithExperience calculates MMR with experience-adjusted K-factor
func (c *Calculator) CalculateWithExperience(result MatchResult, matchesPlayed int) int {
	originalK := c.KFactor
	c.KFactor = c.GetKFactorForExperience(matchesPlayed)
	newMMR := c.Calculate(result)
	c.KFactor = originalK // Restore original K-factor
	return newMMR
}
