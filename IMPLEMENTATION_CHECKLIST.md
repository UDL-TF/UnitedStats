# Implementation Checklist - UnitedStats
**For Sub-Agents: Use this checklist to track implementation progress**

---

## 🎯 Phase 1: Core Backend (Priority: HIGH)

### 1.1 Project Setup
- [ ] Initialize Go module: `go mod init github.com/UDL-TF/UnitedStats`
- [ ] Create folder structure (cmd/, internal/, pkg/, migrations/)
- [ ] Set up `.gitignore` for Go projects
- [ ] Create `go.mod` with dependencies:
  - `github.com/go-redis/redis/v8`
  - `gorm.io/gorm`
  - `gorm.io/driver/postgres`
  - `github.com/gin-gonic/gin`
- [ ] Set up Docker Compose (`docker-compose.yml`)

**Deliverable**: Project skeleton ready for development

---

### 1.2 Database Schema
- [ ] Create `migrations/001_initial_schema.sql`:
  - [ ] `players` table
  - [ ] `player_statistics` table
  - [ ] `weapon_statistics` table
  - [ ] `kill_events` table (temporary)
  - [ ] `match_sessions` table
  - [ ] `servers` table
  - [ ] `performance_snapshots` table
- [ ] Add indexes for performance (steam_id, mmr, timestamp)
- [ ] Test migration on local PostgreSQL

**Deliverable**: Database schema applied and tested

**PR Title**: `[Phase 1] feat(db): Add initial database schema with MMR tables`

---

### 1.3 Event Structs & Models
**File**: `pkg/events/events.go`

- [ ] Define `Event` struct:
  ```go
  type Event struct {
      Type       string    // "kill", "deflect", "weaponstats"
      Timestamp  time.Time
      ServerIP   string
      Gamemode   string
      Player     *Player
      Killer     *Player
      Victim     *Player
      Weapon     string
      PerformanceData *PerformanceMetrics
  }
  ```
- [ ] Define `Player` struct (Name, SteamID, Team, MMR)
- [ ] Define `PerformanceMetrics` struct (timing, angle, speed, distance)
- [ ] Add JSON tags for serialization

**File**: `internal/models/player.go`

- [ ] Define GORM models matching database schema
- [ ] Implement `Player`, `PlayerStatistics`, `WeaponStatistics` models
- [ ] Add database helper functions (GetPlayerByID, UpdateMMR, etc.)

**Deliverable**: Shared event types and database models

**PR Title**: `[Phase 1] feat(models): Add event structs and GORM database models`

---

### 1.4 MMR Calculator
**File**: `internal/mmr/calculator.go`

- [ ] Implement `CalculateKillWeight(killerMMR, victimMMR int) float64`
  - Formula: `1 + 0.5 * LOG2(victimMMR / killerMMR)`
  - Clamp to [0.5, 1.5]
- [ ] Implement `CalculateRankWeight(currentMMR int) float64`
  - Formula: `1.0 - (currentMMR/5000)*0.9`
  - Clamp to [0.1, 1.0]
- [ ] Implement `CalculateRankScore(accuracyScore, kdWeighted float64, currentMMR int) int`
  - Formula: `accuracyScore * (1 + kdWeighted*rankWeight) * 1000`
  - Clamp to [0, 5000]
- [ ] Implement `GetRankTier(mmr int) string`
  - Return "Bronze", "Silver", "Gold", "Platinum", "Diamond", "Master"

**File**: `internal/mmr/calculator_test.go`

- [ ] Test `CalculateKillWeight`:
  - Equal MMR (2000, 2000) → 1.0
  - 2x stronger (2000, 4000) → 1.5
  - 2x weaker (2000, 1000) → 0.5
  - Extreme values (capping behavior)
- [ ] Test `CalculateRankWeight`:
  - Bronze (500) → 0.91
  - Master (4000) → 0.28
- [ ] Test `CalculateRankScore` with realistic player data
- [ ] Test `GetRankTier` boundary values

**Deliverable**: MMR calculation library with 100% test coverage

**PR Title**: `[Phase 1] feat(mmr): Implement kill weighting and rank score calculations`

---

### 1.5 Event Parser
**File**: `internal/parser/parser.go`

- [ ] Implement `ParseLogLine(line string, serverIP string) (*events.Event, error)`
- [ ] Create regex patterns for:
  - [ ] Kill events (with killer_mmr, victim_mmr)
  - [ ] Deflect events (timing, angle, speed, distance)
  - [ ] Weapon stats events (shots, hits, damage)
- [ ] Handle malformed logs gracefully (return error, don't crash)
- [ ] Extract player info (name, SteamID, team)
- [ ] Extract event metadata (weapon, timestamp)

**File**: `internal/parser/parser_test.go`

- [ ] Test parsing kill event:
  ```
  "Scout<45><STEAM_1:0:123456><Red>" killed "Heavy<67><STEAM_1:1:654321><Blue>" (weapon "scattergun") (killer_mmr "2380") (victim_mmr "3700")
  ```
- [ ] Test parsing deflect event:
  ```
  "Pyro<45><STEAM_1:0:123456><Red>" triggered "deflect" (timing_accuracy "1.0") (angle_accuracy "0.95") ...
  ```
- [ ] Test malformed logs (should return error)
- [ ] Test edge cases (missing fields, invalid SteamID)

**Deliverable**: Robust log parser with error handling

**PR Title**: `[Phase 1] feat(parser): Implement log line parser for kill/deflect events`

---

### 1.6 Redis Queue
**File**: `internal/queue/redis.go`

- [ ] Implement `RedisQueue` struct with connection
- [ ] Implement `Push(queueName string, event *events.Event) error`
  - Serialize event to JSON
  - Push to Redis list
- [ ] Implement `Pop(queueName string) (*events.Event, error)`
  - Blocking pop (BRPOP)
  - Deserialize from JSON
- [ ] Implement `GetQueueDepth(queueName string) int`
- [ ] Handle connection errors gracefully (retry logic)

**File**: `internal/queue/redis_test.go`

- [ ] Test push/pop with mock Redis
- [ ] Test serialization/deserialization
- [ ] Test queue depth tracking
- [ ] Test connection failure handling

**Deliverable**: Redis queue abstraction layer

**PR Title**: `[Phase 1] feat(queue): Add Redis queue implementation with JSON serialization`

---

### 1.7 UDP Collector Service
**File**: `cmd/collector/main.go`

- [ ] Implement UDP listener on port 27500
- [ ] Read packets into buffer (4096 bytes)
- [ ] Spawn goroutine for each packet (concurrency)
- [ ] Parse log line using `parser.ParseLogLine()`
- [ ] Push parsed event to Redis queue
- [ ] Add error logging for parse failures
- [ ] Add metrics (packets received, parse errors)
- [ ] Graceful shutdown on SIGTERM

**File**: `cmd/collector/main_test.go`

- [ ] Test UDP packet reception
- [ ] Test concurrent packet handling
- [ ] Test error handling for malformed packets

**Config**: `configs/collector.yaml`

- [ ] UDP port (default: 27500)
- [ ] Redis connection string
- [ ] Log level (info/debug)

**Deliverable**: UDP collector service ready to receive logs

**PR Title**: `[Phase 1] feat(collector): Implement UDP listener service with goroutines`

---

### 1.8 Event Processor Service
**File**: `cmd/processor/main.go`

- [ ] Spawn worker goroutines (10 workers)
- [ ] Each worker:
  - [ ] Pop event from Redis queue
  - [ ] Process based on event type (kill, deflect, weaponstats)
  - [ ] Update in-memory stats (Redis cache)
  - [ ] Mark player as dirty
- [ ] Implement batch writer (goroutine):
  - [ ] Every 30 seconds
  - [ ] Fetch all dirty player stats
  - [ ] Bulk upsert to PostgreSQL
  - [ ] Clear dirty flags
- [ ] Add metrics (events processed/sec, queue depth, batch size)
- [ ] Graceful shutdown (finish processing queue)

**File**: `cmd/processor/kill_handler.go`

- [ ] Implement `processKillEvent(event *Event, db *DB, cache *Redis)`
  - Calculate kill weight
  - Update killer: `weighted_kills += weight`, `total_kills += 1`
  - Update victim: `total_deaths += 1`
  - Recalculate killer MMR
  - Update victim last_seen
  - Mark both players dirty

**File**: `cmd/processor/deflect_handler.go`

- [ ] Implement `processDeflectEvent(event *Event, db *DB, cache *Redis)`
  - Update deflect metrics (running average)
  - Update accuracy score
  - Recalculate MMR
  - Mark player dirty

**File**: `cmd/processor/batch_writer.go`

- [ ] Implement `flushDirtyStats(db *DB, cache *Redis)`
  - Fetch dirty player IDs from Redis set
  - Load stats from Redis
  - Bulk upsert to PostgreSQL (use transaction)
  - Clear dirty set on success

**Deliverable**: Event processor with kill weighting and batch writes

**PR Title**: `[Phase 1] feat(processor): Add event processor workers with MMR calculation`

---

### 1.9 Integration Testing
**File**: `test/integration_test.go`

- [ ] Test end-to-end flow:
  1. Send fake UDP packet to collector
  2. Verify event in Redis queue
  3. Verify stats updated in database
  4. Verify MMR calculated correctly
- [ ] Test batch writer:
  - Send 100 events
  - Wait 30 seconds
  - Verify database has all updates
- [ ] Test concurrency (send 1000 events rapidly)

**Deliverable**: Verified end-to-end data flow

**PR Title**: `[Phase 1] test: Add integration tests for collector → processor → database`

---

## 🎮 Phase 2: SourceMod Plugins (Priority: HIGH)

### 2.1 Core Library
**File**: `sourcemod/scripting/include/superlogs-core.inc`

- [ ] Implement UDP socket connection
- [ ] Implement `SuperLogs_Init()` (create socket)
- [ ] Implement `SuperLogs_SendUDP(const char[] message)` (send to stats server)
- [ ] Implement `SuperLogs_FormatEvent()` (format log lines)
- [ ] Implement `SuperLogs_GetPlayerInfo()` (SteamID, name, team)
- [ ] Add config cvars:
  - `sm_superlogs_server` (default: "stats.udl.tf")
  - `sm_superlogs_port` (default: "27500")
  - `sm_superlogs_enabled` (default: "1")

**Deliverable**: Shared UDP sender library

**PR Title**: `[Phase 2] feat(sourcemod): Add core UDP sender library for SuperLogs`

---

### 2.2 Dodgeball Plugin
**File**: `sourcemod/scripting/superlogs-dodgeball.sp`

- [ ] Hook `OnEntityCreated` (track rockets)
- [ ] Hook `SDKHook_Touch` on rockets (detect deflects)
- [ ] Implement `CalculateTimingAccuracy(rocket, client)`:
  - Predict impact time
  - Compare to optimal deflect window (0.1s)
  - Return 0.0-1.0 score
- [ ] Implement `CalculateAngleAccuracy(rocket, client)`:
  - Get client aim angles
  - Get rocket direction
  - Calculate angle difference
  - Return 0.0-1.0 score
- [ ] Implement `GetRocketSpeed(rocket)`:
  - Detect weapon type (Direct Hit = 1.5, Stock = 1.0)
- [ ] Implement `CalculateDistance(rocket, client)`:
  - Return distance in Hammer Units
  - Convert to distance factor (0.5-1.0)
- [ ] Calculate `DeflectScore` using formula
- [ ] Send UDP event with all metrics
- [ ] Track kill events (call core library)

**Deliverable**: Dodgeball plugin with deflect accuracy tracking

**PR Title**: `[Phase 2] feat(sourcemod): Add Dodgeball SuperLogs plugin with deflect metrics`

---

### 2.3 Default TF2 Plugin
**File**: `sourcemod/scripting/superlogs-default.sp`

- [ ] Hook `player_death` event
- [ ] Fetch killer MMR from API (HTTP request to `/api/v1/players/:steamid/mmr`)
- [ ] Fetch victim MMR from API
- [ ] Format kill event with MMR values
- [ ] Send UDP event
- [ ] Track weapon stats:
  - [ ] Hook `TF2_CalcIsAttackCritical` (shots fired)
  - [ ] Hook `player_hurt` (shots hit, damage)
  - [ ] Send weaponstats every round end
- [ ] Track additional events:
  - [ ] Headshots (customkill == TF_CUSTOM_HEADSHOT)
  - [ ] Backstabs (customkill == TF_CUSTOM_BACKSTAB)
  - [ ] Airshots (detect mid-air kills)

**Deliverable**: Default TF2 plugin with enhanced event tracking

**PR Title**: `[Phase 2] feat(sourcemod): Add default TF2 SuperLogs plugin with MMR-aware kills`

---

### 2.4 Plugin Testing
- [ ] Set up local TF2 server (Docker or local install)
- [ ] Install SourceMod + plugins
- [ ] Configure server to point to localhost:27500
- [ ] Play test matches (kill bots, deflect rockets)
- [ ] Verify UDP packets received by collector
- [ ] Verify events parsed correctly
- [ ] Verify MMR updates in database

**Deliverable**: Plugins tested on live TF2 server

**PR Title**: `[Phase 2] test: Add TF2 server testing documentation and scripts`

---

## 🌐 Phase 3: REST API (Priority: MEDIUM)

### 3.1 API Server Setup
**File**: `cmd/api/main.go`

- [ ] Initialize Gin router
- [ ] Set up CORS middleware
- [ ] Set up logging middleware
- [ ] Connect to PostgreSQL
- [ ] Connect to Redis (for caching)
- [ ] Start server on port 8080
- [ ] Graceful shutdown

**Config**: `configs/api.yaml`

- [ ] Server port
- [ ] Database connection
- [ ] Redis connection
- [ ] CORS allowed origins

**Deliverable**: API server skeleton

**PR Title**: `[Phase 3] feat(api): Initialize Gin API server with database connections`

---

### 3.2 Player Endpoints
**File**: `cmd/api/handlers/player.go`

- [ ] `GET /api/v1/players/:steamid`
  - Fetch player from database
  - Return JSON with MMR, rank, stats
  - Cache response in Redis (5 min TTL)
- [ ] `GET /api/v1/players/:steamid/mmr`
  - Return only MMR value (for SourceMod plugin)
  - Fast response (<10ms)
- [ ] `GET /api/v1/players/:steamid/performance`
  - Fetch performance_snapshots (last 30 days)
  - Return time-series data for graphs
- [ ] `GET /api/v1/players/:steamid/matches`
  - Fetch recent match_sessions
  - Paginated response (limit=20)

**Deliverable**: Player query endpoints

**PR Title**: `[Phase 3] feat(api): Add player profile and MMR query endpoints`

---

### 3.3 Leaderboard Endpoints
**File**: `cmd/api/handlers/leaderboard.go`

- [ ] `GET /api/v1/leaderboard?gamemode=X&rank=Y&limit=50`
  - Filter by gamemode (optional)
  - Filter by rank tier (optional)
  - Sort by MMR descending
  - Return top N players
  - Cache for 1 minute
- [ ] `GET /api/v1/leaderboard/global`
  - Top 100 players across all gamemodes
  - Include rank distribution stats

**Deliverable**: Leaderboard endpoints

**PR Title**: `[Phase 3] feat(api): Add leaderboard endpoints with filtering`

---

### 3.4 Server Endpoints
**File**: `cmd/api/handlers/server.go`

- [ ] `GET /api/v1/servers`
  - List all active servers
  - Include last heartbeat timestamp
- [ ] `GET /api/v1/servers/:id/live`
  - Fetch live match data from Redis
  - Return current players, scores, map

**Deliverable**: Server status endpoints

**PR Title**: `[Phase 3] feat(api): Add server listing and live match endpoints`

---

### 3.5 API Documentation
**File**: `docs/API.md`

- [ ] Document all endpoints with:
  - Method, path, parameters
  - Request/response examples
  - Error codes
- [ ] Add OpenAPI/Swagger spec (optional but nice)
- [ ] Add rate limiting documentation

**Deliverable**: Complete API reference

**PR Title**: `[Phase 3] docs: Add comprehensive API documentation`

---

## 🖥️ Phase 4: Web Interface (Priority: LOW)

### 4.1 Frontend Setup
- [ ] Choose framework (Svelte recommended)
- [ ] Initialize project (`npm create svelte@latest web`)
- [ ] Set up Tailwind CSS
- [ ] Configure API endpoint (env var)
- [ ] Set up routing (SvelteKit routing)

**Deliverable**: Frontend project skeleton

**PR Title**: `[Phase 4] feat(web): Initialize Svelte frontend with Tailwind CSS`

---

### 4.2 Player Profile Page
**Route**: `/player/:steamid`

- [ ] Fetch player data from API
- [ ] Display:
  - [ ] SteamID, name, rank badge
  - [ ] Current MMR (large number)
  - [ ] Rank tier (Bronze/Silver/Gold/etc.)
  - [ ] K/D ratio (weighted and raw)
  - [ ] Accuracy score
  - [ ] Weapon breakdown (top weapons)
- [ ] Add MMR history graph (Chart.js or similar)
- [ ] Add performance metrics graph (deflect scores over time)
- [ ] Mobile responsive design

**Deliverable**: Player profile page

**PR Title**: `[Phase 4] feat(web): Add player profile page with MMR graphs`

---

### 4.3 Leaderboard Page
**Route**: `/leaderboard`

- [ ] Fetch leaderboard data from API
- [ ] Display table:
  - Rank, Name, MMR, K/D, Accuracy
- [ ] Add filters:
  - [ ] Gamemode dropdown (All, Default, Dodgeball)
  - [ ] Rank tier dropdown (All, Bronze, Silver, etc.)
- [ ] Pagination (50 per page)
- [ ] Highlight current user (if logged in)
- [ ] Auto-refresh every 60 seconds

**Deliverable**: Leaderboard page

**PR Title**: `[Phase 4] feat(web): Add leaderboard page with filters and pagination`

---

### 4.4 Server Browser Page
**Route**: `/servers`

- [ ] Fetch server list from API
- [ ] Display server cards:
  - Server name, IP:Port
  - Current map
  - Player count (X/24)
  - Gamemode
- [ ] Click server → Show live match details
- [ ] Auto-refresh every 10 seconds

**Deliverable**: Server browser page

**PR Title**: `[Phase 4] feat(web): Add server browser with live match display`

---

### 4.5 Landing Page
**Route**: `/`

- [ ] Hero section (project description)
- [ ] Feature highlights (kill weighting, MMR system)
- [ ] Top 10 players widget
- [ ] Recent activity feed
- [ ] "Search player" input (autocomplete)

**Deliverable**: Landing page

**PR Title**: `[Phase 4] feat(web): Add landing page with search and top players`

---

## 🚀 Phase 5: Production Deployment (Priority: MEDIUM)

### 5.1 Docker Compose
**File**: `docker-compose.yml`

- [ ] Define services:
  - [ ] `collector` (UDP listener)
  - [ ] `processor` (event workers, 5 replicas)
  - [ ] `api` (REST API)
  - [ ] `postgres` (database)
  - [ ] `redis` (cache/queue)
  - [ ] `frontend` (web UI)
- [ ] Configure networking (internal bridge)
- [ ] Configure volumes (persistent data)
- [ ] Configure health checks
- [ ] Configure restart policies

**Deliverable**: Docker Compose stack

**PR Title**: `[Phase 5] feat(deploy): Add Docker Compose stack for all services`

---

### 5.2 CI/CD Pipeline
**File**: `.github/workflows/build.yml`

- [ ] On push to main:
  - [ ] Run tests (`go test ./...`)
  - [ ] Build Docker images
  - [ ] Push to registry (Docker Hub or GHCR)
- [ ] On PR:
  - [ ] Run tests
  - [ ] Run linters
  - [ ] Check code coverage

**File**: `.github/workflows/deploy.yml`

- [ ] On tag (v*):
  - [ ] Deploy to production server (SSH)
  - [ ] Run database migrations
  - [ ] Restart services

**Deliverable**: Automated build/deploy pipeline

**PR Title**: `[Phase 5] feat(ci): Add GitHub Actions for testing and deployment`

---

### 5.3 Monitoring & Logging
- [ ] Set up Prometheus metrics:
  - Events processed/sec
  - Queue depth
  - API response times
  - Database query latency
- [ ] Set up Grafana dashboards
- [ ] Set up alerting (PagerDuty or similar):
  - Queue backup > 10k events
  - API error rate > 5%
  - Database connection failures
- [ ] Centralized logging (ELK stack or similar)

**Deliverable**: Monitoring infrastructure

**PR Title**: `[Phase 5] feat(monitoring): Add Prometheus metrics and Grafana dashboards`

---

### 5.4 Backups
**File**: `scripts/backup.sh`

- [ ] Automated PostgreSQL backups (daily)
- [ ] Upload to S3 or similar
- [ ] Test restore procedure
- [ ] Document disaster recovery plan

**Deliverable**: Backup automation

**PR Title**: `[Phase 5] feat(backups): Add automated database backup scripts`

---

### 5.5 Production Checklist
- [ ] Domain configured (stats.udl.tf → server IP)
- [ ] SSL certificate installed (Let's Encrypt)
- [ ] Firewall configured (allow 27500/UDP, 443/TCP)
- [ ] Security hardening (disable root SSH, fail2ban)
- [ ] Rate limiting on API (per-IP limits)
- [ ] Database tuning (connection pool, query optimization)
- [ ] Load testing completed (1000 events/min)
- [ ] Documentation updated (deployment guide)

**Deliverable**: Production-ready deployment

**PR Title**: `[Phase 5] docs: Add production deployment guide`

---

## 📊 Progress Tracking

### Overall Progress
- [ ] Phase 1: Core Backend (0/9 tasks)
- [ ] Phase 2: SourceMod Plugins (0/4 tasks)
- [ ] Phase 3: REST API (0/5 tasks)
- [ ] Phase 4: Web Interface (0/5 tasks)
- [ ] Phase 5: Production Deployment (0/5 tasks)

**Total**: 0/28 tasks complete (0%)

---

## 🤖 For Sub-Agents

**How to use this checklist**:

1. **Pick a section** (e.g., "1.4 MMR Calculator")
2. **Read PROJECT_BRIEF.md** for context
3. **Implement the feature** with tests
4. **Check off items** as you complete them
5. **Create PR** using the template (PULL_REQUEST_TEMPLATE.md)
6. **Update this file** to mark tasks complete

**Commit message format**:
```
feat(mmr): implement kill weight calculation

- Add CalculateKillWeight function with LOG2 formula
- Add clamping to [0.5, 1.5] range
- Add comprehensive unit tests
- Update checklist (Phase 1, task 1.4)

Closes #X
```

---

**Last Updated**: January 31, 2026  
**Next Review**: After Phase 1 completion
