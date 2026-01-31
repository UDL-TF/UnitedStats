# UnitedStats - Agent Briefing Document
**Project**: TF2 Skill-Based Statistics & MMR System  
**Repository**: https://github.com/UDL-TF/UnitedStats  
**Status**: Planning Phase → Implementation Starting  
**Tech Stack**: Golang, PostgreSQL, Redis, SourceMod (SourcePawn)

---

## 📋 Project Overview

### What We're Building
A Team Fortress 2 statistics system with **skill-based MMR ranking** that rewards quality kills over quantity. Unlike traditional stat trackers, killing stronger opponents counts for more than farming weak players.

### Why It Exists
- **Previous system (stats.udl.tf) failed** due to database bloat from storing raw logs
- Need **UDP-based event streaming** from TF2 servers
- Want **skill measurement**, not just K/D padding
- Support **multiple gamemodes** (default TF2, Dodgeball, MGE, etc.)

---

## 🎯 Core Concepts

### 1. Kill Weighting System
**Problem**: Traditional K/D can be farmed by playing against weak opponents.

**Solution**: Weight kills based on opponent MMR.

```
KillWeight = CLAMP(1 + 0.5 * LOG2(MMR_ratio), 0.5, 1.5)

Where: MMR_ratio = EnemyMMR / YourMMR

Examples:
- Equal opponent (1:1 ratio) → 1.0 weight
- 2x stronger opponent → 1.5 weight (max)
- 2x weaker opponent → 0.5 weight (min)
```

**Weighted K/D**:
```
WeightedKills = SUM(KillWeight for each kill)
K/D_weighted = WeightedKills / Deaths
```

### 2. Rank Tiers & MMR
```
Bronze       0 – 900
Silver       900 – 1400
Gold         1400 – 1900  ← Median
Platinum     1900 – 2600
Diamond      2600 – 3600
Master       3600+
```

### 3. RankScore Formula
```
RankScore = AccuracyScore * (1 + (K/D_weighted * RankWeight))

Where:
- AccuracyScore = Performance metrics (0.0 - 2.0)
- K/D_weighted = Weighted kills / deaths
- RankWeight = Diminishing factor (1.0 → 0.1 as rank increases)
```

**Rank Weight (Anti-Farming)**:
```
RankWeight = 1.0 - (CurrentMMR / 5000) * 0.9

Bronze (500 MMR):   RW = 0.91  (K/D matters most)
Master (4000 MMR):  RW = 0.28  (K/D matters least)
```

### 4. Performance Metrics (Dodgeball Example)
**Deflect Scoring**:
```
DeflectScore = (TimingAccuracy + AngleAccuracy) * (1 + RocketSpeed*0.1) * (1 + Distance*0.1)

Components:
- TimingAccuracy: 1.0 = perfect deflect window, 0.95 = 5% late
- AngleAccuracy: 1.0 = center aim, 0.97 = 10° off-center
- RocketSpeed: Direct Hit = 1.5, Stock = 1.0
- Distance: <256 HU = 1.0, <512 HU = 0.8, >512 HU = 0.5

Example:
Perfect timing (1.0) + 5° off (0.95) + Direct Hit (1.5) + 200 HU (1.0)
= (1.0 + 0.95) * 1.15 * 1.1 = 2.47 (exceptional)
```

**AccuracyScore** = Running average of last 100 deflects (0.0 - 2.0 range)

---

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  TF2 Game Servers                                           │
│  ├─ SourceMod Plugin: superlogs-default.sp                  │
│  ├─ SourceMod Plugin: superlogs-dodgeball.sp                │
│  └─ UDP sender → stats.udl.tf:27500                         │
└────────────┬────────────────────────────────────────────────┘
             │ UDP packets (structured log events)
             ▼
┌─────────────────────────────────────────────────────────────┐
│  Collector Service (Golang)                                 │
│  - Listens on UDP :27500                                    │
│  - Parses log events (kill, deflect, weaponstats, etc.)     │
│  - Pushes to Redis queue: "events:processing"               │
└────────────┬────────────────────────────────────────────────┘
             │ Redis queue
             ▼
┌─────────────────────────────────────────────────────────────┐
│  Processor Workers (Golang, 5-10 goroutines)                │
│  - Pop from queue → Calculate kill weights → Update stats   │
│  - Recalculate MMR in real-time                             │
│  - Mark dirty stats → Batch write every 30s                 │
└────────────┬────────────────────────────────────────────────┘
             │ Batch writes
             ▼
┌─────────────────────────────────────────────────────────────┐
│  PostgreSQL Database                                        │
│  - Stores AGGREGATED stats only (no raw logs!)              │
│  - Tables: players, player_statistics, weapon_statistics    │
│  - JSONB fields for gamemode-specific metrics               │
└────────────┬────────────────────────────────────────────────┘
             │ Queries
             ▼
┌─────────────────────────────────────────────────────────────┐
│  REST API (Golang - Gin framework)                          │
│  - GET /api/v1/players/:steamid                             │
│  - GET /api/v1/leaderboard?gamemode=dodgeball               │
│  - GET /api/v1/servers/:id/live                             │
└────────────┬────────────────────────────────────────────────┘
             │ JSON responses
             ▼
┌─────────────────────────────────────────────────────────────┐
│  Web Interface (Svelte/Next.js)                             │
│  - Player profiles, MMR graphs, leaderboards                │
└─────────────────────────────────────────────────────────────┘
```

---

## 📦 Tech Stack

### Backend (Golang 1.21+)
- **Collector**: UDP listener with goroutines
- **Processor**: Event workers with RabbitMQ + Watermill
- **API**: Gin framework
- **Libraries**: 
  - `github.com/ThreeDotsLabs/watermill` (pub/sub)
  - `github.com/ThreeDotsLabs/watermill-amqp` (RabbitMQ adapter)
  - `gorm.io/gorm` (ORM)

### Database
- **PostgreSQL 16**: All data storage (stats, tournaments, caching via materialized views)

### Message Queue
- **RabbitMQ (AMQP)**: Event streaming and pub/sub
- **Watermill**: Golang messaging library (persistence, backpressure)

### SourceMod Plugins (SourcePawn)
- **superlogs-default.sp**: Standard TF2 events
- **superlogs-dodgeball.sp**: Deflect tracking with accuracy
- **superlogs-core.inc**: Shared UDP sender, formatters

### Frontend
- **Framework**: Next.js (React + TypeScript)
- **Styling**: Tailwind CSS
- **Features**: SSR for SEO, API routes

### Deployment
- **Docker**: Dockerfiles for all services
- **Kubernetes (Helm)**: Self-hosted K8s deployment
- **CI/CD**: GitHub Actions (auto-build on push)

---

## 🗄️ Database Schema (Key Tables)

### `players`
```sql
CREATE TABLE players (
    player_id BIGSERIAL PRIMARY KEY,
    steam_id VARCHAR(64) UNIQUE NOT NULL,
    steam_name VARCHAR(255),
    current_mmr INT DEFAULT 1000,
    rank_tier VARCHAR(20) DEFAULT 'Bronze',
    first_seen TIMESTAMP DEFAULT NOW(),
    last_seen TIMESTAMP DEFAULT NOW()
);
```

### `player_statistics`
```sql
CREATE TABLE player_statistics (
    stat_id BIGSERIAL PRIMARY KEY,
    player_id BIGINT REFERENCES players(player_id),
    gamemode VARCHAR(64) DEFAULT 'default',
    
    total_kills INT DEFAULT 0,
    weighted_kills DECIMAL(10,2) DEFAULT 0,
    total_deaths INT DEFAULT 0,
    kd_weighted DECIMAL(10,4) DEFAULT 0,
    
    accuracy_score DECIMAL(4,2) DEFAULT 0,  -- 0.0 - 2.0
    rank_score INT DEFAULT 1000,
    rank_weight DECIMAL(4,2) DEFAULT 1.0,
    
    gamemode_metrics JSONB,  -- Flexible storage for deflect scores, etc.
    
    UNIQUE(player_id, gamemode)
);
```

### `kill_events` (temporary, 7-day retention)
```sql
CREATE TABLE kill_events (
    event_id BIGSERIAL PRIMARY KEY,
    killer_id BIGINT REFERENCES players(player_id),
    victim_id BIGINT REFERENCES players(player_id),
    
    killer_mmr INT,
    victim_mmr INT,
    kill_weight DECIMAL(4,2),
    
    weapon_used VARCHAR(128),
    timestamp TIMESTAMP DEFAULT NOW()
);
```

**JSONB Example** (gamemode_metrics for Dodgeball):
```json
{
  "dodgeball": {
    "total_deflects": 5678,
    "avg_timing_accuracy": 0.92,
    "avg_angle_accuracy": 0.88,
    "avg_deflect_score": 1.85,
    "best_deflect_score": 2.47
  }
}
```

---

## 📡 Event Flow Example

### Kill Event (Default TF2)
**SourceMod Plugin Sends**:
```
L 01/31/2026 - 12:34:56: "Scout<45><STEAM_1:0:123456><Red>" killed "Heavy<67><STEAM_1:1:654321><Blue>" 
  (weapon "scattergun")
  (killer_mmr "2380")
  (victim_mmr "3700")
  (is_airshot "1")
```

**Collector Parses** → **Processor Calculates**:
```go
MMR_ratio = 3700 / 2380 = 1.55
KillWeight = 1 + 0.5 * LOG2(1.55) = 1.32

// Update killer stats
WeightedKills += 1.32
Kills += 1

// Update victim stats
Deaths += 1

// Recalculate killer MMR
RankScore = AccuracyScore * (1 + (K/D_weighted * RankWeight))
NewMMR = RankScore * 1000  // Scale factor
```

### Deflect Event (Dodgeball)
**SourceMod Plugin Sends**:
```
L 01/31/2026 - 12:34:56: "Pyro<45><STEAM_1:0:123456><Red>" triggered "deflect"
  (timing_accuracy "1.0")
  (angle_accuracy "0.95")
  (rocket_speed "1.5")
  (distance "210")
  (deflect_score "2.47")
```

**Processor Updates**:
```go
// Update running average
NewAccuracyScore = (OldAccuracyScore * TotalDeflects + 2.47) / (TotalDeflects + 1)

// Recalculate MMR
NewMMR = CalculateRankScore(NewAccuracyScore, KDWeighted, CurrentMMR)
```

---

## 📂 Project Structure

```
unitedstats/
├── cmd/
│   ├── collector/
│   │   └── main.go              # UDP listener service
│   ├── processor/
│   │   └── main.go              # Event processor workers
│   └── api/
│       └── main.go              # REST API server (Gin)
│
├── internal/
│   ├── parser/
│   │   ├── parser.go            # Log line parsing (kill, deflect, etc.)
│   │   └── parser_test.go
│   ├── mmr/
│   │   ├── calculator.go        # MMR formulas (kill weight, rank score)
│   │   └── calculator_test.go
│   ├── performance/
│   │   └── metrics.go           # Deflect scoring, accuracy calculations
│   ├── models/
│   │   ├── player.go            # Database models (GORM)
│   │   ├── statistics.go
│   │   └── events.go
│   └── queue/
│       └── redis.go             # Redis queue interface
│
├── pkg/
│   └── events/
│       └── events.go            # Shared event structs
│
├── sourcemod/
│   └── scripting/
│       ├── superlogs-default.sp
│       ├── superlogs-dodgeball.sp
│       └── include/
│           └── superlogs-core.inc
│
├── migrations/
│   ├── 001_initial_schema.sql
│   ├── 002_add_gamemode_metrics.sql
│   └── ...
│
├── web/
│   ├── src/
│   ├── public/
│   └── package.json
│
├── docker-compose.yml
├── Dockerfile.collector
├── Dockerfile.processor
├── Dockerfile.api
├── go.mod
├── go.sum
└── README.md
```

---

## 🚀 Implementation Phases

### Phase 1: Core Backend (Week 1-2)
**Goal**: Basic UDP → Database pipeline

**Tasks**:
1. ✅ Initialize Golang project (`go mod init github.com/UDL-TF/UnitedStats`)
2. ✅ Implement UDP collector (`cmd/collector/main.go`)
3. ✅ Implement event parser (`internal/parser/parser.go`)
   - Parse kill events
   - Parse deflect events
4. ✅ Implement MMR calculator (`internal/mmr/calculator.go`)
   - `CalculateKillWeight(killerMMR, victimMMR)`
   - `CalculateRankWeight(currentMMR)`
   - `CalculateRankScore(accuracyScore, kdWeighted, currentMMR)`
5. ✅ Set up PostgreSQL schema (`migrations/001_initial_schema.sql`)
6. ✅ Implement Redis queue (`internal/queue/redis.go`)
7. ✅ Implement processor workers (`cmd/processor/main.go`)
8. ✅ Unit tests for MMR formulas

**Deliverable**: Can receive UDP events → Calculate weights → Store in DB

---

### Phase 2: SourceMod Plugins (Week 2-3)
**Goal**: TF2 servers can send events

**Tasks**:
1. ✅ Create `superlogs-core.inc` (shared UDP sender)
2. ✅ Implement `superlogs-dodgeball.sp`
   - Hook rocket deflects
   - Calculate timing accuracy
   - Calculate angle accuracy
   - Send UDP events
3. ✅ Implement `superlogs-default.sp`
   - Hook player deaths
   - Fetch player MMR (HTTP request to API)
   - Send UDP kill events with MMR
4. ✅ Test on local TF2 server

**Deliverable**: Live TF2 server sends events to backend

---

### Phase 3: REST API (Week 3-4)
**Goal**: Query player stats via HTTP

**Tasks**:
1. ✅ Implement Gin API server (`cmd/api/main.go`)
2. ✅ Endpoints:
   - `GET /api/v1/players/:steamid`
   - `GET /api/v1/leaderboard?gamemode=X&rank=Y`
   - `GET /api/v1/servers/:id/live`
3. ✅ Response formatting (JSON)
4. ✅ CORS setup for frontend

**Deliverable**: API returns player MMR, stats, leaderboards

---

### Phase 4: Web Interface (Week 4-5)
**Goal**: User-facing website

**Tasks**:
1. ✅ Choose framework (Svelte recommended)
2. ✅ Player profile page (MMR, rank, stats)
3. ✅ Leaderboard page (filters by gamemode, rank)
4. ✅ Performance graphs (MMR over time, deflect scores)
5. ✅ Server browser (live matches)

**Deliverable**: Public website at stats.udl.tf

---

### Phase 5: Production Deployment (Week 5-6)
**Goal**: Live on production servers

**Tasks**:
1. ✅ Docker Compose setup
2. ✅ Deploy to server (DigitalOcean/AWS)
3. ✅ Configure DNS (stats.udl.tf → server IP)
4. ✅ Install plugins on UDL servers
5. ✅ Monitoring (Prometheus, Grafana)
6. ✅ Backups (automated PostgreSQL dumps)

**Deliverable**: Live system tracking stats for UDL community

---

## 🧪 Testing Requirements

### Unit Tests (Golang)
```go
// internal/mmr/calculator_test.go
func TestKillWeightCalculation(t *testing.T) {
    tests := []struct {
        killerMMR, victimMMR int
        expected             float64
    }{
        {2000, 2000, 1.0},   // Equal
        {2000, 4000, 1.5},   // 2x stronger (capped)
        {2000, 1000, 0.5},   // 2x weaker (capped)
    }
    
    for _, tt := range tests {
        result := CalculateKillWeight(tt.killerMMR, tt.victimMMR)
        assert.InDelta(t, tt.expected, result, 0.01)
    }
}
```

### Integration Tests
1. **UDP → Parser**: Send fake UDP packet, verify parsing
2. **Parser → Processor**: Verify MMR calculation
3. **Processor → Database**: Verify stats update
4. **Database → API**: Verify JSON response

### Load Tests
- **Scenario**: 10 servers × 24 players × 5 events/min = 1200 events/min
- **Target**: <10ms processing latency per event

---

## 📝 Code Examples

### MMR Calculator (`internal/mmr/calculator.go`)
```go
package mmr

import "math"

func CalculateKillWeight(killerMMR, victimMMR int) float64 {
    if killerMMR == 0 || victimMMR == 0 {
        return 1.0
    }
    
    ratio := float64(victimMMR) / float64(killerMMR)
    weight := 1.0 + 0.5*math.Log2(ratio)
    
    // Clamp [0.5, 1.5]
    if weight < 0.5 {
        return 0.5
    }
    if weight > 1.5 {
        return 1.5
    }
    
    return weight
}

func CalculateRankWeight(currentMMR int) float64 {
    rw := 1.0 - (float64(currentMMR)/5000.0)*0.9
    if rw < 0.1 {
        return 0.1
    }
    return rw
}

func CalculateRankScore(accuracyScore, kdWeighted float64, currentMMR int) int {
    rw := CalculateRankWeight(currentMMR)
    rankScore := accuracyScore * (1.0 + (kdWeighted * rw))
    mmr := int(rankScore * 1000)
    
    if mmr < 0 {
        return 0
    }
    if mmr > 5000 {
        return 5000
    }
    
    return mmr
}
```

### UDP Collector (`cmd/collector/main.go`)
```go
package main

import (
    "fmt"
    "net"
)

func main() {
    addr := net.UDPAddr{Port: 27500, IP: net.ParseIP("0.0.0.0")}
    conn, _ := net.ListenUDP("udp", &addr)
    defer conn.Close()
    
    fmt.Println("[Collector] Listening on UDP :27500")
    
    buffer := make([]byte, 4096)
    
    for {
        n, remoteAddr, _ := conn.ReadFromUDP(buffer)
        logLine := string(buffer[:n])
        
        go handleLogLine(logLine, remoteAddr.IP.String())
    }
}

func handleLogLine(logLine, serverIP string) {
    // Parse → Queue → Process
    event := parser.ParseLogLine(logLine, serverIP)
    queue.Push("events:processing", event)
}
```

### SourceMod UDP Sender (`superlogs-core.inc`)
```sourcepawn
#include <socket>

Handle g_hSocket;
char g_szStatsIP[64] = "stats.udl.tf";
int g_iStatsPort = 27500;

void SuperLogs_Init() {
    g_hSocket = SocketCreate(SOCKET_UDP, OnSocketError);
}

void SuperLogs_SendUDP(const char[] message) {
    SocketSendTo(g_hSocket, message, strlen(message), g_szStatsIP, g_iStatsPort);
}
```

---

## 🎯 Success Criteria

### MVP (Minimum Viable Product)
- ✅ 1 TF2 server sending events
- ✅ UDP collector receiving & parsing
- ✅ MMR calculations working correctly
- ✅ Database storing aggregated stats
- ✅ API returning player profiles
- ✅ Basic web interface showing leaderboard

### Production Ready
- ✅ 10+ servers connected
- ✅ Sub-100ms event processing latency
- ✅ 99.9% uptime
- ✅ Automated backups
- ✅ Monitoring & alerting
- ✅ Mobile-responsive UI

---

## 🔧 Development Workflow

### For Agent PRs:
1. **Read this document** + `UnitedStats_SRD_Draft_v2.md` for full specs
2. **Choose a phase/task** from Implementation Phases
3. **Create feature branch**: `git checkout -b feature/mmr-calculator`
4. **Implement with tests**
5. **PR title format**: `[Phase X] Feature: MMR Calculator Implementation`
6. **PR description**: Include test results, code examples

### Git Workflow
```bash
# Clone repo
git clone https://github.com/UDL-TF/UnitedStats.git

# Create feature branch
git checkout -b feature/udp-collector

# Commit with conventional commits
git commit -m "feat(collector): implement UDP listener with goroutines"

# Push and create PR
git push origin feature/udp-collector
```

---

## 📚 Key Documents

1. **This file** (`PROJECT_BRIEF.md`): Quick reference for agents
2. **Full SRD** (`UnitedStats_SRD_Draft_v2.md`): Complete specifications
3. **Database Schema** (`migrations/001_initial_schema.sql`): DB structure
4. **API Docs** (TBD): Endpoint specifications

---

## 🤝 Contributing Guidelines

### Code Style
- **Golang**: `gofmt`, `golint`, standard error handling
- **SourcePawn**: Follow AlliedModders style guide
- **Commits**: Conventional Commits format (`feat:`, `fix:`, `docs:`)

### Testing
- All new functions need unit tests
- Integration tests for critical paths
- Load tests for performance-critical code

### Documentation
- Comment all public functions (GoDoc format)
- Update README.md with new features
- Add migration notes for DB changes

---

## ❓ FAQ for Agents

**Q: Where do I start if implementing the collector?**  
A: Read Phase 1, Task 2. Implement `cmd/collector/main.go` with UDP listener. Reference code example in this doc.

**Q: How do I test MMR calculations without a live server?**  
A: Write unit tests in `internal/mmr/calculator_test.go`. Use table-driven tests with known inputs/outputs.

**Q: What's the priority order for implementation?**  
A: Follow phases sequentially: Backend (Phase 1) → Plugins (Phase 2) → API (Phase 3) → Web (Phase 4).

**Q: How do I handle gamemode-specific metrics?**  
A: Store in JSONB column `gamemode_metrics`. Each gamemode has its own JSON structure (see schema examples).

**Q: Should I use GORM or raw SQL?**  
A: GORM for simple queries, raw SQL for complex aggregations. Performance matters.

---

## 🚨 Critical Constraints

### DO:
✅ Store **aggregated stats only** (no raw logs)  
✅ Use **batch writes** (every 30s) to reduce DB load  
✅ **Clamp kill weights** to [0.5, 1.5] range  
✅ Calculate **MMR in real-time** after each event  
✅ Support **multiple gamemodes** via JSONB  

### DON'T:
❌ Store raw event logs in database (killed the old system)  
❌ Make individual DB writes per event (use batching)  
❌ Allow uncapped kill weights (prevents exploitation)  
❌ Use synchronous processing (use goroutines/workers)  
❌ Hardcode gamemode logic (keep it flexible)  

---

## 📞 Contact & Support

- **Primary Stakeholder**: UDL-TF community
- **Tech Lead**: (You - specify GitHub username)
- **Repository**: https://github.com/UDL-TF/UnitedStats
- **Issues**: Use GitHub Issues for bugs/features
- **Discussions**: GitHub Discussions for questions

---

**Ready to build? Pick a phase, create a branch, and ship it!** 🚀

**Document Version**: 1.0  
**Last Updated**: January 31, 2026  
**Next Review**: After Phase 1 completion
