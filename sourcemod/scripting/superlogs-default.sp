/**
 * SuperLogs Default - Standard TF2 Event Tracking
 * 
 * Tracks kills, deaths, and basic match events for standard TF2 gamemodes.
 * Sends events as JSON via UDP.
 */

#include <sourcemod>
#include <tf2>
#include <tf2_stocks>
#include <superlogs-core>

#pragma semicolon 1
#pragma newdecls required

#define PLUGIN_VERSION "2.0.0"

// ConVars
ConVar g_cvCollectorHost;
ConVar g_cvCollectorPort;
ConVar g_cvGamemode;

// Match tracking
int g_iMatchStartTime;
bool g_bMatchActive;

public Plugin myinfo = {
    name = "SuperLogs - Default TF2",
    author = "UDL Stats Team",
    description = "Logs standard TF2 events to UnitedStats collector (JSON over UDP)",
    version = PLUGIN_VERSION,
    url = "https://github.com/UDL-TF/UnitedStats"
};

public void OnPluginStart() {
    // Create ConVars
    g_cvCollectorHost = CreateConVar("sm_superlogs_host", "stats.udl.tf", "Stats collector hostname");
    g_cvCollectorPort = CreateConVar("sm_superlogs_port", "27500", "Stats collector port");
    g_cvGamemode = CreateConVar("sm_superlogs_gamemode", "default", "Gamemode identifier");
    
    // Hook events
    HookEvent("player_death", Event_PlayerDeath);
    HookEvent("teamplay_round_start", Event_RoundStart);
    HookEvent("teamplay_round_win", Event_RoundWin);
    
    // Auto-exec config
    AutoExecConfig(true, "superlogs-default");
    
    LogMessage("[SuperLogs-Default] Plugin loaded v%s (JSON format)", PLUGIN_VERSION);
}

public void OnConfigsExecuted() {
    // Initialize SuperLogs core
    char host[128], gamemode[32];
    g_cvCollectorHost.GetString(host, sizeof(host));
    g_cvGamemode.GetString(gamemode, sizeof(gamemode));
    
    SuperLogs_Initialize(gamemode, host, g_cvCollectorPort.IntValue);
}

/**
 * Player death event - send kill event
 */
public void Event_PlayerDeath(Event event, const char[] name, bool dontBroadcast) {
    int victim = GetClientOfUserId(event.GetInt("userid"));
    int attacker = GetClientOfUserId(event.GetInt("attacker"));
    
    // Ignore invalid clients, suicides, world deaths
    if (!IsValidClient(victim)) return;
    if (!IsValidClient(attacker)) return;
    if (attacker == victim) return;
    
    // Get weapon
    char weapon[64];
    event.GetString("weapon", weapon, sizeof(weapon));
    
    // Get kill properties
    bool crit = event.GetInt("crit") > 0;
    bool airborne = IsClientAirborne(victim);
    
    // Get weapon item def index (for unusual weapons)
    int weaponEntity = event.GetInt("weapon_def_index");
    
    // Build and send JSON event
    LogEvent()
        .WithEventType("kill")
        .WithPlayer("killer", attacker)
        .WithPlayer("victim", victim)
        .WithWeapon(weapon, weaponEntity)
        .WithBool("crit", crit)
        .WithBool("airborne", airborne)
        .WithPlayerPosition("killer_pos", attacker)
        .WithPlayerPosition("victim_pos", victim)
        .Send();
}

/**
 * Round start event - send match_start
 */
public void Event_RoundStart(Event event, const char[] name, bool dontBroadcast) {
    // Get current map
    char mapName[128];
    GetCurrentMap(mapName, sizeof(mapName));
    
    // Send match_start event
    LogEvent()
        .WithEventType("match_start")
        .WithString("map", mapName)
        .Send();
    
    g_iMatchStartTime = GetTime();
    g_bMatchActive = true;
}

/**
 * Round win event - send match_end
 */
public void Event_RoundWin(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bMatchActive) return;
    
    int winnerTeam = event.GetInt("team");
    int duration = GetTime() - g_iMatchStartTime;
    
    // Send match_end event
    LogEvent()
        .WithEventType("match_end")
        .WithInt("winner_team", winnerTeam)
        .WithInt("duration", duration)
        .Send();
    
    g_bMatchActive = false;
}

/**
 * Check if client is airborne (for airshot detection)
 */
stock bool IsClientAirborne(int client) {
    if (!IsValidClient(client)) return false;
    
    // Get player flags
    int flags = GetEntityFlags(client);
    
    // Check if on ground
    if (flags & FL_ONGROUND) return false;
    
    // Check if in water (doesn't count as airborne)
    if (GetEntityFlags(client) & FL_INWATER) return false;
    
    return true;
}
