#!/bin/bash

# Docker cleanup script for Bengkelin API
# Usage: ./scripts/docker-cleanup.sh [type]
# Type: light, full, nuclear (default: light)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default cleanup type
TYPE=${1:-light}

echo -e "${BLUE}🧹 Docker cleanup for Bengkelin API...${NC}"

# Validate cleanup type
if [[ "$TYPE" != "light" && "$TYPE" != "full" && "$TYPE" != "nuclear" ]]; then
    echo -e "${RED}❌ Invalid cleanup type: $TYPE. Use 'light', 'full', or 'nuclear'${NC}"
    exit 1
fi

# Show current disk usage
echo -e "${BLUE}📊 Current Docker disk usage:${NC}"
docker system df

case $TYPE in
    "light")
        echo -e "${YELLOW}🧽 Performing light cleanup...${NC}"
        
        # Remove stopped containers
        echo -e "${YELLOW}Removing stopped containers...${NC}"
        docker container prune -f
        
        # Remove dangling images
        echo -e "${YELLOW}Removing dangling images...${NC}"
        docker image prune -f
        
        # Remove unused networks
        echo -e "${YELLOW}Removing unused networks...${NC}"
        docker network prune -f
        
        echo -e "${GREEN}✅ Light cleanup completed!${NC}"
        ;;
        
    "full")
        echo -e "${YELLOW}🧽 Performing full cleanup...${NC}"
        
        # Stop all Bengkelin containers
        echo -e "${YELLOW}Stopping all Bengkelin containers...${NC}"
        docker-compose -f docker-compose.yml -f docker-compose.dev.yml -p bengkelin-dev down 2>/dev/null || true
        docker-compose -f docker-compose.yml -f docker-compose.prod.yml -p bengkelin-prod down 2>/dev/null || true
        
        # Remove all stopped containers
        echo -e "${YELLOW}Removing all stopped containers...${NC}"
        docker container prune -f
        
        # Remove all unused images
        echo -e "${YELLOW}Removing all unused images...${NC}"
        docker image prune -a -f
        
        # Remove all unused networks
        echo -e "${YELLOW}Removing all unused networks...${NC}"
        docker network prune -f
        
        # Remove all unused volumes (be careful!)
        echo -e "${YELLOW}Removing all unused volumes...${NC}"
        docker volume prune -f
        
        echo -e "${GREEN}✅ Full cleanup completed!${NC}"
        ;;
        
    "nuclear")
        echo -e "${RED}☢️  NUCLEAR CLEANUP - This will remove EVERYTHING!${NC}"
        echo -e "${YELLOW}This includes:${NC}"
        echo -e "  - All containers (running and stopped)"
        echo -e "  - All images"
        echo -e "  - All volumes (including data!)"
        echo -e "  - All networks"
        echo -e "  - Build cache"
        
        read -p "Are you absolutely sure? Type 'YES' to continue: " -r
        if [[ $REPLY == "YES" ]]; then
            echo -e "${RED}💥 Performing nuclear cleanup...${NC}"
            
            # Stop all containers
            echo -e "${YELLOW}Stopping all containers...${NC}"
            docker stop $(docker ps -aq) 2>/dev/null || true
            
            # Remove everything
            echo -e "${YELLOW}Removing everything...${NC}"
            docker system prune -a -f --volumes
            
            # Remove build cache
            echo -e "${YELLOW}Removing build cache...${NC}"
            docker builder prune -a -f
            
            echo -e "${RED}💥 Nuclear cleanup completed! Everything is gone!${NC}"
        else
            echo -e "${BLUE}❌ Nuclear cleanup cancelled.${NC}"
            exit 0
        fi
        ;;
esac

# Show disk usage after cleanup
echo -e "\n${BLUE}📊 Docker disk usage after cleanup:${NC}"
docker system df

# Show what's left
echo -e "\n${BLUE}🔍 Remaining Docker resources:${NC}"
echo -e "${YELLOW}Images:${NC}"
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | head -10

echo -e "\n${YELLOW}Containers:${NC}"
docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Size}}" | head -10

echo -e "\n${YELLOW}Volumes:${NC}"
docker volume ls --format "table {{.Driver}}\t{{.Name}}" | head -10

echo -e "\n${GREEN}🎉 Cleanup completed!${NC}"

# Show rebuild commands
echo -e "\n${BLUE}💡 To rebuild everything:${NC}"
echo -e "  ${YELLOW}Development:${NC} ./scripts/docker-build.sh dev"
echo -e "  ${YELLOW}Production:${NC} ./scripts/docker-build.sh prod"