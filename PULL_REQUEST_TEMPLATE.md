# Pull Request Template - UnitedStats

## PR Type
<!-- Check the type that applies -->
- [ ] Phase 1: Core Backend
- [ ] Phase 2: SourceMod Plugins
- [ ] Phase 3: REST API
- [ ] Phase 4: Web Interface
- [ ] Phase 5: Production Deployment
- [ ] Bugfix
- [ ] Documentation
- [ ] Other (specify):

---

## Description
<!-- Provide a clear summary of what this PR implements -->

**What does this PR do?**


**Why is this change needed?**


**Related issue(s)**: #(issue number)

---

## Implementation Details

### Files Added/Modified
<!-- List key files with brief descriptions -->

- `path/to/file.go` - Purpose
- `path/to/test.go` - Test coverage
- ...

### Key Functions/Components
<!-- List main functions or components implemented -->

1. **Function/Component Name**
   - Purpose:
   - Input:
   - Output:
   - Example:

---

## Testing

### Unit Tests
<!-- Describe unit tests added -->

- [ ] Tests written for all new functions
- [ ] Tests pass locally (`go test ./...`)
- [ ] Coverage: __%

**Test Results**:
```
# Paste test output here
```

### Integration Tests
<!-- If applicable -->

- [ ] End-to-end flow tested
- [ ] Manual testing completed

**Test Scenario**:


**Expected Behavior**:


**Actual Behavior**:


### Load Testing
<!-- For performance-critical code -->

- [ ] Load tested with X events/second
- [ ] Latency: __ms (target: <10ms)
- [ ] Memory usage: __MB

---

## Code Quality

- [ ] Code follows Golang style guide (`gofmt`, `golint`)
- [ ] All public functions have GoDoc comments
- [ ] Error handling implemented properly
- [ ] No hardcoded values (use config/constants)
- [ ] Logging added for critical paths
- [ ] Code reviewed for performance bottlenecks

---

## Documentation

- [ ] README.md updated (if applicable)
- [ ] API documentation updated (if API changes)
- [ ] Migration guide added (if DB schema changes)
- [ ] Comments added to complex logic

---

## Database Changes

<!-- If this PR modifies the database schema -->

### Migration File
- [ ] Migration file created: `migrations/XXX_description.sql`
- [ ] Rollback script included
- [ ] Migration tested on development database

**Schema Changes**:
```sql
-- Paste migration SQL here
```

---

## Deployment Notes

<!-- Important for production deployment -->

### Environment Variables
<!-- List any new env vars needed -->

```bash
NEW_VAR_NAME=value  # Description
```

### Configuration Changes
<!-- Any config file updates needed -->


### Breaking Changes
<!-- List any breaking changes -->

- [ ] No breaking changes
- [ ] Breaking changes (describe below):


---

## Checklist

### Before Submitting
- [ ] Code compiles without errors (`go build ./...`)
- [ ] All tests pass (`go test ./...`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] Git history is clean (squashed/rebased if needed)
- [ ] Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)

### After Review
- [ ] Feedback addressed
- [ ] Tests updated based on review
- [ ] Documentation updated based on review

---

## Screenshots/Output

<!-- If applicable, add screenshots or command output -->

**Before**:


**After**:


---

## Reviewer Notes

<!-- Any specific areas you want reviewers to focus on -->

**Focus Areas**:
-
-

**Questions for Reviewers**:
-
-

---

## Related PRs

<!-- Link to related PRs if this is part of a larger feature -->

- Depends on: #XX
- Blocks: #XX
- Related: #XX

---

## Performance Impact

<!-- Estimate performance impact -->

- [ ] No performance impact
- [ ] Positive impact (describe):
- [ ] Negative impact (describe + mitigation):

**Benchmarks** (if applicable):
```
# Paste benchmark results
```

---

## Rollback Plan

<!-- How to revert this change if needed -->

**Steps to rollback**:
1.
2.
3.

---

## Post-Deployment Verification

<!-- How to verify this works in production -->

**Verification Steps**:
1.
2.
3.

**Monitoring**:
- Metrics to watch:
- Expected behavior:

---

**Estimated Review Time**: __ hours/days  
**Ready for Merge**: [ ] Yes [ ] No (waiting for: _______)
