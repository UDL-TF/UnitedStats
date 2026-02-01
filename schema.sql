-- UnitedStats Database Schema v3.0
-- PostgreSQL 16+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- PLAYERS & AUTHENTICATION
-- ============================================================================

CREATE TABLE players (
    id BIGSERIAL PRIMARY KEY,
    steam_id VARCHAR(32) UNIQUE NOT NULL,
    name VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- MMR & Rankings
    mmr INTEGER DEFAULT 1000,
    peak_mmr INTEGER DEFAULT 1000,
    mmr_updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Stats summary (updated via triggers)
    total_kills INTEGER DEFAULT 0,
    total_deaths INTEGER DEFAULT 0,
    total_assists INTEGER DEFAULT 0,
    total_airshots INTEGER DEFAULT 0,
    total_headshots INTEGER DEFAULT 0,
    total_backstabs INTEGER DEFAULT 0,
    total_deflects INTEGER DEFAULT 0,
    
    -- Profile
    avatar_url VARCHAR(255),
    country_code VARCHAR(2),
    
    INDEX idx_steam_id (steam_id),
    INDEX idx_mmr (mmr DESC),
    INDEX idx_last_seen (last_seen DESC)
);

-- ============================================================================
-- MATCHES
-- ============================================================================

CREATE TABLE matches (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    
    -- Match info
    server_ip VARCHAR(45) NOT NULL,
    map VARCHAR(64) NOT NULL,
    gamemode VARCHAR(32) NOT NULL,
    
    -- Timing
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ended_at TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    
    -- Results
    winner_team INTEGER, -- 2=RED, 3=BLU, 0=tie
    red_score INTEGER DEFAULT 0,
    blu_score INTEGER DEFAULT 0,
    
    -- Tournament link (nullable)
    tournament_id BIGINT,
    tournament_match_id BIGINT,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_started_at (started_at DESC),
    INDEX idx_server_ip (server_ip),
    INDEX idx_tournament (tournament_id, tournament_match_id)
);

-- ============================================================================
-- PLAYER MATCH PARTICIPATION
-- ============================================================================

CREATE TABLE match_players (
    id BIGSERIAL PRIMARY KEY,
    match_id BIGINT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    
    -- Team & Class
    team INTEGER NOT NULL, -- 2=RED, 3=BLU
    primary_class VARCHAR(32), -- Most played class
    
    -- Stats
    kills INTEGER DEFAULT 0,
    deaths INTEGER DEFAULT 0,
    assists INTEGER DEFAULT 0,
    damage_dealt INTEGER DEFAULT 0,
    healing_done INTEGER DEFAULT 0,
    
    -- Special achievements
    airshots INTEGER DEFAULT 0,
    headshots INTEGER DEFAULT 0,
    backstabs INTEGER DEFAULT 0,
    deflects INTEGER DEFAULT 0,
    
    -- MMR change (calculated at match end)
    mmr_before INTEGER,
    mmr_after INTEGER,
    mmr_change INTEGER,
    
    UNIQUE(match_id, player_id),
    INDEX idx_match_id (match_id),
    INDEX idx_player_id (player_id)
);

-- ============================================================================
-- EVENTS (Raw event log)
-- ============================================================================

CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    
    -- Event metadata
    event_type VARCHAR(32) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Match context
    match_id BIGINT REFERENCES matches(id) ON DELETE SET NULL,
    server_ip VARCHAR(45) NOT NULL,
    gamemode VARCHAR(32) NOT NULL,
    
    -- Raw JSON payload
    payload JSONB NOT NULL,
    
    -- Processed flag
    processed BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_event_type (event_type),
    INDEX idx_timestamp (timestamp DESC),
    INDEX idx_match_id (match_id),
    INDEX idx_processed (processed) WHERE NOT processed,
    INDEX idx_payload_gin (payload) USING GIN
);

-- ============================================================================
-- KILLS (Detailed kill log)
-- ============================================================================

CREATE TABLE kills (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT REFERENCES events(id) ON DELETE CASCADE,
    match_id BIGINT REFERENCES matches(id) ON DELETE CASCADE,
    
    -- Players
    killer_id BIGINT REFERENCES players(id) ON DELETE CASCADE,
    victim_id BIGINT REFERENCES players(id) ON DELETE CASCADE,
    assister_id BIGINT,
    
    -- Weapon & properties
    weapon VARCHAR(64) NOT NULL,
    weapon_item_def_index INTEGER,
    
    crit BOOLEAN DEFAULT FALSE,
    airborne BOOLEAN DEFAULT FALSE,
    headshot BOOLEAN DEFAULT FALSE,
    backstab BOOLEAN DEFAULT FALSE,
    first_blood BOOLEAN DEFAULT FALSE,
    
    -- Position data
    killer_pos_x REAL,
    killer_pos_y REAL,
    killer_pos_z REAL,
    victim_pos_x REAL,
    victim_pos_y REAL,
    victim_pos_z REAL,
    
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    INDEX idx_killer_id (killer_id),
    INDEX idx_victim_id (victim_id),
    INDEX idx_match_id (match_id),
    INDEX idx_timestamp (timestamp DESC),
    INDEX idx_weapon (weapon),
    INDEX idx_headshot (headshot) WHERE headshot = TRUE,
    INDEX idx_backstab (backstab) WHERE backstab = TRUE
);

-- ============================================================================
-- AIRSHOTS
-- ============================================================================

CREATE TABLE airshots (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT REFERENCES events(id) ON DELETE CASCADE,
    match_id BIGINT REFERENCES matches(id) ON DELETE CASCADE,
    
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    victim_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    
    weapon_type VARCHAR(32) NOT NULL, -- rocket, sticky, pipebomb, arrow, flare, stun
    air2air BOOLEAN DEFAULT FALSE,
    
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    INDEX idx_player_id (player_id),
    INDEX idx_weapon_type (weapon_type),
    INDEX idx_air2air (air2air) WHERE air2air = TRUE,
    INDEX idx_timestamp (timestamp DESC)
);

-- ============================================================================
-- DEFLECTS
-- ============================================================================

CREATE TABLE deflects (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT REFERENCES events(id) ON DELETE CASCADE,
    match_id BIGINT REFERENCES matches(id) ON DELETE CASCADE,
    
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    owner_id BIGINT, -- Original projectile owner
    
    projectile_type VARCHAR(32) NOT NULL,
    
    -- Dodgeball specific
    rocket_speed REAL,
    deflect_angle REAL,
    timing_ms INTEGER,
    distance REAL,
    
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    INDEX idx_player_id (player_id),
    INDEX idx_projectile_type (projectile_type),
    INDEX idx_timestamp (timestamp DESC)
);

-- ============================================================================
-- TOURNAMENTS
-- ============================================================================

CREATE TABLE tournaments (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    
    name VARCHAR(128) NOT NULL,
    format VARCHAR(32) NOT NULL, -- swiss, single_elimination, double_elimination
    
    -- Dates
    registration_start TIMESTAMP WITH TIME ZONE,
    registration_end TIMESTAMP WITH TIME ZONE,
    tournament_start TIMESTAMP WITH TIME ZONE,
    tournament_end TIMESTAMP WITH TIME ZONE,
    
    -- Settings
    max_teams INTEGER,
    match_format VARCHAR(32), -- bo1, bo3, bo5
    map_pool TEXT[], -- Array of map names
    
    -- Status
    status VARCHAR(32) DEFAULT 'draft', -- draft, registration, active, completed, cancelled
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_status (status),
    INDEX idx_tournament_start (tournament_start DESC)
);

CREATE TABLE tournament_teams (
    id BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    
    team_name VARCHAR(64) NOT NULL,
    seed INTEGER,
    
    -- Swiss system
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    
    -- Bracket system
    eliminated BOOLEAN DEFAULT FALSE,
    current_bracket VARCHAR(16), -- upper, lower (for double elimination)
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(tournament_id, team_name),
    INDEX idx_tournament_id (tournament_id),
    INDEX idx_wins_losses (wins DESC, losses ASC)
);

CREATE TABLE tournament_matches (
    id BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    
    round_number INTEGER NOT NULL,
    match_number INTEGER NOT NULL,
    
    team_a_id BIGINT REFERENCES tournament_teams(id),
    team_b_id BIGINT REFERENCES tournament_teams(id),
    
    -- Server assignment
    assigned_server VARCHAR(64),
    
    -- Results
    winner_team_id BIGINT REFERENCES tournament_teams(id),
    score_a INTEGER DEFAULT 0,
    score_b INTEGER DEFAULT 0,
    
    -- Link to actual match
    match_id BIGINT REFERENCES matches(id),
    
    -- Status
    status VARCHAR(32) DEFAULT 'pending', -- pending, in_progress, completed, cancelled
    
    scheduled_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(tournament_id, round_number, match_number),
    INDEX idx_tournament_id (tournament_id),
    INDEX idx_status (status)
);

-- ============================================================================
-- MATERIALIZED VIEWS (for performance)
-- ============================================================================

-- Global leaderboard (refreshed every 5 minutes)
CREATE MATERIALIZED VIEW leaderboard AS
SELECT 
    p.id,
    p.steam_id,
    p.name,
    p.mmr,
    p.peak_mmr,
    p.total_kills,
    p.total_deaths,
    CASE 
        WHEN p.total_deaths > 0 THEN ROUND(p.total_kills::NUMERIC / p.total_deaths, 2)
        ELSE p.total_kills::NUMERIC
    END as kd_ratio,
    p.total_airshots,
    p.total_headshots,
    p.total_backstabs,
    p.total_deflects,
    p.last_seen,
    ROW_NUMBER() OVER (ORDER BY p.mmr DESC) as rank
FROM players p
WHERE p.last_seen > NOW() - INTERVAL '30 days'
ORDER BY p.mmr DESC
LIMIT 1000;

CREATE UNIQUE INDEX ON leaderboard (id);
CREATE INDEX ON leaderboard (mmr DESC);
CREATE INDEX ON leaderboard (rank);

-- Refresh function
CREATE OR REPLACE FUNCTION refresh_leaderboard()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY leaderboard;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- TRIGGERS & FUNCTIONS
-- ============================================================================

-- Update player stats on kill insert
CREATE OR REPLACE FUNCTION update_player_stats_on_kill()
RETURNS TRIGGER AS $$
BEGIN
    -- Update killer stats
    IF NEW.killer_id IS NOT NULL THEN
        UPDATE players 
        SET total_kills = total_kills + 1,
            total_headshots = total_headshots + CASE WHEN NEW.headshot THEN 1 ELSE 0 END,
            total_backstabs = total_backstabs + CASE WHEN NEW.backstab THEN 1 ELSE 0 END,
            updated_at = NOW()
        WHERE id = NEW.killer_id;
    END IF;
    
    -- Update victim stats
    IF NEW.victim_id IS NOT NULL THEN
        UPDATE players 
        SET total_deaths = total_deaths + 1,
            updated_at = NOW()
        WHERE id = NEW.victim_id;
    END IF;
    
    -- Update assister stats
    IF NEW.assister_id IS NOT NULL THEN
        UPDATE players 
        SET total_assists = total_assists + 1,
            updated_at = NOW()
        WHERE id = NEW.assister_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_player_stats_on_kill
AFTER INSERT ON kills
FOR EACH ROW
EXECUTE FUNCTION update_player_stats_on_kill();

-- Update player stats on airshot insert
CREATE OR REPLACE FUNCTION update_player_stats_on_airshot()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE players 
    SET total_airshots = total_airshots + 1,
        updated_at = NOW()
    WHERE id = NEW.player_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_player_stats_on_airshot
AFTER INSERT ON airshots
FOR EACH ROW
EXECUTE FUNCTION update_player_stats_on_airshot();

-- Update player stats on deflect insert
CREATE OR REPLACE FUNCTION update_player_stats_on_deflect()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE players 
    SET total_deflects = total_deflects + 1,
        updated_at = NOW()
    WHERE id = NEW.player_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_player_stats_on_deflect
AFTER INSERT ON deflects
FOR EACH ROW
EXECUTE FUNCTION update_player_stats_on_deflect();

-- Update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_players_updated_at BEFORE UPDATE ON players
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tournaments_updated_at BEFORE UPDATE ON tournaments
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- INDEXES FOR COMMON QUERIES
-- ============================================================================

-- Player lookup by steam_id
CREATE INDEX idx_players_steam_id_lookup ON players(steam_id) INCLUDE (name, mmr, total_kills, total_deaths);

-- Recent matches
CREATE INDEX idx_matches_recent ON matches(started_at DESC) INCLUDE (map, gamemode, winner_team);

-- Player match history
CREATE INDEX idx_match_players_history ON match_players(player_id, match_id DESC);

-- Event processing queue
CREATE INDEX idx_events_unprocessed ON events(created_at) WHERE NOT processed;

-- ============================================================================
-- SAMPLE DATA (for testing)
-- ============================================================================

-- Insert a test player
INSERT INTO players (steam_id, name, mmr) 
VALUES ('76561198012345678', 'TestPlayer', 1000)
ON CONFLICT (steam_id) DO NOTHING;

COMMENT ON TABLE players IS 'Player profiles and aggregate statistics';
COMMENT ON TABLE matches IS 'Match records with server and timing information';
COMMENT ON TABLE events IS 'Raw event log from game servers';
COMMENT ON TABLE kills IS 'Detailed kill records with weapon and position data';
COMMENT ON TABLE airshots IS 'Airshot achievements';
COMMENT ON TABLE deflects IS 'Deflect events (airblast and dodgeball)';
COMMENT ON TABLE tournaments IS 'Tournament definitions';
COMMENT ON TABLE tournament_teams IS 'Teams registered for tournaments';
COMMENT ON TABLE tournament_matches IS 'Tournament match pairings and results';
