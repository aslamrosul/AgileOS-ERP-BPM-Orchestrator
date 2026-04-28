# AgileOS Container Monitoring Script (PowerShell)
# Monitors Docker container resource usage in real-time

Write-Host "=========================================" -ForegroundColor Green
Write-Host "AgileOS Container Resource Monitor" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Press Ctrl+C to stop monitoring" -ForegroundColor Yellow
Write-Host ""

# Check if docker is running
try {
    docker info | Out-Null
} catch {
    Write-Host "Error: Docker is not running" -ForegroundColor Red
    exit 1
}

# Monitor containers
while ($true) {
    Clear-Host
    Write-Host "=========================================" -ForegroundColor Green
    Write-Host "AgileOS Container Resource Monitor" -ForegroundColor Green
    Write-Host "Time: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Cyan
    Write-Host "=========================================" -ForegroundColor Green
    Write-Host ""
    
    # Show container stats
    Write-Host "Container Resource Usage:" -ForegroundColor Cyan
    docker stats --no-stream --format "table {{.Name}}`t{{.CPUPerc}}`t{{.MemUsage}}`t{{.MemPerc}}`t{{.NetIO}}" agileos-db agileos-nats 2>$null
    
    Write-Host ""
    Write-Host "=========================================" -ForegroundColor Green
    Write-Host "Container Health Status" -ForegroundColor Cyan
    Write-Host "=========================================" -ForegroundColor Green
    
    # Check health status
    $containers = @("agileos-db", "agileos-nats")
    foreach ($container in $containers) {
        $running = docker ps --filter "name=$container" --format "{{.Names}}" 2>$null
        if ($running) {
            $health = docker inspect --format='{{.State.Health.Status}}' $container 2>$null
            if (-not $health) { $health = "no healthcheck" }
            $status = docker inspect --format='{{.State.Status}}' $container 2>$null
            
            $color = "Green"
            if ($status -ne "running") { $color = "Red" }
            elseif ($health -eq "unhealthy") { $color = "Yellow" }
            
            Write-Host "$container : $status (health: $health)" -ForegroundColor $color
        } else {
            Write-Host "$container : NOT RUNNING" -ForegroundColor Red
        }
    }
    
    Write-Host ""
    Write-Host "=========================================" -ForegroundColor Green
    Write-Host "Backend Health Check" -ForegroundColor Cyan
    Write-Host "=========================================" -ForegroundColor Green
    
    try {
        $health = Invoke-RestMethod -Uri "http://localhost:8081/health" -Method GET -TimeoutSec 2
        Write-Host "Backend Status: $($health.status)" -ForegroundColor Green
        Write-Host "  Database: $($health.database.status) ($($health.database.response_time_ms)ms)" -ForegroundColor Gray
        Write-Host "  NATS: $($health.message_broker.status) ($($health.message_broker.response_time_ms)ms)" -ForegroundColor Gray
        Write-Host "  Uptime: $($health.uptime_seconds)s" -ForegroundColor Gray
    } catch {
        Write-Host "Backend Status: OFFLINE" -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "Refreshing in 5 seconds... (Ctrl+C to stop)" -ForegroundColor Yellow
    Start-Sleep -Seconds 5
}
