package mmr

import (
	"testing"
)

func TestCalculator_Calculate(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name           string
		result         MatchResult
		expectedChange string // "positive", "negative", or specific value
	}{
		{
			name: "equal rating win",
			result: MatchResult{
				PlayerMMR:   1000,
				OpponentMMR: 1000,
				Won:         true,
				TeamSize:    1,
			},
			expectedChange: "positive",
		},
		{
			name: "equal rating loss",
			result: MatchResult{
				PlayerMMR:   1000,
				OpponentMMR: 1000,
				Won:         false,
				TeamSize:    1,
			},
			expectedChange: "negative",
		},
		{
			name: "underdog win (big gain)",
			result: MatchResult{
				PlayerMMR:   1000,
				OpponentMMR: 1400,
				Won:         true,
				TeamSize:    1,
			},
			expectedChange: "positive",
		},
		{
			name: "favorite loss (big loss)",
			result: MatchResult{
				PlayerMMR:   1400,
				OpponentMMR: 1000,
				Won:         false,
				TeamSize:    1,
			},
			expectedChange: "negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newMMR := calc.Calculate(tt.result)
			change := newMMR - tt.result.PlayerMMR

			switch tt.expectedChange {
			case "positive":
				if change <= 0 {
					t.Errorf("Expected positive change, got %d (new: %d, old: %d)",
						change, newMMR, tt.result.PlayerMMR)
				}
			case "negative":
				if change >= 0 {
					t.Errorf("Expected negative change, got %d (new: %d, old: %d)",
						change, newMMR, tt.result.PlayerMMR)
				}
			}

			t.Logf("%s: %d -> %d (change: %+d)",
				tt.name, tt.result.PlayerMMR, newMMR, change)
		})
	}
}

func TestCalculator_UnderdogVsFavorite(t *testing.T) {
	calc := NewCalculator()

	// Underdog (1000) beats favorite (1400)
	underdogWin := calc.Calculate(MatchResult{
		PlayerMMR:   1000,
		OpponentMMR: 1400,
		Won:         true,
		TeamSize:    1,
	})
	underdogGain := underdogWin - 1000

	// Favorite (1400) beats underdog (1000)
	favoriteWin := calc.Calculate(MatchResult{
		PlayerMMR:   1400,
		OpponentMMR: 1000,
		Won:         true,
		TeamSize:    1,
	})
	favoriteGain := favoriteWin - 1400

	// Underdog should gain more for beating favorite than favorite gains for beating underdog
	if underdogGain <= favoriteGain {
		t.Errorf("Underdog gain (%d) should be > favorite gain (%d) for equal wins",
			underdogGain, favoriteGain)
	}

	t.Logf("Underdog beats favorite: +%d MMR", underdogGain)
	t.Logf("Favorite beats underdog: +%d MMR", favoriteGain)
}

func TestCalculator_TeamSizeAdjustment(t *testing.T) {
	calc := NewCalculator()

	// Same match result, different team sizes
	solo := calc.Calculate(MatchResult{
		PlayerMMR:   1000,
		OpponentMMR: 1000,
		Won:         true,
		TeamSize:    1,
	})
	soloChange := solo - 1000

	sixv6 := calc.Calculate(MatchResult{
		PlayerMMR:   1000,
		OpponentMMR: 1000,
		Won:         true,
		TeamSize:    6,
	})
	teamChange := sixv6 - 1000

	// Solo should gain more than team member
	if soloChange <= teamChange {
		t.Errorf("Solo change (%d) should be > team change (%d)", soloChange, teamChange)
	}

	t.Logf("Solo (1v1): +%d MMR", soloChange)
	t.Logf("Team (6v6): +%d MMR", teamChange)
}

func TestCalculator_CalculateTeamMatch(t *testing.T) {
	calc := NewCalculator()

	winningTeam := []int{1000, 1100, 1200}
	losingTeam := []int{1050, 1150, 1250}

	winnerChanges, loserChanges := calc.CalculateTeamMatch(winningTeam, losingTeam)

	// Verify we got changes for all players
	if len(winnerChanges) != len(winningTeam) {
		t.Errorf("Expected %d winner changes, got %d", len(winningTeam), len(winnerChanges))
	}
	if len(loserChanges) != len(losingTeam) {
		t.Errorf("Expected %d loser changes, got %d", len(losingTeam), len(loserChanges))
	}

	// Winners should gain MMR
	for i, change := range winnerChanges {
		if change <= 0 {
			t.Errorf("Winner %d should gain MMR, got %+d", i, change)
		}
		t.Logf("Winner %d: %d -> %d (%+d)",
			i, winningTeam[i], winningTeam[i]+change, change)
	}

	// Losers should lose MMR
	for i, change := range loserChanges {
		if change >= 0 {
			t.Errorf("Loser %d should lose MMR, got %+d", i, change)
		}
		t.Logf("Loser %d: %d -> %d (%+d)",
			i, losingTeam[i], losingTeam[i]+change, change)
	}
}

func TestCalculator_ExperienceAdjustment(t *testing.T) {
	calc := NewCalculator()

	result := MatchResult{
		PlayerMMR:   1000,
		OpponentMMR: 1000,
		Won:         true,
		TeamSize:    1,
	}

	// New player (5 matches)
	newPlayerMMR := calc.CalculateWithExperience(result, 5)
	newPlayerChange := newPlayerMMR - 1000

	// Experienced player (200 matches)
	veteranMMR := calc.CalculateWithExperience(result, 200)
	veteranChange := veteranMMR - 1000

	// New players should have larger rating changes
	if newPlayerChange <= veteranChange {
		t.Errorf("New player change (%d) should be > veteran change (%d)",
			newPlayerChange, veteranChange)
	}

	t.Logf("New player (5 matches): +%d MMR", newPlayerChange)
	t.Logf("Veteran (200 matches): +%d MMR", veteranChange)
}

func TestCalculator_NoNegativeMMR(t *testing.T) {
	calc := NewCalculator()

	// Player at 10 MMR loses to much higher opponent
	result := MatchResult{
		PlayerMMR:   10,
		OpponentMMR: 2000,
		Won:         false,
		TeamSize:    1,
	}

	newMMR := calc.Calculate(result)

	if newMMR < 0 {
		t.Errorf("MMR should not go negative, got %d", newMMR)
	}

	t.Logf("Low MMR player: 10 -> %d", newMMR)
}

func BenchmarkCalculate(b *testing.B) {
	calc := NewCalculator()
	result := MatchResult{
		PlayerMMR:   1000,
		OpponentMMR: 1100,
		Won:         true,
		TeamSize:    6,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calc.Calculate(result)
	}
}

func BenchmarkCalculateTeamMatch(b *testing.B) {
	calc := NewCalculator()
	winningTeam := []int{1000, 1100, 1200, 1300, 1400, 1500}
	losingTeam := []int{1050, 1150, 1250, 1350, 1450, 1550}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calc.CalculateTeamMatch(winningTeam, losingTeam)
	}
}
