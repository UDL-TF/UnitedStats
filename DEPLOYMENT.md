# 🚀 UnitedStats Collector & API - Quick Start

This document explains how to run the complete UnitedStats backend system.

## 📋 Prerequisites

- **Docker** and **Docker Compose** installed
- At least **2GB RAM** available
- Ports available: `5432` (PostgreSQL), `5672` (RabbitMQ), `8080` (API), `27500/udp` (Collector)

## 🏗️ Architecture

```
SourceMod Plugin (TF2 Server)
       ↓ UDP (JSON events)
   COLLECTOR
       ↓ RabbitMQ (AMQP)
   PROCESSOR (×2 replicas)
       ↓ PostgreSQL
   REST API
       ↓ HTTP/JSON
   Frontend / Stats Sites
```

## 🎯 Quick Start

### 1. Start all services

```bash
docker-compose up -d
```

This starts:
- **PostgreSQL** (port 5432) - Database with schema auto-loaded
- **RabbitMQ** (port 5672 + 15672 for management UI) - Message queue
- **Collector** (UDP port 27500) - Receives events from game servers
- **Processor** (×2 replicas) - Processes events and stores to DB
- **API** (port 8080) - REST API for querying data

### 2. Verify services are running

```bash
docker-compose ps
```

All services should show as "Up".

### 3. Check logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f collector
docker-compose logs -f processor
docker-compose logs -f api
```

### 4. Test the API

```bash
# Health check
curl http://localhost:8080/health

# Leaderboard
curl http://localhost:8080/api/v1/leaderboard

# Player stats
curl http://localhost:8080/api/v1/players/76561198012345678
```

### 5. View RabbitMQ Management UI

Open http://localhost:15672 in your browser.

- **Username**: `guest`
- **Password**: `guest`

You can see queues, message rates, and consumer connections.

## 🎮 Configure TF2 Server

### SourceMod Plugin Configuration

Edit `cfg/sourcemod/superlogs-tf2.cfg`:

```
// Collector settings
sm_superlogs_host "your.server.ip"  // IP where collector is running
sm_superlogs_port "27500"
sm_superlogs_gamemode "default"      // or "dodgeball"

// Feature toggles (all default to 1)
sm_superlogs_actions 1
sm_superlogs_teleports 1
sm_superlogs_headshots 1
sm_superlogs_backstabs 1
sm_superlogs_airshots 1
sm_superlogs_jumps 1
sm_superlogs_buildings 1
sm_superlogs_healing 1
sm_superlogs_weaponstats 1
sm_superlogs_loadouts 1
```

### Plugin Files

Copy to your TF2 server:

```
addons/sourcemod/plugins/
  ├── superlogs-tf2.smx        # Main TF2 plugin
  └── superlogs-default.smx    # Fallback/minimal plugin

addons/sourcemod/scripting/include/
  ├── superlogs-core.inc       # Core logging functionality
  ├── json.inc                 # JSON encoding
  └── socket.inc               # UDP sockets
```

## 📊 API Endpoints

### Players

```
GET /api/v1/players/:steam_id
GET /api/v1/players/:steam_id/stats
GET /api/v1/players/:steam_id/matches
```

### Leaderboard

```
GET /api/v1/leaderboard?limit=100&offset=0
```

### Matches

```
GET /api/v1/matches
GET /api/v1/matches/:id
GET /api/v1/matches/:id/events
```

### Stats

```
GET /api/v1/stats/overview
GET /api/v1/stats/weapons
```

## 🗄️ Database Access

### Connect to PostgreSQL

```bash
docker exec -it unitedstats-postgres psql -U unitedstats -d unitedstats
```

### Useful queries

```sql
-- Top players by MMR
SELECT name, steam_id, mmr, total_kills, total_deaths 
FROM players 
ORDER BY mmr DESC 
LIMIT 10;

-- Recent kills
SELECT k.timestamp, pk.name as killer, pv.name as victim, k.weapon
FROM kills k
JOIN players pk ON k.killer_id = pk.id
JOIN players pv ON k.victim_id = pv.id
ORDER BY k.timestamp DESC
LIMIT 20;

-- Event counts by type
SELECT event_type, COUNT(*) as count
FROM events
GROUP BY event_type
ORDER BY count DESC;

-- Refresh leaderboard materialized view
SELECT refresh_leaderboard();
```

## 🛠️ Development

### Build from source

```bash
# Install dependencies
go mod download

# Build all services
go build -o bin/collector ./cmd/collector
go build -o bin/processor ./cmd/processor
go build -o bin/api ./cmd/api

# Run locally (requires PostgreSQL and RabbitMQ running)
export DB_HOST=localhost
export DB_USER=unitedstats
export DB_PASSWORD=unitedstats
export DB_NAME=unitedstats
export RABBITMQ_URL=amqp://guest:guest@localhost:5672/

./bin/collector   # Terminal 1
./bin/processor   # Terminal 2
./bin/api         # Terminal 3
```

### Run tests

```bash
# Parser tests
go test ./internal/parser -v

# Integration tests (requires running services)
go test ./internal/store -v
```

## 🔧 Configuration

### Environment Variables

#### Collector
- `UDP_PORT` - UDP listen port (default: 27500)
- `RABBITMQ_URL` - RabbitMQ connection URL

#### Processor
- `RABBITMQ_URL` - RabbitMQ connection URL
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - PostgreSQL connection

#### API
- `API_PORT` - HTTP listen port (default: 8080)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - PostgreSQL connection

## 📦 Docker Commands

### Stop all services
```bash
docker-compose down
```

### Stop and remove volumes (⚠️ deletes all data)
```bash
docker-compose down -v
```

### Rebuild after code changes
```bash
docker-compose build
docker-compose up -d
```

### Scale processor instances
```bash
docker-compose up -d --scale processor=5
```

### View resource usage
```bash
docker stats
```

## 🐛 Troubleshooting

### Collector not receiving events

1. Check if UDP port is open:
   ```bash
   nc -u -l 27500
   ```

2. Test sending a manual event:
   ```bash
   echo '{"event_type":"kill","timestamp":"2024-01-31T22:00:00Z","gamemode":"default","server_ip":"127.0.0.1","killer":{"steam_id":"76561198012345678","name":"Player1","team":2},"victim":{"steam_id":"76561198087654321","name":"Player2","team":3},"weapon":{"name":"rocketlauncher"},"crit":false,"airborne":true}' | nc -u localhost 27500
   ```

3. Check collector logs:
   ```bash
   docker-compose logs collector
   ```

### Processor not processing events

1. Check RabbitMQ queues:
   - Open http://localhost:15672
   - Check if messages are piling up

2. Check processor logs:
   ```bash
   docker-compose logs processor
   ```

3. Check database connectivity:
   ```bash
   docker exec -it unitedstats-processor sh
   ping postgres
   ```

### API not responding

1. Check if API is running:
   ```bash
   docker-compose ps api
   ```

2. Check API logs:
   ```bash
   docker-compose logs api
   ```

3. Test database connection:
   ```bash
   curl http://localhost:8080/health
   ```

## 📈 Performance Tuning

### Scale processor instances
```bash
docker-compose up -d --scale processor=10
```

### PostgreSQL tuning

Edit `docker-compose.yml` to add PostgreSQL performance settings:

```yaml
postgres:
  command: >
    postgres
    -c shared_buffers=256MB
    -c max_connections=200
    -c effective_cache_size=1GB
```

### RabbitMQ tuning

For high-throughput environments:

```yaml
rabbitmq:
  environment:
    RABBITMQ_VM_MEMORY_HIGH_WATERMARK: 1GB
```

## 📝 License

MIT License - See LICENSE file

## 🤝 Contributing

See CONTRIBUTING.md for guidelines.

## 📞 Support

- GitHub Issues: https://github.com/UDL-TF/UnitedStats/issues
- Discord: [Your Discord]
