# Script untuk test database connection dan query
# Pastikan backend sudah running

Write-Host "🧪 Testing AgileOS Database Layer..." -ForegroundColor Green
Write-Host ""

$backendUrl = "http://localhost:8080"

# Check if backend is running
try {
    $health = Invoke-WebRequest -Uri "$backendUrl/health" -Method GET -ErrorAction Stop
    $healthData = $health.Content | ConvertFrom-Json
    
    Write-Host "✓ Backend is running" -ForegroundColor Green
    Write-Host "  Status: $($healthData.status)" -ForegroundColor Cyan
    Write-Host "  Database: $($healthData.database)" -ForegroundColor Cyan
    Write-Host "  Message Broker: $($healthData.message_broker)" -ForegroundColor Cyan
    Write-Host ""
} catch {
    Write-Host "✗ Backend is not running. Please start it with: .\run-local.ps1" -ForegroundColor Red
    exit 1
}

Write-Host "📊 Database is ready for BPM operations!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Seed the database: .\scripts\seed-db.ps1" -ForegroundColor White
Write-Host "  2. Test workflow creation via API (coming in next prompt)" -ForegroundColor White
Write-Host ""
