package events

import (
	"encoding/json"
	"fmt"
	"time"
)

// FlexibleTime is a custom time type that can parse multiple timestamp formats
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler for flexible timestamp parsing
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Try RFC3339 with timezone first
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		ft.Time = t
		return nil
	}

	// Try without timezone (assume UTC)
	t, err = time.Parse("2006-01-02T15:04:05", s)
	if err == nil {
		ft.Time = t.UTC()
		return nil
	}

	return fmt.Errorf("unable to parse timestamp: %s", s)
}

// MarshalJSON implements json.Marshaler
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ft.Time.Format(time.RFC3339))
}

// EventType represents the type of game event
type EventType string

const (
	// Kill and damage events
	EventTypeKill       EventType = "kill"
	EventTypeAssist     EventType = "assist"
	EventTypeDomination EventType = "domination"
	EventTypeRevenge    EventType = "revenge"
	
	// Action events
	EventTypeHeadshot         EventType = "headshot"
	EventTypeBackstab         EventType = "backstab"
	EventTypeAirshot          EventType = "airshot"
	EventTypeDeflect          EventType = "deflect"
	EventTypeJarate           EventType = "jarate"
	EventTypeMadMilk          EventType = "mad_milk"
	EventTypeShieldBlocked    EventType = "shield_blocked"
	EventTypeStun             EventType = "stun"
	
	// Jump events
	EventTypeRocketJump       EventType = "rocket_jump"
	EventTypeStickyJump       EventType = "sticky_jump"
	EventTypeRocketJumpKill   EventType = "rocket_jump_kill"
	EventTypeStickyJumpKill   EventType = "sticky_jump_kill"
	
	// Teleporter events
	EventTypeTeleport         EventType = "teleport"
	EventTypeTeleportUsed     EventType = "teleport_used"
	
	// Building events
	EventTypeBuiltObject      EventType = "built_object"
	EventTypeKilledObject     EventType = "killed_object"
	EventTypeObjectDestroyed  EventType = "object_destroyed"
	
	// Medic events
	EventTypeHealed           EventType = "healed"
	EventTypeDefendedMedic    EventType = "defended_medic"
	EventTypeUberDeployed     EventType = "uber_deployed"
	EventTypeUberDropped      EventType = "uber_dropped"
	
	// Buff events
	EventTypeBuffDeployed     EventType = "buff_deployed"
	
	// Food events
	EventTypeSandvich         EventType = "sandvich"
	EventTypeDalokohs         EventType = "dalokohs"
	EventTypeSteak            EventType = "steak"
	
	// Round/Match events
	EventTypeMatchStart       EventType = "match_start"
	EventTypeMatchEnd         EventType = "match_end"
	EventTypeRoundStart       EventType = "round_start"
	EventTypeRoundEnd         EventType = "round_end"
	
	// MVP events
	EventTypeMVP1             EventType = "mvp1"
	EventTypeMVP2             EventType = "mvp2"
	EventTypeMVP3             EventType = "mvp3"
	
	// Player events
	EventTypePlayerLoadout    EventType = "player_loadout"
	EventTypePlayerSpawn      EventType = "player_spawn"
	EventTypePlayerDisconnect EventType = "player_disconnect"
	EventTypeClassChange      EventType = "class_change"
	
	// Weapon stats
	EventTypeWeaponStats      EventType = "weapon_stats"
	
	// First blood
	EventTypeFirstBlood       EventType = "first_blood"
)

// Player represents a player in an event
type Player struct {
	SteamID string `json:"steam_id"`
	Name    string `json:"name"`
	Team    int    `json:"team"`
	Class   string `json:"class,omitempty"`
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

// ObjectInfo represents building information
type ObjectInfo struct {
	Type  string `json:"type"`  // "dispenser", "sentry", "teleporter_entrance", etc.
	Level int    `json:"level,omitempty"`
}

// WeaponStatistics contains weapon usage statistics
type WeaponStatistics struct {
	Weapon    string `json:"weapon"`
	Shots     int    `json:"shots"`
	Hits      int    `json:"hits"`
	Kills     int    `json:"kills"`
	Headshots int    `json:"headshots"`
	Teamkills int    `json:"teamkills"`
	Damage    int    `json:"damage"`
	Deaths    int    `json:"deaths"`
}

// PlayerLoadout represents a player's equipped items
type PlayerLoadout struct {
	Primary   int `json:"primary"`
	Secondary int `json:"secondary"`
	Melee     int `json:"melee"`
	PDA       int `json:"pda,omitempty"`
	PDA2      int `json:"pda2,omitempty"`
	Building  int `json:"building,omitempty"`
	Head      int `json:"head,omitempty"`
	Misc      int `json:"misc,omitempty"`
}

// BaseEvent contains fields common to all events
type BaseEvent struct {
	Timestamp FlexibleTime `json:"timestamp"`
	Gamemode  string       `json:"gamemode"`
	ServerIP  string       `json:"server_ip"`
	EventType EventType    `json:"event_type"`
}

// KillEvent represents a player kill
type KillEvent struct {
	BaseEvent
	Killer       Player    `json:"killer"`
	Victim       Player    `json:"victim"`
	Assister     *Player   `json:"assister,omitempty"`
	Weapon       Weapon    `json:"weapon"`
	Crit         bool      `json:"crit"`
	Airborne     bool      `json:"airborne"`
	Headshot     bool      `json:"headshot"`
	Backstab     bool      `json:"backstab"`
	FirstBlood   bool      `json:"first_blood,omitempty"`
	Domination   bool      `json:"domination,omitempty"`
	Revenge      bool      `json:"revenge,omitempty"`
	KillerPos    *Position `json:"killer_pos,omitempty"`
	VictimPos    *Position `json:"victim_pos,omitempty"`
	CustomKill   int       `json:"custom_kill,omitempty"`
}

// AirshotEvent represents various airshot achievements
type AirshotEvent struct {
	BaseEvent
	Player     Player    `json:"player"`
	Victim     Player    `json:"victim"`
	WeaponType string    `json:"weapon_type"` // "rocket", "sticky", "pipebomb", "arrow", "flare", "stun"
	Air2Air    bool      `json:"air2air"`     // Both players airborne
	PlayerPos  *Position `json:"player_pos,omitempty"`
	VictimPos  *Position `json:"victim_pos,omitempty"`
}

// DeflectEvent represents a deflect (both dodgeball and standard airblast)
type DeflectEvent struct {
	BaseEvent
	Player         Player    `json:"player"`
	Owner          *Player   `json:"owner,omitempty"` // Original projectile owner
	ProjectileType string    `json:"projectile_type"` // "rocket", "pipebomb", "flare", "arrow", "player"
	RocketSpeed    float64   `json:"rocket_speed,omitempty"`
	DeflectAngle   float64   `json:"deflect_angle,omitempty"`
	TimingMs       int       `json:"timing_ms,omitempty"`
	Distance       float64   `json:"distance,omitempty"`
	PlayerPos      *Position `json:"player_pos,omitempty"`
}

// StunEvent represents a player being stunned (sandman ball)
type StunEvent struct {
	BaseEvent
	Stunner       Player    `json:"stunner"`
	Victim        Player    `json:"victim"`
	VictimCapping bool      `json:"victim_capping"`
	BigStun       bool      `json:"big_stun"` // Max charge stun
	Airshot       bool      `json:"airshot"`
	StunnerPos    *Position `json:"stunner_pos,omitempty"`
	VictimPos     *Position `json:"victim_pos,omitempty"`
}

// JumpEvent represents a rocket/sticky jump
type JumpEvent struct {
	BaseEvent
	Player    Player    `json:"player"`
	JumpType  string    `json:"jump_type"` // "rocket" or "sticky"
	PlayerPos *Position `json:"player_pos,omitempty"`
}

// JumpKillEvent represents a kill while rocket/sticky jumping
type JumpKillEvent struct {
	BaseEvent
	Player    Player    `json:"player"`
	Victim    Player    `json:"victim"`
	JumpType  string    `json:"jump_type"` // "rocket" or "sticky"
	PlayerPos *Position `json:"player_pos,omitempty"`
	VictimPos *Position `json:"victim_pos,omitempty"`
}

// TeleportEvent represents teleporter usage
type TeleportEvent struct {
	BaseEvent
	Builder   Player `json:"builder"`
	User      *Player `json:"user,omitempty"` // nil if self-teleport
	SelfUsed  bool   `json:"self_used"`
	Repeated  bool   `json:"repeated"` // Used same teleporter within 10 seconds
}

// BuildingEvent represents building creation/destruction
type BuildingEvent struct {
	BaseEvent
	Player   Player      `json:"player"`
	Object   ObjectInfo  `json:"object"`
	Position *Position   `json:"position,omitempty"`
}

// KilledObjectEvent represents a player destroying a building
type KilledObjectEvent struct {
	BaseEvent
	Attacker Player      `json:"attacker"`
	Owner    Player      `json:"owner"`
	Object   ObjectInfo  `json:"object"`
	Weapon   Weapon      `json:"weapon"`
	Position *Position   `json:"position,omitempty"`
}

// HealedEvent represents healing done by medic
type HealedEvent struct {
	BaseEvent
	Medic       Player `json:"medic"`
	HealPoints  int    `json:"heal_points"`
	Reason      string `json:"reason,omitempty"` // "death", "spawn", "disconnect"
}

// MedicEvent represents medic-specific actions
type MedicEvent struct {
	BaseEvent
	Medic      Player  `json:"medic"`
	Patient    *Player `json:"patient,omitempty"`
	ActionType string  `json:"action_type"` // "uber_deployed", "uber_dropped", "defended_medic"
	UberCharge float64 `json:"uber_charge,omitempty"`
}

// BuffEvent represents buff banner deployment
type BuffEvent struct {
	BaseEvent
	Player   Player `json:"player"`
	BuffType string `json:"buff_type"` // "buff", "backup", "conch"
}

// FoodEvent represents eating sandvich/dalokohs/steak
type FoodEvent struct {
	BaseEvent
	Player     Player `json:"player"`
	FoodType   string `json:"food_type"` // "sandvich", "dalokohs", "steak"
	HealedSelf bool   `json:"healed_self"`
	Thrown     bool   `json:"thrown,omitempty"`
}

// JarateEvent represents jarate/mad milk application
type JarateEvent struct {
	BaseEvent
	Attacker Player `json:"attacker"`
	Victim   Player `json:"victim"`
	JarType  string `json:"jar_type"` // "jarate" or "mad_milk"
}

// ShieldBlockEvent represents demoman shield blocking damage
type ShieldBlockEvent struct {
	BaseEvent
	Blocker  Player `json:"blocker"`
	Attacker Player `json:"attacker"`
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

// MVPEvent represents top 3 players at round end
type MVPEvent struct {
	BaseEvent
	Player   Player `json:"player"`
	Position int    `json:"position"` // 1, 2, or 3
}

// PlayerLoadoutEvent represents player's equipped items
type PlayerLoadoutEvent struct {
	BaseEvent
	Player  Player        `json:"player"`
	Loadout PlayerLoadout `json:"loadout"`
}

// WeaponStatsEvent represents weapon statistics dump
type WeaponStatsEvent struct {
	BaseEvent
	Player Player             `json:"player"`
	Weapon WeaponStatistics   `json:"weapon"`
}

// ClassChangeEvent represents a player changing class
type ClassChangeEvent struct {
	BaseEvent
	Player   Player `json:"player"`
	OldClass string `json:"old_class"`
	NewClass string `json:"new_class"`
}

// Event is a union type for all event types
type Event struct {
	Type EventType
	
	// Combat events
	Kill         *KillEvent
	Airshot      *AirshotEvent
	Deflect      *DeflectEvent
	Stun         *StunEvent
	Jarate       *JarateEvent
	ShieldBlock  *ShieldBlockEvent
	
	// Movement events
	Jump         *JumpEvent
	JumpKill     *JumpKillEvent
	Teleport     *TeleportEvent
	
	// Building events
	Building     *BuildingEvent
	KilledObject *KilledObjectEvent
	
	// Medic events
	Healed       *HealedEvent
	Medic        *MedicEvent
	
	// Support events
	Buff         *BuffEvent
	Food         *FoodEvent
	
	// Match events
	MatchStart   *MatchStartEvent
	MatchEnd     *MatchEndEvent
	MVP          *MVPEvent
	
	// Player events
	PlayerLoadout *PlayerLoadoutEvent
	WeaponStats   *WeaponStatsEvent
	ClassChange   *ClassChangeEvent
}
