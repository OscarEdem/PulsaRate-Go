## Description

<!-- Briefly describe what this PR does and why. Link to a related issue if applicable. -->

Fixes # (issue)

---

## Type of Change

<!-- Check all that apply -->

- [ ] 🐛 Bug fix (non-breaking change that fixes an issue)
- [ ] ✨ New feature (non-breaking change that adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 🏎️ Performance improvement
- [ ] 📝 Documentation update
- [ ] 🔧 Refactor / code quality improvement

---

## Zero-Allocation Invariant Checklist

> Any modification to `pkg/limiter/bucket.go` or the core decision path MUST preserve the following. Check each item:

- [ ] **0 B/op allocations** — `go test -bench=BenchmarkAtomicBucket_Allow -benchmem ./pkg/limiter/...` shows `0 B/op`
- [ ] **No mutex/lock in hot path** — core evaluation uses only `sync/atomic` operations
- [ ] **Race detector clean** — `go test -v -race ./...` passes with zero race warnings

---

## Testing Checklist

- [ ] Unit tests added / updated for changed behavior
- [ ] `go test -v -race ./...` passes locally
- [ ] `go test -bench=. -benchmem ./pkg/limiter/...` shows no regression vs. baseline
- [ ] Integration tests pass (`go test -v ./test/integration/...`) if Redis-layer code was modified
- [ ] `golangci-lint run` produces no new warnings

---

## Documentation

- [ ] Code is self-documenting with Go docstrings for all exported symbols
- [ ] `README.md` updated (if user-facing behavior or API changed)
- [ ] `CHANGELOG.md` entry added under `[Unreleased]`
- [ ] `SYSTEM_DESIGN.md` or `ARCHITECTURE_DECISIONS.md` updated (if architecture changed)

---

## Additional Notes

<!-- Any context reviewers should know: edge cases considered, alternatives rejected, trade-offs made. -->
