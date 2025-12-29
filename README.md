# Bengkelin Service

Bengkelin Service adalah sebuah aplikasi berbasis Golang yang menyediakan layanan manajemen bengkel. Aplikasi ini memungkinkan pengguna untuk mengelola berbagai aspek dari sebuah bengkel seperti pencatatan servis, manajemen inventaris, dan penjadwalan.

## 🚀 Quick Start dengan Docker (Direkomendasikan)

### Prasyarat
- Docker & Docker Compose
- Git

### Menjalankan Development Environment

**MySQL (Default):**

**Windows:**
```cmd
# Clone repository
git clone https://github.com/your-username/bengkelin-service.git
cd bengkelin-service

# Setup environment
copy .env.example .env
# Edit .env sesuai konfigurasi Anda

# Build dan jalankan
scripts\docker-build.bat dev mysql
scripts\docker-run.bat dev mysql up
```

**Linux/Mac:**
```bash
# Clone repository
git clone https://github.com/your-username/bengkelin-service.git
cd bengkelin-service

# Setup environment
cp .env.example .env
# Edit .env sesuai konfigurasi Anda

# Build dan jalankan
./scripts/docker-build.sh dev mysql
./scripts/docker-run.sh dev mysql up
```

**PostgreSQL:**

**Windows:**
```cmd
# Setup environment dengan PostgreSQL
# Edit .env: DATABASE_DRIVER=postgres

# Build dan jalankan
scripts\docker-build.bat dev postgres
scripts\docker-run.bat dev postgres up
```

**Linux/Mac:**
```bash
# Setup environment dengan PostgreSQL
# Edit .env: DATABASE_DRIVER=postgres

# Build dan jalankan
./scripts/docker-build.sh dev postgres
./scripts/docker-run.sh dev postgres up
```

### Akses Layanan Development

**MySQL:**
- **API**: http://localhost:3000
- **Swagger Documentation**: http://localhost:3000/swagger/index.html
- **Health Check**: http://localhost:3000/health
- **Metrics (Prometheus)**: http://localhost:3000/metrics
- **Application Metrics**: http://localhost:3000/metrics/app
- **phpMyAdmin**: http://localhost:8080 (user: root, password: lihat .env)
- **Redis Commander**: http://localhost:8081
- **MailHog**: http://localhost:8025

**PostgreSQL:**
- **API**: http://localhost:3000
- **Swagger Documentation**: http://localhost:3000/swagger/index.html
- **Health Check**: http://localhost:3000/health
- **Metrics (Prometheus)**: http://localhost:3000/metrics
- **Application Metrics**: http://localhost:3000/metrics/app
- **pgAdmin**: http://localhost:8082 (admin@bengkelin.com / admin123)
- **Redis Commander**: http://localhost:8081
- **MailHog**: http://localhost:8025

## 🏭 Production Deployment

```bash
# MySQL production
./scripts/docker-build.sh prod mysql
./scripts/docker-run.sh prod mysql up

# PostgreSQL production
./scripts/docker-build.sh prod postgres
./scripts/docker-run.sh prod postgres up

# Akses via Nginx: https://localhost
```

## 🛠️ Manual Installation

### Prasyarat
- [Golang](https://golang.org/dl/) versi 1.21 atau lebih baru
- [MySQL](https://dev.mysql.com/downloads/) 8.0+
- [Redis](https://redis.io/download) (untuk caching dan rate limiting)
- [Git](https://git-scm.com/)

### Langkah-langkah

1. **Clone Repository**
```bash
git clone https://github.com/your-username/bengkelin-service.git
cd bengkelin-service
```

2. **Install Dependencies**
```bash
go mod tidy
```

3. **Setup Environment**
```bash
cp .env.example .env
# Edit .env dengan konfigurasi database dan Redis Anda
```

4. **Menjalankan Aplikasi**
```bash
make run
# atau
go run cmd/app/main.go
```

## 🏗️ Arsitektur & Fitur

### Fitur Utama
- ✅ **JWT Authentication** dengan refresh token dan automatic rotation
- ✅ **Multi-tier Rate Limiting** (general, auth, strict)
- ✅ **Input Validation** dengan security-focused rules (XSS, SQL injection prevention)
- ✅ **Clean Architecture** dengan service layer dan dependency injection
- ✅ **Structured Logging** dengan context awareness
- ✅ **API Documentation** dengan Swagger/OpenAPI 3.0
- ✅ **Health Monitoring** dengan comprehensive health checks
- ✅ **Metrics Collection** dengan Prometheus integration
- ✅ **Docker Support** dengan multi-stage build untuk production
- ✅ **Indonesian-specific Validation** (nomor telepon, plat kendaraan, hari dalam bahasa Indonesia)

### Teknologi Stack
- **Backend**: Go 1.21, Gin Framework
- **Database**: MySQL 8.0 atau PostgreSQL 15 dengan GORM
- **Cache**: Redis untuk rate limiting dan caching
- **Authentication**: JWT dengan refresh token
- **Containerization**: Docker dengan multi-stage build
- **Reverse Proxy**: Nginx dengan SSL support
- **Monitoring**: Health checks dan structured logging

### Struktur Proyek
```
├── cmd/app/                 # Application entry point
├── internal/
│   ├── api/                 # API layer (handlers, middleware, router)
│   └── pkg/                 # Internal packages (models, services, repositories)
├── pkg/                     # Shared packages (crypto, validation, logging)
├── config/                  # Configuration files (nginx, redis, mysql)
├── scripts/                 # Build dan deployment scripts
├── docs/                    # Documentation
└── tests/                   # Test files
```

## 📚 Dokumentasi

- [Docker Setup Guide](docs/DOCKER_SETUP.md) - Panduan lengkap Docker setup
- [PostgreSQL Setup Guide](docs/POSTGRESQL_SETUP.md) - Panduan setup PostgreSQL
- [JWT Implementation](docs/JWT_IMPLEMENTATION.md) - Detail implementasi JWT
- [Rate Limiting](docs/RATE_LIMITING.md) - Konfigurasi rate limiting
- [Input Validation](docs/INPUT_VALIDATION.md) - Sistem validasi input
- [Architecture Improvements](docs/ARCHITECTURE_IMPROVEMENTS.md) - Peningkatan arsitektur
- [Logging Enhancement](docs/LOGGING_ENHANCEMENT.md) - Sistem logging
- [Filename Standardization](docs/FILENAME_STANDARDIZATION.md) - Standarisasi nama file

## 🔧 Development Tools

### Swagger Documentation
```bash
# Install Swagger CLI
make swagger-install

# Generate Swagger documentation
make swagger-gen

# View documentation info
make swagger-serve

# Clean generated files
make swagger-clean
```

### Docker Commands
```bash
# MySQL Development
scripts/docker-build.sh dev mysql     # Build development image
scripts/docker-run.sh dev mysql up    # Start development environment
scripts/docker-run.sh dev mysql logs  # View logs

# PostgreSQL Development
scripts/docker-build.sh dev postgres     # Build development image
scripts/docker-run.sh dev postgres up    # Start development environment
scripts/docker-run.sh dev postgres logs  # View logs

# MySQL Production
scripts/docker-build.sh prod mysql    # Build production image
scripts/docker-run.sh prod mysql up   # Start production environment

# PostgreSQL Production
scripts/docker-build.sh prod postgres    # Build production image
scripts/docker-run.sh prod postgres up   # Start production environment

# Cleanup
scripts/docker-cleanup.sh light # Light cleanup
scripts/docker-cleanup.sh full  # Full cleanup
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific test
go test ./tests/jwt_test.go
go test ./tests/validation_test.go
go test ./tests/rate_limit_test.go
```

### Monitoring & Health Checks
```bash
# Health check endpoints
curl http://localhost:3000/health      # Comprehensive health check
curl http://localhost:3000/ready       # Readiness check (K8s ready)
curl http://localhost:3000/live        # Liveness check (K8s liveness)

# Metrics endpoints
curl http://localhost:3000/metrics     # Prometheus metrics
curl http://localhost:3000/metrics/app # Application-specific metrics

# API Documentation
curl http://localhost:3000/swagger/index.html # Swagger UI
```

### Build
```bash
# Development build
go build -o main cmd/app/main.go

# Production build (optimized)
CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o main cmd/app/main.go
```

## 🔒 Security Features

- **JWT dengan Refresh Token**: Automatic token rotation dan revocation
- **Rate Limiting**: Multi-tier protection (100/min general, 10/min auth, 5/min strict)
- **Input Validation**: XSS dan SQL injection prevention
- **Security Headers**: Comprehensive security headers via Nginx
- **HTTPS**: SSL termination dengan modern TLS configuration
- **Database Security**: Prepared statements dan input sanitization

## 🚀 Production Ready

Aplikasi ini telah dioptimalkan untuk production dengan:

- **Multi-stage Docker Build**: Image size ~15MB
- **Resource Limits**: CPU dan memory limits untuk setiap service
- **Health Checks**: Comprehensive health monitoring
- **Logging**: Structured logging dengan log levels
- **Monitoring**: Ready untuk integrasi dengan monitoring tools
- **SSL/TLS**: Production-ready HTTPS configuration
- **Database Optimization**: Connection pooling dan query optimization

## 🤝 Contributing

1. Fork repository
2. Buat feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push ke branch (`git push origin feature/amazing-feature`)
5. Buat Pull Request

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

## 📞 Contact

Project Link: [https://github.com/your-username/bengkelin-service](https://github.com/your-username/bengkelin-service)