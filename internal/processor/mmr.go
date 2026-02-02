package processor

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
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
	if err := p.updateTeamMMR(ctx, match.ID, winnerIDs, winnerMMRs, winnerChanges, "winner"); err != nil {
		return err
	}

	// Update loser MMRs
	if err := p.updateTeamMMR(ctx, match.ID, loserIDs, loserMMRs, loserChanges, "loser"); err != nil {
		return err
	}

	// End the match
	return p.store.EndMatch(ctx, match.ID, winnerTeam, 0, 0)
}

// updateTeamMMR updates MMR for all players on a team
func (p *Processor) updateTeamMMR(ctx context.Context, matchID int64, playerIDs []int64, oldMMRs, changes []int, team string) error {
	for i, playerID := range playerIDs {
		oldMMR := oldMMRs[i]
		newMMR := oldMMR + changes[i]

		// Update player MMR
		if err := p.store.UpdatePlayerMMR(ctx, playerID, newMMR); err != nil {
			p.logger.Error("Failed to update player MMR", err, watermill.LogFields{
				"team":      team,
				"player_id": playerID,
			})
			continue
		}

		// Update match_player MMR tracking
		if err := p.store.UpdateMatchPlayerMMR(ctx, matchID, playerID, oldMMR, newMMR); err != nil {
			p.logger.Error("Failed to update match player MMR", err, watermill.LogFields{
				"team":      team,
				"player_id": playerID,
			})
		}

		p.logger.Debug("Team MMR updated", watermill.LogFields{
			"team":      team,
			"player_id": playerID,
			"old_mmr":   oldMMR,
			"new_mmr":   newMMR,
			"change":    changes[i],
		})
	}
	return nil
}
