package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/UDL-TF/UnitedStats/pkg/events"
)

// Store handles all database operations
type Store struct {
	db *sql.DB
}

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// New creates a new database store
func New(cfg Config) (*Store, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &Store{db: db}, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// ============================================================================
// PLAYERS
// ============================================================================

// Player represents a player record
type Player struct {
	ID        int64
	SteamID   string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	LastSeen  time.Time

	MMR          int
	PeakMMR      int
	MMRUpdatedAt time.Time

	TotalKills     int
	TotalDeaths    int
	TotalAssists   int
	TotalAirshots  int
	TotalHeadshots int
	TotalBackstabs int
	TotalDeflects  int

	AvatarURL   sql.NullString
	CountryCode sql.NullString
}

// GetOrCreatePlayer gets or creates a player by steam ID
func (s *Store) GetOrCreatePlayer(ctx context.Context, steamID, name string) (*Player, error) {
	var player Player

	// Try to get existing player
	err := s.db.QueryRowContext(ctx, `
		SELECT id, steam_id, name, created_at, updated_at, last_seen,
		       mmr, peak_mmr, mmr_updated_at,
		       total_kills, total_deaths, total_assists,
		       total_airshots, total_headshots, total_backstabs, total_deflects,
		       avatar_url, country_code
		FROM players
		WHERE steam_id = $1
	`, steamID).Scan(
		&player.ID, &player.SteamID, &player.Name, &player.CreatedAt, &player.UpdatedAt, &player.LastSeen,
		&player.MMR, &player.PeakMMR, &player.MMRUpdatedAt,
		&player.TotalKills, &player.TotalDeaths, &player.TotalAssists,
		&player.TotalAirshots, &player.TotalHeadshots, &player.TotalBackstabs, &player.TotalDeflects,
		&player.AvatarURL, &player.CountryCode,
	)

	if err == nil {
		// Update name and last seen
		_, err = s.db.ExecContext(ctx, `
			UPDATE players 
			SET name = $1, last_seen = NOW(), updated_at = NOW()
			WHERE steam_id = $2
		`, name, steamID)
		return &player, err
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query player: %w", err)
	}

	// Create new player
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO players (steam_id, name)
		VALUES ($1, $2)
		RETURNING id, steam_id, name, created_at, updated_at, last_seen,
		          mmr, peak_mmr, mmr_updated_at,
		          total_kills, total_deaths, total_assists,
		          total_airshots, total_headshots, total_backstabs, total_deflects,
		          avatar_url, country_code
	`, steamID, name).Scan(
		&player.ID, &player.SteamID, &player.Name, &player.CreatedAt, &player.UpdatedAt, &player.LastSeen,
		&player.MMR, &player.PeakMMR, &player.MMRUpdatedAt,
		&player.TotalKills, &player.TotalDeaths, &player.TotalAssists,
		&player.TotalAirshots, &player.TotalHeadshots, &player.TotalBackstabs, &player.TotalDeflects,
		&player.AvatarURL, &player.CountryCode,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	return &player, nil
}

// GetPlayerBySteamID gets a player by steam ID
func (s *Store) GetPlayerBySteamID(ctx context.Context, steamID string) (*Player, error) {
	var player Player

	err := s.db.QueryRowContext(ctx, `
		SELECT id, steam_id, name, created_at, updated_at, last_seen,
		       mmr, peak_mmr, mmr_updated_at,
		       total_kills, total_deaths, total_assists,
		       total_airshots, total_headshots, total_backstabs, total_deflects,
		       avatar_url, country_code
		FROM players
		WHERE steam_id = $1
	`, steamID).Scan(
		&player.ID, &player.SteamID, &player.Name, &player.CreatedAt, &player.UpdatedAt, &player.LastSeen,
		&player.MMR, &player.PeakMMR, &player.MMRUpdatedAt,
		&player.TotalKills, &player.TotalDeaths, &player.TotalAssists,
		&player.TotalAirshots, &player.TotalHeadshots, &player.TotalBackstabs, &player.TotalDeflects,
		&player.AvatarURL, &player.CountryCode,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	return &player, nil
}

// ============================================================================
// MATCHES
// ============================================================================

// Match represents a match record
type Match struct {
	ID        int64
	UUID      string
	ServerIP  string
	Map       string
	Gamemode  string
	StartedAt time.Time
	EndedAt   sql.NullTime

	DurationSeconds sql.NullInt32
	WinnerTeam      sql.NullInt32
	RedScore        int
	BluScore        int

	TournamentID      sql.NullInt64
	TournamentMatchID sql.NullInt64

	CreatedAt time.Time
}

// CreateMatch creates a new match
func (s *Store) CreateMatch(ctx context.Context, serverIP, mapName, gamemode string, startedAt time.Time) (*Match, error) {
	var match Match

	err := s.db.QueryRowContext(ctx, `
		INSERT INTO matches (server_ip, map, gamemode, started_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, uuid, server_ip, map, gamemode, started_at, created_at
	`, serverIP, mapName, gamemode, startedAt).Scan(
		&match.ID, &match.UUID, &match.ServerIP, &match.Map, &match.Gamemode,
		&match.StartedAt, &match.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create match: %w", err)
	}

	return &match, nil
}

// GetOrCreateActiveMatch gets or creates an active match for a server
func (s *Store) GetOrCreateActiveMatch(ctx context.Context, serverIP, mapName, gamemode string) (*Match, error) {
	var match Match

	// Try to get active match (ended_at IS NULL)
	err := s.db.QueryRowContext(ctx, `
		SELECT id, uuid, server_ip, map, gamemode, started_at, ended_at,
		       duration_seconds, winner_team, red_score, blu_score,
		       tournament_id, tournament_match_id, created_at
		FROM matches
		WHERE server_ip = $1 AND ended_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1
	`, serverIP).Scan(
		&match.ID, &match.UUID, &match.ServerIP, &match.Map, &match.Gamemode,
		&match.StartedAt, &match.EndedAt, &match.DurationSeconds, &match.WinnerTeam,
		&match.RedScore, &match.BluScore, &match.TournamentID, &match.TournamentMatchID,
		&match.CreatedAt,
	)

	if err == nil {
		return &match, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query active match: %w", err)
	}

	// Create new match
	return s.CreateMatch(ctx, serverIP, mapName, gamemode, time.Now())
}

// EndMatch marks a match as ended
func (s *Store) EndMatch(ctx context.Context, matchID int64, winnerTeam, redScore, bluScore int) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE matches
		SET ended_at = NOW(),
		    duration_seconds = EXTRACT(EPOCH FROM (NOW() - started_at))::INTEGER,
		    winner_team = $2,
		    red_score = $3,
		    blu_score = $4
		WHERE id = $1
	`, matchID, winnerTeam, redScore, bluScore)

	if err != nil {
		return fmt.Errorf("failed to end match: %w", err)
	}

	return nil
}

// ============================================================================
// EVENTS
// ============================================================================

// InsertRawEvent inserts a raw event into the events table
func (s *Store) InsertRawEvent(ctx context.Context, eventType string, timestamp time.Time, serverIP, gamemode string, payload json.RawMessage) (int64, error) {
	var eventID int64

	err := s.db.QueryRowContext(ctx, `
		INSERT INTO events (event_type, timestamp, server_ip, gamemode, payload)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, eventType, timestamp, serverIP, gamemode, payload).Scan(&eventID)

	if err != nil {
		return 0, fmt.Errorf("failed to insert event: %w", err)
	}

	return eventID, nil
}

// MarkEventProcessed marks an event as processed
func (s *Store) MarkEventProcessed(ctx context.Context, eventID int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE events SET processed = TRUE WHERE id = $1
	`, eventID)
	return err
}

// ============================================================================
// KILLS
// ============================================================================

// InsertKill inserts a kill event
func (s *Store) InsertKill(ctx context.Context, kill *events.KillEvent, eventID, matchID int64) error {
	// Get or create players
	killer, err := s.GetOrCreatePlayer(ctx, kill.Killer.SteamID, kill.Killer.Name)
	if err != nil {
		return fmt.Errorf("failed to get/create killer: %w", err)
	}

	victim, err := s.GetOrCreatePlayer(ctx, kill.Victim.SteamID, kill.Victim.Name)
	if err != nil {
		return fmt.Errorf("failed to get/create victim: %w", err)
	}

	var assisterID sql.NullInt64
	if kill.Assister != nil {
		assister, err := s.GetOrCreatePlayer(ctx, kill.Assister.SteamID, kill.Assister.Name)
		if err != nil {
			return fmt.Errorf("failed to get/create assister: %w", err)
		}
		assisterID = sql.NullInt64{Int64: assister.ID, Valid: true}
	}

	// Insert kill
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO kills (
			event_id, match_id, killer_id, victim_id, assister_id,
			weapon, weapon_item_def_index, crit, airborne, headshot, backstab, first_blood,
			killer_pos_x, killer_pos_y, killer_pos_z,
			victim_pos_x, victim_pos_y, victim_pos_z,
			timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`,
		eventID, matchID, killer.ID, victim.ID, assisterID,
		kill.Weapon.Name, kill.Weapon.ItemDefIndex, kill.Crit, kill.Airborne,
		kill.Headshot, kill.Backstab, kill.FirstBlood,
		getPosVal(kill.KillerPos, "x"), getPosVal(kill.KillerPos, "y"), getPosVal(kill.KillerPos, "z"),
		getPosVal(kill.VictimPos, "x"), getPosVal(kill.VictimPos, "y"), getPosVal(kill.VictimPos, "z"),
		kill.Timestamp,
	)

	return err
}

// ============================================================================
// AIRSHOTS
// ============================================================================

// InsertAirshot inserts an airshot event
func (s *Store) InsertAirshot(ctx context.Context, airshot *events.AirshotEvent, eventID, matchID int64) error {
	player, err := s.GetOrCreatePlayer(ctx, airshot.Player.SteamID, airshot.Player.Name)
	if err != nil {
		return err
	}

	victim, err := s.GetOrCreatePlayer(ctx, airshot.Victim.SteamID, airshot.Victim.Name)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO airshots (event_id, match_id, player_id, victim_id, weapon_type, air2air, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, eventID, matchID, player.ID, victim.ID, airshot.WeaponType, airshot.Air2Air, airshot.Timestamp)

	return err
}

// ============================================================================
// DEFLECTS
// ============================================================================

// InsertDeflect inserts a deflect event
func (s *Store) InsertDeflect(ctx context.Context, deflect *events.DeflectEvent, eventID, matchID int64) error {
	player, err := s.GetOrCreatePlayer(ctx, deflect.Player.SteamID, deflect.Player.Name)
	if err != nil {
		return err
	}

	var ownerID sql.NullInt64
	if deflect.Owner != nil {
		owner, err := s.GetOrCreatePlayer(ctx, deflect.Owner.SteamID, deflect.Owner.Name)
		if err != nil {
			return err
		}
		ownerID = sql.NullInt64{Int64: owner.ID, Valid: true}
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO deflects (
			event_id, match_id, player_id, owner_id, projectile_type,
			rocket_speed, deflect_angle, timing_ms, distance, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, eventID, matchID, player.ID, ownerID, deflect.ProjectileType,
		floatPtr(deflect.RocketSpeed), floatPtr(deflect.DeflectAngle),
		intPtr(deflect.TimingMs), floatPtr(deflect.Distance), deflect.Timestamp)

	return err
}

// ============================================================================
// LEADERBOARD
// ============================================================================

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	Rank           int
	SteamID        string
	Name           string
	MMR            int
	PeakMMR        int
	TotalKills     int
	TotalDeaths    int
	KDRatio        float64
	TotalAirshots  int
	TotalHeadshots int
	TotalBackstabs int
	TotalDeflects  int
	LastSeen       time.Time
}

// GetLeaderboard gets the top players from the materialized view
func (s *Store) GetLeaderboard(ctx context.Context, limit, offset int) ([]*LeaderboardEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT rank, steam_id, name, mmr, peak_mmr, total_kills, total_deaths, kd_ratio,
		       total_airshots, total_headshots, total_backstabs, total_deflects, last_seen
		FROM leaderboard
		ORDER BY rank
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []*LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		err := rows.Scan(
			&entry.Rank, &entry.SteamID, &entry.Name, &entry.MMR, &entry.PeakMMR,
			&entry.TotalKills, &entry.TotalDeaths, &entry.KDRatio,
			&entry.TotalAirshots, &entry.TotalHeadshots, &entry.TotalBackstabs,
			&entry.TotalDeflects, &entry.LastSeen,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}
		entries = append(entries, &entry)
	}

	return entries, rows.Err()
}

// RefreshLeaderboard refreshes the leaderboard materialized view
func (s *Store) RefreshLeaderboard(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `SELECT refresh_leaderboard()`)
	return err
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

func getPosVal(pos *events.Position, axis string) sql.NullFloat64 {
	if pos == nil {
		return sql.NullFloat64{Valid: false}
	}

	var val float64
	switch axis {
	case "x":
		val = pos.X
	case "y":
		val = pos.Y
	case "z":
		val = pos.Z
	}

	return sql.NullFloat64{Float64: val, Valid: true}
}

func floatPtr(f float64) sql.NullFloat64 {
	if f == 0 {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}

func intPtr(i int) sql.NullInt32 {
	if i == 0 {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: int32(i), Valid: true}
}
