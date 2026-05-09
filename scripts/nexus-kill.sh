#!/bin/bash

# NEXUS CYBER - EMERGENCY SHUTDOWN SCRIPT
# Digunakan untuk mematikan semua layanan Nexus secara bersih.

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${RED}Shutting down Nexus Cyber Infrastructure...${NC}"

# Matikan Go Gateway
pkill -f "go run cmd/gateway/main.go" && echo -e "${GREEN}✔ Gateway Stopped${NC}"

# Matikan Python OJK Portal
pkill -f "python app.py" && echo -e "${GREEN}✔ OJK Portal Stopped${NC}"

# Matikan Next.js Dashboard
pkill -f "next-dev" && echo -e "${GREEN}✔ Dashboard Stopped${NC}"
pkill -f "next"

# Matikan Redis
docker-compose stop redis && echo -e "${GREEN}✔ Redis Container Stopped${NC}"

echo -e "${RED}ALL SYSTEMS OFFLINE.${NC}"
