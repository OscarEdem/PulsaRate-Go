# Contributing to PulsaRate-Go
<!-- -------------------------------------------------------------------------------------------------------------------                                                                                                                                                                               #*eddiere -->

Thank you for contributing to **PulsaRate-Go**. We welcome contributions from the community to help keep PulsaRate fast, efficient, and robust.

---

## Code of Conduct & Principles

1. **Zero-Allocation Core Engine:** Code path modifications in `pkg/limiter/bucket.go` MUST preserve $0\text{ B/op}$ memory allocations and lock-free execution (`sync/atomic`).
2. **Race Safety:** All pull requests must pass the Go race detector without warnings.
3. **Comprehensive Test Coverage:** New features or bug fixes must include unit tests and, where applicable, benchmark tests.

---

## Local Development Setup

### Prerequisites
* **Go 1.22+**
* **Docker & Docker Compose** (for running Redis during integration tests)
* **git**

### Environment Initialization
```bash
# Clone repository
git clone https://github.com/pulsarate/pulsarate-go.git
cd pulsarate-go

# Download dependencies
go mod download

# Start local Redis container
docker compose up -d redis
```

---

## Testing Guidelines

### 1. Unit Tests & Race Detection
Run all unit tests with the race detector enabled:
```bash
go test -v -race ./pkg/...
```

### 2. Memory & Throughput Benchmarks
Verify that performance hasn't degraded:
```bash
go test -bench=. -benchmem ./pkg/limiter/...
```
*Expected Result: 0 B/op allocations for atomic bucket evaluation.*

### 3. Integration & Chaos Tests
Run end-to-end integration tests against the local Redis instance:
```bash
go test -v ./test/integration/...
```

---

## Coding Style & Formatting

* Run `gofmt -s -w .` and `golangci-lint run` before submitting code.
* Ensure exported structs, functions, and interfaces are well-documented with Go docstrings.
* Maintain clean error handling using Go 1.13+ `fmt.Errorf("...: %w", err)` wrapping.

---

## Pull Request Workflow

1. Fork the repository and create your branch from `main`:
   ```bash
   git checkout -b feature/my-amazing-feature
   ```
2. Commit your changes with a clear, concise commit message.
3. Verify all tests pass locally (`go test -v -race ./...`).
4. Push to your branch and open a Pull Request against `main`.
