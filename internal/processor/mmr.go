package processor

import (
	"context"
	"fmt"

	"github.com/UDL-TF/UnitedStats/internal/mmr"
	"github.com/UDL-TF/UnitedStats/pkg/events"
)

// processMatchEndEvent processes a match end event and calculates MMR
func (p *Processor) processMatchEndEvent(ctx context.Context, matchEnd *events.MatchEndEvent) error {
	// Get active match
	match, err := p.store.GetOrCreateActiveMatch(ctx, matchEnd.ServerIP, "", matchEnd.Gamemode)
	if err != nil {
		return err
	}

	// Determine winner team (2=RED, 3=BLU)
	winnerTeam := matchEnd.WinnerTeam
	loserTeam := 3 // BLU
	if winnerTeam == 3 {
		loserTeam = 2 // RED
	}

	// Get players from both teams
	winnerIDs, winnerMMRs, err := p.store.GetMatchTeamPlayers(ctx, match.ID, winnerTeam)
	if err != nil {
		return fmt.Errorf("failed to get winner team: %w", err)
	}

	loserIDs, loserMMRs, err := p.store.GetMatchTeamPlayers(ctx, match.ID, loserTeam)
	if err != nil {
		return fmt.Errorf("failed to get loser team: %w", err)
	}

	// Skip MMR calculation if teams are empty
	if len(winnerIDs) == 0 || len(loserIDs) == 0 {
		p.logger.Info("Skipping MMR calculation - empty teams", nil)
		return p.store.EndMatch(ctx, match.ID, winnerTeam, 0, 0)
	}

	// Calculate MMR changes
	calc := mmr.NewCalculator()
	winnerChanges, loserChanges := calc.CalculateTeamMatch(winnerMMRs, loserMMRs)

	// Update winner MMRs
	for i, playerID := range winnerIDs {
		oldMMR := winnerMMRs[i]
		newMMR := oldMMR + winnerChanges[i]

		// Update player MMR
		if err := p.store.UpdatePlayerMMR(ctx, playerID, newMMR); err != nil {
			p.logger.Error("Failed to update winner MMR", err, nil)
			continue
		}

		// Update match_player MMR tracking
		if err := p.store.UpdateMatchPlayerMMR(ctx, match.ID, playerID, oldMMR, newMMR); err != nil {
			p.logger.Error("Failed to update match player MMR", err, nil)
		}

		p.logger.Debug("Winner MMR updated", map[string]interface{}{
			"player_id": playerID,
			"old_mmr":   oldMMR,
			"new_mmr":   newMMR,
			"change":    winnerChanges[i],
		})
	}

	// Update loser MMRs
	for i, playerID := range loserIDs {
		oldMMR := loserMMRs[i]
		newMMR := oldMMR + loserChanges[i]

		// Update player MMR
		if err := p.store.UpdatePlayerMMR(ctx, playerID, newMMR); err != nil {
			p.logger.Error("Failed to update loser MMR", err, nil)
			continue
		}

		// Update match_player MMR tracking
		if err := p.store.UpdateMatchPlayerMMR(ctx, match.ID, playerID, oldMMR, newMMR); err != nil {
			p.logger.Error("Failed to update match player MMR", err, nil)
		}

		p.logger.Debug("Loser MMR updated", map[string]interface{}{
			"player_id": playerID,
			"old_mmr":   oldMMR,
			"new_mmr":   newMMR,
			"change":    loserChanges[i],
		})
	}

	// End the match
	return p.store.EndMatch(ctx, match.ID, winnerTeam, 0, 0)
}
