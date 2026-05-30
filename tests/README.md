# Bengkelin Service - Testing Infrastructure

This directory contains comprehensive testing infrastructure for the Bengkelin service, organized into different testing categories.

## 📁 Test Structure

```
tests/
├── unit/                    # Unit tests for individual components
│   ├── handlers/           # Handler unit tests
│   ├── services/           # Service layer unit tests
│   ├── repositories/       # Repository unit tests
│   └── utils/              # Utility function tests
├── integration/            # Integration tests
│   ├── api/               # API endpoint integration tests
│   └── database/          # Database integration tests
├── performance/           # Performance and load tests
│   └── load_tests/        # Load testing scenarios
├── fixtures/              # Test data and fixtures
│   ├── data/              # Test data files
│   └── mocks/             # Mock implementations
└── helpers/               # Test helper functions and utilities
```

## 🧪 Testing Categories

### Unit Tests (`/unit/`)
- **Purpose**: Test individual components in isolation
- **Coverage**: Handlers, Services, Repositories, Utilities
- **Mocking**: External dependencies are mocked
- **Speed**: Fast execution (< 1s per test)

### Integration Tests (`/integration/`)
- **Purpose**: Test component interactions and API endpoints
- **Coverage**: Full API workflows, Database operations
- **Dependencies**: Real database connections (test DB)
- **Speed**: Moderate execution (1-5s per test)

### Performance Tests (`/performance/`)
- **Purpose**: Test system performance under load
- **Coverage**: API throughput, Response times, Resource usage
- **Tools**: Custom load testing, Benchmarking
- **Speed**: Longer execution (30s-5min per test)

## 🚀 Running Tests

### All Tests
```bash
make test-all
```

### Unit Tests Only
```bash
make test-unit
```

### Integration Tests Only
```bash
make test-integration
```

### Performance Tests Only
```bash
make test-performance
```

### With Coverage Report
```bash
make test-coverage
```

## 📊 Test Coverage Goals

- **Unit Tests**: 80%+ coverage
- **Integration Tests**: All API endpoints
- **Performance Tests**: Key user journeys
- **Overall Coverage**: 75%+ combined coverage

## 🛠️ Test Configuration

### Environment Variables
```bash
# Test database configuration
TEST_DB_HOST=localhost
TEST_DB_PORT=5432
TEST_DB_NAME=bengkelin_test
TEST_DB_USER=test_user
TEST_DB_PASSWORD=test_password

# Test Redis configuration
TEST_REDIS_HOST=localhost
TEST_REDIS_PORT=6379
TEST_REDIS_DB=1

# Test RabbitMQ configuration
TEST_RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

### Test Database Setup
```bash
# Create test database
createdb bengkelin_test

# Run migrations
make migrate-test

# Seed test data
make seed-test
```

## 📝 Writing Tests

### Unit Test Example
```go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := &mocks.UserRepository{}
    service := service.NewUserService(mockRepo)
    
    // Act
    result, err := service.CreateUser(testUser)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockRepo.AssertExpectations(t)
}
```

### Integration Test Example
```go
func TestUserAPI_CreateUser(t *testing.T) {
    // Setup test server
    router := setupTestRouter()
    
    // Make request
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/v1/users", body)
    router.ServeHTTP(w, req)
    
    // Assert response
    assert.Equal(t, 201, w.Code)
}
```

## 🔧 Test Utilities

### Database Helpers
- `SetupTestDB()`: Initialize test database
- `CleanupTestDB()`: Clean test database
- `SeedTestData()`: Insert test fixtures

### HTTP Helpers
- `SetupTestRouter()`: Initialize test router
- `MakeAuthenticatedRequest()`: Create authenticated requests
- `AssertJSONResponse()`: Validate JSON responses

### Mock Helpers
- `MockUserRepository`: Mock user repository
- `MockAuthService`: Mock authentication service
- `MockRedisClient`: Mock Redis client

## 📈 Continuous Integration

### GitHub Actions Integration
```yaml
- name: Run Tests
  run: |
    make test-unit
    make test-integration
    make test-coverage
```

### Test Reports
- Coverage reports generated in `coverage/`
- Test results in JUnit format for CI integration
- Performance benchmarks tracked over time

## 🐛 Debugging Tests

### Verbose Output
```bash
go test -v ./tests/...
```

### Run Specific Test
```bash
go test -run TestUserService_CreateUser ./tests/unit/services/
```

### Debug with Delve
```bash
dlv test ./tests/unit/services/ -- -test.run TestUserService_CreateUser
```

## 📋 Test Checklist

### Before Committing
- [ ] All tests pass locally
- [ ] Coverage meets minimum threshold
- [ ] No test data leakage between tests
- [ ] Integration tests use test database
- [ ] Performance tests don't exceed time limits

### Code Review
- [ ] Tests cover happy path and error cases
- [ ] Mocks are properly configured
- [ ] Test names are descriptive
- [ ] Test data is realistic
- [ ] Cleanup is properly handled

---

**Last Updated**: January 3, 2026  
**Maintained By**: Bengkelin Development Team