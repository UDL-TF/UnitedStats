package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/UDL-TF/UnitedStats/pkg/events"
)

// ParseError represents a parsing error
type ParseError struct {
	Line   string
	Reason string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error: %s (line: %s)", e.Reason, e.Line)
}

// ParseLine parses a single JSON log line into an Event
func ParseLine(line string) (*events.Event, error) {
	// Trim whitespace
	line = strings.TrimSpace(line)
	
	// Skip empty lines
	if line == "" {
		return nil, nil
	}
	
	// Skip comment lines
	if strings.HasPrefix(line, "#") {
		return nil, nil
	}
	
	// First, parse just to get the event_type
	var baseEvent struct {
		EventType events.EventType `json:"event_type"`
	}
	
	if err := json.Unmarshal([]byte(line), &baseEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid JSON: %v", err)}
	}
	
	// Parse based on event type
	switch baseEvent.EventType {
	// Combat events
	case events.EventTypeKill:
		return parseKillEvent(line)
	case events.EventTypeAirshot:
		return parseAirshotEvent(line)
	case events.EventTypeDeflect:
		return parseDeflectEvent(line)
	case events.EventTypeStun:
		return parseStunEvent(line)
	case events.EventTypeJarate:
		return parseJarateEvent(line)
	case events.EventTypeShieldBlocked:
		return parseShieldBlockEvent(line)
	
	// Movement events
	case events.EventTypeRocketJump, events.EventTypeStickyJump:
		return parseJumpEvent(line)
	case events.EventTypeRocketJumpKill, events.EventTypeStickyJumpKill:
		return parseJumpKillEvent(line)
	case events.EventTypeTeleport, events.EventTypeTeleportUsed:
		return parseTeleportEvent(line)
	
	// Building events
	case events.EventTypeBuiltObject:
		return parseBuildingEvent(line)
	case events.EventTypeKilledObject:
		return parseKilledObjectEvent(line)
	
	// Medic events
	case events.EventTypeHealed:
		return parseHealedEvent(line)
	case events.EventTypeUberDeployed, events.EventTypeUberDropped, events.EventTypeDefendedMedic:
		return parseMedicEvent(line)
	
	// Support events
	case events.EventTypeBuffDeployed:
		return parseBuffEvent(line)
	case events.EventTypeSandvich, events.EventTypeDalokohs, events.EventTypeSteak:
		return parseFoodEvent(line)
	
	// Match events
	case events.EventTypeMatchStart, events.EventTypeRoundStart:
		return parseMatchStartEvent(line)
	case events.EventTypeMatchEnd, events.EventTypeRoundEnd:
		return parseMatchEndEvent(line)
	case events.EventTypeMVP1, events.EventTypeMVP2, events.EventTypeMVP3:
		return parseMVPEvent(line)
	
	// Player events
	case events.EventTypePlayerLoadout:
		return parsePlayerLoadoutEvent(line)
	case events.EventTypeWeaponStats:
		return parseWeaponStatsEvent(line)
	case events.EventTypeClassChange:
		return parseClassChangeEvent(line)
	
	default:
		// Unknown event type - skip but don't error
		return nil, nil
	}
}

// parseKillEvent parses a kill event JSON
func parseKillEvent(line string) (*events.Event, error) {
	var killEvent events.KillEvent
	
	if err := json.Unmarshal([]byte(line), &killEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid kill event: %v", err)}
	}
	
	return &events.Event{
		Type: events.EventTypeKill,
		Kill: &killEvent,
	}, nil
}

// parseAirshotEvent parses an airshot event JSON
func parseAirshotEvent(line string) (*events.Event, error) {
	var airshotEvent events.AirshotEvent
	
	if err := json.Unmarshal([]byte(line), &airshotEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid airshot event: %v", err)}
	}
	
	return &events.Event{
		Type:    events.EventTypeAirshot,
		Airshot: &airshotEvent,
	}, nil
}

// parseDeflectEvent parses a deflect event JSON
func parseDeflectEvent(line string) (*events.Event, error) {
	var deflectEvent events.DeflectEvent
	
	if err := json.Unmarshal([]byte(line), &deflectEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid deflect event: %v", err)}
	}
	
	return &events.Event{
		Type:    events.EventTypeDeflect,
		Deflect: &deflectEvent,
	}, nil
}

// parseStunEvent parses a stun event JSON
func parseStunEvent(line string) (*events.Event, error) {
	var stunEvent events.StunEvent
	
	if err := json.Unmarshal([]byte(line), &stunEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid stun event: %v", err)}
	}
	
	return &events.Event{
		Type: events.EventTypeStun,
		Stun: &stunEvent,
	}, nil
}

// parseJarateEvent parses a jarate/mad milk event JSON
func parseJarateEvent(line string) (*events.Event, error) {
	var jarateEvent events.JarateEvent
	
	if err := json.Unmarshal([]byte(line), &jarateEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid jarate event: %v", err)}
	}
	
	return &events.Event{
		Type:   events.EventTypeJarate,
		Jarate: &jarateEvent,
	}, nil
}

// parseShieldBlockEvent parses a shield block event JSON
func parseShieldBlockEvent(line string) (*events.Event, error) {
	var shieldBlockEvent events.ShieldBlockEvent
	
	if err := json.Unmarshal([]byte(line), &shieldBlockEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid shield_block event: %v", err)}
	}
	
	return &events.Event{
		Type:        events.EventTypeShieldBlocked,
		ShieldBlock: &shieldBlockEvent,
	}, nil
}

// parseJumpEvent parses a rocket/sticky jump event JSON
func parseJumpEvent(line string) (*events.Event, error) {
	var jumpEvent events.JumpEvent
	
	if err := json.Unmarshal([]byte(line), &jumpEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid jump event: %v", err)}
	}
	
	return &events.Event{
		Type: jumpEvent.EventType,
		Jump: &jumpEvent,
	}, nil
}

// parseJumpKillEvent parses a jump kill event JSON
func parseJumpKillEvent(line string) (*events.Event, error) {
	var jumpKillEvent events.JumpKillEvent
	
	if err := json.Unmarshal([]byte(line), &jumpKillEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid jump_kill event: %v", err)}
	}
	
	return &events.Event{
		Type:     jumpKillEvent.EventType,
		JumpKill: &jumpKillEvent,
	}, nil
}

// parseTeleportEvent parses a teleport event JSON
func parseTeleportEvent(line string) (*events.Event, error) {
	var teleportEvent events.TeleportEvent
	
	if err := json.Unmarshal([]byte(line), &teleportEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid teleport event: %v", err)}
	}
	
	return &events.Event{
		Type:     teleportEvent.EventType,
		Teleport: &teleportEvent,
	}, nil
}

// parseBuildingEvent parses a building built event JSON
func parseBuildingEvent(line string) (*events.Event, error) {
	var buildingEvent events.BuildingEvent
	
	if err := json.Unmarshal([]byte(line), &buildingEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid building event: %v", err)}
	}
	
	return &events.Event{
		Type:     events.EventTypeBuiltObject,
		Building: &buildingEvent,
	}, nil
}

// parseKilledObjectEvent parses a killed object event JSON
func parseKilledObjectEvent(line string) (*events.Event, error) {
	var killedObjectEvent events.KilledObjectEvent
	
	if err := json.Unmarshal([]byte(line), &killedObjectEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid killed_object event: %v", err)}
	}
	
	return &events.Event{
		Type:         events.EventTypeKilledObject,
		KilledObject: &killedObjectEvent,
	}, nil
}

// parseHealedEvent parses a healed event JSON
func parseHealedEvent(line string) (*events.Event, error) {
	var healedEvent events.HealedEvent
	
	if err := json.Unmarshal([]byte(line), &healedEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid healed event: %v", err)}
	}
	
	return &events.Event{
		Type:   events.EventTypeHealed,
		Healed: &healedEvent,
	}, nil
}

// parseMedicEvent parses a medic event JSON
func parseMedicEvent(line string) (*events.Event, error) {
	var medicEvent events.MedicEvent
	
	if err := json.Unmarshal([]byte(line), &medicEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid medic event: %v", err)}
	}
	
	return &events.Event{
		Type:  medicEvent.EventType,
		Medic: &medicEvent,
	}, nil
}

// parseBuffEvent parses a buff deployed event JSON
func parseBuffEvent(line string) (*events.Event, error) {
	var buffEvent events.BuffEvent
	
	if err := json.Unmarshal([]byte(line), &buffEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid buff event: %v", err)}
	}
	
	return &events.Event{
		Type: events.EventTypeBuffDeployed,
		Buff: &buffEvent,
	}, nil
}

// parseFoodEvent parses a food event JSON
func parseFoodEvent(line string) (*events.Event, error) {
	var foodEvent events.FoodEvent
	
	if err := json.Unmarshal([]byte(line), &foodEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid food event: %v", err)}
	}
	
	return &events.Event{
		Type: foodEvent.EventType,
		Food: &foodEvent,
	}, nil
}

// parseMatchStartEvent parses a match_start event JSON
func parseMatchStartEvent(line string) (*events.Event, error) {
	var matchStartEvent events.MatchStartEvent
	
	if err := json.Unmarshal([]byte(line), &matchStartEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid match_start event: %v", err)}
	}
	
	return &events.Event{
		Type:       matchStartEvent.EventType,
		MatchStart: &matchStartEvent,
	}, nil
}

// parseMatchEndEvent parses a match_end event JSON
func parseMatchEndEvent(line string) (*events.Event, error) {
	var matchEndEvent events.MatchEndEvent
	
	if err := json.Unmarshal([]byte(line), &matchEndEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid match_end event: %v", err)}
	}
	
	return &events.Event{
		Type:     matchEndEvent.EventType,
		MatchEnd: &matchEndEvent,
	}, nil
}

// parseMVPEvent parses an MVP event JSON
func parseMVPEvent(line string) (*events.Event, error) {
	var mvpEvent events.MVPEvent
	
	if err := json.Unmarshal([]byte(line), &mvpEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid mvp event: %v", err)}
	}
	
	return &events.Event{
		Type: mvpEvent.EventType,
		MVP:  &mvpEvent,
	}, nil
}

// parsePlayerLoadoutEvent parses a player loadout event JSON
func parsePlayerLoadoutEvent(line string) (*events.Event, error) {
	var playerLoadoutEvent events.PlayerLoadoutEvent
	
	if err := json.Unmarshal([]byte(line), &playerLoadoutEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid player_loadout event: %v", err)}
	}
	
	return &events.Event{
		Type:          events.EventTypePlayerLoadout,
		PlayerLoadout: &playerLoadoutEvent,
	}, nil
}

// parseWeaponStatsEvent parses a weapon stats event JSON
func parseWeaponStatsEvent(line string) (*events.Event, error) {
	var weaponStatsEvent events.WeaponStatsEvent
	
	if err := json.Unmarshal([]byte(line), &weaponStatsEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid weapon_stats event: %v", err)}
	}
	
	return &events.Event{
		Type:        events.EventTypeWeaponStats,
		WeaponStats: &weaponStatsEvent,
	}, nil
}

// parseClassChangeEvent parses a class change event JSON
func parseClassChangeEvent(line string) (*events.Event, error) {
	var classChangeEvent events.ClassChangeEvent
	
	if err := json.Unmarshal([]byte(line), &classChangeEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid class_change event: %v", err)}
	}
	
	return &events.Event{
		Type:        events.EventTypeClassChange,
		ClassChange: &classChangeEvent,
	}, nil
}
