#!/bin/bash

# Docker build script for Bengkelin API
# Usage: ./scripts/docker-build.sh [environment] [database]
# Environment: dev, prod (default: dev)
# Database: mysql, postgres (default: mysql)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENV=${1:-dev}
DB=${2:-mysql}

echo -e "${BLUE}🐳 Building Bengkelin API Docker images for ${ENV} environment with ${DB}...${NC}"

# Validate environment
if [[ "$ENV" != "dev" && "$ENV" != "prod" ]]; then
    echo -e "${RED}❌ Invalid environment: $ENV. Use 'dev' or 'prod'${NC}"
    exit 1
fi

# Validate database
if [[ "$DB" != "mysql" && "$DB" != "postgres" ]]; then
    echo -e "${RED}❌ Invalid database: $DB. Use 'mysql' or 'postgres'${NC}"
    exit 1
fi

# Build based on environment and database
case $ENV in
    "dev")
        echo -e "${YELLOW}📦 Building development image...${NC}"
        docker build -f Dockerfile.dev -t bengkelin-api:dev .
        echo -e "${GREEN}✅ Development image built successfully!${NC}"
        
        echo -e "${YELLOW}🔧 Building development services with ${DB}...${NC}"
        if [[ "$DB" == "postgres" ]]; then
            docker-compose -f docker-compose.yml -f docker-compose.dev.yml -f docker-compose.postgres-dev.yml build
        else
            docker-compose -f docker-compose.yml -f docker-compose.dev.yml build
        fi
        echo -e "${GREEN}✅ Development services built successfully!${NC}"
        ;;
        
    "prod")
        echo -e "${YELLOW}📦 Building production image...${NC}"
        docker build -f Dockerfile -t bengkelin-api:prod .
        echo -e "${GREEN}✅ Production image built successfully!${NC}"
        
        echo -e "${YELLOW}🔧 Building production services with ${DB}...${NC}"
        if [[ "$DB" == "postgres" ]]; then
            docker-compose -f docker-compose.yml -f docker-compose.prod.yml -f docker-compose.postgres-prod.yml build
        else
            docker-compose -f docker-compose.yml -f docker-compose.prod.yml build
        fi
        echo -e "${GREEN}✅ Production services built successfully!${NC}"
        ;;
esac

# Show image sizes
echo -e "${BLUE}📊 Image sizes:${NC}"
docker images | grep bengkelin-api

echo -e "${GREEN}🎉 Build completed for ${ENV} environment with ${DB}!${NC}"

# Show next steps
echo -e "${BLUE}📋 Next steps:${NC}"
case $ENV in
    "dev")
        echo -e "  ${YELLOW}Start development:${NC} ./scripts/docker-run.sh dev $DB"
        if [[ "$DB" == "postgres" ]]; then
            echo -e "  ${YELLOW}View logs:${NC} docker-compose -f docker-compose.yml -f docker-compose.dev.yml -f docker-compose.postgres-dev.yml logs -f"
        else
            echo -e "  ${YELLOW}View logs:${NC} docker-compose -f docker-compose.yml -f docker-compose.dev.yml logs -f"
        fi
        ;;
    "prod")
        echo -e "  ${YELLOW}Start production:${NC} ./scripts/docker-run.sh prod $DB"
        if [[ "$DB" == "postgres" ]]; then
            echo -e "  ${YELLOW}View logs:${NC} docker-compose -f docker-compose.yml -f docker-compose.prod.yml -f docker-compose.postgres-prod.yml logs -f"
        else
            echo -e "  ${YELLOW}View logs:${NC} docker-compose -f docker-compose.yml -f docker-compose.prod.yml logs -f"
        fi
        ;;
esac