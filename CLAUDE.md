# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- **Build**: `task build` (builds binary to `backend/bin/smart-alert`)
- **Run**: `task run` or `go run backend/cmd/server/main.go`
- **Dev**: `task dev` (runs with air for hot reload)
- **Test**: `task test` (all tests) or `task test:short` (unit tests only)
  - Single test: `go test -v ./path/to/pkg -run TestName`
- **Lint**: `task lint` (golangci-lint)
- **Format**: `task fmt` (go fmt)
- **Dependencies**: `task deps` or `task tidy`
- **Database**:
  - Start DB: `task db:up`
  - Stop DB: `task db:down`
  - Migrations Up: `task goose:up`
  - Migrations Status: `task goose:status`
  - Create Migration: `task goose:create -- migration_name`

## Architecture

This is a Go backend service following Clean Architecture principles:

- **Entry Point**: `backend/cmd/server/main.go`
- **Core Business Logic**: `backend/internal/domain/`
  - `models/`: Domain entities and interfaces
  - `repository/`: Data access interfaces
  - `service/`: Business logic implementation
- **Adapters/Infrastructure**:
  - `backend/internal/handler/`: HTTP handlers (Echo framework)
  - `backend/internal/scheduler/`: Cron jobs
- **Configuration**: `backend/pkg/config/` and `backend/configs/`
- **Database**: PostgreSQL managed via Docker Compose and Goose migrations

## Style & Patterns

- **Framework**: Echo v4 for HTTP server
- **Database Access**: `database/sql` with PostgreSQL
- **Configuration**: Environment variables loaded via `godotenv`
- **Project Root**: All Go commands should be run relative to `backend/` directory or using `task` from root if configured, but currently `Taskfile.yml` is in `backend/`.
  - *Note*: The `Taskfile.yml` is located in `backend/`, so `task` commands generally need to be run from `backend/` directory or using `task -d backend`.

## Environment Setup

- Requires Go 1.25+ and Docker
- Uses `Task` (modern make) for command automation
- Uses `goose` for database migrations
