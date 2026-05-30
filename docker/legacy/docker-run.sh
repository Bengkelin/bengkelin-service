#!/bin/bash

# Docker run script for Bengkelin API
# Usage: ./scripts/docker-run.sh [environment] [database] [action]
# Environment: dev, prod (default: dev)
# Database: mysql, postgres (default: mysql)
# Action: up, down, restart, logs, status (default: up)

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
ACTION=${3:-up}

echo -e "${BLUE}🐳 Managing Bengkelin API Docker containers...${NC}"

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

# Validate action
if [[ "$ACTION" != "up" && "$ACTION" != "down" && "$ACTION" != "restart" && "$ACTION" != "logs" && "$ACTION" != "status" ]]; then
    echo -e "${RED}❌ Invalid action: $ACTION. Use 'up', 'down', 'restart', 'logs', or 'status'${NC}"
    exit 1
fi

# Set compose files based on environment and database
if [[ "$ENV" == "dev" ]]; then
    if [[ "$DB" == "postgres" ]]; then
        COMPOSE_FILES="-f docker-compose.yml -f docker-compose.dev.yml -f docker-compose.postgres-dev.yml"
        PROJECT_NAME="bengkelin-dev-postgres"
    else
        COMPOSE_FILES="-f docker-compose.yml -f docker-compose.dev.yml"
        PROJECT_NAME="bengkelin-dev"
    fi
else
    if [[ "$DB" == "postgres" ]]; then
        COMPOSE_FILES="-f docker-compose.yml -f docker-compose.prod.yml -f docker-compose.postgres-prod.yml"
        PROJECT_NAME="bengkelin-prod-postgres"
    else
        COMPOSE_FILES="-f docker-compose.yml -f docker-compose.prod.yml"
        PROJECT_NAME="bengkelin-prod"
    fi
fi

# Execute action
case $ACTION in
    "up")
        echo -e "${YELLOW}🚀 Starting ${ENV} environment with ${DB}...${NC}"
        
        # Check if .env file exists
        if [[ ! -f .env ]]; then
            echo -e "${YELLOW}⚠️  .env file not found. Creating from .env.example...${NC}"
            cp .env.example .env
            echo -e "${YELLOW}📝 Please edit .env file with your configuration before running again.${NC}"
            exit 1
        fi
        
        # Start services
        docker-compose $COMPOSE_FILES -p $PROJECT_NAME up -d
        
        echo -e "${GREEN}✅ ${ENV} environment with ${DB} started successfully!${NC}"
        
        # Show status
        echo -e "${BLUE}📊 Container status:${NC}"
        docker-compose $COMPOSE_FILES -p $PROJECT_NAME ps
        
        # Show URLs
        echo -e "${BLUE}🌐 Available services:${NC}"
        if [[ "$ENV" == "dev" ]]; then
            echo -e "  ${GREEN}API:${NC} http://localhost:3000"
            if [[ "$DB" == "postgres" ]]; then
                echo -e "  ${GREEN}pgAdmin:${NC} http://localhost:8082 (admin@bengkelin.com / admin123)"
                echo -e "  ${GREEN}PostgreSQL:${NC} localhost:5432"
            else
                echo -e "  ${GREEN}phpMyAdmin:${NC} http://localhost:8080"
                echo -e "  ${GREEN}MySQL:${NC} localhost:3306"
            fi
            echo -e "  ${GREEN}Redis Commander:${NC} http://localhost:8081"
            echo -e "  ${GREEN}MailHog:${NC} http://localhost:8025"
        else
            echo -e "  ${GREEN}API:${NC} https://localhost (via Nginx)"
            echo -e "  ${GREEN}API Direct:${NC} http://localhost:3000"
        fi
        ;;
        
    "down")
        echo -e "${YELLOW}🛑 Stopping ${ENV} environment with ${DB}...${NC}"
        docker-compose $COMPOSE_FILES -p $PROJECT_NAME down
        echo -e "${GREEN}✅ ${ENV} environment stopped successfully!${NC}"
        ;;
        
    "restart")
        echo -e "${YELLOW}🔄 Restarting ${ENV} environment with ${DB}...${NC}"
        docker-compose $COMPOSE_FILES -p $PROJECT_NAME restart
        echo -e "${GREEN}✅ ${ENV} environment restarted successfully!${NC}"
        ;;
        
    "logs")
        echo -e "${YELLOW}📋 Showing logs for ${ENV} environment with ${DB}...${NC}"
        docker-compose $COMPOSE_FILES -p $PROJECT_NAME logs -f
        ;;
        
    "status")
        echo -e "${BLUE}📊 Status for ${ENV} environment with ${DB}:${NC}"
        docker-compose $COMPOSE_FILES -p $PROJECT_NAME ps
        
        echo -e "\n${BLUE}💾 Volume usage:${NC}"
        docker system df
        
        echo -e "\n${BLUE}🔍 Health checks:${NC}"
        docker-compose $COMPOSE_FILES -p $PROJECT_NAME ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
        ;;
esac

# Show helpful commands
echo -e "\n${BLUE}💡 Helpful commands:${NC}"
echo -e "  ${YELLOW}View logs:${NC} ./scripts/docker-run.sh $ENV $DB logs"
echo -e "  ${YELLOW}Check status:${NC} ./scripts/docker-run.sh $ENV $DB status"
echo -e "  ${YELLOW}Restart:${NC} ./scripts/docker-run.sh $ENV $DB restart"
echo -e "  ${YELLOW}Stop:${NC} ./scripts/docker-run.sh $ENV $DB down"
echo -e "  ${YELLOW}Shell into API:${NC} docker exec -it ${PROJECT_NAME}_bengkelin-api_1 sh"