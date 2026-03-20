#!/bin/bash

# NEXUS CYBER - SCAFFOLDING SCRIPT (Bash)
# Digunakan untuk inisiasi cepat arsitektur folder modular.

echo "🚀 Initiating Nexus Cyber Scaffolding..."

# 1. Create Core Gateway Structure (Go)
mkdir -p nexus-core-gateway/cmd/gateway
mkdir -p nexus-core-gateway/internal/ai
mkdir -p nexus-core-gateway/internal/mtd
mkdir -p nexus-core-gateway/internal/crypto
mkdir -p nexus-core-gateway/internal/proxy
mkdir -p nexus-core-gateway/internal/repair
mkdir -p nexus-core-gateway/pkg
mkdir -p nexus-core-gateway/configs

# 2. Create Admin Dashboard Structure (Next.js Boilerplate placeholder)
mkdir -p nexus-admin-dashboard/app
mkdir -p nexus-admin-dashboard/components
mkdir -p nexus-admin-dashboard/lib
mkdir -p nexus-admin-dashboard/public

# 3. Create Supporting Folders
mkdir -p scripts
mkdir -p docs

# 4. Create placeholders
touch nexus-core-gateway/README.md
touch nexus-admin-dashboard/README.md
touch docs/DATABASE_SCHEMA.md

echo "✅ Scaffolding complete. Arsitektur modular siap digunakan."
