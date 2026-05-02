#!/usr/bin/env pwsh

Write-Host "🚀 AgileOS WebSocket Quick Start" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan

# Check if Docker is running
Write-Host "`n1. Checking Docker..." -ForegroundColor Yellow
try {
    docker version | Out-Null
    Write-Host "✓ Docker is running" -ForegroundColor Green
} catch {
    Write-Host "✗ Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

# Start backend services
Write-Host "`n2. Starting backend services..." -ForegroundColor Yellow
Set-Location "agile-os"
docker-compose up -d agileos-db agileos-nats

# Wait for services to be ready
Write-Host "   Waiting for services to start..." -ForegroundColor Gray
Start-Sleep -Seconds 10

# Build and start backend
Write-Host "`n3. Building and starting Go backend..." -ForegroundColor Yellow
Set-Location "backend-go"
go mod tidy
go build -o agileos-backend .
Start-Process -FilePath "./agileos-backend" -WindowStyle Hidden

# Wait for backend to start
Write-Host "   Waiting for backend to start..." -ForegroundColor Gray
Start-Sleep -Seconds 5

# Seed database
Write-Host "`n4. Seeding database..." -ForegroundColor Yellow
try {
    .\scripts\seed-db.ps1
    .\scripts\seed-users.ps1
    Write-Host "✓ Database seeded successfully" -ForegroundColor Green
} catch {
    Write-Host "⚠ Database seeding may have failed, continuing..." -ForegroundColor Yellow
}

# Start frontend
Write-Host "`n5. Starting Next.js frontend..." -ForegroundColor Yellow
Set-Location "..\frontend-next"
Start-Process -FilePath "npm" -ArgumentList "run", "dev" -WindowStyle Hidden

Write-Host "`n6. Testing WebSocket functionality..." -ForegroundColor Yellow
Set-Location "..\backend-go"
Start-Sleep -Seconds 10
.\scripts\test-websocket.ps1

Write-Host "`n🎉 WebSocket Quick Start Complete!" -ForegroundColor Green
Write-Host "=================================" -ForegroundColor Green
Write-Host "`nNext Steps:" -ForegroundColor Cyan
Write-Host "1. Open http://localhost:3001 in your browser" -ForegroundColor White
Write-Host "2. Look for the connection status and notification bell" -ForegroundColor White
Write-Host "3. Open Developer Tools to see WebSocket logs" -ForegroundColor White
Write-Host "4. The test script created a sample task - check for notifications!" -ForegroundColor White

Write-Host "`nServices Running:" -ForegroundColor Cyan
Write-Host "- Backend API: http://localhost:8081" -ForegroundColor Gray
Write-Host "- Frontend: http://localhost:3001" -ForegroundColor Gray
Write-Host "- WebSocket: ws://localhost:8081/ws" -ForegroundColor Gray
Write-Host "- SurrealDB: http://localhost:8002" -ForegroundColor Gray
Write-Host "- NATS: nats://localhost:4223" -ForegroundColor Gray

Write-Host "`nTo stop services:" -ForegroundColor Yellow
Write-Host "docker-compose down" -ForegroundColor Gray