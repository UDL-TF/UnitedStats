# CI/CD Documentation

## Overview

UnitedStats uses GitHub Actions for continuous integration and deployment. The CI/CD pipeline automatically tests code, builds Docker images, and releases new versions.

---

## 🔄 Workflows

### 1. CI Workflow (`ci.yml`)

**Triggers:**
- Push to `main` or `develop`
- Pull requests to `main` or `develop`
- Manual dispatch

**Jobs:**

#### Lint
- Runs `golangci-lint` with 20+ linters
- Checks code quality, security, and style
- Timeout: 5 minutes

#### Test
- Runs all tests with race detector
- Spins up PostgreSQL and RabbitMQ services
- Generates coverage report
- Uploads to Codecov

#### Build
- Compiles binaries for all 3 services
- Tests both amd64 and arm64 architectures
- Uploads artifacts for verification

#### Docker Build
- Tests Docker image builds
- No push (verification only)
- Uses BuildKit cache

**Runtime:** ~5 minutes

---

### 2. Service Workflows (collector.yml, processor.yml, api.yml)

**Triggers:**
- Push to `main` (with path filters)
- Pull requests (with path filters)
- Manual dispatch

**Path Filters:**
Each workflow only runs when relevant code changes:

```yaml
# Collector
- cmd/collector/**
- internal/collector/**
- internal/queue/**

# Processor  
- cmd/processor/**
- internal/processor/**
- internal/store/**
- internal/parser/**
- internal/mmr/**

# API
- cmd/api/**
- internal/api/**
- internal/store/**
```

**Jobs:**

#### Test
- Runs service-specific tests
- PostgreSQL service for database tests
- Coverage uploaded to Codecov with flags

#### Build (main branch only)
- Multi-platform Docker build (amd64, arm64)
- Pushes to `ghcr.io/udl-tf/unitedstats/{service}`
- Tags: `main`, `main-<sha>`, `latest`
- Build attestation for supply chain security

**Runtime:** ~3-4 minutes per service

---

### 3. Release Workflow (`release.yml`)

**Triggers:**
- Tag push matching `v*.*.*` (e.g., `v1.0.0`)
- Manual dispatch with version input

**Jobs:**

#### Create Release
- Generates changelog (from CHANGELOG.md if exists)
- Creates GitHub release
- Attaches release notes

#### Build and Push Images
- Builds all 3 services (matrix strategy)
- Multi-platform: linux/amd64, linux/arm64
- Pushes to GitHub Container Registry
- Tags:
  - `v1.2.3` (exact version)
  - `v1.2` (minor version)
  - `v1` (major version)
  - `latest` (latest release)
- Build attestation attached

#### Build Binaries
- Cross-compiles for multiple platforms:
  - Linux: amd64, arm64
  - macOS: amd64, arm64
  - Windows: amd64
- Creates archives (.tar.gz for unix, .zip for windows)
- Uploads to GitHub release

**Runtime:** ~15 minutes

---

## 📦 Docker Images

### Registries

**GitHub Container Registry (ghcr.io):**
```
ghcr.io/udl-tf/unitedstats/collector
ghcr.io/udl-tf/unitedstats/processor
ghcr.io/udl-tf/unitedstats/api
```

### Tags

**Main branch builds:**
- `main` - Latest from main branch
- `main-abc1234` - Specific commit SHA
- `latest` - Alias for main

**Release builds:**
- `v1.2.3` - Exact semver
- `v1.2` - Minor version (receives patches)
- `v1` - Major version (receives minor updates)
- `latest` - Latest stable release

### Pull Images

```bash
# Latest development (main branch)
docker pull ghcr.io/udl-tf/unitedstats/collector:latest

# Specific version
docker pull ghcr.io/udl-tf/unitedstats/collector:v1.0.0

# Specific commit
docker pull ghcr.io/udl-tf/unitedstats/collector:main-abc1234
```

### Multi-Platform Support

All images support:
- `linux/amd64` (x86_64)
- `linux/arm64` (ARM 64-bit, e.g., Apple Silicon, AWS Graviton)

Docker automatically pulls the correct architecture.

---

## 🧪 Testing

### Running Tests Locally

```bash
# All tests
go test -v ./...

# With race detector
go test -v -race ./...

# With coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Specific package
go test -v ./internal/mmr/...
```

### Integration Tests

Some tests require services:

```bash
# Start services
docker-compose up -d postgres rabbitmq

# Run tests with environment
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=unitedstats
export DB_PASSWORD=unitedstats
export DB_NAME=unitedstats
export RABBITMQ_URL=amqp://guest:guest@localhost:5672/

go test -v ./internal/store/...
```

---

## 🔍 Code Quality

### Linters

`golangci-lint` runs with these linters:

**Bugs & Errors:**
- `errcheck` - Check error handling
- `gosec` - Security issues
- `staticcheck` - Static analysis

**Code Quality:**
- `gocyclo` - Cyclomatic complexity
- `gocritic` - Code review suggestions
- `dupl` - Code duplication

**Style:**
- `gofmt` - Formatting
- `goimports` - Import organization
- `misspell` - Spelling mistakes
- `whitespace` - Whitespace issues

### Running Locally

```bash
# Install golangci-lint
brew install golangci-lint  # macOS
# or download from https://golangci-lint.run/

# Run all linters
golangci-lint run

# Auto-fix issues
golangci-lint run --fix

# Run specific linters
golangci-lint run --enable=gosec
```

---

## 🚀 Releasing

### Creating a Release

**Via Git Tag:**

```bash
# Create and push tag
git tag v1.0.0
git push origin v1.0.0

# Release workflow triggers automatically
```

**Via GitHub CLI:**

```bash
gh release create v1.0.0 --generate-notes
```

**Via GitHub Web:**

1. Go to "Releases" → "Draft a new release"
2. Choose tag: `v1.0.0` (create new)
3. Generate release notes
4. Publish release

### Versioning

Follow [Semantic Versioning](https://semver.org/):

- `v1.0.0` - Major release (breaking changes)
- `v1.1.0` - Minor release (new features, backwards compatible)
- `v1.1.1` - Patch release (bug fixes)

### Release Checklist

- [ ] Update CHANGELOG.md
- [ ] Update version in code (if hardcoded)
- [ ] Run tests locally: `go test ./...`
- [ ] Build Docker images locally: `docker-compose build`
- [ ] Create Git tag: `git tag v1.0.0`
- [ ] Push tag: `git push origin v1.0.0`
- [ ] Monitor GitHub Actions for build status
- [ ] Verify Docker images were pushed
- [ ] Verify GitHub release was created
- [ ] Test deployed images

---

## 🔐 Security

### Supply Chain Security

**Build Attestation:**
- All Docker images include build provenance
- Verifies image was built by GitHub Actions
- Provides transparency into build process

**Verify an image:**

```bash
# Install cosign
brew install cosign

# Verify attestation
cosign verify-attestation \
  --type slsaprovenance \
  ghcr.io/udl-tf/unitedstats/collector:v1.0.0
```

### Dependency Management

**Dependabot:**
- Automatically checks for dependency updates
- Weekly schedule (Mondays)
- Creates PRs for:
  - Go module updates
  - GitHub Action updates
  - Docker base image updates

**Review Process:**
1. Dependabot creates PR
2. CI runs automatically
3. Review changes and test results
4. Merge if tests pass

---

## 📊 Coverage

### Codecov Integration

Coverage reports are uploaded to [Codecov](https://codecov.io/gh/UDL-TF/UnitedStats).

**Flags:**
- `unittests` - Overall coverage (CI workflow)
- `collector` - Collector service coverage
- `processor` - Processor service coverage
- `api` - API service coverage

**View Coverage:**

```bash
# Badge in README.md shows overall coverage
# Click badge to view detailed report on Codecov

# Local HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 🐛 Troubleshooting

### Workflow Failed

**Check logs:**
1. Go to "Actions" tab on GitHub
2. Click the failed workflow
3. Click the failed job
4. Expand the failed step
5. Review error messages

**Common issues:**

**Tests failing:**
- Check if code changes broke existing tests
- Review error messages in test output
- Run tests locally to reproduce

**Docker build failing:**
- Check Dockerfile syntax
- Verify all dependencies are available
- Check for base image issues

**Lint failing:**
- Run `golangci-lint run` locally
- Fix reported issues
- Push fix and re-run workflow

### Manual Workflow Trigger

**Via GitHub CLI:**

```bash
# Trigger CI
gh workflow run ci.yml

# Trigger release with version
gh workflow run release.yml -f version=v1.0.0
```

**Via GitHub Web:**

1. Go to "Actions" tab
2. Select workflow
3. Click "Run workflow"
4. Fill in inputs (if required)
5. Click "Run workflow"

---

## 📝 Best Practices

### Pull Requests

1. **Create feature branch** from `main`
2. **Make changes** and commit
3. **Push branch** and open PR
4. **CI runs automatically** on PR
5. **Review CI results** (must pass)
6. **Request code review**
7. **Merge** when approved and CI passes

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add MMR calculation
fix: correct database connection leak
docs: update API documentation
ci: add Docker build caching
test: add MMR edge case tests
refactor: simplify event parser
```

### Branch Protection

**Recommended settings:**

- [ ] Require pull request reviews (1 approver)
- [ ] Require status checks to pass (CI)
- [ ] Require branches to be up to date
- [ ] Require conversation resolution
- [ ] Do not allow bypassing

---

## 🔄 Continuous Deployment (Future)

### Staging Environment

**Auto-deploy from `develop` branch:**

```yaml
# .github/workflows/deploy-staging.yml
on:
  push:
    branches: [develop]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to staging
        run: |
          # Deploy logic here
          # kubectl apply, terraform, etc.
```

### Production Environment

**Auto-deploy from version tags:**

```yaml
# .github/workflows/deploy-production.yml
on:
  push:
    tags: ['v*.*.*']

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: production
    steps:
      - name: Deploy to production
        run: |
          # Deploy logic here
```

---

## 📚 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker BuildKit](https://docs.docker.com/build/buildkit/)
- [golangci-lint Linters](https://golangci-lint.run/usage/linters/)
- [Dependabot Configuration](https://docs.github.com/en/code-security/dependabot)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)

---

**CI/CD is ready to use! Every push and release is automatically tested and built. 🚀**
