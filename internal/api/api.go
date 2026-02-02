package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/UDL-TF/UnitedStats/internal/store"
	"github.com/gin-gonic/gin"
)

// API provides the REST API server
type API struct {
	store  *store.Store
	router *gin.Engine
	server *http.Server
}

// Config holds API configuration
type Config struct {
	Store *store.Store
	Port  int
}

// New creates a new API server
func New(cfg Config) *API {
	router := gin.Default()

	api := &API{
		store:  cfg.Store,
		router: router,
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.Port),
			Handler:           router,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
		},
	}

	api.setupRoutes()
	return api
}

// setupRoutes configures all API routes
func (a *API) setupRoutes() {
	// Health check
	a.router.GET("/health", a.healthCheck)

	// API v1
	v1 := a.router.Group("/api/v1")

	// Players
	players := v1.Group("/players")
	players.GET("/:steam_id", a.getPlayer)
	players.GET("/:steam_id/stats", a.getPlayerStats)
	players.GET("/:steam_id/matches", a.getPlayerMatches)

	// Leaderboard
	v1.GET("/leaderboard", a.getLeaderboard)

	// Matches
	matches := v1.Group("/matches")
	matches.GET("", a.getMatches)
	matches.GET("/:id", a.getMatch)
	matches.GET("/:id/events", a.getMatchEvents)

	// Stats
	stats := v1.Group("/stats")
	stats.GET("/overview", a.getStatsOverview)
	stats.GET("/weapons", a.getWeaponStats)
}

// Start starts the API server
func (a *API) Start() error {
	return a.server.ListenAndServe()
}

// Shutdown gracefully shuts down the API server
func (a *API) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

// ============================================================================
// HANDLERS
// ============================================================================

// healthCheck returns the health status
func (a *API) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

// getLeaderboard returns the top players
func (a *API) getLeaderboard(c *gin.Context) {
	// Parse query parameters
	limit := 100
	offset := 0

	if limitStr := c.DefaultQuery("limit", "100"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	if offsetStr := c.DefaultQuery("offset", "0"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil {
			offset = val
		}
	}

	// Limit max results
	if limit > 500 {
		limit = 500
	}

	entries, err := a.store.GetLeaderboard(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": entries,
		"limit":       limit,
		"offset":      offset,
		"count":       len(entries),
	})
}

// getPlayer returns player information
func (a *API) getPlayer(c *gin.Context) {
	steamID := c.Param("steam_id")

	player, err := a.store.GetPlayerBySteamID(c.Request.Context(), steamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"player": player,
	})
}

// getPlayerStats returns detailed player statistics
func (a *API) getPlayerStats(c *gin.Context) {
	steamID := c.Param("steam_id")

	player, err := a.store.GetPlayerBySteamID(c.Request.Context(), steamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}

	// Calculate additional stats
	kd := float64(0)
	if player.TotalDeaths > 0 {
		kd = float64(player.TotalKills) / float64(player.TotalDeaths)
	}

	c.JSON(http.StatusOK, gin.H{
		"steam_id": player.SteamID,
		"name":     player.Name,
		"mmr": gin.H{
			"current": player.MMR,
			"peak":    player.PeakMMR,
		},
		"stats": gin.H{
			"kills":     player.TotalKills,
			"deaths":    player.TotalDeaths,
			"assists":   player.TotalAssists,
			"kd_ratio":  kd,
			"airshots":  player.TotalAirshots,
			"headshots": player.TotalHeadshots,
			"backstabs": player.TotalBackstabs,
			"deflects":  player.TotalDeflects,
		},
		"last_seen": player.LastSeen,
	})
}

// getPlayerMatches returns a player's match history
func (a *API) getPlayerMatches(c *gin.Context) {
	steamID := c.Param("steam_id")
	limit := 20
	offset := 0

	if limitStr := c.DefaultQuery("limit", "20"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	if offsetStr := c.DefaultQuery("offset", "0"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil {
			offset = val
		}
	}

	if limit > 100 {
		limit = 100
	}

	player, err := a.store.GetPlayerBySteamID(c.Request.Context(), steamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}

	matches, err := a.store.GetPlayerMatchHistory(c.Request.Context(), player.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch match history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"player_id": player.ID,
		"steam_id":  player.SteamID,
		"matches":   matches,
		"count":     len(matches),
		"limit":     limit,
		"offset":    offset,
	})
}

// getMatches returns recent matches
func (a *API) getMatches(c *gin.Context) {
	limit := 50
	offset := 0

	if limitStr := c.DefaultQuery("limit", "50"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	if offsetStr := c.DefaultQuery("offset", "0"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil {
			offset = val
		}
	}

	if limit > 200 {
		limit = 200
	}

	matches, err := a.store.GetRecentMatches(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch matches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"matches": matches,
		"count":   len(matches),
		"limit":   limit,
		"offset":  offset,
	})
}

// getMatch returns a specific match
func (a *API) getMatch(c *gin.Context) {
	matchIDStr := c.Param("id")
	matchID, err := strconv.ParseInt(matchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID"})
		return
	}

	match, err := a.store.GetMatchByID(c.Request.Context(), matchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"match": match,
	})
}

// getMatchEvents returns events for a match
func (a *API) getMatchEvents(c *gin.Context) {
	matchID := c.Param("id")

	// TODO: Implement match events query
	c.JSON(http.StatusOK, gin.H{
		"match_id": matchID,
		"events":   []interface{}{},
		"count":    0,
	})
}

// getStatsOverview returns overall statistics
func (a *API) getStatsOverview(c *gin.Context) {
	stats, err := a.store.GetStatsOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// getWeaponStats returns weapon statistics
func (a *API) getWeaponStats(c *gin.Context) {
	limit := 50

	if limitStr := c.DefaultQuery("limit", "50"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	if limit > 200 {
		limit = 200
	}

	stats, err := a.store.GetWeaponStats(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weapon stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"weapons": stats,
		"count":   len(stats),
	})
}
