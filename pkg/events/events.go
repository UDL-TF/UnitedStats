package events

import "time"

// EventType represents the type of game event
type EventType string

const (
	EventTypeKill       EventType = "KILL"
	EventTypeDeflect    EventType = "DEFLECT"
	EventTypeMatchStart EventType = "MATCH_START"
	EventTypeMatchEnd   EventType = "MATCH_END"
)

// BaseEvent contains fields common to all events
type BaseEvent struct {
	Timestamp time.Time
	Gamemode  string
	ServerIP  string
}

// KillEvent represents a player kill
type KillEvent struct {
	BaseEvent
	KillerSteamID string
	KillerName    string
	VictimSteamID string
	VictimName    string
	Weapon        string
	Crit          bool
	Airborne      bool
}

// DeflectEvent represents a deflect in dodgeball
type DeflectEvent struct {
	BaseEvent
	PlayerSteamID string
	PlayerName    string
	RocketSpeed   float64
	DeflectAngle  float64
	TimingMs      int
	Distance      float64
}

// MatchStartEvent represents match/round start
type MatchStartEvent struct {
	BaseEvent
	MapName string
}

// MatchEndEvent represents match/round end
type MatchEndEvent struct {
	BaseEvent
	WinnerTeam int // 2=RED, 3=BLU, 0=tie
	Duration   int // seconds
}

// Event is a union type for all event types
type Event struct {
	Type EventType
	
	Kill       *KillEvent
	Deflect    *DeflectEvent
	MatchStart *MatchStartEvent
	MatchEnd   *MatchEndEvent
}
