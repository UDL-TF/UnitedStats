# 🔄 UnitedStats v3.0 - Architecture Changes Summary

**Date**: January 31, 2026  
**Version**: 3.0  
**Status**: Major Update - Tournament System Added

---

## 🎯 What Changed (v2 → v3)

### 1. ❌ **REMOVED: Redis**
**Old Architecture** (v2):
- Redis for queuing events
- Redis for caching player stats
- `go-redis` library

**Why Removed**:
- Added complexity (extra service to manage)
- RabbitMQ provides better message persistence
- PostgreSQL materialized views can handle caching

---

### 2. ✅ **ADDED: RabbitMQ + Watermill**

**New Message Queue**: RabbitMQ (AMQP protocol)

**Why RabbitMQ**:
- Message persistence (survives crashes)
- Better backpressure handling
- Message acknowledgment (at-least-once delivery)
- Industry standard for event streaming
- Easier horizontal scaling

**Watermill Library** ([watermill.io](https://watermill.io/)):
- Clean pub/sub abstraction
- Multiple backend support (AMQP, Kafka, Redis, etc.)
- Built-in middleware (metrics, retries, poison queue)
- Golang-native

**Example Usage**:
```go
// Collector publishes
publisher.Publish("events.kill", killEvent)

// Processor subscribes
subscriber.Subscribe("events.kill", func(event *Event) {
    processKillEvent(event, db)
})
```

---

### 3. ✅ **CONFIRMED: PostgreSQL Only**

**All Data Storage**:
- Public ladder stats
- Tournament data
- Player profiles
- Match history

**Caching Strategy**:
- Materialized views for leaderboards (refresh every 5 minutes)
- Database-level query caching
- No external cache needed (PostgreSQL is fast enough)

**Why No Redis Cache**:
- PostgreSQL can handle the load
- Simpler infrastructure
- Less data synchronization issues

---

### 4. ✅ **CONFIRMED: Next.js Frontend**

**Framework**: Next.js 14 (React + TypeScript)

**Key Features**:
- Server-side rendering (SSR) for SEO
- API routes (can proxy to Golang API)
- Static site generation for leaderboards
- TypeScript for type safety
- Tailwind CSS for styling

**Why Next.js**:
- Best React framework for production
- Great developer experience
- Built-in optimizations
- Strong community support

---

### 5. ✅ **ADDED: Helm Charts for Kubernetes**

**Self-Hosted Deployment**:
- Docker images for all services
- Helm charts for Kubernetes orchestration
- Persistent volumes for PostgreSQL + RabbitMQ

**Services**:
- Collector (2 replicas for HA)
- Processor (5 replicas for load distribution)
- API (3 replicas behind load balancer)
- Frontend (2 replicas)
- PostgreSQL (StatefulSet)
- RabbitMQ (StatefulSet)

**Why Kubernetes**:
- Automatic scaling
- Self-healing
- Rolling updates
- Easy monitoring

---

### 6. ✅ **CONFIRMED: Configuration**

| Setting | Value |
|---------|-------|
| MMR Decay | Yes (inactive players lose MMR) |
| Privacy | Public (all stats visible) |
| Authentication | Not required for viewing |
| Frontend | Next.js |
| Hosting | Self-hosted (Docker + K8s) |

---

## 🏆 NEW: Tournament System

### Core Features

#### 1. **Tournament Formats Supported**
- **Swiss System** (pairing based on W-L record)
- **Single Elimination** (bracket, one loss = out)
- **Double Elimination** (upper/lower bracket)

#### 2. **Tournament Workflow**
```
Admin creates tournament
  ↓
Players register
  ↓
Bracket/pairings generated
  ↓
Servers assigned to matches
  ↓
Matches played (events tagged with tournament_id)
  ↓
Results reported
  ↓
Bracket advances
  ↓
Tournament completes
  ↓
Final standings published
```

#### 3. **Server Integration**
- **On-demand server assignment**: Matches assigned to specific servers
- **Tournament tagging**: All events (kills, deflects) tagged with `tournament_id` + `match_id`
- **Stats isolation**: Tournament stats separate from public ladder
- **Live tracking**: Real-time bracket updates during matches

#### 4. **Database Schema**
**New Tables**:
- `tournaments` - Tournament metadata
- `tournament_participants` - Player/team registrations
- `tournament_matches` - Match schedule and results
- `tournament_stats` - Player performance within tournament

**Updated Tables**:
- `kill_events` - Added `tournament_id`, `match_id` columns
- `player_statistics` - Can now have gamemode="tournament_X"

---

### Tournament API Endpoints

```
POST   /api/v1/tournaments                          # Create tournament
GET    /api/v1/tournaments                          # List tournaments
GET    /api/v1/tournaments/:id                      # Tournament details
POST   /api/v1/tournaments/:id/register             # Player registration
POST   /api/v1/tournaments/:id/generate_bracket     # Generate pairings
GET    /api/v1/tournaments/:id/bracket              # Get bracket data
PUT    /api/v1/tournaments/:id/matches/:match_id/assign_server
POST   /api/v1/tournaments/:id/matches/:match_id/result
GET    /api/v1/tournaments/:id/leaderboard          # Tournament standings
GET    /api/v1/servers/:server_id/current_match    # Server polls for assigned match
```

---

### Frontend Pages (Next.js)

**New Pages**:
- `/tournaments` - List of all tournaments
- `/tournaments/[id]` - Tournament detail (bracket view)
- `/tournaments/[id]/matches/[matchId]` - Match detail
- `/admin/tournaments/create` - Admin tournament creation

**Components**:
- `BracketVisualization.tsx` - SVG bracket renderer (single/double elim)
- `SwissStandings.tsx` - Table showing W-L records and pairings
- `MatchCard.tsx` - Live match scores
- `TournamentCard.tsx` - Tournament summary cards

---

## 📦 Updated Tech Stack

| Component | Old (v2) | New (v3) | Status |
|-----------|----------|----------|--------|
| Backend | Golang | Golang | ✅ Same |
| Database | PostgreSQL + Redis | PostgreSQL only | 🔄 Simplified |
| Message Queue | Redis | RabbitMQ (AMQP) | 🆕 Changed |
| Pub/Sub Library | go-redis | Watermill | 🆕 New |
| API | Gin | Gin | ✅ Same |
| Frontend | Svelte/Next.js (TBD) | Next.js | ✅ Confirmed |
| Deployment | Docker Compose | Docker + Helm (K8s) | 🆕 Enhanced |
| Caching | Redis | PostgreSQL materialized views | 🔄 Changed |

---

## 🗂️ Updated Project Structure

```
unitedstats/
├── cmd/
│   ├── collector/         # UDP listener → RabbitMQ publisher
│   ├── processor/         # RabbitMQ subscriber → DB writer
│   ├── api/               # REST API (Gin)
│   └── tournament/        # 🆕 Tournament manager service
│
├── internal/
│   ├── parser/            # Log line parsing
│   ├── mmr/               # MMR calculations
│   ├── performance/       # Performance metrics
│   ├── models/            # GORM models
│   ├── queue/             # 🔄 Watermill pub/sub (replaces Redis)
│   └── tournament/        # 🆕 Tournament logic
│       ├── bracket.go     # Bracket generation
│       ├── swiss.go       # Swiss pairing
│       └── manager.go     # Tournament lifecycle
│
├── pkg/
│   └── events/            # Shared event structs
│
├── sourcemod/
│   └── scripting/
│       ├── superlogs-default.sp
│       ├── superlogs-dodgeball.sp
│       └── include/
│           └── superlogs-core.inc
│
├── frontend/              # 🆕 Next.js app
│   ├── app/
│   │   ├── tournaments/
│   │   │   ├── page.tsx
│   │   │   └── [id]/
│   │   │       ├── page.tsx
│   │   │       └── matches/[matchId]/page.tsx
│   │   ├── leaderboard/
│   │   └── players/
│   ├── components/
│   │   ├── BracketVisualization.tsx
│   │   ├── SwissStandings.tsx
│   │   └── MatchCard.tsx
│   └── package.json
│
├── migrations/
│   ├── 001_initial_schema.sql
│   ├── 002_add_gamemode_metrics.sql
│   └── 003_tournament_system.sql     # 🆕 Tournament tables
│
├── helm/                  # 🆕 Kubernetes deployment
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
│       ├── collector-deployment.yaml
│       ├── processor-deployment.yaml
│       ├── api-deployment.yaml
│       ├── frontend-deployment.yaml
│       ├── postgres-statefulset.yaml
│       └── rabbitmq-statefulset.yaml
│
├── docker/                # 🆕 Dockerfiles
│   ├── Dockerfile.collector
│   ├── Dockerfile.processor
│   ├── Dockerfile.api
│   └── Dockerfile.frontend
│
├── docker-compose.yml     # Local development
├── go.mod
└── README.md
```

---

## 🔄 Migration Guide (v2 → v3)

### For Agents Working on Backend:

#### 1. Replace Redis Queue with Watermill
**Before (v2)**:
```go
import "github.com/go-redis/redis/v8"

// Push to Redis
redisClient.RPush(ctx, "events:processing", eventJSON)

// Pop from Redis
result := redisClient.BRPop(ctx, 0, "events:processing")
```

**After (v3)**:
```go
import (
    "github.com/ThreeDotsLabs/watermill"
    "github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
)

// Publish to RabbitMQ
publisher.Publish("events.kill", msg)

// Subscribe to RabbitMQ
messages, _ := subscriber.Subscribe(ctx, "events.kill")
for msg := range messages {
    processEvent(msg)
    msg.Ack()
}
```

**Dependencies**:
```bash
go get github.com/ThreeDotsLabs/watermill
go get github.com/ThreeDotsLabs/watermill-amqp
```

#### 2. Remove Redis Caching Logic
**Before (v2)**:
```go
// Cache player stats in Redis
redisClient.HSet(ctx, fmt.Sprintf("player:%s", steamID), "mmr", mmr)

// Get from cache
cached := redisClient.HGet(ctx, fmt.Sprintf("player:%s", steamID), "mmr")
```

**After (v3)**:
```go
// Query PostgreSQL directly (with materialized views for leaderboards)
db.Raw("SELECT * FROM player_leaderboard_mv WHERE steam_id = ?", steamID).Scan(&player)

// Or use GORM
db.Where("steam_id = ?", steamID).First(&player)
```

**Create Materialized View**:
```sql
CREATE MATERIALIZED VIEW player_leaderboard_mv AS
SELECT 
    steam_id,
    steam_name,
    current_mmr,
    rank_tier,
    kd_weighted
FROM players
JOIN player_statistics USING (player_id)
ORDER BY current_mmr DESC
LIMIT 1000;

-- Refresh every 5 minutes (cron or trigger)
REFRESH MATERIALIZED VIEW player_leaderboard_mv;
```

---

### For Agents Working on Frontend:

#### Next.js Setup
```bash
npx create-next-app@latest frontend --typescript --tailwind --app

cd frontend
npm install
npm run dev
```

#### Example Tournament Page (`app/tournaments/[id]/page.tsx`)
```tsx
import { BracketVisualization } from '@/components/BracketVisualization';

export default async function TournamentPage({ params }: { params: { id: string } }) {
  const res = await fetch(`http://api:8080/api/v1/tournaments/${params.id}`);
  const tournament = await res.json();
  
  return (
    <div>
      <h1>{tournament.name}</h1>
      <BracketVisualization bracket={tournament.bracket} />
    </div>
  );
}
```

---

### For Agents Working on Deployment:

#### Docker Compose (Development)
```yaml
version: '3.8'

services:
  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"   # AMQP
      - "15672:15672" # Management UI
    environment:
      RABBITMQ_DEFAULT_USER: stats
      RABBITMQ_DEFAULT_PASS: password

  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: unitedstats
      POSTGRES_USER: stats_user
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  collector:
    build:
      context: .
      dockerfile: docker/Dockerfile.collector
    ports:
      - "27500:27500/udp"
    environment:
      AMQP_URI: amqp://stats:password@rabbitmq:5672/

  processor:
    build:
      context: .
      dockerfile: docker/Dockerfile.processor
    environment:
      AMQP_URI: amqp://stats:password@rabbitmq:5672/
      DB_URI: postgres://stats_user:password@postgres:5432/unitedstats
    deploy:
      replicas: 3

volumes:
  postgres_data:
```

#### Helm Chart (`helm/values.yaml`)
```yaml
collector:
  replicas: 2
  image:
    repository: unitedstats/collector
    tag: latest

processor:
  replicas: 5
  image:
    repository: unitedstats/processor
    tag: latest

api:
  replicas: 3
  image:
    repository: unitedstats/api
    tag: latest

frontend:
  replicas: 2
  image:
    repository: unitedstats/frontend
    tag: latest

postgres:
  persistence:
    size: 100Gi
    storageClass: local-path

rabbitmq:
  persistence:
    size: 20Gi
    storageClass: local-path
```

---

## 📊 Updated Implementation Phases

### Phase 1: Core Backend (Updated)
1. ✅ Project setup
2. ✅ Database schema
3. ✅ Event structs
4. ✅ MMR calculator
5. ✅ Event parser
6. 🆕 **RabbitMQ + Watermill setup** (replaces Redis)
7. ✅ UDP collector (now publishes to RabbitMQ)
8. ✅ Event processor (now subscribes to RabbitMQ)
9. ✅ Integration tests

---

### Phase 2: SourceMod Plugins (Same)
1. ✅ Core library
2. ✅ Dodgeball plugin
3. ✅ Default TF2 plugin
4. ✅ Testing

---

### Phase 2.5: Tournament System (NEW)
1. 🆕 **Database schema** (tournaments, matches, participants)
2. 🆕 **Bracket generation** (single/double elimination)
3. 🆕 **Swiss pairing algorithm**
4. 🆕 **Tournament manager** (lifecycle, server assignment)
5. 🆕 **Tournament API endpoints**

---

### Phase 3: REST API (Updated)
1. ✅ API server setup
2. ✅ Player endpoints
3. ✅ Leaderboard endpoints
4. ✅ Server endpoints
5. ✅ API documentation
6. 🆕 **Tournament endpoints**

---

### Phase 4: Web Interface (Updated to Next.js)
1. 🔄 **Frontend setup** (Next.js instead of Svelte)
2. ✅ Player profile page
3. ✅ Leaderboard page
4. ✅ Server browser
5. ✅ Landing page
6. 🆕 **Tournament pages** (bracket view, Swiss standings)

---

### Phase 5: Production Deployment (Enhanced)
1. ✅ Docker Compose
2. ✅ CI/CD pipeline
3. ✅ Monitoring & logging
4. ✅ Backups
5. ✅ Production checklist
6. 🆕 **Helm charts** (Kubernetes deployment)

---

## 🎯 What Agents Should Do Next

### Priority 1: Update Core Infrastructure
1. **Remove Redis dependencies** from collector/processor
2. **Implement Watermill publisher** (`internal/queue/publisher.go`)
3. **Implement Watermill subscriber** (`internal/queue/subscriber.go`)
4. **Update collector** to publish to RabbitMQ
5. **Update processor** to subscribe from RabbitMQ
6. **Test message flow** (UDP → Collector → RabbitMQ → Processor → DB)

### Priority 2: Add Tournament System
1. **Create tournament database schema** (`migrations/003_tournament_system.sql`)
2. **Implement bracket generation** (`internal/tournament/bracket.go`)
3. **Implement Swiss pairing** (`internal/tournament/swiss.go`)
4. **Create tournament API endpoints** (`cmd/api/handlers/tournament.go`)
5. **Test tournament creation flow**

### Priority 3: Build Next.js Frontend
1. **Initialize Next.js project** (`npx create-next-app frontend`)
2. **Create tournament list page** (`app/tournaments/page.tsx`)
3. **Create bracket visualization component** (`components/BracketVisualization.tsx`)
4. **Create Swiss standings component** (`components/SwissStandings.tsx`)
5. **Test with mock data**

### Priority 4: Deployment Setup
1. **Write Dockerfiles** for all services
2. **Create Helm chart** (`helm/Chart.yaml` + templates)
3. **Test local deployment** with Docker Compose
4. **Document deployment process**

---

## 🔧 Code Migration Examples

### Example 1: Collector (Redis → RabbitMQ)

**Before (v2)**:
```go
// cmd/collector/main.go
func handleLogLine(logLine, serverIP string, redisQueue *queue.RedisQueue) {
    event, err := parser.ParseLogLine(logLine, serverIP)
    if err != nil {
        log.Printf("Parse error: %v", err)
        return
    }
    
    redisQueue.Push("events:processing", event)
}
```

**After (v3)**:
```go
// cmd/collector/main.go
func handleLogLine(logLine, serverIP string, publisher *queue.Publisher) {
    event, err := parser.ParseLogLine(logLine, serverIP)
    if err != nil {
        log.Printf("Parse error: %v", err)
        return
    }
    
    topic := fmt.Sprintf("events.%s", event.Type)  // e.g., "events.kill"
    publisher.Publish(topic, event)
}
```

---

### Example 2: Processor (Redis → RabbitMQ)

**Before (v2)**:
```go
// cmd/processor/main.go
func main() {
    redisQueue := queue.NewRedisQueue("localhost:6379")
    
    for {
        event := redisQueue.Pop("events:processing")
        processEvent(event, db)
    }
}
```

**After (v3)**:
```go
// cmd/processor/main.go
func main() {
    subscriber := queue.NewSubscriber("amqp://localhost:5672")
    
    subscriber.Subscribe("events.kill", func(event *Event) {
        processKillEvent(event, db)
    })
    
    subscriber.Subscribe("events.deflect", func(event *Event) {
        processDeflectEvent(event, db)
    })
    
    select {}  // Run forever
}
```

---

## ✅ Open Questions RESOLVED

| Question | Answer |
|----------|--------|
| Database choice? | ✅ PostgreSQL only |
| Frontend framework? | ✅ Next.js |
| Hosting approach? | ✅ Self-hosted (Docker + Helm) |
| MMR decay for inactive? | ✅ Yes |
| Privacy model? | ✅ Everything public |
| Message queue? | ✅ RabbitMQ (AMQP) |

### NEW Open Questions (v3):
1. **MMR decay rate**: How much MMR per week of inactivity? (Suggest: -10 to -20)
2. **Tournament MMR**: Separate rating or affect public ladder?
3. **Server provisioning**: Manual assignment or auto-spawn?
4. **Match check-in**: Require player confirmation before match?
5. **Bye rounds**: How to handle odd participant count in brackets?

---

## 📚 Updated Documentation

**Documents to Update**:
- ✅ `UnitedStats_SRD_Draft_v3.md` - Full spec with tournament system
- 🔄 `PROJECT_BRIEF.md` - Update tech stack section
- 🔄 `IMPLEMENTATION_CHECKLIST.md` - Add Phase 2.5 tasks
- 🔄 `QUICKSTART_AGENTS.md` - Update setup instructions for RabbitMQ
- 🔄 `README.md` - Update tech stack table

---

## 🚀 Timeline Update

**Original Estimate (v2)**: 4-6 weeks for production

**New Estimate (v3)**: 8-10 weeks for production (added tournament system)

**MVP Estimate**: 6-8 weeks (basic tournament with single elimination)

---

**Document Version**: 3.0  
**Status**: Ready for Implementation  
**Updated**: January 31, 2026

---

## Summary of Major Changes

1. ❌ **Removed Redis** → Simplified infrastructure
2. ✅ **Added RabbitMQ + Watermill** → Better message queue
3. ✅ **Confirmed Next.js** → Frontend framework decided
4. ✅ **Added Tournament System** → Major new feature
5. ✅ **Added Helm Charts** → Kubernetes deployment
6. ✅ **Confirmed Self-Hosted** → No cloud dependencies

**Total New Tasks**: ~15 additional tasks across all phases

**Impact on Timeline**: +2-4 weeks due to tournament system complexity

**Agents should prioritize**: RabbitMQ migration first, then tournament system
