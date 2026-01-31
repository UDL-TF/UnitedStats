package events

import "time"

// EventType represents the type of game event
type EventType string

const (
	EventTypeKill       EventType = "kill"
	EventTypeDeflect    EventType = "deflect"
	EventTypeMatchStart EventType = "match_start"
	EventTypeMatchEnd   EventType = "match_end"
)

// Player represents a player in an event
type Player struct {
	SteamID string `json:"steam_id"`
	Name    string `json:"name"`
	Team    int    `json:"team"`
}

// Position represents a 3D position
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// Weapon represents weapon information
type Weapon struct {
	Name         string `json:"name"`
	ItemDefIndex int    `json:"item_def_index,omitempty"`
}

// BaseEvent contains fields common to all events
type BaseEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Gamemode  string    `json:"gamemode"`
	ServerIP  string    `json:"server_ip"`
	EventType EventType `json:"event_type"`
}

// KillEvent represents a player kill
type KillEvent struct {
	BaseEvent
	Killer     Player    `json:"killer"`
	Victim     Player    `json:"victim"`
	Weapon     Weapon    `json:"weapon"`
	Crit       bool      `json:"crit"`
	Airborne   bool      `json:"airborne"`
	KillerPos  *Position `json:"killer_pos,omitempty"`
	VictimPos  *Position `json:"victim_pos,omitempty"`
}

// DeflectEvent represents a deflect in dodgeball
type DeflectEvent struct {
	BaseEvent
	Player       Player    `json:"player"`
	RocketSpeed  float64   `json:"rocket_speed"`
	DeflectAngle float64   `json:"deflect_angle"`
	TimingMs     int       `json:"timing_ms"`
	Distance     float64   `json:"distance"`
	PlayerPos    *Position `json:"player_pos,omitempty"`
}

// MatchStartEvent represents match/round start
type MatchStartEvent struct {
	BaseEvent
	Map string `json:"map"`
}

// MatchEndEvent represents match/round end
type MatchEndEvent struct {
	BaseEvent
	WinnerTeam int `json:"winner_team"` // 2=RED, 3=BLU, 0=tie
	Duration   int `json:"duration"`    // seconds
}

// Event is a union type for all event types
type Event struct {
	Type EventType
	
	Kill       *KillEvent
	Deflect    *DeflectEvent
	MatchStart *MatchStartEvent
	MatchEnd   *MatchEndEvent
}
