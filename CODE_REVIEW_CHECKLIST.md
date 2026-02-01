# Code Review Checklist for PR #3

## 🔍 Review Areas

### 1. Architecture & Design
- [ ] Microservices separation is clean (collector/processor/api)
- [ ] Event-driven architecture makes sense for the use case
- [ ] Message queue usage is appropriate
- [ ] Database schema is well-designed

### 2. MMR System
- [ ] **Formula correctness**: Review `internal/mmr/mmr.go`
  - Elo formula implementation
  - Team size adjustment (1/√teamSize)
  - Experience-based K-factor logic
- [ ] **Test coverage**: Check `internal/mmr/mmr_test.go`
  - Are edge cases covered?
  - Do the test assertions make sense?
- [ ] **Integration**: Review `internal/processor/mmr.go`
  - Match-end processing logic
  - Database update order
  - Error handling

### 3. Database Layer (`internal/store/store.go`)
- [ ] SQL queries are safe (no injection vulnerabilities)
- [ ] Indexes match query patterns
- [ ] Transactions where needed
- [ ] Connection pooling configured properly
- [ ] NULL handling is correct

### 4. API (`internal/api/api.go`)
- [ ] All endpoints have proper error handling
- [ ] Input validation is present
- [ ] Response formats are consistent
- [ ] Pagination limits are reasonable
- [ ] No sensitive data leaking

### 5. Collector (`internal/collector/collector.go`)
- [ ] UDP packet handling is robust
- [ ] JSON parsing has error handling
- [ ] No blocking operations
- [ ] Graceful shutdown works

### 6. Processor (`internal/processor/processor.go`)
- [ ] Message acknowledgment is correct
- [ ] Event parsing handles all types
- [ ] Database errors don't crash the service
- [ ] Player creation logic is idempotent

### 7. Docker & Deployment
- [ ] `docker-compose.yml` is production-ready
- [ ] Dockerfiles use multi-stage builds
- [ ] Health checks are configured
- [ ] Volumes for persistence
- [ ] Environment variables documented

### 8. Documentation
- [ ] README is clear and complete
- [ ] DEPLOYMENT.md has accurate instructions
- [ ] MMR_SYSTEM.md explains the system well
- [ ] Code comments are helpful

### 9. Testing
- [ ] Run MMR tests: `cd internal/mmr && go test -v`
- [ ] Run parser tests: `cd internal/parser && go test -v`
- [ ] Check test coverage is reasonable

### 10. Performance Concerns
- [ ] No N+1 query problems
- [ ] Batch operations where possible
- [ ] Connection pooling limits are set
- [ ] Materialized view refresh strategy

---

## 🧪 Manual Testing Checklist

### Start Services
```bash
cd unitedstats
docker-compose up -d
docker-compose ps  # All services "Up"?
```

### Test Collector
```bash
# Send a test kill event
echo '{"event_type":"kill","timestamp":"2024-02-01T10:00:00Z","gamemode":"default","server_ip":"127.0.0.1","killer":{"steam_id":"76561198012345678","name":"TestPlayer1","team":2},"victim":{"steam_id":"76561198087654321","name":"TestPlayer2","team":3},"weapon":{"name":"rocketlauncher"},"crit":false,"airborne":true}' | nc -u localhost 27500

# Check RabbitMQ
open http://localhost:15672  # See message in queue?
```

### Test Database
```bash
docker exec -it unitedstats-postgres psql -U unitedstats

# Check event was stored
SELECT * FROM events ORDER BY created_at DESC LIMIT 1;

# Check player was created
SELECT * FROM players ORDER BY created_at DESC LIMIT 2;

# Check kill was recorded
SELECT * FROM kills ORDER BY timestamp DESC LIMIT 1;
```

### Test API
```bash
# Health check
curl http://localhost:8080/health

# Leaderboard
curl http://localhost:8080/api/v1/leaderboard

# Player stats
curl http://localhost:8080/api/v1/players/76561198012345678

# Recent matches
curl http://localhost:8080/api/v1/matches
```

### Test MMR Calculation
```bash
# Send match_start event
echo '{"event_type":"match_start","timestamp":"2024-02-01T10:00:00Z","gamemode":"default","server_ip":"127.0.0.1","map":"cp_badlands"}' | nc -u localhost 27500

# Send some kills...

# Send match_end event
echo '{"event_type":"match_end","timestamp":"2024-02-01T10:30:00Z","gamemode":"default","server_ip":"127.0.0.1","winner_team":2}' | nc -u localhost 27500

# Check MMR was updated
docker exec -it unitedstats-postgres psql -U unitedstats -c "SELECT name, mmr, peak_mmr FROM players ORDER BY id DESC LIMIT 5;"

# Check match_players has MMR changes
docker exec -it unitedstats-postgres psql -U unitedstats -c "SELECT mp.mmr_before, mp.mmr_after, mp.mmr_change, p.name FROM match_players mp JOIN players p ON mp.player_id = p.id ORDER BY mp.id DESC LIMIT 10;"
```

---

## ⚠️ Potential Issues to Check

### Code Quality
- [ ] Are there any TODOs left?
- [ ] Is error handling consistent?
- [ ] Are there any hardcoded values that should be config?
- [ ] Is logging adequate for debugging?

### Security
- [ ] SQL injection prevention (using parameterized queries?)
- [ ] Input validation on API endpoints
- [ ] No secrets hardcoded
- [ ] CORS not too permissive

### Performance
- [ ] Database query optimization
- [ ] Index usage
- [ ] Connection pool sizing
- [ ] Memory leaks (goroutine leaks?)

### Reliability
- [ ] Graceful shutdown
- [ ] Message acknowledgment
- [ ] Database transaction handling
- [ ] Error recovery

---

## 🎯 Merge Criteria

Before merging, ensure:

1. ✅ All tests pass
2. ✅ Docker Compose starts successfully
3. ✅ Manual testing shows basic functionality works
4. ✅ Documentation is clear
5. ✅ No obvious security issues
6. ✅ Code quality is acceptable
7. ✅ MMR calculations are mathematically sound

---

## 🚀 After Merge

1. Build Docker images
2. Tag release version
3. Deploy to staging environment
4. Monitor for issues
5. Tune MMR K-factor based on real data
6. Plan frontend development

---

**PR Link**: https://github.com/UDL-TF/UnitedStats/pull/3
