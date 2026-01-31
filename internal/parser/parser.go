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
	case events.EventTypeKill:
		return parseKillEvent(line)
	case events.EventTypeDeflect:
		return parseDeflectEvent(line)
	case events.EventTypeMatchStart:
		return parseMatchStartEvent(line)
	case events.EventTypeMatchEnd:
		return parseMatchEndEvent(line)
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

// parseMatchStartEvent parses a match_start event JSON
func parseMatchStartEvent(line string) (*events.Event, error) {
	var matchStartEvent events.MatchStartEvent
	
	if err := json.Unmarshal([]byte(line), &matchStartEvent); err != nil {
		return nil, &ParseError{Line: line, Reason: fmt.Sprintf("invalid match_start event: %v", err)}
	}
	
	return &events.Event{
		Type:       events.EventTypeMatchStart,
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
		Type:     events.EventTypeMatchEnd,
		MatchEnd: &matchEndEvent,
	}, nil
}
