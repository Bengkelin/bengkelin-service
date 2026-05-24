# Bengkelin Service

> Go-based REST API backend for a bengkel (auto-repair workshop) management platform.

## Quick Links

| Resource                        | Path                                                                                             |
| ------------------------------- | ------------------------------------------------------------------------------------------------ |
| 📖 **Codebase Overview**        | [`docs/CODEBASE_OVERVIEW.md`](docs/CODEBASE_OVERVIEW.md)                                         |
| 🚀 **Full Setup & Usage Guide** | [`docs/README.full.md`](docs/README.full.md)                                                     |
| 🏗️ **Architecture**             | [`docs/architecture/07-Architecture.md`](docs/architecture/07-Architecture.md)                   |
| 🗄️ **Database Schema**          | [`docs/database/02-ERD.md`](docs/database/02-ERD.md)                                             |
| 📡 **API Documentation**        | [`docs/api/api-documentation.md`](docs/api/api-documentation.md)                                 |
| 🐳 **Docker Setup**             | [`docs/deployment/DOCKER_SETUP.md`](docs/deployment/DOCKER_SETUP.md)                             |
| 🧪 **Testing Guide**            | [`docs/deployment/TESTING_GUIDE.md`](docs/deployment/TESTING_GUIDE.md)                           |
| 🔐 **JWT Implementation**       | [`docs/implementation/JWT_IMPLEMENTATION.md`](docs/implementation/JWT_IMPLEMENTATION.md)         |
| 📊 **Technical Skills**         | [`docs/portfolio/TECHNICAL_SKILLS_ASSESSMENT.md`](docs/portfolio/TECHNICAL_SKILLS_ASSESSMENT.md) |

## Tech Stack

**Go 1.23 · Gin · GORM · PostgreSQL (with pg_trgm) · Redis · RabbitMQ · JWT · WebSocket · Prometheus · Docker**

- **sync.Once** singleton pattern for all 16 repository instances (thread-safe)

## Architecture

```
HTTP Request
    │
    ▼
┌─────────┐    ┌─────────┐    ┌────────────┐    ┌────┐
│ Handler │───▶│ Service │───▶│ Repository │───▶│ DB │
└─────────┘    └─────────┘    └────────────┘    └────┘
    │               │               │
    │          (business        (data access
    │           logic)           only)
    │
  (parse request,
   validate, format response)
```

**Layer rules:**

| Layer | Responsibility |
|-------|---------------|
| **Handler** (`internal/api/handlers/`) | Parse HTTP requests, call service, format response. No business logic. |
| **Service** (`internal/pkg/service/`) | Business logic, validation rules, orchestrate multiple repos. |
| **Repository** (`internal/pkg/repository/`) | Data access only — GORM queries, no business logic. |
| **Container** (`internal/pkg/container/`) | Dependency injection — wires 15 repos + 7 services at startup. |

**Key packages:**

| Package | Purpose |
|---------|---------|
| `internal/pkg/container/` | DI container with constructor-based injection |
| `internal/pkg/dto/` | Request/response DTOs — decoupled from DB models |
| `internal/pkg/errors/` | Structured error types for consistent error handling |
| `pkg/response/` | Standard response envelope (`BuildSuccessResponse`, `BuildPaginatedResponse`) |

## Getting Started

```bash
# 1. Clone & configure
git clone https://github.com/your-username/bengkelin-service.git
cd bengkelin-service
cp .env.example .env   # edit .env with your values

# 2. Run database migrations
psql -d bengkelin -f scripts/migrations/add_missing_indexes.sql
psql -d bengkelin -f scripts/migrations/add_performance_indexes.sql

# 3. Run
make run

# Or with Docker
./scripts/docker-build.sh dev mysql
./scripts/docker-run.sh dev mysql up
```

**API**: http://localhost:3000
**Swagger**: http://localhost:3000/swagger/index.html
**Health**: http://localhost:3000/health

**Common commands:**

```bash
make build              # Build binary
make run                # Run application
make test-unit          # Run unit tests
make test-integration   # Run integration tests
make test-coverage      # Run tests with coverage
make lint               # Run linter
make swagger-gen        # Regenerate Swagger docs
```

## Database

- **PostgreSQL** with `pg_trgm` extension for trigram-based `ILIKE` search
- Run `scripts/migrations/add_missing_indexes.sql` before first run — adds trigram indexes for search columns (`bengkel_name`, `nama_service`, `full_address`, `city`, `province`), foreign key indexes, and composite indexes for common query patterns
- Run `scripts/migrations/add_performance_indexes.sql` for additional performance indexes

## Project Structure

```
bengkelin-service/
├── cmd/app/                    # Application entrypoint
├── internal/
│   ├── api/
│   │   ├── handlers/           # HTTP handlers (auth, bengkel, user, chat, order, etc.)
│   │   ├── middleware/         # JWT, CORS, logging, validation, error handler
│   │   └── router/             # v1/ and v2/ route definitions
│   └── pkg/
│       ├── config/             # App configuration
│       ├── constants/          # Shared constants
│       ├── container/          # DI container (wires 15 repos + 7 services)
│       ├── db/                 # Database connection + transaction helper
│       ├── dto/                # Request/response DTOs
│       ├── errors/             # Structured error types
│       ├── events/             # Event definitions
│       ├── models/             # GORM model definitions
│       ├── rabbitmq/           # RabbitMQ producer/consumer
│       ├── redis/              # Redis client
│       ├── repository/         # Data access layer (16 repos, sync.Once singletons)
│       ├── service/            # Business logic layer (7+ services)
│       └── validator/          # Input validation
├── pkg/
│   ├── helpers/                # Shared utilities
│   └── response/               # Standard response envelope
├── scripts/migrations/         # SQL migration scripts
└── tests/                      # Unit, integration, performance tests
```

**Services:** AuthService, UserService, BengkelService, OrderService, ChatService, MitraService, AdminFeeService, FileUploadService

## API Response Format

All endpoints return a standard envelope:

```json
// Success (single item)
{
  "success": true,
  "message": "Data retrieved successfully",
  "errors": null,
  "data": { ... }
}

// Success (paginated list)
{
  "success": true,
  "message": "Data retrieved successfully",
  "errors": null,
  "data": {
    "items": [ ... ],
    "total": 100,
    "page": 1,
    "limit": 10,
    "total_pages": 10
  }
}

// Error
{
  "success": false,
  "message": "Validation failed",
  "errors": { "field": "error message" },
  "data": null
}
```

---

For full documentation see [`docs/`](docs/).

## Contributing / Development Notes

**Architecture rules (from Phase 1 & 2a refactors):**

- Handlers must not contain business logic — delegate to services
- Services orchestrate repositories; repositories do data access only
- Use `db.WithTransaction` for multi-step write operations
- Use `FileUploadService` for all file uploads (don't duplicate upload logic in handlers)
- All response DTOs must use the standard `pkg/response/` envelope
- Repository singletons use `sync.Once` — don't bypass with direct struct instantiation
- New database queries that use `ILIKE` require a trigram index (see `scripts/migrations/add_missing_indexes.sql`)

**Adding new features:**

1. Define model in `internal/pkg/models/`
2. Create DTOs in `internal/pkg/dto/`
3. Add repository in `internal/pkg/repository/` with `sync.Once` singleton
4. Add service in `internal/pkg/service/` implementing an interface from `interfaces.go`
5. Wire into `internal/pkg/container/container.go`
6. Add handler in `internal/api/handlers/`
7. Register route in `internal/api/router/`

**Refactor history:** See [`plan-claude-code/`](plan-claude-code/) for Phase 1 (clean architecture) and Phase 2a (database performance & thread safety) details.
