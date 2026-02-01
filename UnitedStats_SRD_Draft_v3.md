# UnitedStats v3.0 - Updated System Requirements

**MAJOR UPDATES - January 31, 2026**

This document reflects the latest architectural decisions and new tournament system requirements.

---

## 🔄 Architecture Changes (v2 → v3)

### ❌ **REMOVED: Redis**
- No longer using Redis for queuing or caching
- Simplifies infrastructure (one less service)

### ✅ **ADDED: AMQP with Watermill**
- **Message broker**: RabbitMQ (AMQP protocol)
- **Golang library**: [Watermill](https://watermill.io/) for pub/sub messaging
- **Benefits**:
  - Built-in persistence (messages survive crashes)
  - Better backpressure handling
  - Message acknowledgment (at-least-once delivery)
  - Easier horizontal scaling

### ✅ **Database: PostgreSQL Only**
- **Primary database**: PostgreSQL 16
- **Caching**: Use PostgreSQL materialized views + built-in caching
- **Session storage**: PostgreSQL (if needed for auth)

### ✅ **Frontend: Next.js (Confirmed)**
- React-based framework
- Server-side rendering for SEO
- API routes for backend integration
- TypeScript support

### ✅ **Deployment: Self-Hosted**
- **Dockerfiles** for all services
- **Helm charts** for Kubernetes deployment
- No cloud provider dependencies

### ✅ **MMR Decay: Enabled**
- Inactive players lose MMR over time
- Decay rate: TBD (e.g., -10 MMR per week of inactivity)

### ✅ **Privacy: Everything Public**
- No private profiles
- All stats, leaderboards, match history publicly visible
- No authentication required for viewing

---

## 🏆 NEW: Tournament System

### Core Requirement
**Integrated tournament management** within the stats system (not a separate platform).

**Supported formats**:
1. **Swiss System** (round-robin with pairing)
2. **Single Elimination** (bracket, one loss = out)
3. **Double Elimination** (bracket, two losses = out)

---

### Tournament Features

#### 1. League/Tournament Creation
- **Admin interface** to create tournaments
- **Settings**:
  - Tournament name, format (Swiss/Single/Double)
  - Start/end dates
  - Player registration period
  - Server pool assignment
  - Match rules (best-of-X, map pool)

#### 2. Dynamic Server Allocation
- **On-demand server creation** for tournament matches
- **Server reporting**: Each server knows which tournament match it's hosting
- **Event tagging**: All events (kills, deflects) tagged with `tournament_id` + `match_id`
- **Stats isolation**: Tournament stats separate from public ladder stats

#### 3. Match Management
- **Scheduling**: Bracket generation (automatic pairing)
- **Server assignment**: "Match 1: Team A vs Team B → Server udl-tournament-1"
- **Score reporting**: Servers send results to tournament API
- **Validation**: Ensure correct teams/players on correct server

#### 4. Bracket/Swiss Progression
- **Swiss**: Auto-pair based on current standings (W-L record)
- **Single Elimination**: Winner advances, loser eliminated
- **Double Elimination**: Upper/lower bracket progression

#### 5. Tournament Leaderboards
- **Separate from public MMR**
- **Tournament-specific stats**: K/D, performance, placement
- **Finals results**: Champion, runner-up, 3rd place

---

### Database Schema Additions

#### `tournaments` Table
```sql
CREATE TABLE tournaments (
    tournament_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    format VARCHAR(20) NOT NULL,  -- 'swiss', 'single_elim', 'double_elim'
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    registration_deadline TIMESTAMP,
    max_participants INT,
    status VARCHAR(20) DEFAULT 'registration',  -- 'registration', 'active', 'completed'
    rules JSONB,  -- Match format, map pool, etc.
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### `tournament_participants` Table
```sql
CREATE TABLE tournament_participants (
    participant_id SERIAL PRIMARY KEY,
    tournament_id INT REFERENCES tournaments(tournament_id),
    player_id BIGINT REFERENCES players(player_id),
    team_name VARCHAR(255),  -- Optional for team tournaments
    seed INT,  -- Seeding for brackets
    status VARCHAR(20) DEFAULT 'active',  -- 'active', 'eliminated'
    registered_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tournament_id, player_id)
);
```

#### `tournament_matches` Table
```sql
CREATE TABLE tournament_matches (
    match_id SERIAL PRIMARY KEY,
    tournament_id INT REFERENCES tournaments(tournament_id),
    round_number INT,  -- Swiss round or bracket round
    match_number INT,  -- Position in bracket/round
    
    participant_a_id INT REFERENCES tournament_participants(participant_id),
    participant_b_id INT REFERENCES tournament_participants(participant_id),
    
    server_id INT REFERENCES servers(server_id),  -- Assigned server
    
    score_a INT DEFAULT 0,
    score_b INT DEFAULT 0,
    winner_id INT,  -- References participant_id
    
    status VARCHAR(20) DEFAULT 'pending',  -- 'pending', 'in_progress', 'completed'
    scheduled_time TIMESTAMP,
    completed_at TIMESTAMP,
    
    match_data JSONB  -- Detailed stats, map results, etc.
);
```

#### `tournament_stats` Table
```sql
CREATE TABLE tournament_stats (
    stat_id SERIAL PRIMARY KEY,
    tournament_id INT REFERENCES tournaments(tournament_id),
    player_id BIGINT REFERENCES players(player_id),
    
    matches_played INT DEFAULT 0,
    wins INT DEFAULT 0,
    losses INT DEFAULT 0,
    
    total_kills INT DEFAULT 0,
    total_deaths INT DEFAULT 0,
    kd_ratio DECIMAL(10,4) DEFAULT 0,
    
    -- Tournament-specific performance
    tournament_score DECIMAL(10,2) DEFAULT 0,  -- Custom scoring for placement
    placement INT,  -- Final placement (1st, 2nd, 3rd, etc.)
    
    UNIQUE(tournament_id, player_id)
);
```

#### Update `kill_events` and Other Event Tables
```sql
ALTER TABLE kill_events ADD COLUMN tournament_id INT REFERENCES tournaments(tournament_id);
ALTER TABLE kill_events ADD COLUMN match_id INT REFERENCES tournament_matches(match_id);

-- Same for deflect events, weapon stats, etc.
```

---

## 🏗️ Updated Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  TF2 Game Servers (Dynamic Pool)                            │
│  ├─ Public Servers (stats.udl.tf:27500)                     │
│  ├─ Tournament Servers (stats.udl.tf:27500)                 │
│  │   - Tagged with tournament_id + match_id                 │
│  │   - Receives match assignment from API                   │
│  └─ SuperLogs Plugin (sends UDP + tournament metadata)      │
└────────────┬────────────────────────────────────────────────┘
             │ UDP packets
             ▼
┌─────────────────────────────────────────────────────────────┐
│  Collector Service (Golang)                                 │
│  - Receives UDP events                                      │
│  - Parses events (with tournament_id/match_id if present)   │
│  - Publishes to RabbitMQ via Watermill                      │
└────────────┬────────────────────────────────────────────────┘
             │ Watermill Pub/Sub
             ▼
┌─────────────────────────────────────────────────────────────┐
│  RabbitMQ (AMQP Broker)                                     │
│  - Queues: events.kill, events.deflect, events.match       │
│  - Persistent messages (survives crashes)                   │
│  - Fanout for multiple subscribers                          │
└────────────┬────────────────────────────────────────────────┘
             │ Watermill Subscribe
             ▼
┌─────────────────────────────────────────────────────────────┐
│  Processor Workers (Golang, multiple goroutines)            │
│  - Subscribe to RabbitMQ topics                             │
│  - Calculate MMR, kill weights                              │
│  - Update public stats OR tournament stats                  │
│  - Batch write to PostgreSQL every 30s                      │
└────────────┬────────────────────────────────────────────────┘
             │ Writes
             ▼
┌─────────────────────────────────────────────────────────────┐
│  PostgreSQL 16                                              │
│  - Public ladder stats                                      │
│  - Tournament data (matches, brackets, stats)               │
│  - Materialized views for leaderboards (fast queries)       │
└────────────┬────────────────────────────────────────────────┘
             │ Queries
             ▼
┌─────────────────────────────────────────────────────────────┐
│  REST API (Golang - Gin framework)                          │
│  - Public stats endpoints                                   │
│  - Tournament endpoints:                                    │
│    - POST /api/v1/tournaments (create)                      │
│    - GET /api/v1/tournaments/:id/bracket                    │
│    - POST /api/v1/tournaments/:id/matches/:match_id/result  │
│    - GET /api/v1/tournaments/:id/leaderboard                │
│  - Server assignment API (for match hosting)                │
└────────────┬────────────────────────────────────────────────┘
             │ HTTP/JSON
             ▼
┌─────────────────────────────────────────────────────────────┐
│  Next.js Frontend                                           │
│  - Public leaderboards                                      │
│  - Player profiles                                          │
│  - Tournament pages:                                        │
│    - Bracket visualization                                  │
│    - Swiss standings                                        │
│    - Live match tracker                                     │
│  - Admin panel (tournament creation)                        │
└─────────────────────────────────────────────────────────────┘
```

---

## 📦 Updated Tech Stack

| Component | Technology | Change |
|-----------|-----------|--------|
| Backend | Golang 1.21+ | ✅ Same |
| Database | PostgreSQL 16 | ✅ Confirmed |
| Message Queue | RabbitMQ (AMQP) | 🆕 Replaces Redis |
| Messaging Library | Watermill (Go) | 🆕 New |
| API Framework | Gin | ✅ Same |
| Frontend | Next.js (React) | 🆕 Confirmed |
| TF2 Plugins | SourceMod | ✅ Same |
| Deployment | Docker + Helm | 🆕 Added K8s support |
| Caching | PostgreSQL (materialized views) | 🆕 Replaces Redis cache |

---

## 🎯 Tournament System - Detailed Design

### Workflow: Creating a Tournament

#### Admin Flow
1. **Create Tournament**
   - POST `/api/v1/tournaments`
   ```json
   {
     "name": "UDL Winter Championship 2026",
     "format": "double_elim",
     "start_date": "2026-02-15T18:00:00Z",
     "registration_deadline": "2026-02-10T23:59:59Z",
     "max_participants": 32,
     "rules": {
       "match_format": "bo3",
       "map_pool": ["cp_process", "cp_gullywash", "cp_snakewater"]
     }
   }
   ```

2. **Players Register**
   - POST `/api/v1/tournaments/:id/register`
   - Players sign up (auto-seeded by MMR)

3. **Generate Bracket**
   - POST `/api/v1/tournaments/:id/generate_bracket`
   - System creates `tournament_matches` entries
   - Assigns match numbers, round numbers

4. **Assign Servers**
   - Admin (or auto-scheduler) assigns servers to matches
   - PUT `/api/v1/tournaments/:id/matches/:match_id/assign_server`
   ```json
   {
     "server_id": 5,
     "scheduled_time": "2026-02-15T18:30:00Z"
   }
   ```

5. **Server Receives Assignment**
   - Server polls API: GET `/api/v1/servers/:server_id/current_match`
   - Response:
   ```json
   {
     "match_id": 42,
     "tournament_id": 7,
     "participant_a": "Team Alpha",
     "participant_b": "Team Bravo",
     "map": "cp_process",
     "scheduled_time": "2026-02-15T18:30:00Z"
   }
   ```

6. **Match Starts**
   - Server loads map, sets tournament metadata
   - SuperLogs plugin tags all events with `tournament_id=7, match_id=42`

7. **Events Processed**
   - Collector receives UDP events with tournament tags
   - Processor updates `tournament_stats` (separate from public ladder)
   - MMR changes do **not** apply (or apply to tournament-only rating)

8. **Match Ends**
   - Server sends final score:
   - POST `/api/v1/tournaments/:id/matches/:match_id/result`
   ```json
   {
     "score_a": 3,
     "score_b": 1,
     "winner_id": 101
   }
   ```

9. **Bracket Updates**
   - System advances winner to next round
   - Creates new match entry for next round
   - Assigns server (if scheduled)

10. **Tournament Completes**
    - All matches finished
    - Final placements calculated
    - POST `/api/v1/tournaments/:id/finalize`
    - Awards distributed (if applicable)

---

### Swiss System Specifics

#### Pairing Algorithm
```go
func GenerateSwissRound(tournamentID int, roundNumber int) {
    // 1. Fetch all participants with current W-L records
    participants := GetParticipants(tournamentID)
    
    // 2. Sort by wins (descending), then by MMR (tiebreaker)
    sort.Slice(participants, func(i, j int) bool {
        if participants[i].Wins == participants[j].Wins {
            return participants[i].MMR > participants[j].MMR
        }
        return participants[i].Wins > participants[j].Wins
    })
    
    // 3. Pair adjacent players (avoid repeat pairings)
    for i := 0; i < len(participants); i += 2 {
        if i+1 < len(participants) {
            CreateMatch(tournamentID, roundNumber, participants[i], participants[i+1])
        }
    }
}
```

#### Example Swiss Tournament (8 players, 3 rounds)
```
Round 1:
  Match 1: Player A vs Player B
  Match 2: Player C vs Player D
  Match 3: Player E vs Player F
  Match 4: Player G vs Player H

After Round 1:
  1-0: A, C, E, G
  0-1: B, D, F, H

Round 2 (pair by record):
  Match 5: A vs C  (both 1-0)
  Match 6: E vs G  (both 1-0)
  Match 7: B vs D  (both 0-1)
  Match 8: F vs H  (both 0-1)

... and so on for Round 3
```

---

### Bracket System (Single/Double Elimination)

#### Bracket Generation (Single Elimination, 16 players)
```
Round 1 (8 matches):
  Match 1: Seed 1 vs Seed 16
  Match 2: Seed 8 vs Seed 9
  Match 3: Seed 4 vs Seed 13
  Match 4: Seed 5 vs Seed 12
  Match 5: Seed 2 vs Seed 15
  Match 6: Seed 7 vs Seed 10
  Match 7: Seed 3 vs Seed 14
  Match 8: Seed 6 vs Seed 11

Round 2 (4 matches):
  Match 9: Winner(M1) vs Winner(M2)
  Match 10: Winner(M3) vs Winner(M4)
  Match 11: Winner(M5) vs Winner(M6)
  Match 12: Winner(M7) vs Winner(M8)

Round 3 (2 matches):
  Match 13: Winner(M9) vs Winner(M10)
  Match 14: Winner(M11) vs Winner(M12)

Finals:
  Match 15: Winner(M13) vs Winner(M14)
```

#### Double Elimination (Upper + Lower Bracket)
- **Upper bracket**: Same as single elimination
- **Lower bracket**: Losers drop down, play each other
- **Grand Finals**: Upper bracket winner vs Lower bracket winner

---

### Server Assignment Strategy

#### Option 1: Manual Assignment
- Admin assigns servers to matches via UI
- Good for small tournaments

#### Option 2: Auto-Assignment
- System picks available server from pool
- Checks server availability (not currently hosting another match)
- Prefers servers with low latency to participants

#### Option 3: On-Demand Provisioning
- Spin up new TF2 server containers (Docker/K8s)
- Configure with tournament settings
- Shut down after match completes

**Recommendation**: Start with Option 1 (manual), add Option 2 later, Option 3 for advanced setups.

---

## 🔧 Implementation Changes

### Phase 1 Updates: Replace Redis with RabbitMQ + Watermill

#### Remove Redis Queue (`internal/queue/redis.go`)
- ❌ Delete this file

#### Add Watermill Publisher (`internal/queue/publisher.go`)
```go
package queue

import (
    "github.com/ThreeDotsLabs/watermill"
    "github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
    "github.com/ThreeDotsLabs/watermill/message"
)

type Publisher struct {
    pub message.Publisher
}

func NewPublisher(amqpURI string) (*Publisher, error) {
    amqpConfig := amqp.NewDurableQueueConfig(amqpURI)
    publisher, err := amqp.NewPublisher(amqpConfig, watermill.NewStdLogger(false, false))
    if err != nil {
        return nil, err
    }
    
    return &Publisher{pub: publisher}, nil
}

func (p *Publisher) Publish(topic string, event *Event) error {
    msg := message.NewMessage(watermill.NewUUID(), event.Marshal())
    return p.pub.Publish(topic, msg)
}
```

#### Add Watermill Subscriber (`internal/queue/subscriber.go`)
```go
package queue

import (
    "github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
    "github.com/ThreeDotsLabs/watermill/message"
)

type Subscriber struct {
    sub message.Subscriber
}

func NewSubscriber(amqpURI string) (*Subscriber, error) {
    amqpConfig := amqp.NewDurableQueueConfig(amqpURI)
    subscriber, err := amqp.NewSubscriber(amqpConfig, watermill.NewStdLogger(false, false))
    if err != nil {
        return nil, err
    }
    
    return &Subscriber{sub: subscriber}, nil
}

func (s *Subscriber) Subscribe(topic string, handler func(*Event)) error {
    messages, err := s.sub.Subscribe(context.Background(), topic)
    if err != nil {
        return err
    }
    
    for msg := range messages {
        event := UnmarshalEvent(msg.Payload)
        handler(event)
        msg.Ack()
    }
    
    return nil
}
```

#### Update Collector (`cmd/collector/main.go`)
```go
// Old: Push to Redis
// queue.Push("events:processing", event)

// New: Publish to RabbitMQ via Watermill
publisher.Publish("events.kill", event)  // or events.deflect, events.match
```

#### Update Processor (`cmd/processor/main.go`)
```go
// Old: Pop from Redis queue
// event := queue.Pop("events:processing")

// New: Subscribe to RabbitMQ topics
subscriber.Subscribe("events.kill", func(event *Event) {
    processKillEvent(event, db)
})
```

---

### Phase 2 Updates: Tournament Tables & Logic

#### New Files:
- `internal/tournament/bracket.go` - Bracket generation logic
- `internal/tournament/swiss.go` - Swiss pairing algorithm
- `internal/tournament/manager.go` - Tournament lifecycle management
- `cmd/api/handlers/tournament.go` - Tournament API endpoints

#### Database Migrations:
- `migrations/003_tournament_system.sql` - Add tournament tables

---

### Phase 4 Updates: Next.js Frontend

#### New Pages:
- `/tournaments` - List of active/past tournaments
- `/tournaments/[id]` - Tournament detail page (bracket view)
- `/tournaments/[id]/matches/[matchId]` - Match detail page
- `/admin/tournaments/create` - Admin tournament creation form

#### Components:
- `BracketVisualization.tsx` - SVG bracket renderer
- `SwissStandings.tsx` - Table for Swiss standings
- `MatchCard.tsx` - Live match score display

---

### Phase 5 Updates: Docker + Helm

#### Dockerfiles:
- `Dockerfile.collector`
- `Dockerfile.processor`
- `Dockerfile.api`
- `Dockerfile.frontend` (Next.js)

#### Helm Chart Structure:
```
helm/
├── Chart.yaml
├── values.yaml
└── templates/
    ├── collector-deployment.yaml
    ├── processor-deployment.yaml
    ├── api-deployment.yaml
    ├── frontend-deployment.yaml
    ├── postgres-statefulset.yaml
    ├── rabbitmq-statefulset.yaml
    ├── services.yaml
    └── ingress.yaml
```

#### Example `values.yaml`:
```yaml
collector:
  replicas: 2
  image: unitedstats/collector:latest
  udpPort: 27500

processor:
  replicas: 5
  image: unitedstats/processor:latest

api:
  replicas: 3
  image: unitedstats/api:latest

frontend:
  replicas: 2
  image: unitedstats/frontend:latest

postgres:
  persistence:
    size: 100Gi

rabbitmq:
  persistence:
    size: 20Gi
```

---

## 📊 Updated Implementation Checklist

### NEW Phase 1 Tasks:
- [ ] 1.10: Replace Redis with RabbitMQ + Watermill
  - Install RabbitMQ Docker image
  - Implement Watermill publisher
  - Implement Watermill subscriber
  - Update collector to publish
  - Update processor to subscribe
  - Test message persistence

### NEW Phase 2.5: Tournament System
- [ ] 2.5.1: Database schema (tournament tables)
- [ ] 2.5.2: Bracket generation (single/double elimination)
- [ ] 2.5.3: Swiss pairing algorithm
- [ ] 2.5.4: Tournament manager (lifecycle)
- [ ] 2.5.5: Server assignment logic

### NEW Phase 3 Tasks:
- [ ] 3.6: Tournament API endpoints
  - POST /tournaments (create)
  - POST /tournaments/:id/register
  - POST /tournaments/:id/generate_bracket
  - GET /tournaments/:id/bracket
  - POST /tournaments/:id/matches/:match_id/result
  - GET /tournaments/:id/leaderboard

### NEW Phase 4 Tasks:
- [ ] 4.6: Tournament pages (Next.js)
  - Tournament list page
  - Bracket visualization component
  - Swiss standings component
  - Live match tracker

### NEW Phase 5 Tasks:
- [ ] 5.6: Helm charts for Kubernetes
  - Write deployment manifests
  - Configure persistent volumes
  - Set up ingress (for frontend)
  - Document deployment process

---

## 🎯 Open Questions RESOLVED

1. ✅ **Database**: PostgreSQL confirmed
2. ✅ **Frontend**: Next.js confirmed
3. ✅ **Hosting**: Self-hosted (Docker + Helm)
4. ✅ **MMR Decay**: Yes, inactive players lose MMR
5. ✅ **Privacy**: Everything public
6. ✅ **Message Queue**: RabbitMQ (AMQP) via Watermill

### NEW Open Questions:
1. **MMR Decay Rate**: How much MMR lost per week? (suggestion: -10 to -20)
2. **Tournament MMR**: Separate rating for tournaments, or affect public MMR?
3. **Server Provisioning**: Manual assignment or auto-provisioning?
4. **Match Validation**: Require all players to check-in before match starts?
5. **Bracket Bye Rounds**: How to handle odd number of participants?

---

## 🚀 Updated Timeline

### MVP (6-8 weeks with 3-5 agents):
- ✅ Public ladder stats (Phases 1-3)
- ✅ Basic tournament system (single elimination only)
- ✅ Next.js frontend with bracket view
- ✅ Docker deployment

### Production (10-12 weeks):
- ✅ All tournament formats (Swiss, Single, Double)
- ✅ Kubernetes deployment (Helm charts)
- ✅ Admin panel for tournament management
- ✅ MMR decay system
- ✅ Advanced bracket visualization

---

## 📝 Next Steps for Agents

1. **Update IMPLEMENTATION_CHECKLIST.md** with new tasks
2. **Implement RabbitMQ + Watermill** (Phase 1.10)
3. **Design tournament database schema** (Phase 2.5.1)
4. **Build tournament API** (Phase 3.6)
5. **Create Next.js bracket components** (Phase 4.6)
6. **Write Helm charts** (Phase 5.6)

---

**Document Version**: 3.0  
**Status**: Ready for Implementation  
**Updated**: January 31, 2026

**Major Changes**:
- Removed Redis, added RabbitMQ + Watermill
- Confirmed Next.js frontend
- Added comprehensive tournament system
- Added Helm chart requirements
- Confirmed self-hosted deployment
