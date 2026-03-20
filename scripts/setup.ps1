# NEXUS CYBER - SCAFFOLDING SCRIPT (PowerShell)
# Cocok untuk OS Windows.

Write-Host "🚀 Initiating Nexus Cyber Scaffolding..." -ForegroundColor Cyan

$folders = @(
    "nexus-core-gateway/cmd/gateway",
    "nexus-core-gateway/internal/ai",
    "nexus-core-gateway/internal/mtd",
    "nexus-core-gateway/internal/crypto",
    "nexus-core-gateway/internal/proxy",
    "nexus-core-gateway/internal/repair",
    "nexus-core-gateway/pkg",
    "nexus-core-gateway/configs",
    "nexus-admin-dashboard/app",
    "nexus-admin-dashboard/components",
    "nexus-admin-dashboard/lib",
    "nexus-admin-dashboard/public",
    "scripts",
    "docs"
)

foreach ($folder in $folders) {
    if (!(Test-Path $folder)) {
        New-Item -ItemType Directory -Path $folder -Force | Out-Null
        Write-Host "Created: $folder"
    }
}

New-Item -ItemType File -Path "nexus-core-gateway/README.md" -Force | Out-Null
New-Item -ItemType File -Path "nexus-admin-dashboard/README.md" -Force | Out-Null
New-Item -ItemType File -Path "docs/DATABASE_SCHEMA.md" -Force | Out-Null

Write-Host "✅ Scaffolding complete. Arsitektur modular siap digunakan." -ForegroundColor Green
