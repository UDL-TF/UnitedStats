# MMR System Documentation

## Overview

UnitedStats uses an **Elo-based MMR (Matchmaking Rating)** system to rank players based on match performance. The system is designed for team-based games (6v6, 9v9, etc.) with automatic adjustment for team sizes.

## How MMR Works

### Base Formula

The system uses the standard Elo formula:

```
NewRating = OldRating + K × (ActualScore - ExpectedScore)
```

Where:
- **K-factor**: Determines how much ratings change per match (default: 32)
- **ActualScore**: 1 for win, 0 for loss
- **ExpectedScore**: Probability of winning calculated from rating difference

### Expected Score Calculation

```
ExpectedScore = 1 / (1 + 10^((OpponentRating - PlayerRating) / 400))
```

This means:
- Equal ratings (1000 vs 1000) = 50% win probability
- +400 rating difference = ~90% win probability for favorite
- -400 rating difference = ~10% win probability for underdog

### Team Adjustment

For team games, individual rating changes are adjusted by:

```
TeamFactor = 1 / √(TeamSize)
```

This means:
- **1v1**: Full K-factor (32)
- **6v6**: K-factor × 0.408 ≈ 13 points
- **9v9**: K-factor × 0.333 ≈ 11 points

**Rationale**: Larger teams mean individual impact is diluted, so ratings should change more slowly.

## Rating Changes Examples

### Scenario 1: Equal Teams (6v6)
- **Team A** (avg 1200) vs **Team B** (avg 1200)
- Expected outcome: 50/50
- **If Team A wins**: Each player gains ~13 MMR
- **If Team A loses**: Each player loses ~13 MMR

### Scenario 2: Upset Win
- **Underdog** (1000) vs **Favorite** (1400)
- Expected win probability: ~10%
- **If underdog wins**: Gains ~29 MMR (huge upset bonus!)
- **If favorite wins**: Gains ~3 MMR (expected outcome)

### Scenario 3: Large Team Game (9v9)
- **Team A** (avg 1300) vs **Team B** (avg 1100)
- Team A expected to win: ~76%
- **If Team A wins**: Each player gains ~3 MMR
- **If Team A loses**: Each player loses ~22 MMR (major upset)

## Experience-Based K-Factor

New players have higher volatility for faster calibration:

| Matches Played | K-Factor | Rationale |
|---------------|----------|-----------|
| 0-9           | 50       | Placement matches - find true skill quickly |
| 10-29         | 40       | Calibration - still adjusting |
| 30-99         | 32       | Standard - settled rating |
| 100+          | 24       | Veteran - stable, experienced rating |

## Default MMR

- **New players start at**: 1000 MMR
- **Cannot go below**: 0 MMR
- **Peak MMR**: Tracked separately (lifetime best)

## Match Processing Flow

### 1. Match Start
```
- Match created in database
- Players registered to match_players table
- MMR snapshot taken (mmr_before)
```

### 2. During Match
```
- Events logged (kills, airshots, etc.)
- Stats accumulated
- No MMR changes yet
```

### 3. Match End
```
- Winner determined (team 2 or 3)
- Team average MMRs calculated
- Individual MMR changes computed
- Player MMRs updated
- match_players updated with mmr_after and mmr_change
- Peak MMR updated if new high
```

## Implementation Details

### Core Calculator (`internal/mmr/mmr.go`)

```go
calc := mmr.NewCalculator()

// Single match result
result := mmr.MatchResult{
    PlayerMMR:   1000,
    OpponentMMR: 1100,
    Won:         true,
    TeamSize:    6,
}
newMMR := calc.Calculate(result)

// Team match (6v6)
winningTeam := []int{1000, 1100, 1200, 1300, 1400, 1500}
losingTeam  := []int{1050, 1150, 1250, 1350, 1450, 1550}
winnerChanges, loserChanges := calc.CalculateTeamMatch(winningTeam, losingTeam)
```

### Database Schema

```sql
-- Players table
CREATE TABLE players (
    id BIGSERIAL PRIMARY KEY,
    mmr INTEGER DEFAULT 1000,
    peak_mmr INTEGER DEFAULT 1000,
    mmr_updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Match player tracking
CREATE TABLE match_players (
    match_id BIGINT,
    player_id BIGINT,
    team INTEGER,
    mmr_before INTEGER,
    mmr_after INTEGER,
    mmr_change INTEGER  -- Calculated as mmr_after - mmr_before
);
```

### Processor Integration

The event processor automatically calculates MMR when it receives a `match_end` event:

```go
// internal/processor/mmr.go
func (p *Processor) processMatchEndEvent(ctx, matchEnd) error {
    // 1. Get players from winning and losing teams
    // 2. Calculate average team MMRs
    // 3. Compute individual MMR changes
    // 4. Update player MMRs
    // 5. Record changes in match_players
    // 6. Update peak MMR if needed
}
```

## API Endpoints

### Get Player MMR
```bash
GET /api/v1/players/:steam_id/stats

Response:
{
  "mmr": {
    "current": 1234,
    "peak": 1567
  }
}
```

### Get Match MMR Changes
```bash
GET /api/v1/matches/:id

Response:
{
  "match": {...},
  "players": [
    {
      "steam_id": "76561198012345678",
      "name": "Player1",
      "team": 2,
      "mmr_before": 1200,
      "mmr_after": 1213,
      "mmr_change": +13
    }
  ]
}
```

### Leaderboard
```bash
GET /api/v1/leaderboard?limit=100

Response:
{
  "leaderboard": [
    {
      "rank": 1,
      "name": "TopPlayer",
      "mmr": 2345,
      "peak_mmr": 2456
    }
  ]
}
```

## Rating Tiers (Suggested)

| Tier | MMR Range | Description |
|------|-----------|-------------|
| 🥉 Bronze | 0-799 | Learning the game |
| 🥈 Silver | 800-999 | Developing skills |
| 🥇 Gold | 1000-1199 | Average player (starting MMR) |
| 💎 Platinum | 1200-1399 | Above average |
| 💠 Diamond | 1400-1599 | Skilled player |
| ⭐ Master | 1600-1799 | Very skilled |
| 👑 Grandmaster | 1800-1999 | Elite player |
| 🏆 Legend | 2000+ | Top tier |

*Note: Tier thresholds can be adjusted based on actual player distribution.*

## Advantages of This System

### 1. **Fair to Underdogs**
Beating a higher-rated team gives bigger MMR gains than beating equals.

### 2. **Team Size Aware**
6v6 competitive and 9v9 casual matches both work correctly.

### 3. **Fast Calibration**
New players reach their true rating quickly with high K-factor.

### 4. **Upset Protection**
Established players don't lose huge amounts from a single bad game.

### 5. **Peak Tracking**
Players can see their all-time best, even if current rating is lower.

## Potential Future Enhancements

### 1. **Class-Specific MMR**
Track separate ratings for Soldier, Scout, Medic, etc.

### 2. **Decay System**
Reduce MMR for inactive players (e.g., -5 per month of inactivity).

### 3. **Uncertainty Factor**
Add RD (rating deviation) like in Glicko-2 for players with few matches.

### 4. **Anti-Smurf Detection**
If a "new" player crushes opponents, boost their K-factor even more.

### 5. **Match Quality Calculation**
Predict match balance: `0.50 ± 0.15` = fair, `0.90` = stomp expected.

### 6. **Performance Bonuses**
Bonus MMR for MVP, carry performances, or exceptional stats.

### 7. **Map-Specific Ratings**
Some players are better on certain maps (cp_badlands vs koth_viaduct).

## Testing

Run MMR tests:

```bash
cd internal/mmr
go test -v

# Example output:
=== RUN   TestCalculator_Calculate
    mmr_test.go:60: equal rating win: 1000 -> 1016 (change: +16)
    mmr_test.go:60: equal rating loss: 1000 -> 984 (change: -16)
    mmr_test.go:60: underdog win (big gain): 1000 -> 1029 (change: +29)
=== RUN   TestCalculator_TeamSizeAdjustment
    mmr_test.go:119: Solo (1v1): +16 MMR
    mmr_test.go:120: Team (6v6): +7 MMR
```

## Configuration

### Adjust K-Factor

```go
// For more volatile ratings (faster movement)
calc := mmr.NewCalculator()
calc.KFactor = 40.0

// For more stable ratings (slower movement)
calc.KFactor = 24.0
```

### Change Default MMR

```go
calc := mmr.NewCalculator()
calc.DefaultMMR = 1200  // Start higher than 1000
```

### Modify Experience Brackets

Edit `internal/mmr/mmr.go`:

```go
func (c *Calculator) GetKFactorForExperience(matchesPlayed int) float64 {
    switch {
    case matchesPlayed < 5:
        return 60.0  // Even faster calibration
    case matchesPlayed < 20:
        return 45.0
    // ...
    }
}
```

## Mathematical Properties

### Zero-Sum
Total MMR in the system stays constant. When one team gains +50 MMR total, the other team loses -50 MMR total.

### Convergence
Players will converge to their "true skill" rating over time, with confidence increasing as matches played increases.

### Rating Difference vs Win Probability

| Rating Diff | Win % |
|-------------|-------|
| 0           | 50%   |
| +100        | 64%   |
| +200        | 76%   |
| +300        | 85%   |
| +400        | 91%   |
| +500        | 95%   |

## Debugging

### Check MMR Calculation
```sql
-- See recent MMR changes
SELECT 
    p.name,
    mp.mmr_before,
    mp.mmr_after,
    mp.mmr_change,
    m.winner_team,
    mp.team
FROM match_players mp
JOIN players p ON mp.player_id = p.id
JOIN matches m ON mp.match_id = m.id
WHERE mp.mmr_change IS NOT NULL
ORDER BY m.ended_at DESC
LIMIT 20;
```

### Verify MMR Distribution
```sql
-- MMR histogram
SELECT 
    FLOOR(mmr / 100) * 100 as mmr_bucket,
    COUNT(*) as player_count
FROM players
WHERE last_seen > NOW() - INTERVAL '30 days'
GROUP BY mmr_bucket
ORDER BY mmr_bucket;
```

---

**Built with fairness in mind. May your MMR rise! 🚀**
