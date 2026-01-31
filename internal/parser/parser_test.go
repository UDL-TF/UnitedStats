package parser

import (
	"bufio"
	"os"
	"testing"
	"time"

	"github.com/UDL-TF/UnitedStats/pkg/events"
)

// TestParseKillEvent tests basic KILL event parsing
func TestParseKillEvent(t *testing.T) {
	line := "KILL|1706745600|default|192.168.1.100|76561198012345678|Player1|76561198087654321|Player2|scattergun|0|0"
	
	event, err := ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine() error = %v", err)
	}
	
	if event == nil {
		t.Fatal("ParseLine() returned nil event")
	}
	
	if event.Type != events.EventTypeKill {
		t.Errorf("event.Type = %v, want %v", event.Type, events.EventTypeKill)
	}
	
	kill := event.Kill
	if kill == nil {
		t.Fatal("event.Kill is nil")
	}
	
	// Check timestamp
	expectedTime := time.Unix(1706745600, 0)
	if !kill.Timestamp.Equal(expectedTime) {
		t.Errorf("Timestamp = %v, want %v", kill.Timestamp, expectedTime)
	}
	
	// Check fields
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Gamemode", kill.Gamemode, "default"},
		{"ServerIP", kill.ServerIP, "192.168.1.100"},
		{"KillerSteamID", kill.KillerSteamID, "76561198012345678"},
		{"KillerName", kill.KillerName, "Player1"},
		{"VictimSteamID", kill.VictimSteamID, "76561198087654321"},
		{"VictimName", kill.VictimName, "Player2"},
		{"Weapon", kill.Weapon, "scattergun"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
	
	if kill.Crit {
		t.Error("Crit = true, want false")
	}
	
	if kill.Airborne {
		t.Error("Airborne = true, want false")
	}
}

// TestParseCriticalKill tests critical hit kill
func TestParseCriticalKill(t *testing.T) {
	line := "KILL|1706745601|default|192.168.1.100|76561198012345678|Player1|76561198087654321|Player2|rocket_launcher|1|0"
	
	event, err := ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine() error = %v", err)
	}
	
	if !event.Kill.Crit {
		t.Error("Crit = false, want true")
	}
}

// TestParseAirshot tests airshot kill
func TestParseAirshot(t *testing.T) {
	line := "KILL|1706745602|default|192.168.1.100|76561198012345678|SniperPro|76561198087654321|ScoutMain|tf_projectile_pipe_remote|0|1"
	
	event, err := ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine() error = %v", err)
	}
	
	if !event.Kill.Airborne {
		t.Error("Airborne = false, want true")
	}
}

// TestParseEscapedNames tests special character escaping
func TestParseEscapedNames(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		wantKillerName string
		wantVictimName string
	}{
		{
			name: "Escaped pipe",
			line: "KILL|1706745604|default|192.168.1.100|76561198012345678|Player\\pOne|76561198087654321|NormalName|minigun|0|0",
			wantKillerName: "Player|One",
			wantVictimName: "NormalName",
		},
		{
			name: "Multiple escaped pipes",
			line: "KILL|1706745605|default|192.168.1.100|76561198012345678|[TF2]\\pBot\\p2000|76561198087654321|l33t\\pplayer|shotgun_primary|0|0",
			wantKillerName: "[TF2]|Bot|2000",
			wantVictimName: "l33t|player",
		},
		{
			name: "Escaped backslash",
			line: "KILL|1706745606|default|192.168.1.100|76561198012345678|Player\\\\One|76561198087654321|Victim|smg|0|0",
			wantKillerName: "Player\\One",
			wantVictimName: "Victim",
		},
		{
			name: "Escaped newline",
			line: "KILL|1706745607|default|192.168.1.100|76561198012345678|Multi\\nLine\\nName|76561198087654321|TestPlayer|flamethrower|0|0",
			wantKillerName: "Multi\nLine\nName",
			wantVictimName: "TestPlayer",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ParseLine(tt.line)
			if err != nil {
				t.Fatalf("ParseLine() error = %v", err)
			}
			
			if event.Kill.KillerName != tt.wantKillerName {
				t.Errorf("KillerName = %q, want %q", event.Kill.KillerName, tt.wantKillerName)
			}
			
			if event.Kill.VictimName != tt.wantVictimName {
				t.Errorf("VictimName = %q, want %q", event.Kill.VictimName, tt.wantVictimName)
			}
		})
	}
}

// TestParseDeflectEvent tests DEFLECT event parsing
func TestParseDeflectEvent(t *testing.T) {
	line := "DEFLECT|1706745700|dodgeball|192.168.1.100|76561198012345678|DodgeballPro|1500.50|1.0000|50|100.00"
	
	event, err := ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine() error = %v", err)
	}
	
	if event.Type != events.EventTypeDeflect {
		t.Errorf("event.Type = %v, want %v", event.Type, events.EventTypeDeflect)
	}
	
	deflect := event.Deflect
	if deflect == nil {
		t.Fatal("event.Deflect is nil")
	}
	
	if deflect.Gamemode != "dodgeball" {
		t.Errorf("Gamemode = %v, want dodgeball", deflect.Gamemode)
	}
	
	if deflect.PlayerSteamID != "76561198012345678" {
		t.Errorf("PlayerSteamID = %v, want 76561198012345678", deflect.PlayerSteamID)
	}
	
	if deflect.PlayerName != "DodgeballPro" {
		t.Errorf("PlayerName = %v, want DodgeballPro", deflect.PlayerName)
	}
	
	if deflect.RocketSpeed != 1500.50 {
		t.Errorf("RocketSpeed = %v, want 1500.50", deflect.RocketSpeed)
	}
	
	if deflect.DeflectAngle != 1.0000 {
		t.Errorf("DeflectAngle = %v, want 1.0000", deflect.DeflectAngle)
	}
	
	if deflect.TimingMs != 50 {
		t.Errorf("TimingMs = %v, want 50", deflect.TimingMs)
	}
	
	if deflect.Distance != 100.00 {
		t.Errorf("Distance = %v, want 100.00", deflect.Distance)
	}
}

// TestParseMatchStartEvent tests MATCH_START event parsing
func TestParseMatchStartEvent(t *testing.T) {
	line := "MATCH_START|1706745800|default|192.168.1.100|cp_process_final"
	
	event, err := ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine() error = %v", err)
	}
	
	if event.Type != events.EventTypeMatchStart {
		t.Errorf("event.Type = %v, want %v", event.Type, events.EventTypeMatchStart)
	}
	
	matchStart := event.MatchStart
	if matchStart == nil {
		t.Fatal("event.MatchStart is nil")
	}
	
	if matchStart.MapName != "cp_process_final" {
		t.Errorf("MapName = %v, want cp_process_final", matchStart.MapName)
	}
}

// TestParseMatchEndEvent tests MATCH_END event parsing
func TestParseMatchEndEvent(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantTeam   int
		wantDuration int
	}{
		{
			name:       "RED wins",
			line:       "MATCH_END|1706745900|default|192.168.1.100|2|600",
			wantTeam:   2,
			wantDuration: 600,
		},
		{
			name:       "BLU wins",
			line:       "MATCH_END|1706745901|default|192.168.1.100|3|450",
			wantTeam:   3,
			wantDuration: 450,
		},
		{
			name:       "TIE",
			line:       "MATCH_END|1706745902|default|192.168.1.100|0|300",
			wantTeam:   0,
			wantDuration: 300,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ParseLine(tt.line)
			if err != nil {
				t.Fatalf("ParseLine() error = %v", err)
			}
			
			if event.Type != events.EventTypeMatchEnd {
				t.Errorf("event.Type = %v, want %v", event.Type, events.EventTypeMatchEnd)
			}
			
			matchEnd := event.MatchEnd
			if matchEnd.WinnerTeam != tt.wantTeam {
				t.Errorf("WinnerTeam = %v, want %v", matchEnd.WinnerTeam, tt.wantTeam)
			}
			
			if matchEnd.Duration != tt.wantDuration {
				t.Errorf("Duration = %v, want %v", matchEnd.Duration, tt.wantDuration)
			}
		})
	}
}

// TestParseInvalidLines tests error handling for malformed lines
func TestParseInvalidLines(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantErr bool
	}{
		{
			name:    "Empty line",
			line:    "",
			wantErr: false, // Should skip, not error
		},
		{
			name:    "Comment line",
			line:    "# This is a comment",
			wantErr: false, // Should skip, not error
		},
		{
			name:    "Missing fields",
			line:    "KILL|1706746200|default|192.168.1.100|76561198012345678",
			wantErr: true,
		},
		{
			name:    "Invalid timestamp",
			line:    "KILL|not_a_number|default|192.168.1.100|76561198012345678|Player1|76561198087654321|Player2|scattergun|0|0",
			wantErr: true,
		},
		{
			name:    "Unknown event type",
			line:    "UNKNOWN_EVENT|1706746202|default|192.168.1.100|data|more|data",
			wantErr: false, // Should skip, not error
		},
		{
			name:    "Only separators",
			line:    "||||||||",
			wantErr: true, // Invalid timestamp
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ParseLine(tt.line)
			
			if tt.wantErr {
				if err == nil {
					t.Error("ParseLine() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("ParseLine() error = %v, want nil", err)
				}
				// For skipped lines, event should be nil
				if tt.line == "" || tt.line[0] == '#' || tt.line == "UNKNOWN_EVENT|1706746202|default|192.168.1.100|data|more|data" {
					if event != nil {
						t.Errorf("ParseLine() event = %v, want nil (should skip)", event)
					}
				}
			}
		})
	}
}

// TestParseAllFixtures tests parsing the entire test fixture file
func TestParseAllFixtures(t *testing.T) {
	file, err := os.Open("../../test/fixtures/sample_logs.txt")
	if err != nil {
		t.Skipf("Skipping fixture test: %v", err)
		return
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	validEvents := 0
	skippedLines := 0
	errorCount := 0
	
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		
		event, err := ParseLine(line)
		
		if err != nil {
			// Some lines are intentionally malformed for testing
			t.Logf("Line %d error (expected for some): %v", lineNum, err)
			errorCount++
			continue
		}
		
		if event == nil {
			// Skipped line (comment, empty, unknown type)
			skippedLines++
			continue
		}
		
		validEvents++
		
		// Basic validation: all events should have timestamp and gamemode
		var baseEvent events.BaseEvent
		switch event.Type {
		case events.EventTypeKill:
			baseEvent = event.Kill.BaseEvent
		case events.EventTypeDeflect:
			baseEvent = event.Deflect.BaseEvent
		case events.EventTypeMatchStart:
			baseEvent = event.MatchStart.BaseEvent
		case events.EventTypeMatchEnd:
			baseEvent = event.MatchEnd.BaseEvent
		}
		
		if baseEvent.Gamemode == "" {
			t.Errorf("Line %d: event has empty gamemode", lineNum)
		}
		
		if baseEvent.ServerIP == "" {
			t.Errorf("Line %d: event has empty server IP", lineNum)
		}
	}
	
	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading fixture file: %v", err)
	}
	
	t.Logf("Parsed %d lines: %d valid events, %d skipped, %d errors", 
		lineNum, validEvents, skippedLines, errorCount)
	
	// We expect a good number of valid events
	if validEvents < 50 {
		t.Errorf("Expected at least 50 valid events, got %d", validEvents)
	}
}

// BenchmarkParseKillEvent benchmarks KILL event parsing
func BenchmarkParseKillEvent(b *testing.B) {
	line := "KILL|1706745600|default|192.168.1.100|76561198012345678|Player1|76561198087654321|Player2|scattergun|0|0"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseLine(line)
	}
}

// BenchmarkParseDeflectEvent benchmarks DEFLECT event parsing
func BenchmarkParseDeflectEvent(b *testing.B) {
	line := "DEFLECT|1706745700|dodgeball|192.168.1.100|76561198012345678|DodgeballPro|1500.50|1.0000|50|100.00"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseLine(line)
	}
}
