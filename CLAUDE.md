# CLAUDE.md — Project Guide for AI Assistants

## Project Overview
- **bengkelin-service**: Go/Gin backend for bengkel (auto workshop) management platform
- Tech stack: Go 1.23+, Gin, GORM, PostgreSQL (pg_trgm enabled), Redis, RabbitMQ, WebSocket
- Domain entities: User, Mitra, Bengkel, Order, Vehicle, Chat, AdminFee

## Architecture — Strict Layer Rules
Handler → Service → Repository → DB

- `internal/api/handlers/`: HTTP only — parse request, call service, return response. Zero business logic, zero direct repo calls.
- `internal/service/`: ALL business logic lives here. No direct DB calls — use injected repos only.
- `internal/repository/`: Data access only. No business logic. All singletons use `sync.Once`.
- `internal/container/`: Single source of truth for all dependencies. 15 repos + 7 services wired.
- `internal/response/`, `internal/crypto/`, `internal/helpers/`, `internal/validation/`, `internal/websocket/`: Shared utilities.

## Dependency Injection Rules
- ALL handlers must get services from container — never instantiate inline
- ALL services receive repos via `ServiceDependencies` struct in `internal/service/interfaces.go`
- Adding a new repo: implement interface → add to container → inject via `ServiceDependencies`
- Adding a new service: implement interface → add to `interfaces.go` → wire in container

## Key Commands
- `make run` — Start the application
- `make build` — Build the binary
- `make test` — Run all tests
- `make test-unit` — Run unit tests only (`tests/unit/...`)
- `make test-integration` — Run integration tests (`tests/integration/...`)
- `make test-coverage` — Run tests with HTML coverage report
- `make vet` — Run go vet
- `make lint` — Run golangci-lint
- `make swagger-gen` — Regenerate Swagger docs
- `make fmt` — Format code

## Response Standards — ALWAYS follow these
```go
// Single item
response.BuildSuccessResponse(message, data)

// Paginated list
response.BuildPaginatedResponse(message, items, total, page, limit)

// Error — via BaseHandler
h.HandleError(c, err)
```
- Never return raw GORM models — always map to DTO from `internal/dto/`
- Never use `c.AbortWithStatusJSON` directly — use `h.HandleError`

## File Upload — ALWAYS use shared service
- Use `FileUploadService` from `internal/service/file_upload.go`
- Presets: `AvatarUploadConfig` (5MB), `PhotoUploadConfig` (10MB), `VehicleUploadConfig` (5MB)
- Never write inline file upload logic in handlers

## Transaction Rules
- Any operation touching 2+ tables MUST use `db.WithTransaction(fn)` from `internal/db/database.go`
- Example: `OrderService.CreateOrderWithServices` wraps order + order_service creation atomically

## Database & Performance
- Search columns use trigram indexes (pg_trgm) — do not use raw `LIKE`, use `ILIKE` with indexed columns
- Distance queries use SQL-level Haversine — never calculate distance in application code
- All repo singletons use `sync.Once` — never use bare nil-check initialization
- Redis operations must pass request context: use `WithContext` variants, never package-level ctx

## Security
- Admin endpoints protected by `AuthJWTAdmin` middleware (JWT token + role == "admin")
- Never use header-based secret for auth
- Sensitive model fields (Password, Role) must have `json:"-"`
- All responses use DTOs — raw models never reach HTTP layer

## Migration Files
- Location: `scripts/migrations/`
- Naming: `descriptive_name.sql` (check existing files for convention)
- Always include UP and DOWN
- Run `add_missing_indexes.sql` and `add_user_role_column.sql` before first run

## Testing
- Unit tests → `tests/unit/services/` or `tests/unit/handlers/`
- Integration tests → `tests/integration/api/`
- Mocks → `tests/fixtures/mocks/repositories.go` (already has all repo mocks)
- All services are fully injectable — use mock repos for unit tests

## Known Partial Issues
- Error handling: most handlers use `BaseHandler.HandleError`, but some still use direct `c.AbortWithStatusJSON` — fix when touching those files
