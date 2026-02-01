package parser

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/UDL-TF/UnitedStats/pkg/events"
)

// TestParseKillEvent tests basic KILL event parsing
func TestParseKillEvent(t *testing.T) {
	line := `{"timestamp":"2024-02-01T12:00:00","gamemode":"default","server_ip":"192.168.1.100","event_type":"kill","killer":{"steam_id":"76561198012345678","name":"Player1","team":2},"victim":{"steam_id":"76561198087654321","name":"Player2","team":3},"weapon":{"name":"scattergun"},"crit":false,"airborne":false}`

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

	// Check fields
	if kill.Gamemode != "default" {
		t.Errorf("Gamemode = %v, want default", kill.Gamemode)
	}

	if kill.ServerIP != "192.168.1.100" {
		t.Errorf("ServerIP = %v, want 192.168.1.100", kill.ServerIP)
	}

	if kill.Killer.SteamID != "76561198012345678" {
		t.Errorf("Killer.SteamID = %v, want 76561198012345678", kill.Killer.SteamID)
	}

	if kill.Killer.Name != "Player1" {
		t.Errorf("Killer.Name = %v, want Player1", kill.Killer.Name)
	}

	if kill.Killer.Team != 2 {
		t.Errorf("Killer.Team = %v, want 2", kill.Killer.Team)
	}

	if kill.Victim.SteamID != "76561198087654321" {
		t.Errorf("Victim.SteamID = %v, want 76561198087654321", kill.Victim.SteamID)
	}

	if kill.Victim.Name != "Player2" {
		t.Errorf("Victim.Name = %v, want Player2", kill.Victim.Name)
	}

	if kill.Victim.Team != 3 {
		t.Errorf("Victim.Team = %v, want 3", kill.Victim.Team)
	}

	if kill.Weapon.Name != "scattergun" {
		t.Errorf("Weapon.Name = %v, want scattergun", kill.Weapon.Name)
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
	line := `{"timestamp":"2024-02-01T12:00:01","gamemode":"default","server_ip":"192.168.1.100","event_type":"kill","killer":{"steam_id":"76561198012345678","name":"Player1","team":2},"victim":{"steam_id":"76561198087654321","name":"Player2","team":3},"weapon":{"name":"rocket_launcher"},"crit":true,"airborne":false}`

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
	line := `{"timestamp":"2024-02-01T12:00:02","gamemode":"default","server_ip":"192.168.1.100","event_type":"kill","killer":{"steam_id":"76561198012345678","name":"SniperPro","team":2},"victim":{"steam_id":"76561198087654321","name":"ScoutMain","team":3},"weapon":{"name":"tf_projectile_pipe_remote"},"crit":false,"airborne":true}`

	event, err := ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine() error = %v", err)
	}

	if !event.Kill.Airborne {
		t.Error("Airborne = false, want true")
	}
}

// TestParseKillWithPositions tests kill with position data
func TestParseKillWithPositions(t *testing.T) {
	line := `{"timestamp":"2024-02-01T12:00:00","gamemode":"default","server_ip":"192.168.1.100","event_type":"kill","killer":{"steam_id":"76561198012345678","name":"Player1","team":2},"victim":{"steam_id":"76561198087654321","name":"Player2","team":3},"weapon":{"name":"scattergun"},"crit":false,"airborne":false,"killer_pos":{"x":100.5,"y":200.3,"z":50.0},"victim_pos":{"x":150.2,"y":210.8,"z":52.1}}`

	event, err := ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine() error = %v", err)
	}

	kill := event.Kill

	if kill.KillerPos == nil {
		t.Fatal("KillerPos is nil")
	}

	if kill.KillerPos.X != 100.5 {
		t.Errorf("KillerPos.X = %v, want 100.5", kill.KillerPos.X)
	}

	if kill.VictimPos == nil {
		t.Fatal("VictimPos is nil")
	}

	if kill.VictimPos.X != 150.2 {
		t.Errorf("VictimPos.X = %v, want 150.2", kill.VictimPos.X)
	}
}

// TestParseSpecialCharacters tests names with special characters
func TestParseSpecialCharacters(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantKiller string
		wantVictim string
	}{
		{
			name:       "Pipe in name",
			line:       `{"timestamp":"2024-02-01T12:00:04","gamemode":"default","server_ip":"192.168.1.100","event_type":"kill","killer":{"steam_id":"76561198012345678","name":"Player|One","team":2},"victim":{"steam_id":"76561198087654321","name":"NormalName","team":3},"weapon":{"name":"minigun"},"crit":false,"airborne":false}`,
			wantKiller: "Player|One",
			wantVictim: "NormalName",
		},
		{
			name:       "Multiple special chars",
			line:       `{"timestamp":"2024-02-01T12:00:05","gamemode":"default","server_ip":"192.168.1.100","event_type":"kill","killer":{"steam_id":"76561198012345678","name":"[TF2]|Bot|2000","team":2},"victim":{"steam_id":"76561198087654321","name":"l33t|player","team":3},"weapon":{"name":"shotgun_primary"},"crit":false,"airborne":false}`,
			wantKiller: "[TF2]|Bot|2000",
			wantVictim: "l33t|player",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ParseLine(tt.line)
			if err != nil {
				t.Fatalf("ParseLine() error = %v", err)
			}

			if event.Kill.Killer.Name != tt.wantKiller {
				t.Errorf("Killer.Name = %q, want %q", event.Kill.Killer.Name, tt.wantKiller)
			}

			if event.Kill.Victim.Name != tt.wantVictim {
				t.Errorf("Victim.Name = %q, want %q", event.Kill.Victim.Name, tt.wantVictim)
			}
		})
	}
}

// TestParseDeflectEvent tests DEFLECT event parsing
func TestParseDeflectEvent(t *testing.T) {
	line := `{"timestamp":"2024-02-01T12:05:00","gamemode":"dodgeball","server_ip":"192.168.1.100","event_type":"deflect","player":{"steam_id":"76561198012345678","name":"DodgeballPro","team":2},"rocket_speed":1500.5,"deflect_angle":1.0000,"timing_ms":50,"distance":100.0}`

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

	if deflect.Player.SteamID != "76561198012345678" {
		t.Errorf("Player.SteamID = %v, want 76561198012345678", deflect.Player.SteamID)
	}

	if deflect.Player.Name != "DodgeballPro" {
		t.Errorf("Player.Name = %v, want DodgeballPro", deflect.Player.Name)
	}

	if deflect.RocketSpeed != 1500.5 {
		t.Errorf("RocketSpeed = %v, want 1500.5", deflect.RocketSpeed)
	}

	if deflect.DeflectAngle != 1.0000 {
		t.Errorf("DeflectAngle = %v, want 1.0000", deflect.DeflectAngle)
	}

	if deflect.TimingMs != 50 {
		t.Errorf("TimingMs = %v, want 50", deflect.TimingMs)
	}

	if deflect.Distance != 100.0 {
		t.Errorf("Distance = %v, want 100.0", deflect.Distance)
	}
}

// TestParseMatchStartEvent tests MATCH_START event parsing
func TestParseMatchStartEvent(t *testing.T) {
	line := `{"timestamp":"2024-02-01T12:00:00","gamemode":"default","server_ip":"192.168.1.100","event_type":"match_start","map":"cp_process_final"}`

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

	if matchStart.Map != "cp_process_final" {
		t.Errorf("Map = %v, want cp_process_final", matchStart.Map)
	}
}

// TestParseMatchEndEvent tests MATCH_END event parsing
func TestParseMatchEndEvent(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		wantTeam     int
		wantDuration int
	}{
		{
			name:         "RED wins",
			line:         `{"timestamp":"2024-02-01T12:10:00","gamemode":"default","server_ip":"192.168.1.100","event_type":"match_end","winner_team":2,"duration":600}`,
			wantTeam:     2,
			wantDuration: 600,
		},
		{
			name:         "BLU wins",
			line:         `{"timestamp":"2024-02-01T12:20:00","gamemode":"default","server_ip":"192.168.1.100","event_type":"match_end","winner_team":3,"duration":450}`,
			wantTeam:     3,
			wantDuration: 450,
		},
		{
			name:         "TIE",
			line:         `{"timestamp":"2024-02-01T12:30:00","gamemode":"default","server_ip":"192.168.1.100","event_type":"match_end","winner_team":0,"duration":300}`,
			wantTeam:     0,
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
			name:    "Invalid JSON",
			line:    `{"invalid json`,
			wantErr: true,
		},
		{
			name:    "Unknown event type",
			line:    `{"timestamp":"2024-02-01T12:00:00","gamemode":"default","server_ip":"192.168.1.100","event_type":"unknown"}`,
			wantErr: false, // Should skip, not error
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
				if tt.line == "" || tt.line[0] == '#' || strings.Contains(tt.line, `"event_type":"unknown"`) {
					if event != nil {
						t.Errorf("ParseLine() event = %v, want nil (should skip)", event)
					}
				}
			}
		})
	}
}

// TestParseAllFixtures tests parsing the entire JSON fixture file
func TestParseAllFixtures(t *testing.T) {
	file, err := os.Open("../../test/fixtures/sample_logs_json.txt")
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
			t.Logf("Line %d error: %v", lineNum, err)
			errorCount++
			continue
		}

		if event == nil {
			// Skipped line (comment, empty, unknown type)
			skippedLines++
			continue
		}

		validEvents++

		// Basic validation: all events should have gamemode and server_ip
		var gamemode, serverIP string
		switch event.Type {
		case events.EventTypeKill:
			gamemode = event.Kill.Gamemode
			serverIP = event.Kill.ServerIP
		case events.EventTypeDeflect:
			gamemode = event.Deflect.Gamemode
			serverIP = event.Deflect.ServerIP
		case events.EventTypeMatchStart:
			gamemode = event.MatchStart.Gamemode
			serverIP = event.MatchStart.ServerIP
		case events.EventTypeMatchEnd:
			gamemode = event.MatchEnd.Gamemode
			serverIP = event.MatchEnd.ServerIP
		}

		if gamemode == "" {
			t.Errorf("Line %d: event has empty gamemode", lineNum)
		}

		if serverIP == "" {
			t.Errorf("Line %d: event has empty server IP", lineNum)
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading fixture file: %v", err)
	}

	t.Logf("Parsed %d lines: %d valid events, %d skipped, %d errors",
		lineNum, validEvents, skippedLines, errorCount)

	// We expect a good number of valid events
	if validEvents < 20 {
		t.Errorf("Expected at least 20 valid events, got %d", validEvents)
	}
}

// BenchmarkParseKillEvent benchmarks KILL event parsing
func BenchmarkParseKillEvent(b *testing.B) {
	line := `{"timestamp":"2024-02-01T12:00:00","gamemode":"default","server_ip":"192.168.1.100","event_type":"kill","killer":{"steam_id":"76561198012345678","name":"Player1","team":2},"victim":{"steam_id":"76561198087654321","name":"Player2","team":3},"weapon":{"name":"scattergun"},"crit":false,"airborne":false}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseLine(line)
	}
}

// BenchmarkParseDeflectEvent benchmarks DEFLECT event parsing
func BenchmarkParseDeflectEvent(b *testing.B) {
	line := `{"timestamp":"2024-02-01T12:05:00","gamemode":"dodgeball","server_ip":"192.168.1.100","event_type":"deflect","player":{"steam_id":"76561198012345678","name":"DodgeballPro","team":2},"rocket_speed":1500.5,"deflect_angle":1.0000,"timing_ms":50,"distance":100.0}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseLine(line)
	}
}
