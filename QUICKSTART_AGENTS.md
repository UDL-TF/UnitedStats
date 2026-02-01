# Quick Start Guide for Sub-Agents

**Goal**: Get an agent up and running on UnitedStats development in < 5 minutes

---

## 🚀 30-Second Overview

**What**: TF2 MMR ranking system with skill-based kill weighting  
**Why**: Old system failed (database bloat). New system uses UDP streaming + aggregated stats  
**How**: Golang backend + SourceMod plugins + PostgreSQL + Redis  

**Key Innovation**: Killing a 2x stronger opponent = 1.5 kills, killing 2x weaker = 0.5 kills

---

## 📖 Essential Reading (Pick ONE)

**For Backend Work**:
1. Read `PROJECT_BRIEF.md` (sections 1-6) - 15 min
2. Read `IMPLEMENTATION_CHECKLIST.md` (Phase 1) - 5 min

**For SourceMod Plugin Work**:
1. Read `PROJECT_BRIEF.md` (sections 1, 5) - 10 min
2. Read `IMPLEMENTATION_CHECKLIST.md` (Phase 2) - 5 min

**For API/Frontend Work**:
1. Read `PROJECT_BRIEF.md` (sections 1, 7) - 10 min
2. Read `IMPLEMENTATION_CHECKLIST.md` (Phase 3 or 4) - 5 min

**For a Complete Understanding**:
- Read full `UnitedStats_SRD_Draft_v2.md` - 30 min

---

## 🎯 Choose Your Task (Priority Order)

### High Priority (Start Here)
1. **MMR Calculator** (`internal/mmr/calculator.go`)
   - Complexity: Easy
   - Impact: Critical
   - Time: 1-2 hours
   - See: Checklist 1.4

2. **Event Parser** (`internal/parser/parser.go`)
   - Complexity: Medium
   - Impact: Critical
   - Time: 2-3 hours
   - See: Checklist 1.5

3. **UDP Collector** (`cmd/collector/main.go`)
   - Complexity: Medium
   - Impact: Critical
   - Time: 2-4 hours
   - See: Checklist 1.7

### Medium Priority
4. **Event Processor** (`cmd/processor/main.go`)
   - Complexity: Hard
   - Impact: Critical
   - Time: 4-6 hours
   - See: Checklist 1.8

5. **Dodgeball Plugin** (`superlogs-dodgeball.sp`)
   - Complexity: Medium
   - Impact: High
   - Time: 3-5 hours
   - See: Checklist 2.2

6. **REST API** (`cmd/api/main.go`)
   - Complexity: Easy
   - Impact: Medium
   - Time: 2-4 hours
   - See: Checklist 3.1-3.3

### Low Priority (Do Later)
7. **Web Frontend** (Svelte app)
   - Complexity: Medium
   - Impact: Low (can use API directly first)
   - Time: 6-10 hours
   - See: Checklist 4.1-4.5

---

## 🛠️ Quick Setup (Local Development)

### Prerequisites
```bash
# Install Go 1.21+
go version

# Install PostgreSQL 16
psql --version

# Install Redis 7
redis-cli --version

# Install Docker (optional but recommended)
docker --version
```

### Clone & Initialize
```bash
# Clone repo
git clone https://github.com/UDL-TF/UnitedStats.git
cd UnitedStats

# Initialize Go module (if not done)
go mod init github.com/UDL-TF/UnitedStats

# Install dependencies
go mod tidy

# Create project structure
mkdir -p cmd/{collector,processor,api}
mkdir -p internal/{parser,mmr,performance,models,queue}
mkdir -p pkg/events
mkdir -p sourcemod/scripting/include
mkdir -p migrations
mkdir -p test
```

### Start Database (Docker)
```bash
# Start PostgreSQL + Redis
docker-compose up -d postgres redis

# Apply migrations
psql -U stats_user -d unitedstats -f migrations/001_initial_schema.sql
```

### Run Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/mmr/...
```

---

## 📝 Example: Implement MMR Calculator (Task 1.4)

**Step 1: Create file**
```bash
touch internal/mmr/calculator.go
touch internal/mmr/calculator_test.go
```

**Step 2: Implement functions**
```go
// internal/mmr/calculator.go
package mmr

import "math"

func CalculateKillWeight(killerMMR, victimMMR int) float64 {
    if killerMMR == 0 || victimMMR == 0 {
        return 1.0
    }
    ratio := float64(victimMMR) / float64(killerMMR)
    weight := 1.0 + 0.5*math.Log2(ratio)
    
    if weight < 0.5 { return 0.5 }
    if weight > 1.5 { return 1.5 }
    return weight
}

// Add CalculateRankWeight, CalculateRankScore, GetRankTier...
```

**Step 3: Write tests**
```go
// internal/mmr/calculator_test.go
package mmr

import "testing"

func TestCalculateKillWeight(t *testing.T) {
    tests := []struct {
        killerMMR, victimMMR int
        expected             float64
    }{
        {2000, 2000, 1.0},
        {2000, 4000, 1.5},
        {2000, 1000, 0.5},
    }
    
    for _, tt := range tests {
        result := CalculateKillWeight(tt.killerMMR, tt.victimMMR)
        if math.Abs(result-tt.expected) > 0.01 {
            t.Errorf("Got %.2f, want %.2f", result, tt.expected)
        }
    }
}
```

**Step 4: Run tests**
```bash
go test ./internal/mmr/...
```

**Step 5: Create PR**
```bash
git checkout -b feature/mmr-calculator
git add internal/mmr/
git commit -m "feat(mmr): implement kill weight calculation with LOG2 formula"
git push origin feature/mmr-calculator
```

Use `PULL_REQUEST_TEMPLATE.md` to create the PR on GitHub.

---

## 🧪 Testing Your Work

### Unit Tests (Always Required)
```bash
# Run tests for your package
go test ./internal/mmr/...

# With coverage
go test -cover ./internal/mmr/...

# With verbose output
go test -v ./internal/mmr/...
```

### Integration Tests (For End-to-End Features)
```bash
# Run integration tests
go test ./test/...

# Test with Docker services running
docker-compose up -d
go test ./test/...
```

### Manual Testing (For Services)
```bash
# Test UDP collector
# Terminal 1: Start collector
go run cmd/collector/main.go

# Terminal 2: Send test packet
echo "L 01/31/2026 - 12:34:56: Test event" | nc -u localhost 27500

# Check logs for received packet
```

---

## 🐛 Common Issues

### "cannot find package"
```bash
go mod tidy
go mod download
```

### "connection refused" (Redis/Postgres)
```bash
# Start services
docker-compose up -d

# Check status
docker-compose ps
```

### Tests failing
```bash
# Clean test cache
go clean -testcache

# Run with verbose
go test -v ./...
```

---

## 📚 Key Formulas (Quick Reference)

### Kill Weight
```
KillWeight = CLAMP(1 + 0.5 * LOG2(VictimMMR / KillerMMR), 0.5, 1.5)
```

### Rank Weight (Diminishing Returns)
```
RankWeight = 1.0 - (CurrentMMR / 5000) * 0.9
```

### Rank Score
```
RankScore = AccuracyScore * (1 + (K/D_weighted * RankWeight))
MMR = RankScore * 1000
```

### Deflect Score (Dodgeball)
```
DeflectScore = (TimingAccuracy + AngleAccuracy) * (1 + RocketSpeed*0.1) * (1 + Distance*0.1)
```

---

## 🔗 Important Links

- **Full Spec**: `UnitedStats_SRD_Draft_v2.md`
- **Project Brief**: `PROJECT_BRIEF.md`
- **Checklist**: `IMPLEMENTATION_CHECKLIST.md`
- **PR Template**: `PULL_REQUEST_TEMPLATE.md`
- **Repository**: https://github.com/UDL-TF/UnitedStats

---

## 🤝 Workflow

1. **Pick a task** from `IMPLEMENTATION_CHECKLIST.md`
2. **Create feature branch**: `git checkout -b feature/task-name`
3. **Implement with tests**
4. **Run tests**: `go test ./...`
5. **Commit**: `git commit -m "feat(scope): description"`
6. **Push**: `git push origin feature/task-name`
7. **Create PR** using template
8. **Update checklist** (mark task complete)

---

## 💡 Tips for Success

### DO:
✅ Read the relevant section of PROJECT_BRIEF.md first  
✅ Write tests alongside code (not after)  
✅ Use table-driven tests for formulas  
✅ Add comments to complex logic  
✅ Run `go fmt` and `golint` before committing  
✅ Keep PRs focused (one feature per PR)  

### DON'T:
❌ Skip reading the documentation  
❌ Commit code without tests  
❌ Make massive PRs (split into smaller ones)  
❌ Hardcode values (use config/constants)  
❌ Ignore linter warnings  

---

## 🎓 Learning Resources

### Golang
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Testing in Go](https://go.dev/doc/tutorial/add-a-test)

### SourceMod/SourcePawn
- [SourceMod Documentation](https://sm.alliedmods.net/new-api/)
- [SourcePawn Syntax](https://wiki.alliedmods.net/SourcePawn_Syntax)

### PostgreSQL
- [GORM Documentation](https://gorm.io/docs/)
- [PostgreSQL JSON/JSONB](https://www.postgresql.org/docs/current/datatype-json.html)

### Redis
- [go-redis Documentation](https://redis.uptrace.dev/)

---

## 🚨 Get Help

1. **Check documentation** (this file, PROJECT_BRIEF.md, SRD)
2. **Search issues** on GitHub
3. **Ask in PR comments** (tag reviewers)
4. **Create discussion** on GitHub Discussions

---

**Ready to code? Pick a task from IMPLEMENTATION_CHECKLIST.md and ship it!** 🚀

**Estimated Time to First PR**: 2-4 hours (for MMR calculator or parser)

**Good Luck!** 🎯
