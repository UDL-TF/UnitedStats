# TF2 Comprehensive Event Logging - Implementation Summary

## Overview

We've implemented a comprehensive TF2 event logging system based on the HLstatsX SuperLogs-TF2 plugin by psychonic and CmptrWz. The system captures all major TF2 gameplay events and sends them as structured JSON events via UDP to the UnitedStats collector.

## Event Types Implemented

### Combat Events (11 types)
- **kill** - Player kills with full details (weapon, crit, airborne, headshot, backstab, first blood, domination, revenge)
- **assist** - Kill assists
- **headshot** - Sniper/Ambassador headshots
- **backstab** - Spy backstabs
- **airshot** - Airborne victim hits (rocket, sticky, pipebomb, arrow, flare, stun)
- **deflect** - Pyro airblast deflections (rockets, pipes, flares, arrows, jarate, players)
- **jarate** - Jarate and Mad Milk application
- **mad_milk** - Mad Milk application (via jarate event with type)
- **shield_blocked** - Demoman shield blocking damage
- **stun** - Sandman ball stuns
- **domination** - Player domination
- **revenge** - Revenge kills

### Movement Events (6 types)
- **rocket_jump** - Soldier rocket jumps
- **sticky_jump** - Demoman sticky jumps
- **rocket_jump_kill** - Kills while rocket jumping
- **sticky_jump_kill** - Kills while sticky jumping
- **teleport** - Teleporter usage (with repeat detection)
- **teleport_used** - Player using teleporter

### Building Events (3 types)
- **built_object** - Building construction (sentry, dispenser, teleporter)
- **killed_object** - Building destruction by enemy
- **object_destroyed** - Building destroyed (owner perspective)

### Medic Events (4 types)
- **healed** - Healing points accumulated
- **defended_medic** - Defender bonus for protecting medic
- **uber_deployed** - ÜberCharge deployment
- **uber_dropped** - ÜberCharge dropped on death

### Support Events (2 types)
- **buff_deployed** - Battalion's Backup/Buff Banner deployment
- **food** events:
  - **sandvich** - Heavy eating Sandvich
  - **dalokohs** - Heavy eating Dalokohs Bar
  - **steak** - Heavy eating Buffalo Steak Sandvich

### Match/Round Events (6 types)
- **match_start** - Match/round start
- **match_end** - Match/round end with winner and duration
- **round_start** - Round start
- **round_end** - Round end
- **mvp1**, **mvp2**, **mvp3** - Top 3 players at round end

### Player Events (4 types)
- **player_loadout** - Player's equipped items (all 8 slots)
- **player_spawn** - Player spawning
- **player_disconnect** - Player disconnect
- **class_change** - Player changing class

### Statistics Events (2 types)
- **weapon_stats** - Weapon usage statistics (shots, hits, kills, headshots, damage, deaths)
- **first_blood** - First kill of the round

## Total: 40+ unique event types

## Data Structure

Each event contains:

### Base Fields (all events)
```json
{
  "timestamp": "2024-01-31T22:00:00",
  "gamemode": "default",
  "server_ip": "192.168.1.100",
  "event_type": "kill"
}
```

### Player Object
```json
{
  "steam_id": "76561198012345678",
  "name": "PlayerName",
  "team": 2,
  "class": "soldier"
}
```

### Position Object
```json
{
  "x": 1234.56,
  "y": 7890.12,
  "z": 345.67
}
```

### Weapon Object
```json
{
  "name": "rocketlauncher",
  "item_def_index": 205
}
```

## Example Events

### Kill Event
```json
{
  "timestamp": "2024-01-31T22:00:00",
  "gamemode": "default",
  "server_ip": "192.168.1.100",
  "event_type": "kill",
  "killer": {
    "steam_id": "76561198012345678",
    "name": "Player1",
    "team": 2,
    "class": "soldier"
  },
  "victim": {
    "steam_id": "76561198087654321",
    "name": "Player2",
    "team": 3,
    "class": "scout"
  },
  "weapon": {
    "name": "rocketlauncher",
    "item_def_index": 205
  },
  "crit": false,
  "airborne": true,
  "headshot": false,
  "backstab": false,
  "first_blood": false,
  "killer_pos": {"x": 1234.56, "y": 7890.12, "z": 345.67},
  "victim_pos": {"x": 2345.67, "y": 8901.23, "z": 456.78}
}
```

### Airshot Event
```json
{
  "timestamp": "2024-01-31T22:00:00",
  "gamemode": "default",
  "server_ip": "192.168.1.100",
  "event_type": "airshot",
  "player": {
    "steam_id": "76561198012345678",
    "name": "Player1",
    "team": 2,
    "class": "soldier"
  },
  "victim": {
    "steam_id": "76561198087654321",
    "name": "Player2",
    "team": 3,
    "class": "scout"
  },
  "weapon_type": "rocket",
  "air2air": true,
  "player_pos": {"x": 1234.56, "y": 7890.12, "z": 345.67},
  "victim_pos": {"x": 2345.67, "y": 8901.23, "z": 456.78}
}
```

### Deflect Event
```json
{
  "timestamp": "2024-01-31T22:00:00",
  "gamemode": "default",
  "server_ip": "192.168.1.100",
  "event_type": "deflect",
  "player": {
    "steam_id": "76561198012345678",
    "name": "PyroPlayer",
    "team": 2,
    "class": "pyro"
  },
  "owner": {
    "steam_id": "76561198087654321",
    "name": "SoldierPlayer",
    "team": 3,
    "class": "soldier"
  },
  "projectile_type": "rocket",
  "player_pos": {"x": 1234.56, "y": 7890.12, "z": 345.67}
}
```

## Feature Toggles

The plugin supports granular control via ConVars:

```
sm_superlogs_actions 1        // Player actions (stuns, deflects, etc.)
sm_superlogs_teleports 1      // Teleporter usage
sm_superlogs_headshots 1      // Headshot events
sm_superlogs_backstabs 1      // Backstab events
sm_superlogs_airshots 1       // Airshot detection
sm_superlogs_jumps 1          // Rocket/sticky jumps
sm_superlogs_buildings 1      // Building events
sm_superlogs_healing 1        // Medic healing
sm_superlogs_weaponstats 1    // Weapon statistics
sm_superlogs_loadouts 1       // Player loadouts
```

## Advanced Features

### Airshot Detection
- Detects when victim is airborne (not on ground, not in water)
- Tracks weapon type (rocket, sticky, pipebomb, arrow, flare)
- Detects air-to-air shots (both players airborne)
- Position tracking for both players

### Jump Tracking
- Distinguishes between rocket and sticky jumps
- Detects taunt kill fake jumps
- Tracks kills while jumping
- Resets on landing or death

### Teleporter Anti-Padding
- Tracks last teleport time per builder/user pair
- Marks repeated uses within 10 seconds
- Detects self-teleports

### Weapon Statistics
- Tracks shots, hits, kills, headshots, damage, deaths per weapon
- Dumps stats on death, class change, round end, disconnect
- Placeholder for full HLstatsX-style weapon trie system

### Player Loadout Tracking
- Captures all 8 equipment slots
- Tracks item definition indices
- Logs on spawn and equipment changes

## Technical Implementation

### Go Parser (internal/parser/parser.go)
- Parses JSON event lines
- Type-safe event unmarshaling
- Error handling with detailed parse errors
- Skips unknown event types gracefully

### Event Structures (pkg/events/events.go)
- Type-safe event definitions
- Union type for all events
- Comprehensive field documentation
- Optional fields with proper JSON omitempty tags

### SourceMod Plugin (sourcemod/scripting/superlogs-tf2.sp)
- ~1000 lines of event tracking code
- SDKHooks integration for damage tracking
- UserMessage hooks for jarate/shield blocks
- Timer-based loadout detection
- State tracking for jumps, healing, teleports

## Files Modified/Created

1. **pkg/events/events.go** - Complete event type definitions
2. **internal/parser/parser.go** - Parser for all event types
3. **sourcemod/scripting/superlogs-tf2.sp** - Comprehensive TF2 plugin

## Based On

- **HLstatsX SuperLogs-TF2** by psychonic & CmptrWz
- https://github.com/NomisCZ/hlstatsx-community-edition/blob/master/sourcemod/scripting/superlogs-tf2.sp
- Adapted from file-based logging to JSON/UDP streaming
- Maintains compatibility with HLstatsX event semantics

## Next Steps

1. Test plugin compilation with SourceMod 1.11+
2. Add SM-JSON dependency documentation
3. Implement weapon trie system for full weapon stats
4. Add dodgeball-specific events (separate plugin)
5. Create unit tests for all event parsers
6. Add integration tests with sample event data

## Performance Considerations

- Event hooks only registered when features enabled
- Minimal state tracking (jump status, heal points, teleport times)
- Weapon stats use fixed-size arrays (not dynamic)
- Position tracking optional via ConVars
- UDP fire-and-forget (no blocking I/O)

## Compatibility

- **SourceMod**: 1.11 or higher
- **TF2**: All game modes
- **Required Extensions**: SDKHooks, SM-JSON
- **Optional**: socket extension for UDP
