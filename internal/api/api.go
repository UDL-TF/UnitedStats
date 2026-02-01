package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/UDL-TF/UnitedStats/internal/store"
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
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: router,
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
	{
		// Players
		players := v1.Group("/players")
		{
			players.GET("/:steam_id", a.getPlayer)
			players.GET("/:steam_id/stats", a.getPlayerStats)
			players.GET("/:steam_id/matches", a.getPlayerMatches)
		}

		// Leaderboard
		v1.GET("/leaderboard", a.getLeaderboard)

		// Matches
		matches := v1.Group("/matches")
		{
			matches.GET("", a.getMatches)
			matches.GET("/:id", a.getMatch)
			matches.GET("/:id/events", a.getMatchEvents)
		}

		// Stats
		stats := v1.Group("/stats")
		{
			stats.GET("/overview", a.getStatsOverview)
			stats.GET("/weapons", a.getWeaponStats)
		}
	}
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
		"status": "ok",
		"timestamp": time.Now().Unix(),
	})
}

// getLeaderboard returns the top players
func (a *API) getLeaderboard(c *gin.Context) {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

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

	player, err := a.store.GetPlayerBySteamID(c.Request.Context(), steamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}

	// TODO: Implement match history query
	c.JSON(http.StatusOK, gin.H{
		"player_id": player.ID,
		"matches":   []interface{}{},
		"count":     0,
	})
}

// getMatches returns recent matches
func (a *API) getMatches(c *gin.Context) {
	// TODO: Implement matches query
	c.JSON(http.StatusOK, gin.H{
		"matches": []interface{}{},
		"count":   0,
	})
}

// getMatch returns a specific match
func (a *API) getMatch(c *gin.Context) {
	matchID := c.Param("id")

	// TODO: Implement match detail query
	c.JSON(http.StatusOK, gin.H{
		"match_id": matchID,
		"match":    nil,
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
	// TODO: Implement stats overview
	c.JSON(http.StatusOK, gin.H{
		"total_players": 0,
		"total_matches": 0,
		"total_kills":   0,
	})
}

// getWeaponStats returns weapon statistics
func (a *API) getWeaponStats(c *gin.Context) {
	// TODO: Implement weapon stats query
	c.JSON(http.StatusOK, gin.H{
		"weapons": []interface{}{},
	})
}
