/**
 * SuperLogs TF2 - Comprehensive TF2 Event Tracking
 * 
 * Based on HLstatsX SuperLogs-TF2 plugin by psychonic & CmptrWz
 * Adapted to JSON/UDP format for UnitedStats
 * 
 * Tracks all standard TF2 events including:
 * - Kills, deaths, assists
 * - Airshots, headshots, backstabs
 * - Rocket/sticky jumps and jump kills
 * - Teleporter usage
 * - Building construction/destruction
 * - Medic healing and ubers
 * - Weapon statistics
 * - Player loadouts
 * - And much more...
 */

#include <sourcemod>
#include <tf2>
#include <tf2_stocks>
#include <sdkhooks>
#include <superlogs-core>

#pragma semicolon 1
#pragma newdecls required

#define PLUGIN_VERSION "2.0.0"
#define TELEPORT_AGAIN_TIME 10.0

// Object types
#define OBJ_DISPENSER 0
#define OBJ_TELEPORTER_ENTRANCE 1
#define OBJ_TELEPORTER_EXIT 2
#define OBJ_SENTRYGUN 3
#define OBJ_ATTACHMENT_SAPPER 4

// Jump status
#define JUMP_NONE 0
#define JUMP_ROCKET_START 1
#define JUMP_ROCKET 2
#define JUMP_STICKY 3

// ConVars
ConVar g_cvCollectorHost;
ConVar g_cvCollectorPort;
ConVar g_cvGamemode;

// Feature toggles
ConVar g_cvLogActions;
ConVar g_cvLogTeleports;
ConVar g_cvLogHeadshots;
ConVar g_cvLogBackstabs;
ConVar g_cvLogAirshots;
ConVar g_cvLogJumps;
ConVar g_cvLogBuildings;
ConVar g_cvLogHealing;
ConVar g_cvLogWeaponStats;
ConVar g_cvLogLoadouts;

// Feature states
bool g_bLogActions;
bool g_bLogTeleports;
bool g_bLogHeadshots;
bool g_bLogBackstabs;
bool g_bLogAirshots;
bool g_bLogJumps;
bool g_bLogBuildings;
bool g_bLogHealing;
bool g_bLogWeaponStats;
bool g_bLogLoadouts;

// Player state tracking
int g_iJumpStatus[MAXPLAYERS+1];
int g_iHealPoints[MAXPLAYERS+1];
float g_fLastTeleport[MAXPLAYERS+1][MAXPLAYERS+1];
TFClassType g_iPlayerClass[MAXPLAYERS+1];

// Weapon stats tracking
int g_iWeaponShots[MAXPLAYERS+1][32];
int g_iWeaponHits[MAXPLAYERS+1][32];
int g_iWeaponKills[MAXPLAYERS+1][32];
int g_iWeaponHeadshots[MAXPLAYERS+1][32];
int g_iWeaponDamage[MAXPLAYERS+1][32];
int g_iWeaponDeaths[MAXPLAYERS+1][32];

// Match tracking
int g_iMatchStartTime;
bool g_bMatchActive;

public Plugin myinfo = {
    name = "SuperLogs - TF2 Complete",
    author = "UDL Stats Team (based on psychonic & CmptrWz work)",
    description = "Comprehensive TF2 event logging to UnitedStats collector (JSON/UDP)",
    version = PLUGIN_VERSION,
    url = "https://github.com/UDL-TF/UnitedStats"
};

public void OnPluginStart() {
    // Core config
    g_cvCollectorHost = CreateConVar("sm_superlogs_host", "stats.udl.tf", "Stats collector hostname");
    g_cvCollectorPort = CreateConVar("sm_superlogs_port", "27500", "Stats collector port");
    g_cvGamemode = CreateConVar("sm_superlogs_gamemode", "default", "Gamemode identifier");
    
    // Feature toggles
    g_cvLogActions = CreateConVar("sm_superlogs_actions", "1", "Log player actions (stuns, deflects, etc.)");
    g_cvLogTeleports = CreateConVar("sm_superlogs_teleports", "1", "Log teleporter usage");
    g_cvLogHeadshots = CreateConVar("sm_superlogs_headshots", "1", "Log headshot events");
    g_cvLogBackstabs = CreateConVar("sm_superlogs_backstabs", "1", "Log backstab events");
    g_cvLogAirshots = CreateConVar("sm_superlogs_airshots", "1", "Log airshot events");
    g_cvLogJumps = CreateConVar("sm_superlogs_jumps", "1", "Log rocket/sticky jumps");
    g_cvLogBuildings = CreateConVar("sm_superlogs_buildings", "1", "Log building events");
    g_cvLogHealing = CreateConVar("sm_superlogs_healing", "1", "Log medic healing");
    g_cvLogWeaponStats = CreateConVar("sm_superlogs_weaponstats", "1", "Log weapon statistics");
    g_cvLogLoadouts = CreateConVar("sm_superlogs_loadouts", "1", "Log player loadouts");
    
    // Hook ConVar changes
    g_cvLogActions.AddChangeHook(OnConVarChanged);
    g_cvLogTeleports.AddChangeHook(OnConVarChanged);
    g_cvLogHeadshots.AddChangeHook(OnConVarChanged);
    g_cvLogBackstabs.AddChangeHook(OnConVarChanged);
    g_cvLogAirshots.AddChangeHook(OnConVarChanged);
    g_cvLogJumps.AddChangeHook(OnConVarChanged);
    g_cvLogBuildings.AddChangeHook(OnConVarChanged);
    g_cvLogHealing.AddChangeHook(OnConVarChanged);
    g_cvLogWeaponStats.AddChangeHook(OnConVarChanged);
    g_cvLogLoadouts.AddChangeHook(OnConVarChanged);
    
    // Hook core events (always on)
    HookEvent("player_death", Event_PlayerDeath);
    HookEvent("teamplay_round_start", Event_RoundStart);
    HookEvent("teamplay_round_win", Event_RoundWin);
    HookEvent("arena_win_panel", Event_WinPanel);
    HookEvent("teamplay_win_panel", Event_WinPanel);
    HookEvent("player_spawn", Event_PlayerSpawn);
    HookEvent("player_disconnect", Event_PlayerDisconnect, EventHookMode_Pre);
    
    AutoExecConfig(true, "superlogs-tf2");
    
    LogMessage("[SuperLogs-TF2] Plugin loaded v%s", PLUGIN_VERSION);
}

public void OnConfigsExecuted() {
    // Initialize SuperLogs core
    char host[128], gamemode[32];
    g_cvCollectorHost.GetString(host, sizeof(host));
    g_cvGamemode.GetString(gamemode, sizeof(gamemode));
    
    SuperLogs_Initialize(gamemode, host, g_cvCollectorPort.IntValue);
    
    // Update feature states
    UpdateFeatureStates();
    
    // Hook/unhook events based on config
    UpdateEventHooks();
}

void UpdateFeatureStates() {
    g_bLogActions = g_cvLogActions.BoolValue;
    g_bLogTeleports = g_cvLogTeleports.BoolValue;
    g_bLogHeadshots = g_cvLogHeadshots.BoolValue;
    g_bLogBackstabs = g_cvLogBackstabs.BoolValue;
    g_bLogAirshots = g_cvLogAirshots.BoolValue;
    g_bLogJumps = g_cvLogJumps.BoolValue;
    g_bLogBuildings = g_cvLogBuildings.BoolValue;
    g_bLogHealing = g_cvLogHealing.BoolValue;
    g_bLogWeaponStats = g_cvLogWeaponStats.BoolValue;
    g_bLogLoadouts = g_cvLogLoadouts.BoolValue;
}

void UpdateEventHooks() {
    // Actions
    if (g_bLogActions) {
        HookEvent("player_stunned", Event_PlayerStunned);
        HookEvent("object_deflected", Event_ObjectDeflected);
        HookEvent("deploy_buff_banner", Event_DeployBuff);
        HookEvent("medic_defended", Event_MedicDefended);
        HookUserMessage(GetUserMessageId("PlayerJarated"), UserMsg_PlayerJarated);
        HookUserMessage(GetUserMessageId("PlayerShieldBlocked"), UserMsg_PlayerShieldBlocked);
    }
    
    // Teleports
    if (g_bLogTeleports) {
        HookEvent("player_teleported", Event_PlayerTeleported);
    }
    
    // Jumps
    if (g_bLogJumps) {
        HookEvent("rocket_jump", Event_RocketJump);
        HookEvent("rocket_jump_landed", Event_JumpLanded);
        HookEvent("sticky_jump", Event_StickyJump);
        HookEvent("sticky_jump_landed", Event_JumpLanded);
    }
    
    // Buildings
    if (g_bLogBuildings) {
        HookEvent("player_builtobject", Event_PlayerBuiltObject);
        HookEvent("object_destroyed", Event_ObjectDestroyed);
    }
    
    // Healing
    if (g_bLogHealing) {
        HookEvent("medic_death", Event_MedicDeath);
    }
    
    // Loadouts
    if (g_bLogLoadouts) {
        HookEvent("post_inventory_application", Event_PostInventoryApplication);
    }
}

public void OnConVarChanged(ConVar convar, const char[] oldValue, const char[] newValue) {
    UpdateFeatureStates();
}

public void OnClientPutInServer(int client) {
    if (!IsValidClient(client)) return;
    
    // Reset tracking
    g_iJumpStatus[client] = JUMP_NONE;
    g_iPlayerClass[client] = TFClass_Unknown;
    g_iHealPoints[client] = 0;
    
    for (int i = 0; i <= MaxClients; i++) {
        g_fLastTeleport[client][i] = 0.0;
        g_fLastTeleport[i][client] = 0.0;
    }
    
    // Reset weapon stats
    for (int i = 0; i < 32; i++) {
        g_iWeaponShots[client][i] = 0;
        g_iWeaponHits[client][i] = 0;
        g_iWeaponKills[client][i] = 0;
        g_iWeaponHeadshots[client][i] = 0;
        g_iWeaponDamage[client][i] = 0;
        g_iWeaponDeaths[client][i] = 0;
    }
    
    // Hook damage for airshot/weapon stats
    SDKHook(client, SDKHook_OnTakeDamage, OnTakeDamage);
}

//=============================================================================
// CORE EVENTS - Always logged
//=============================================================================

/**
 * Player death - send kill event with full details
 */
public void Event_PlayerDeath(Event event, const char[] name, bool dontBroadcast) {
    int victim = GetClientOfUserId(event.GetInt("userid"));
    int attacker = GetClientOfUserId(event.GetInt("attacker"));
    int assister = GetClientOfUserId(event.GetInt("assister"));
    
    if (!IsValidClient(victim)) return;
    
    // Get weapon
    char weapon[64];
    event.GetString("weapon_logclassname", weapon, sizeof(weapon));
    int weaponDefIndex = event.GetInt("weapon_def_index");
    
    // Get kill properties
    int customKill = event.GetInt("customkill");
    int deathFlags = event.GetInt("death_flags");
    bool crit = (event.GetInt("damagebits") & DMG_ACID) > 0;
    bool airborne = IsClientAirborne(victim);
    bool headshot = (customKill == TF_CUSTOM_HEADSHOT);
    bool backstab = (customKill == TF_CUSTOM_BACKSTAB);
    bool firstBlood = (deathFlags & TF_DEATHFLAG_FIRSTBLOOD) > 0;
    bool domination = event.GetBool("dominated");
    bool revenge = event.GetBool("revenge");
    
    // Build kill event
    LogEvent event = LogEvent()
        .WithEventType("kill")
        .WithPlayer("victim", victim)
        .WithWeapon(weapon, weaponDefIndex)
        .WithBool("crit", crit)
        .WithBool("airborne", airborne)
        .WithBool("headshot", headshot)
        .WithBool("backstab", backstab)
        .WithPlayerPosition("victim_pos", victim)
        .WithInt("custom_kill", customKill);
    
    if (IsValidClient(attacker)) {
        event.WithPlayer("killer", attacker);
        event.WithPlayerPosition("killer_pos", attacker);
        
        if (firstBlood) event.WithBool("first_blood", true);
        if (domination) event.WithBool("domination", true);
        if (revenge) event.WithBool("revenge", true);
        
        // Log additional action events
        if (g_bLogHeadshots && headshot) {
            LogEvent().WithEventType("headshot")
                .WithPlayer("player", attacker)
                .WithPlayer("victim", victim)
                .Send();
        }
        
        if (g_bLogBackstabs && backstab) {
            LogEvent().WithEventType("backstab")
                .WithPlayer("player", attacker)
                .WithPlayer("victim", victim)
                .Send();
        }
        
        // Check for jump kills
        if (g_bLogJumps && g_iJumpStatus[attacker] != JUMP_NONE) {
            char jumpType[16];
            strcopy(jumpType, sizeof(jumpType), g_iJumpStatus[attacker] == JUMP_ROCKET ? "rocket" : "sticky");
            
            LogEvent().WithEventType(g_iJumpStatus[attacker] == JUMP_ROCKET ? "rocket_jump_kill" : "sticky_jump_kill")
                .WithPlayer("player", attacker)
                .WithPlayer("victim", victim)
                .WithString("jump_type", jumpType)
                .WithPlayerPosition("player_pos", attacker)
                .WithPlayerPosition("victim_pos", victim)
                .Send();
        }
    }
    
    if (IsValidClient(assister)) {
        event.WithPlayer("assister", assister);
    }
    
    event.Send();
    
    // Reset jump status on death
    g_iJumpStatus[victim] = JUMP_NONE;
    
    // Dump healing stats if medic
    if (g_bLogHealing && TF2_GetPlayerClass(victim) == TFClass_Medic) {
        DumpHealingStats(victim, "death");
    }
    
    // Dump weapon stats
    if (g_bLogWeaponStats) {
        DumpWeaponStats(victim);
    }
}

/**
 * Round start
 */
public void Event_RoundStart(Event event, const char[] name, bool dontBroadcast) {
    char mapName[128];
    GetCurrentMap(mapName, sizeof(mapName));
    
    LogEvent()
        .WithEventType("round_start")
        .WithString("map", mapName)
        .Send();
    
    g_iMatchStartTime = GetTime();
    g_bMatchActive = true;
}

/**
 * Round end
 */
public void Event_RoundWin(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bMatchActive) return;
    
    int winnerTeam = event.GetInt("team");
    int duration = GetTime() - g_iMatchStartTime;
    
    LogEvent()
        .WithEventType("round_end")
        .WithInt("winner_team", winnerTeam)
        .WithInt("duration", duration)
        .Send();
    
    g_bMatchActive = false;
    
    // Dump all weapon stats at round end
    if (g_bLogWeaponStats) {
        for (int i = 1; i <= MaxClients; i++) {
            if (IsValidClient(i)) {
                DumpWeaponStats(i);
            }
        }
    }
}

/**
 * Win panel (MVP tracking)
 */
public void Event_WinPanel(Event event, const char[] name, bool dontBroadcast) {
    int player1 = event.GetInt("player_1");
    int player2 = event.GetInt("player_2");
    int player3 = event.GetInt("player_3");
    
    if (player1 > 0) {
        LogEvent().WithEventType("mvp1")
            .WithPlayer("player", player1)
            .WithInt("position", 1)
            .Send();
    }
    
    if (player2 > 0) {
        LogEvent().WithEventType("mvp2")
            .WithPlayer("player", player2)
            .WithInt("position", 2)
            .Send();
    }
    
    if (player3 > 0) {
        LogEvent().WithEventType("mvp3")
            .WithPlayer("player", player3)
            .WithInt("position", 3)
            .Send();
    }
}

/**
 * Player spawn
 */
public void Event_PlayerSpawn(Event event, const char[] name, bool dontBroadcast) {
    int client = GetClientOfUserId(event.GetInt("userid"));
    if (!IsValidClient(client)) return;
    
    TFClassType newClass = TF2_GetPlayerClass(client);
    TFClassType oldClass = g_iPlayerClass[client];
    
    // Reset jump status
    g_iJumpStatus[client] = JUMP_NONE;
    
    // Log class change
    if (oldClass != newClass && oldClass != TFClass_Unknown) {
        char oldClassName[32], newClassName[32];
        TF2_GetClassName(oldClass, oldClassName, sizeof(oldClassName));
        TF2_GetClassName(newClass, newClassName, sizeof(newClassName));
        
        LogEvent()
            .WithEventType("class_change")
            .WithPlayer("player", client)
            .WithString("old_class", oldClassName)
            .WithString("new_class", newClassName)
            .Send();
        
        // Dump weapon stats on class change
        if (g_bLogWeaponStats) {
            DumpWeaponStats(client);
        }
    }
    
    g_iPlayerClass[client] = newClass;
    
    // Dump healing stats on spawn
    if (g_bLogHealing) {
        DumpHealingStats(client, "spawn");
    }
}

/**
 * Player disconnect
 */
public void Event_PlayerDisconnect(Event event, const char[] name, bool dontBroadcast) {
    int client = GetClientOfUserId(event.GetInt("userid"));
    if (!IsValidClient(client)) return;
    
    // Dump stats before disconnect
    if (g_bLogWeaponStats) {
        DumpWeaponStats(client);
    }
    
    if (g_bLogHealing) {
        DumpHealingStats(client, "disconnect");
    }
}

//=============================================================================
// ACTION EVENTS
//=============================================================================

/**
 * Player stunned (sandman ball)
 */
public void Event_PlayerStunned(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bLogActions) return;
    
    int stunner = GetClientOfUserId(event.GetInt("stunner"));
    int victim = GetClientOfUserId(event.GetInt("victim"));
    
    if (!IsValidClient(stunner) || !IsValidClient(victim)) return;
    
    bool victimCapping = event.GetBool("victim_capping");
    bool bigStun = event.GetBool("big_stun");
    bool airshot = IsClientAirborne(victim);
    
    LogEvent()
        .WithEventType("stun")
        .WithPlayer("stunner", stunner)
        .WithPlayer("victim", victim)
        .WithBool("victim_capping", victimCapping)
        .WithBool("big_stun", bigStun)
        .WithBool("airshot", airshot)
        .WithPlayerPosition("stunner_pos", stunner)
        .WithPlayerPosition("victim_pos", victim)
        .Send();
}

/**
 * Object deflected (airblast)
 */
public void Event_ObjectDeflected(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bLogActions) return;
    
    int client = GetClientOfUserId(event.GetInt("userid"));
    int owner = GetClientOfUserId(event.GetInt("ownerid"));
    int weaponID = event.GetInt("weaponid");
    
    if (!IsValidClient(client)) return;
    
    char projectileType[32];
    switch (weaponID) {
        case TF_WEAPON_NONE: strcopy(projectileType, sizeof(projectileType), "player");
        case TF_WEAPON_ROCKETLAUNCHER, TF_WEAPON_DIRECTHIT: strcopy(projectileType, sizeof(projectileType), "rocket");
        case TF_WEAPON_GRENADE_DEMOMAN: strcopy(projectileType, sizeof(projectileType), "pipebomb");
        case TF_WEAPON_FLAREGUN: strcopy(projectileType, sizeof(projectileType), "flare");
        case TF_WEAPON_COMPOUND_BOW: strcopy(projectileType, sizeof(projectileType), "arrow");
        case TF_WEAPON_JAR: strcopy(projectileType, sizeof(projectileType), "jarate");
        case TF_WEAPON_GRENADE_STUNBALL: strcopy(projectileType, sizeof(projectileType), "stunball");
        default: strcopy(projectileType, sizeof(projectileType), "unknown");
    }
    
    LogEvent event = LogEvent()
        .WithEventType("deflect")
        .WithPlayer("player", client)
        .WithString("projectile_type", projectileType)
        .WithPlayerPosition("player_pos", client);
    
    if (IsValidClient(owner)) {
        event.WithPlayer("owner", owner);
    }
    
    event.Send();
}

/**
 * Buff deployed (Battalion's Backup, etc.)
 */
public void Event_DeployBuff(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bLogActions) return;
    
    int client = GetClientOfUserId(event.GetInt("buff_owner"));
    if (!IsValidClient(client)) return;
    
    LogEvent()
        .WithEventType("buff_deployed")
        .WithPlayer("player", client)
        .WithString("buff_type", "buff")
        .Send();
}

/**
 * Medic defended
 */
public void Event_MedicDefended(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bLogActions) return;
    
    int client = GetClientOfUserId(event.GetInt("userid"));
    if (!IsValidClient(client)) return;
    
    LogEvent()
        .WithEventType("defended_medic")
        .WithPlayer("player", client)
        .Send();
}

/**
 * Player jarated/milked (user message)
 */
public Action UserMsg_PlayerJarated(UserMsg msg_id, Handle bf, const int[] players, int playersNum, bool reliable, bool init) {
    if (!g_bLogActions) return Plugin_Continue;
    
    int attacker = BfReadByte(bf);
    int victim = BfReadByte(bf);
    
    if (!IsValidClient(attacker) || !IsValidClient(victim)) return Plugin_Continue;
    
    char jarType[16];
    if (TF2_IsPlayerInCondition(victim, TFCond_Jarated)) {
        strcopy(jarType, sizeof(jarType), "jarate");
    } else if (TF2_IsPlayerInCondition(victim, TFCond_Milked)) {
        strcopy(jarType, sizeof(jarType), "mad_milk");
    } else {
        return Plugin_Continue;
    }
    
    LogEvent()
        .WithEventType("jarate")
        .WithPlayer("attacker", attacker)
        .WithPlayer("victim", victim)
        .WithString("jar_type", jarType)
        .Send();
    
    return Plugin_Continue;
}

/**
 * Shield blocked damage (user message)
 */
public Action UserMsg_PlayerShieldBlocked(UserMsg msg_id, Handle bf, const int[] players, int playersNum, bool reliable, bool init) {
    if (!g_bLogActions) return Plugin_Continue;
    
    int victim = BfReadByte(bf);
    int attacker = BfReadByte(bf);
    
    if (!IsValidClient(attacker) || !IsValidClient(victim)) return Plugin_Continue;
    
    LogEvent()
        .WithEventType("shield_blocked")
        .WithPlayer("blocker", victim)
        .WithPlayer("attacker", attacker)
        .Send();
    
    return Plugin_Continue;
}

//=============================================================================
// JUMP EVENTS
//=============================================================================

public void Event_RocketJump(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bLogJumps) return;
    
    int client = GetClientOfUserId(event.GetInt("userid"));
    if (!IsValidClient(client)) return;
    
    // Taunt kills trigger one event, real jumps trigger two
    if (g_iJumpStatus[client] == JUMP_ROCKET_START) {
        g_iJumpStatus[client] = JUMP_ROCKET;
        
        LogEvent()
            .WithEventType("rocket_jump")
            .WithPlayer("player", client)
            .WithString("jump_type", "rocket")
            .WithPlayerPosition("player_pos", client)
            .Send();
    } else if (g_iJumpStatus[client] != JUMP_ROCKET) {
        g_iJumpStatus[client] = JUMP_ROCKET_START;
    }
}

public void Event_StickyJump(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bLogJumps) return;
    
    int client = GetClientOfUserId(event.GetInt("userid"));
    if (!IsValidClient(client)) return;
    
    if (g_iJumpStatus[client] != JUMP_STICKY) {
        g_iJumpStatus[client] = JUMP_STICKY;
        
        LogEvent()
            .WithEventType("sticky_jump")
            .WithPlayer("player", client)
            .WithString("jump_type", "sticky")
            .WithPlayerPosition("player_pos", client)
            .Send();
    }
}

public void Event_JumpLanded(Event event, const char[] name, bool dontBroadcast) {
    int client = GetClientOfUserId(event.GetInt("userid"));
    if (!IsValidClient(client)) return;
    
    g_iJumpStatus[client] = JUMP_NONE;
}

//=============================================================================
// TELEPORTER EVENTS
//=============================================================================

public void Event_PlayerTeleported(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bLogTeleports) return;
    
    int builder = GetClientOfUserId(event.GetInt("builderid"));
    int user = GetClientOfUserId(event.GetInt("userid"));
    
    if (!IsValidClient(builder)) return;
    
    float curTime = GetGameTime();
    bool repeated = (g_fLastTeleport[builder][user] > curTime - TELEPORT_AGAIN_TIME);
    bool selfUsed = (builder == user);
    
    LogEvent event = LogEvent()
        .WithEventType("teleport")
        .WithPlayer("builder", builder)
        .WithBool("self_used", selfUsed)
        .WithBool("repeated", repeated);
    
    if (!selfUsed && IsValidClient(user)) {
        event.WithPlayer("user", user);
    }
    
    event.Send();
    
    g_fLastTeleport[builder][user] = curTime;
}

//=============================================================================
// BUILDING EVENTS
//=============================================================================

public void Event_PlayerBuiltObject(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bLogBuildings) return;
    
    int client = GetClientOfUserId(event.GetInt("userid"));
    if (!IsValidClient(client)) return;
    
    int objectType = event.GetInt("object");
    int objectIndex = event.GetInt("index");
    
    char objectName[32];
    GetObjectName(objectType, objectName, sizeof(objectName));
    
    LogEvent()
        .WithEventType("built_object")
        .WithPlayer("player", client)
        .WithString("object.type", objectName)
        .Send();
}

public void Event_ObjectDestroyed(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bLogBuildings) return;
    
    int attacker = GetClientOfUserId(event.GetInt("attacker"));
    int owner = GetClientOfUserId(event.GetInt("userid"));
    
    if (!IsValidClient(attacker) || !IsValidClient(owner)) return;
    
    int objectType = event.GetInt("objecttype");
    char weapon[64];
    event.GetString("weapon", weapon, sizeof(weapon));
    
    char objectName[32];
    GetObjectName(objectType, objectName, sizeof(objectName));
    
    LogEvent()
        .WithEventType("killed_object")
        .WithPlayer("attacker", attacker)
        .WithPlayer("owner", owner)
        .WithString("object.type", objectName)
        .WithString("weapon.name", weapon)
        .Send();
}

//=============================================================================
// MEDIC/HEALING EVENTS
//=============================================================================

public void Event_MedicDeath(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bLogHealing) return;
    
    int client = GetClientOfUserId(event.GetInt("userid"));
    if (!IsValidClient(client)) return;
    
    g_iHealPoints[client] = GetEntProp(client, Prop_Send, "m_iHealPoints");
}

void DumpHealingStats(int client, const char[] reason) {
    if (!IsValidClient(client)) return;
    
    int currentHeals = GetEntProp(client, Prop_Send, "m_iHealPoints");
    int healsDone = currentHeals - g_iHealPoints[client];
    
    if (healsDone > 0) {
        LogEvent()
            .WithEventType("healed")
            .WithPlayer("medic", client)
            .WithInt("heal_points", healsDone)
            .WithString("reason", reason)
            .Send();
    }
    
    g_iHealPoints[client] = currentHeals;
}

//=============================================================================
// LOADOUT EVENTS
//=============================================================================

public void Event_PostInventoryApplication(Event event, const char[] name, bool dontBroadcast) {
    if (!g_bLogLoadouts) return;
    
    int client = GetClientOfUserId(event.GetInt("userid"));
    if (!IsValidClient(client)) return;
    
    // Delay to let items settle
    CreateTimer(0.2, Timer_LogLoadout, GetClientUserId(client));
}

public Action Timer_LogLoadout(Handle timer, int userid) {
    int client = GetClientOfUserId(userid);
    if (!IsValidClient(client)) return Plugin_Stop;
    
    // Get item definition indices for all slots
    int primary = GetPlayerWeaponSlotItemIndex(client, TFWeaponSlot_Primary);
    int secondary = GetPlayerWeaponSlotItemIndex(client, TFWeaponSlot_Secondary);
    int melee = GetPlayerWeaponSlotItemIndex(client, TFWeaponSlot_Melee);
    int pda = GetPlayerWeaponSlotItemIndex(client, TFWeaponSlot_PDA);
    int pda2 = GetPlayerWeaponSlotItemIndex(client, TFWeaponSlot_Building);
    
    LogEvent()
        .WithEventType("player_loadout")
        .WithPlayer("player", client)
        .WithInt("loadout.primary", primary)
        .WithInt("loadout.secondary", secondary)
        .WithInt("loadout.melee", melee)
        .WithInt("loadout.pda", pda)
        .WithInt("loadout.pda2", pda2)
        .Send();
    
    return Plugin_Stop;
}

int GetPlayerWeaponSlotItemIndex(int client, int slot) {
    int weapon = GetPlayerWeaponSlot(client, slot);
    if (weapon == -1) return -1;
    
    return GetEntProp(weapon, Prop_Send, "m_iItemDefinitionIndex");
}

//=============================================================================
// WEAPON STATS & AIRSHOTS
//=============================================================================

public Action OnTakeDamage(int victim, int &attacker, int &inflictor, float &damage, int &damagetype) {
    if (!IsValidClient(attacker) || attacker == victim) {
        return Plugin_Continue;
    }
    
    // Airshot detection
    if (g_bLogAirshots && IsClientAirborne(victim)) {
        DetectAirshot(attacker, victim, inflictor, damage);
    }
    
    // Weapon stats tracking
    if (g_bLogWeaponStats) {
        // Track hit (simplified - would need weapon index mapping)
        // This is a placeholder - full implementation would match HLstatsX weapon tracking
    }
    
    return Plugin_Continue;
}

void DetectAirshot(int attacker, int victim, int inflictor, float damage) {
    if (!IsValidEntity(inflictor)) return;
    
    char className[64];
    GetEdictClassname(inflictor, className, sizeof(className));
    
    char weaponType[32];
    bool isAirshot = false;
    
    if (StrContains(className, "tf_projectile_rocket") != -1) {
        strcopy(weaponType, sizeof(weaponType), "rocket");
        isAirshot = true;
    } else if (StrContains(className, "tf_projectile_pipe") != -1) {
        if (StrContains(className, "remote") != -1) {
            strcopy(weaponType, sizeof(weaponType), "sticky");
        } else {
            strcopy(weaponType, sizeof(weaponType), "pipebomb");
        }
        isAirshot = true;
    } else if (StrContains(className, "tf_projectile_arrow") != -1) {
        strcopy(weaponType, sizeof(weaponType), "arrow");
        isAirshot = true;
    } else if (StrContains(className, "tf_projectile_flare") != -1 && damage > 10.0) {
        strcopy(weaponType, sizeof(weaponType), "flare");
        isAirshot = true;
    }
    
    if (isAirshot) {
        bool air2air = (g_iJumpStatus[attacker] != JUMP_NONE);
        
        LogEvent()
            .WithEventType("airshot")
            .WithPlayer("player", attacker)
            .WithPlayer("victim", victim)
            .WithString("weapon_type", weaponType)
            .WithBool("air2air", air2air)
            .WithPlayerPosition("player_pos", attacker)
            .WithPlayerPosition("victim_pos", victim)
            .Send();
    }
}

void DumpWeaponStats(int client) {
    // Placeholder - full weapon stats tracking would require extensive weapon index mapping
    // Similar to HLstatsX's weapon trie system
}

//=============================================================================
// UTILITY FUNCTIONS
//=============================================================================

bool IsClientAirborne(int client) {
    if (!IsValidClient(client)) return false;
    
    int flags = GetEntityFlags(client);
    
    if (flags & FL_ONGROUND) return false;
    if (flags & FL_INWATER) return false;
    
    return true;
}

void GetObjectName(int objectType, char[] buffer, int maxlen) {
    switch (objectType) {
        case OBJ_DISPENSER: strcopy(buffer, maxlen, "dispenser");
        case OBJ_TELEPORTER_ENTRANCE: strcopy(buffer, maxlen, "teleporter_entrance");
        case OBJ_TELEPORTER_EXIT: strcopy(buffer, maxlen, "teleporter_exit");
        case OBJ_SENTRYGUN: strcopy(buffer, maxlen, "sentry");
        case OBJ_ATTACHMENT_SAPPER: strcopy(buffer, maxlen, "sapper");
        default: strcopy(buffer, maxlen, "unknown");
    }
}

void TF2_GetClassName(TFClassType class, char[] buffer, int maxlen) {
    switch (class) {
        case TFClass_Scout: strcopy(buffer, maxlen, "scout");
        case TFClass_Sniper: strcopy(buffer, maxlen, "sniper");
        case TFClass_Soldier: strcopy(buffer, maxlen, "soldier");
        case TFClass_DemoMan: strcopy(buffer, maxlen, "demoman");
        case TFClass_Medic: strcopy(buffer, maxlen, "medic");
        case TFClass_Heavy: strcopy(buffer, maxlen, "heavy");
        case TFClass_Pyro: strcopy(buffer, maxlen, "pyro");
        case TFClass_Spy: strcopy(buffer, maxlen, "spy");
        case TFClass_Engineer: strcopy(buffer, maxlen, "engineer");
        default: strcopy(buffer, maxlen, "unknown");
    }
}
