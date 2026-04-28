# AgileOS BPM - Quick Start Script
# Quickly test production setup locally

Write-Host "AgileOS BPM - Local Production Test" -ForegroundColor Green
Write-Host "====================================" -ForegroundColor Green
Write-Host ""

# Check if .env exists
if (-not (Test-Path ".env")) {
    Write-Host "Creating .env from template..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env"
    Write-Host "Please edit .env file with your configuration" -ForegroundColor Yellow
    Write-Host "Press Enter to continue after editing .env..." -ForegroundColor Yellow
    Read-Host
}

# Build and start containers
Write-Host "Building and starting containers..." -ForegroundColor Cyan
docker-compose -f docker-compose.prod.yml up --build -d

# Wait for services to be ready
Write-Host "Waiting for services to start..." -ForegroundColor Cyan
Start-Sleep -Seconds 15

# Check health
Write-Host ""
Write-Host "Checking service health..." -ForegroundColor Cyan

$services = @(
    @{Name="Nginx"; URL="http://localhost:8090/health"},
    @{Name="Backend"; URL="http://localhost:8090/api/v1/health"},
    @{Name="Frontend"; URL="http://localhost:3001"},
    @{Name="SurrealDB"; URL="http://localhost:8002/health"},
    @{Name="NATS"; URL="http://localhost:8223/healthz"}
)

foreach ($service in $services) {
    try {
        $response = Invoke-WebRequest -Uri $service.URL -UseBasicParsing -TimeoutSec 5
        if ($response.StatusCode -eq 200) {
            Write-Host "  $($service.Name): OK" -ForegroundColor Green
        }
    } catch {
        Write-Host "  $($service.Name): FAILED" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "====================================" -ForegroundColor Green
Write-Host "AgileOS BPM is running!" -ForegroundColor Green
Write-Host "====================================" -ForegroundColor Green
Write-Host ""
Write-Host "Access points:"
Write-Host "  Frontend:  http://localhost:8090" -ForegroundColor Cyan
Write-Host "  Backend:   http://localhost:8090/api" -ForegroundColor Cyan
Write-Host "  Database:  http://localhost:8002" -ForegroundColor Cyan
Write-Host "  NATS:      http://localhost:8223" -ForegroundColor Cyan
Write-Host ""
Write-Host "To stop: docker-compose -f docker-compose.prod.yml down" -ForegroundColor Yellow
Write-Host "To view logs: docker-compose -f docker-compose.prod.yml logs -f" -ForegroundColor Yellow
Write-Host ""
