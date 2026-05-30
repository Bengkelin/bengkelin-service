.PHONY: run build test test-coverage swagger-install swagger-gen swagger-serve test-all test-unit test-integration test-performance test-setup test-cleanup docker-dev docker-dev-down docker-prod docker-prod-down

# Build and run commands
run:
	go run cmd/app/main.go

build:
	go build -o cmd/app cmd/app/main.go

# Legacy test commands (keeping for compatibility)
test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Comprehensive test commands
test-all:
	@echo "Running all tests..."
	go test -v ./tests/...

test-unit:
	@echo "Running unit tests..."
	go test -v ./tests/unit/...

test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/integration/...

test-performance:
	@echo "Running performance tests..."
	go test -v ./tests/performance/...

test-coverage-new:
	@echo "Running tests with coverage..."
	@mkdir -p coverage
	go test -v -coverprofile=coverage/coverage.out ./tests/...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report generated at coverage/coverage.html"

test-coverage-func:
	@echo "Running tests with function coverage..."
	@mkdir -p coverage
	go test -v -coverprofile=coverage/coverage.out ./tests/...
	go tool cover -func=coverage/coverage.out

test-benchmark:
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./tests/performance/...

test-setup:
	@echo "Setting up test environment..."
	@mkdir -p coverage
	@echo "Creating test database..."
	createdb bengkelin_test || echo "Test database already exists"

test-cleanup:
	@echo "Cleaning up test environment..."
	dropdb bengkelin_test || echo "Test database doesn't exist"
	rm -rf coverage/

test-watch:
	@echo "Running tests in watch mode..."
	find . -name "*.go" | entr -r make test-unit

# Test database commands
migrate-test:
	@echo "Running test database migrations..."
	migrate -path ./migrations -database "postgres://postgres:password@localhost/bengkelin_test?sslmode=disable" up

seed-test:
	@echo "Seeding test database..."
	go run scripts/seed_test_data.go

# Code quality commands
lint:
	@echo "Running linter..."
	golangci-lint run

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...

# Security scanning
security-scan:
	@echo "Running security scan..."
	gosec ./...

# Dependency management
deps-download:
	@echo "Downloading dependencies..."
	go mod download

deps-tidy:
	@echo "Tidying dependencies..."
	go mod tidy

deps-verify:
	@echo "Verifying dependencies..."
	go mod verify

# CI/CD commands
ci-test:
	@echo "Running CI tests..."
	make test-setup
	make test-coverage-new
	make test-cleanup

ci-build:
	@echo "Running CI build..."
	make deps-download
	make lint
	make vet
	make build

# Development workflow
dev-setup:
	@echo "Setting up development environment..."
	make deps-download
	make test-setup

dev-test:
	@echo "Running development tests..."
	make test-unit
	make test-integration

# Clean commands
clean:
	go clean
	rm -f cmd/app/main
	rm -rf coverage/

clean-all: clean test-cleanup
	@echo "Cleaned all artifacts and test data"

# Docker commands
docker-dev:
	docker compose -f docker/docker-compose.dev.yml up --build

docker-dev-down:
	docker compose -f docker/docker-compose.dev.yml down

docker-dev-logs:
	docker compose -f docker/docker-compose.dev.yml logs -f

docker-prod:
	docker compose -f docker/docker-compose.prod.yml up -d --build

docker-prod-down:
	docker compose -f docker/docker-compose.prod.yml down

docker-prod-logs:
	docker compose -f docker/docker-compose.prod.yml logs -f

# Swagger documentation
swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest

swagger-gen:
	swag init -g main.go -o docs --parseDependency --parseInternal

swagger-serve:
	@echo "Swagger documentation available at: http://localhost:3000/swagger/index.html"
	@echo "Make sure to run 'make swagger-gen' first to generate the documentation"
	@echo "Then start the server with 'make run'"

swagger-clean:
	rm -f docs/docs.go docs/swagger.json docs/swagger.yaml

# Help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  build           - Build the application"
	@echo "  run             - Run the application"
	@echo ""
	@echo "Testing:"
	@echo "  test            - Run legacy tests (all packages)"
	@echo "  test-all        - Run all new structured tests"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-performance - Run performance tests only"
	@echo "  test-coverage   - Run legacy coverage"
	@echo "  test-coverage-new - Run new structured coverage"
	@echo "  test-benchmark  - Run benchmark tests"
	@echo "  test-setup      - Setup test environment"
	@echo "  test-cleanup    - Cleanup test environment"
	@echo "  test-watch      - Run tests in watch mode"
	@echo ""
	@echo "Database:"
	@echo "  migrate-test    - Run test database migrations"
	@echo "  seed-test       - Seed test database"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo "  vet             - Run go vet"
	@echo "  security-scan   - Run security scan"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps-download   - Download dependencies"
	@echo "  deps-tidy       - Tidy dependencies"
	@echo "  deps-verify     - Verify dependencies"
	@echo ""
	@echo "CI/CD:"
	@echo "  ci-test         - Run CI tests"
	@echo "  ci-build        - Run CI build"
	@echo ""
	@echo "Development:"
	@echo "  dev-setup       - Setup development environment"
	@echo "  dev-test        - Run development tests"
	@echo ""
	@echo "Swagger:"
	@echo "  swagger-install - Install Swagger CLI"
	@echo "  swagger-gen     - Generate Swagger docs"
	@echo "  swagger-serve   - Show Swagger serve info"
	@echo "  swagger-clean   - Clean Swagger files"
	@echo ""
	@echo "Cleanup:"
	@echo "  clean           - Clean build artifacts"
	@echo "  clean-all       - Clean all artifacts and test data"
	@echo ""
	@echo "Docker:"
	@echo "  docker-dev      - Start development environment (PostgreSQL + Redis + pgAdmin)"
	@echo "  docker-dev-down - Stop development environment"
	@echo "  docker-dev-logs - View development logs"
	@echo "  docker-prod     - Start production environment"
	@echo "  docker-prod-down- Stop production environment"
	@echo "  docker-prod-logs- View production logs"