# Script untuk seed database dengan sample workflow
# Pastikan SurrealDB sudah running (docker-compose up -d)

Write-Host "🌱 Seeding AgileOS Database..." -ForegroundColor Green
Write-Host ""

$surrealUrl = "http://localhost:8000"
$user = "root"
$pass = "root"

# Check if SurrealDB is running
try {
    $health = Invoke-WebRequest -Uri "$surrealUrl/health" -Method GET -ErrorAction Stop
    Write-Host "✓ SurrealDB is running" -ForegroundColor Green
} catch {
    Write-Host "✗ SurrealDB is not running. Please start it with: docker-compose up -d" -ForegroundColor Red
    exit 1
}

# Read seed file
$seedFile = Join-Path $PSScriptRoot "..\database\seed.surql"
if (-not (Test-Path $seedFile)) {
    Write-Host "✗ Seed file not found: $seedFile" -ForegroundColor Red
    exit 1
}

$seedContent = Get-Content $seedFile -Raw

Write-Host "📄 Seed file loaded: $seedFile" -ForegroundColor Cyan
Write-Host ""
Write-Host "⚠️  Manual Step Required:" -ForegroundColor Yellow
Write-Host "   1. Open SurrealDB Dashboard: $surrealUrl" -ForegroundColor White
Write-Host "   2. Login with username: $user, password: $pass" -ForegroundColor White
Write-Host "   3. Copy and paste the content from: database/seed.surql" -ForegroundColor White
Write-Host "   4. Execute the queries" -ForegroundColor White
Write-Host ""
Write-Host "Or use SurrealDB CLI if installed:" -ForegroundColor Cyan
Write-Host "   surreal sql --conn $surrealUrl --user $user --pass $pass --ns agileos --db main --file database/seed.surql" -ForegroundColor White
Write-Host ""
Write-Host "After seeding, verify with:" -ForegroundColor Cyan
Write-Host '   SELECT * FROM workflow;' -ForegroundColor White
Write-Host '   SELECT * FROM step;' -ForegroundColor White
Write-Host ""
