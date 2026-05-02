# Development script untuk menjalankan backend di local
# Services (SurrealDB & NATS) harus sudah running via docker-compose

$env:SURREAL_URL="ws://localhost:8002/rpc"
$env:SURREAL_USER="root"
$env:SURREAL_PASS="root"
$env:NATS_URL="nats://localhost:4223"
$env:PORT="8080"

Write-Host "🚀 Starting AgileOS Engine (Local Development Mode)" -ForegroundColor Green
Write-Host "   Database: $env:SURREAL_URL" -ForegroundColor Cyan
Write-Host "   NATS: $env:NATS_URL" -ForegroundColor Cyan
Write-Host "   Port: $env:PORT" -ForegroundColor Cyan
Write-Host ""

go run main.go
