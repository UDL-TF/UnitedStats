/**
 * SuperLogs Core - UDP Event Sender Library
 * 
 * Provides shared functionality for sending game events via UDP to the collector.
 * 
 * Usage:
 *   #include <superlogs-core>
 *   
 *   public void OnPluginStart() {
 *       SuperLogs_Initialize("default", "stats.udl.tf", 27500);
 *   }
 *   
 *   SuperLogs_SendKill(killerClient, victimClient, weapon, crit, airborne);
 */

#if defined _superlogs_core_included
  #endinput
#endif
#define _superlogs_core_included

// Configuration
#define SUPERLOGS_MAX_LINE_LENGTH 512
#define SUPERLOGS_VERSION "1.0.0"

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
 * Send raw log line via UDP
 * 
 * @param line         Formatted log line to send
 */
stock void SuperLogs_SendRaw(const char[] line) {
    // Create socket if needed
    if (g_hSocket == INVALID_HANDLE) {
        g_hSocket = SocketCreate(SOCKET_UDP, OnSocketError);
    }
    
    // Send UDP packet
    SocketSetOption(g_hSocket, SocketReuseAddr, 1);
    SocketSendTo(g_hSocket, line, strlen(line), g_sCollectorHost, g_iCollectorPort);
    
    #if defined DEBUG
    LogMessage("[SuperLogs] Sent: %s", line);
    #endif
}

/**
 * Escape special characters in strings (pipes, newlines, etc.)
 * 
 * @param input        Input string
 * @param output       Output buffer
 * @param maxlen       Max length of output buffer
 */
stock void SuperLogs_EscapeString(const char[] input, char[] output, int maxlen) {
    int outPos = 0;
    int len = strlen(input);
    
    for (int i = 0; i < len && outPos < maxlen - 2; i++) {
        char c = input[i];
        
        // Escape special characters
        if (c == '|') {
            output[outPos++] = '\\';
            output[outPos++] = 'p';
        } else if (c == '\n') {
            output[outPos++] = '\\';
            output[outPos++] = 'n';
        } else if (c == '\r') {
            output[outPos++] = '\\';
            output[outPos++] = 'r';
        } else if (c == '\\') {
            output[outPos++] = '\\';
            output[outPos++] = '\\';
        } else {
            output[outPos++] = c;
        }
    }
    
    output[outPos] = '\0';
}

/**
 * Get current server IP address
 * 
 * @param buffer       Output buffer for IP address
 * @param maxlen       Max length of buffer
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
 * Get current Unix timestamp
 * 
 * @return Unix timestamp (seconds since epoch)
 */
stock int SuperLogs_GetTimestamp() {
    return GetTime();
}

/**
 * Send KILL event
 * 
 * Format: KILL|timestamp|gamemode|server_ip|killer_steamid|killer_name|victim_steamid|victim_name|weapon|crit|airborne
 * 
 * @param killerClient   Client index of killer
 * @param victimClient   Client index of victim
 * @param weapon         Weapon name
 * @param crit           Was critical hit (0/1)
 * @param airborne       Was victim airborne (0/1)
 */
stock void SuperLogs_SendKill(int killerClient, int victimClient, const char[] weapon, int crit, int airborne) {
    char line[SUPERLOGS_MAX_LINE_LENGTH];
    char serverIP[32];
    char killerSteam[32], victimSteam[32];
    char killerName[64], victimName[64];
    char killerNameEsc[128], victimNameEsc[128];
    char weaponEsc[64];
    
    // Get server IP
    SuperLogs_GetServerIP(serverIP, sizeof(serverIP));
    
    // Get SteamIDs
    GetClientAuthId(killerClient, AuthId_SteamID64, killerSteam, sizeof(killerSteam));
    GetClientAuthId(victimClient, AuthId_SteamID64, victimSteam, sizeof(victimSteam));
    
    // Get player names and escape
    GetClientName(killerClient, killerName, sizeof(killerName));
    GetClientName(victimClient, victimName, sizeof(victimName));
    SuperLogs_EscapeString(killerName, killerNameEsc, sizeof(killerNameEsc));
    SuperLogs_EscapeString(victimName, victimNameEsc, sizeof(victimNameEsc));
    SuperLogs_EscapeString(weapon, weaponEsc, sizeof(weaponEsc));
    
    // Format log line
    Format(line, sizeof(line), "KILL|%d|%s|%s|%s|%s|%s|%s|%s|%d|%d",
        SuperLogs_GetTimestamp(),
        g_sGamemode,
        serverIP,
        killerSteam,
        killerNameEsc,
        victimSteam,
        victimNameEsc,
        weaponEsc,
        crit,
        airborne
    );
    
    SuperLogs_SendRaw(line);
}

/**
 * Send DEFLECT event (for dodgeball)
 * 
 * Format: DEFLECT|timestamp|gamemode|server_ip|player_steamid|player_name|rocket_speed|deflect_angle|timing_ms|distance
 * 
 * @param client         Client index
 * @param rocketSpeed    Rocket speed at deflect (units/sec)
 * @param angle          Deflect angle accuracy (0.0-1.0)
 * @param timing         Timing in milliseconds
 * @param distance       Distance from rocket
 */
stock void SuperLogs_SendDeflect(int client, float rocketSpeed, float angle, int timing, float distance) {
    char line[SUPERLOGS_MAX_LINE_LENGTH];
    char serverIP[32];
    char steamID[32];
    char playerName[64];
    char playerNameEsc[128];
    
    SuperLogs_GetServerIP(serverIP, sizeof(serverIP));
    GetClientAuthId(client, AuthId_SteamID64, steamID, sizeof(steamID));
    GetClientName(client, playerName, sizeof(playerName));
    SuperLogs_EscapeString(playerName, playerNameEsc, sizeof(playerNameEsc));
    
    Format(line, sizeof(line), "DEFLECT|%d|%s|%s|%s|%s|%.2f|%.4f|%d|%.2f",
        SuperLogs_GetTimestamp(),
        g_sGamemode,
        serverIP,
        steamID,
        playerNameEsc,
        rocketSpeed,
        angle,
        timing,
        distance
    );
    
    SuperLogs_SendRaw(line);
}

/**
 * Send MATCH_START event
 * 
 * Format: MATCH_START|timestamp|gamemode|server_ip|map_name
 * 
 * @param mapName        Current map name
 */
stock void SuperLogs_SendMatchStart(const char[] mapName) {
    char line[SUPERLOGS_MAX_LINE_LENGTH];
    char serverIP[32];
    char mapEsc[128];
    
    SuperLogs_GetServerIP(serverIP, sizeof(serverIP));
    SuperLogs_EscapeString(mapName, mapEsc, sizeof(mapEsc));
    
    Format(line, sizeof(line), "MATCH_START|%d|%s|%s|%s",
        SuperLogs_GetTimestamp(),
        g_sGamemode,
        serverIP,
        mapEsc
    );
    
    SuperLogs_SendRaw(line);
}

/**
 * Send MATCH_END event
 * 
 * Format: MATCH_END|timestamp|gamemode|server_ip|winner_team|duration
 * 
 * @param winnerTeam     Winning team (2=RED, 3=BLU, 0=tie)
 * @param duration       Match duration in seconds
 */
stock void SuperLogs_SendMatchEnd(int winnerTeam, int duration) {
    char line[SUPERLOGS_MAX_LINE_LENGTH];
    char serverIP[32];
    
    SuperLogs_GetServerIP(serverIP, sizeof(serverIP));
    
    Format(line, sizeof(line), "MATCH_END|%d|%s|%s|%d|%d",
        SuperLogs_GetTimestamp(),
        g_sGamemode,
        serverIP,
        winnerTeam,
        duration
    );
    
    SuperLogs_SendRaw(line);
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
