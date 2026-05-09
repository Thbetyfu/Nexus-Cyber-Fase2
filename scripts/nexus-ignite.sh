#!/bin/bash

# NEXUS CYBER - AUTO-IGNITION SCRIPT v1.0
# Digunakan untuk menyalakan seluruh infrastruktur dalam satu perintah.

# Warna untuk output terminal
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}          NEXUS CYBER - INITIALIZING MATRIX       ${NC}"
echo -e "${BLUE}==================================================${NC}"

# 1. Pastikan Redis Menyala
echo -e "${GREEN}[1/4] Memastikan Redis Cache (MTD Matrix) aktif...${NC}"
docker-compose up -d redis > /dev/null 2>&1

# 2. Nyalakan Portal OJK (Target)
echo -e "${GREEN}[2/4] Meluncurkan Portal OJK (Protected Asset)...${NC}"
cd playground/vulnerable-ojk-portal
./venv/bin/python app.py > ../../logs/ojk.log 2>&1 &
cd ../..

# 3. Nyalakan Nexus Core Gateway (Shield)
echo -e "${GREEN}[3/4] Mengaktifkan Nexus AI Gateway (Shield Layer)...${NC}"
export GOROOT=/home/taqy/Downloads/go
export PATH=$GOROOT/bin:$PATH
cd nexus-core-gateway
go run ./cmd/gateway/... > ../logs/gateway.log 2>&1 &
cd ..

# 4. Nyalakan Dashboard (Command Center)
echo -e "${GREEN}[4/4] Menyiapkan Nexus SOC Dashboard...${NC}"
export PATH=/home/taqy/Downloads/node/bin:$PATH
cd nexus-admin-dashboard
npm run dev > ../logs/dashboard.log 2>&1 &
cd ..

echo -e "${BLUE}==================================================${NC}"
echo -e "${GREEN}      SISTEM NEXUS BERHASIL DINYALAKAN!          ${NC}"
echo -e "${BLUE}==================================================${NC}"
echo -e "Akses Dashboard : ${GREEN}http://localhost:3000${NC}"
echo -e "Akses Gateway   : ${GREEN}http://localhost:8080${NC}"
echo -e "Cek Log Gateway : ${BLUE}tail -f logs/gateway.log${NC}"
echo -e "Cek Log OJK     : ${BLUE}tail -f logs/ojk.log${NC}"
echo -e "${BLUE}==================================================${NC}"
echo -e "Gunakan ${RED}./nexus-kill.sh${NC} untuk mematikan semua layanan."
