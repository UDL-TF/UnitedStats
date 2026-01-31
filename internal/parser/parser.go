package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

// UnescapeString reverses the escaping done by SourceMod plugin
func UnescapeString(s string) string {
	s = strings.ReplaceAll(s, "\\p", "|")
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

// ParseLine parses a single log line into an Event
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
	
	// Split by pipe
	parts := strings.Split(line, "|")
	if len(parts) < 4 {
		return nil, &ParseError{Line: line, Reason: "insufficient fields"}
	}
	
	eventType := events.EventType(parts[0])
	
	// Parse based on event type
	switch eventType {
	case events.EventTypeKill:
		return parseKillEvent(line, parts)
	case events.EventTypeDeflect:
		return parseDeflectEvent(line, parts)
	case events.EventTypeMatchStart:
		return parseMatchStartEvent(line, parts)
	case events.EventTypeMatchEnd:
		return parseMatchEndEvent(line, parts)
	default:
		// Unknown event type - skip but don't error
		return nil, nil
	}
}

// parseBaseFields extracts common fields (timestamp, gamemode, server_ip)
func parseBaseFields(line string, parts []string) (events.BaseEvent, error) {
	if len(parts) < 4 {
		return events.BaseEvent{}, &ParseError{Line: line, Reason: "missing base fields"}
	}
	
	// Parse timestamp
	timestampInt, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return events.BaseEvent{}, &ParseError{Line: line, Reason: fmt.Sprintf("invalid timestamp: %v", err)}
	}
	timestamp := time.Unix(timestampInt, 0)
	
	return events.BaseEvent{
		Timestamp: timestamp,
		Gamemode:  parts[2],
		ServerIP:  parts[3],
	}, nil
}

// parseKillEvent parses KILL event
// Format: KILL|timestamp|gamemode|server_ip|killer_steamid|killer_name|victim_steamid|victim_name|weapon|crit|airborne
func parseKillEvent(line string, parts []string) (*events.Event, error) {
	if len(parts) < 11 {
		return nil, &ParseError{Line: line, Reason: "KILL event needs 11 fields"}
	}
	
	base, err := parseBaseFields(line, parts)
	if err != nil {
		return nil, err
	}
	
	// Parse crit (0/1)
	crit, err := strconv.Atoi(parts[9])
	if err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid crit value: %v", err)}
	}
	
	// Parse airborne (0/1)
	airborne, err := strconv.Atoi(parts[10])
	if err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid airborne value: %v", err)}
	}
	
	killEvent := &events.KillEvent{
		BaseEvent:     base,
		KillerSteamID: parts[4],
		KillerName:    UnescapeString(parts[5]),
		VictimSteamID: parts[6],
		VictimName:    UnescapeString(parts[7]),
		Weapon:        UnescapeString(parts[8]),
		Crit:          crit == 1,
		Airborne:      airborne == 1,
	}
	
	return &events.Event{
		Type: events.EventTypeKill,
		Kill: killEvent,
	}, nil
}

// parseDeflectEvent parses DEFLECT event
// Format: DEFLECT|timestamp|gamemode|server_ip|player_steamid|player_name|rocket_speed|deflect_angle|timing_ms|distance
func parseDeflectEvent(line string, parts []string) (*events.Event, error) {
	if len(parts) < 10 {
		return nil, &ParseError{Line: line, Reason: "DEFLECT event needs 10 fields"}
	}
	
	base, err := parseBaseFields(line, parts)
	if err != nil {
		return nil, err
	}
	
	// Parse rocket_speed
	rocketSpeed, err := strconv.ParseFloat(parts[6], 64)
	if err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid rocket_speed: %v", err)}
	}
	
	// Parse deflect_angle
	deflectAngle, err := strconv.ParseFloat(parts[7], 64)
	if err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid deflect_angle: %v", err)}
	}
	
	// Parse timing_ms
	timingMs, err := strconv.Atoi(parts[8])
	if err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid timing_ms: %v", err)}
	}
	
	// Parse distance
	distance, err := strconv.ParseFloat(parts[9], 64)
	if err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid distance: %v", err)}
	}
	
	deflectEvent := &events.DeflectEvent{
		BaseEvent:     base,
		PlayerSteamID: parts[4],
		PlayerName:    UnescapeString(parts[5]),
		RocketSpeed:   rocketSpeed,
		DeflectAngle:  deflectAngle,
		TimingMs:      timingMs,
		Distance:      distance,
	}
	
	return &events.Event{
		Type:    events.EventTypeDeflect,
		Deflect: deflectEvent,
	}, nil
}

// parseMatchStartEvent parses MATCH_START event
// Format: MATCH_START|timestamp|gamemode|server_ip|map_name
func parseMatchStartEvent(line string, parts []string) (*events.Event, error) {
	if len(parts) < 5 {
		return nil, &ParseError{Line: line, Reason: "MATCH_START event needs 5 fields"}
	}
	
	base, err := parseBaseFields(line, parts)
	if err != nil {
		return nil, err
	}
	
	matchStartEvent := &events.MatchStartEvent{
		BaseEvent: base,
		MapName:   UnescapeString(parts[4]),
	}
	
	return &events.Event{
		Type:       events.EventTypeMatchStart,
		MatchStart: matchStartEvent,
	}, nil
}

// parseMatchEndEvent parses MATCH_END event
// Format: MATCH_END|timestamp|gamemode|server_ip|winner_team|duration
func parseMatchEndEvent(line string, parts []string) (*events.Event, error) {
	if len(parts) < 6 {
		return nil, &ParseError{Line: line, Reason: "MATCH_END event needs 6 fields"}
	}
	
	base, err := parseBaseFields(line, parts)
	if err != nil {
		return nil, err
	}
	
	// Parse winner_team
	winnerTeam, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid winner_team: %v", err)}
	}
	
	// Parse duration
	duration, err := strconv.Atoi(parts[5])
	if err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid duration: %v", err)}
	}
	
	matchEndEvent := &events.MatchEndEvent{
		BaseEvent:  base,
		WinnerTeam: winnerTeam,
		Duration:   duration,
	}
	
	return &events.Event{
		Type:     events.EventTypeMatchEnd,
		MatchEnd: matchEndEvent,
	}, nil
}
