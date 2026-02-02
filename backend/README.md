# Smart Alert - Go Backend

A long-running Go service for automated investment alerts based on value averaging and market valuation strategies.

## Prerequisites

- Go 1.21+
- Docker & Docker Compose
- [Task](https://taskfile.dev/) (modern Makefile alternative)
- [goose](https://github.com/pressly/goose) (for database migrations)

## Getting Started

### 1. Start PostgreSQL

```bash
task db:up
```

### 2. Run Database Migrations

```bash
# Install goose if not already installed
task goose:install

# Run all pending migrations
task goose:up

# Check migration status
task goose:status
```

### 3. Run the Application

```bash
# Development mode with hot reload
task dev

# Or run directly
task run
```

## Project Structure

```
backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── app/            # App initialization and lifecycle
│   ├── domain/         # Business logic
│   │   ├── models/     # Domain models
│   │   ├── repository/ # Data access layer
│   │   └── service/    # Business services
│   ├── handler/        # HTTP handlers
│   │   ├── health/     # Health check endpoint
│   │   └── middleware/ # HTTP middleware
│   └── scheduler/      # Cron job scheduling
├── pkg/
│   ├── config/         # Configuration utilities
│   ├── logger/         # Logging utilities
│   └── utils/          # General utilities
├── configs/            # Configuration files
├── migrations/         # Database migrations
├── scripts/            # Helper scripts
├── Taskfile.yml        # Build automation
├── docker-compose.yml  # Docker services
└── go.mod             # Go module definition
```

## Available Tasks

```bash
task build          # Build the binary
task run            # Run the application
task dev            # Run with hot reload
task test           # Run tests
task lint           # Run linter
task tidy           # Tidy go modules
task db:up          # Start PostgreSQL
task db:down        # Stop PostgreSQL
task goose:up         # Run all pending migrations
task goose:down       # Rollback one migration
task goose:up-by-one  # Run one migration
task goose:redo       # Rollback then re-run last migration
task goose:reset      # Rollback all migrations
task goose:status     # Show migration status
task goose:version    # Show current version
task goose:create --  # Create new migration
```
