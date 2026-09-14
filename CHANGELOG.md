# Changelog
<!-- -------------------------------------------------------------------------------------------------------------------                                                                                                                                                                               #*eddiere -->

All notable changes to **PulsaRate-Go** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

> Changes staged for the next release.

### Added
- GitHub Actions CI workflow (`ci.yml`) — automated race detection, linting, and benchmark regression
- Pull Request template with zero-allocation invariant checklist
- Issue templates for bug reports and feature requests
- `CODE_OF_CONDUCT.md` — Contributor Covenant 2.1
- `SECURITY.md` — coordinated vulnerability disclosure policy
- `SUPPORT.md` — community support channels

---

## [v0.1.0] — 2026-09-14

### Added
- **Core atomic token bucket engine** (`pkg/limiter/bucket.go`)
  - Lock-free `sync/atomic` operations with **0 B/op** memory allocations
  - ~29ns local CPU evaluation latency on AMD EPYC 7763 @ 2.45GHz
- **Lazy-batched Redis lease worker** (`pkg/limiter/`)
  - Async background goroutine pool for Redis `EVAL` lease requests
  - Configurable `BatchSize` (tokens per lease) and `LeaseTTL`
  - Pipelined batch requests — achieves **94%+ reduction** in Redis RPS
- **Circuit breaker & graceful degradation**
  - Automatic fallback to local sliding-window limits when Redis latency exceeds 50ms
  - Configurable `CircuitOpenDuration` and `LatencyThreshold`
- **Adaptive PID controller**
  - Dynamically scales `BatchSize` and host backpressure based on p99 tail latency
- **Middleware integrations**
  - `middleware.GinRateLimit(engine)` — Gin Web Framework middleware
  - `middleware.RateLimit(engine)` — Standard Go `net/http` handler wrapper
  - Envoy gRPC `ExtAuthz` sidecar support
- **Observability**
  - Prometheus metrics: `rate_limit_allowed_total`, `rate_limit_rejected_total`, `redis_lease_duration_seconds`
  - OpenTelemetry distributed tracing spans
- **Documentation**
  - `README.md` — Quick start, architecture diagram, performance benchmarks
  - `SYSTEM_DESIGN.md` — Technical blueprint, Lua script specs, sequence diagrams
  - `USE_CASES.md` — Enterprise production use cases (API gateway, fintech, gaming)
  - `DEPLOYMENT_GUIDE.md` — AWS EC2 & GCP production deployment guide
  - `ARCHITECTURE_DECISIONS.md` — Architectural Decision Records (ADRs)
  - `CONTRIBUTING.md` — Contribution guidelines, testing requirements

### Performance Benchmarks (v0.1.0)

| Metric | Value |
| :--- | :--- |
| Local decision latency | ~29.08 ns/op |
| Memory allocations | **0 B/op** |
| Throughput | 52,800,000+ RPS (local) |
| Redis RPS (vs central EVAL) | 450 req/s (**94% reduction**) |

---

## Legend

| Symbol | Meaning |
| :--- | :--- |
| **Added** | New features |
| **Changed** | Changes in existing functionality |
| **Deprecated** | Soon-to-be removed features |
| **Removed** | Removed features |
| **Fixed** | Bug fixes |
| **Security** | Vulnerability fixes |

[Unreleased]: https://github.com/aeroforge-labs/PulsaRate-Go/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/aeroforge-labs/PulsaRate-Go/releases/tag/v0.1.0
