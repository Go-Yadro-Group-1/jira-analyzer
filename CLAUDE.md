# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o jira-analyzer ./cmd

# Run all tests with coverage
go test -v -cover ./...

# Run a single test
go test -v -run TestName ./path/to/package/...

# Run integration tests (requires Docker for testcontainers)
go test -tags integration -v -cover ./...

# Lint (requires golangci-lint v2.1+)
golangci-lint run
```

The CI pipeline runs: build → lint → unit tests → integration tests (sequential). Build tags: `integration`, `unit`.

## Architecture

**Jira Analyzer** — gRPC-сервер для аналитики Jira-проектов с CLI-обёрткой (Cobra).

### Layers

```
cmd/                        CLI (Cobra) + gRPC server bootstrap
  internal/cli/server/      server startup, config loading, DB connect
  internal/config/          Config struct (Viper + validator), GetConfig() singleton
internal/service/           Business logic: Service orchestrates Repository + Cache
  histogram.go              Duration/status histogram builders with auto unit selection
  model.go                  Domain models (histograms, charts)
internal/repository/        Interfaces (Repository, CacheRepository)
  postgres/                 PostgreSQL impl; SQL in embedded queries/ dir
  memory/                   Generic in-memory cache (sync.RWMutex)
migrations/                 golang-migrate up/down SQL files (raw + analytics schemas)
```

### Key design decisions

- **Interface-driven DI**: `Service` depends on `Repository` and `Cache` interfaces (defined in `internal/service/service.go`), implementations injected at startup
- **Generic cache**: `CacheRepository[K1, K2]` uses type parameters; staleness detected by comparing DB vs cache timestamps
- **Embedded SQL**: queries loaded via `//go:embed queries` at compile time (`internal/repository/postgres/`)
- **Integration tests**: use `testcontainers-go` to spin up PostgreSQL; gated behind `//go:build integration`
- **Custom errors**: `ProjectQueryError` with `Unwrap()` for error chain inspection
