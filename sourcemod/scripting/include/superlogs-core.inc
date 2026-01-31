/**
 * SuperLogs Core - JSON Event Sender via UDP
 * 
 * Based on TFDB logging pattern but sends JSON events via UDP instead of writing to files.
 * 
 * Usage:
 *   #include <superlogs-core>
 *   
 *   public void OnPluginStart() {
 *       SuperLogs_Initialize("default", "stats.udl.tf", 27500);
 *   }
 *   
 *   // Send a kill event
 *   LogEvent()
 *       .WithEventType("kill")
 *       .WithPlayer("killer", killerClient)
 *       .WithPlayer("victim", victimClient)
 *       .WithString("weapon", weapon)
 *       .WithBool("crit", crit)
 *       .WithBool("airborne", airborne)
 *       .Send();
 */

#if defined _superlogs_core_included
  #endinput
#endif
#define _superlogs_core_included

#include <sourcemod>
#include <json>

// Configuration
#define SUPERLOGS_VERSION "2.0.0"

// Global state
char g_sGamemode[32];
char g_sCollectorHost[128];
int g_iCollectorPort;
Handle g_hSocket;

/**
 * Initialize SuperLogs system
 * 
 * @param gamemode     Gamemode identifier (e.g., "default", "dodgeball")
 * @param host         Collector hostname/IP
 * @param port         Collector UDP port
 */
stock void SuperLogs_Initialize(const char[] gamemode, const char[] host, int port) {
    strcopy(g_sGamemode, sizeof(g_sGamemode), gamemode);
    strcopy(g_sCollectorHost, sizeof(g_sCollectorHost), host);
    g_iCollectorPort = port;
    
    LogMessage("[SuperLogs] Initialized: gamemode=%s, collector=%s:%d", gamemode, host, port);
}

/**
 * Get current server IP address
 */
stock void SuperLogs_GetServerIP(char[] buffer, int maxlen) {
    int pieces[4];
    int ip = GetConVarInt(FindConVar("hostip"));
    
    pieces[0] = (ip >> 24) & 0xFF;
    pieces[1] = (ip >> 16) & 0xFF;
    pieces[2] = (ip >> 8) & 0xFF;
    pieces[3] = ip & 0xFF;
    
    Format(buffer, maxlen, "%d.%d.%d.%d", pieces[0], pieces[1], pieces[2], pieces[3]);
}

/**
 * Send JSON event via UDP
 */
stock void SuperLogs_SendJSON(JSON_Object event) {
    // Create socket if needed
    if (g_hSocket == INVALID_HANDLE) {
        g_hSocket = SocketCreate(SOCKET_UDP, OnSocketError);
    }
    
    // Encode to JSON string
    char buffer[4096];
    event.Encode(buffer, sizeof(buffer));
    
    // Send UDP packet
    SocketSetOption(g_hSocket, SocketReuseAddr, 1);
    SocketSendTo(g_hSocket, buffer, strlen(buffer), g_sCollectorHost, g_iCollectorPort);
    
    #if defined DEBUG
    LogMessage("[SuperLogs] Sent: %s", buffer);
    #endif
}

/**
 * Socket error callback
 */
public int OnSocketError(Handle socket, const int errorType, const int errorNum, any hFile) {
    LogError("[SuperLogs] Socket error %d (type %d)", errorNum, errorType);
    CloseHandle(socket);
    g_hSocket = INVALID_HANDLE;
    return 0;
}

/**
 * LogEvent methodmap - builds structured JSON events
 */
methodmap LogEvent < JSON_Object {
    /**
     * Creates a new LogEvent
     */
    public LogEvent() {
        LogEvent event = view_as<LogEvent>(new JSON_Object());
        
        // Add timestamp
        char timeStr[64];
        FormatTime(timeStr, sizeof(timeStr), "%Y-%m-%dT%H:%M:%S", GetTime());
        event.SetString("timestamp", timeStr);
        
        // Add gamemode
        event.SetString("gamemode", g_sGamemode);
        
        // Add server IP
        char serverIP[32];
        SuperLogs_GetServerIP(serverIP, sizeof(serverIP));
        event.SetString("server_ip", serverIP);
        
        return event;
    }
    
    /**
     * Sets the event type (e.g., "kill", "deflect", "match_start")
     */
    public LogEvent WithEventType(const char[] eventType) {
        this.SetString("event_type", eventType);
        return this;
    }
    
    /**
     * Adds a string value
     */
    public LogEvent WithString(const char[] key, const char[] value) {
        this.SetString(key, value);
        return this;
    }
    
    /**
     * Adds an integer value
     */
    public LogEvent WithInt(const char[] key, int value) {
        this.SetInt(key, value);
        return this;
    }
    
    /**
     * Adds a float value
     */
    public LogEvent WithFloat(const char[] key, float value) {
        this.SetFloat(key, value);
        return this;
    }
    
    /**
     * Adds a boolean value
     */
    public LogEvent WithBool(const char[] key, bool value) {
        this.SetBool(key, value);
        return this;
    }
    
    /**
     * Adds player information (SteamID + name) as a nested object
     * 
     * @param role         Player role ("killer", "victim", "player")
     * @param client       Client index
     */
    public LogEvent WithPlayer(const char[] role, int client) {
        if (!IsValidClient(client)) {
            return this;
        }
        
        JSON_Object playerObj = new JSON_Object();
        
        // Get SteamID
        char steamID[32];
        GetClientAuthId(client, AuthId_SteamID64, steamID, sizeof(steamID));
        playerObj.SetString("steam_id", steamID);
        
        // Get player name
        char playerName[64];
        GetClientName(client, playerName, sizeof(playerName));
        playerObj.SetString("name", playerName);
        
        // Get team
        int team = GetClientTeam(client);
        playerObj.SetInt("team", team);
        
        // Attach to event
        this.SetObject(role, playerObj);
        
        return this;
    }
    
    /**
     * Adds player position (x, y, z) as a nested object
     */
    public LogEvent WithPlayerPosition(const char[] key, int client) {
        if (!IsValidClient(client)) {
            return this;
        }
        
        float pos[3];
        GetClientAbsOrigin(client, pos);
        
        JSON_Object posObj = new JSON_Object();
        posObj.SetFloat("x", pos[0]);
        posObj.SetFloat("y", pos[1]);
        posObj.SetFloat("z", pos[2]);
        
        this.SetObject(key, posObj);
        
        return this;
    }
    
    /**
     * Adds weapon information as a nested object
     */
    public LogEvent WithWeapon(const char[] weapon, int itemDefIndex = -1) {
        JSON_Object weaponObj = new JSON_Object();
        weaponObj.SetString("name", weapon);
        
        if (itemDefIndex != -1) {
            weaponObj.SetInt("item_def_index", itemDefIndex);
        }
        
        this.SetObject("weapon", weaponObj);
        
        return this;
    }
    
    /**
     * Sends the event via UDP and cleans up
     */
    public void Send() {
        SuperLogs_SendJSON(this);
        
        // Cleanup
        json_cleanup_and_delete(this);
    }
}

/**
 * Helper: Check if client is valid player
 */
stock bool IsValidClient(int client) {
    return (client > 0 && 
            client <= MaxClients && 
            IsClientInGame(client) && 
            !IsFakeClient(client));
}

/**
 * Starts a new event
 * Usage: LogEvent().WithEventType("kill").WithPlayer("killer", client).Send();
 */
stock LogEvent LogEvent() {
    return new LogEvent();
}
