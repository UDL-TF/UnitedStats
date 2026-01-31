/**
 * SuperLogs Default - Standard TF2 Event Tracking
 * 
 * Tracks kills, deaths, and basic match events for standard TF2 gamemodes.
 * 
 * Sends events:
 * - KILL: Player kills
 * - MATCH_START: Round start
 * - MATCH_END: Round end
 */

#include <sourcemod>
#include <tf2>
#include <tf2_stocks>
#include <superlogs-core>

#pragma semicolon 1
#pragma newdecls required

#define PLUGIN_VERSION "1.0.0"

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
    description = "Logs standard TF2 events to UnitedStats collector",
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
    
    LogMessage("[SuperLogs-Default] Plugin loaded v%s", PLUGIN_VERSION);
}

public void OnConfigsExecuted() {
    // Initialize SuperLogs core
    char host[128], gamemode[32];
    g_cvCollectorHost.GetString(host, sizeof(host));
    g_cvGamemode.GetString(gamemode, sizeof(gamemode));
    
    SuperLogs_Initialize(gamemode, host, g_cvCollectorPort.IntValue);
}

/**
 * Player death event - send KILL event
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
    int crit = event.GetInt("crit");
    int airborne = IsClientAirborne(victim) ? 1 : 0;
    
    // Send KILL event
    SuperLogs_SendKill(attacker, victim, weapon, crit, airborne);
}

/**
 * Round start event - send MATCH_START
 */
public void Event_RoundStart(Event event, const char[] name, bool dontBroadcast) {
    // Get current map
    char mapName[128];
    GetCurrentMap(mapName, sizeof(mapName));
    
    // Send MATCH_START
    SuperLogs_SendMatchStart(mapName);
    
    g_iMatchStartTime = GetTime();
    g_bMatchActive = true;
}

/**
 * Round win event - send MATCH_END
 */
public void Event_RoundWin(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bMatchActive) return;
    
    int winnerTeam = event.GetInt("team");
    int duration = GetTime() - g_iMatchStartTime;
    
    // Send MATCH_END
    SuperLogs_SendMatchEnd(winnerTeam, duration);
    
    g_bMatchActive = false;
}

/**
 * Check if client is valid player
 */
stock bool IsValidClient(int client) {
    return (client > 0 && 
            client <= MaxClients && 
            IsClientInGame(client) && 
            !IsFakeClient(client));
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
