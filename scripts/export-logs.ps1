# Export logs from Docker containers for Azure App Insights or analysis

param(
    [string]$OutputDir = ".\logs",
    [int]$Lines = 1000
)

$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$logDir = Join-Path $OutputDir $timestamp

# Create output directory
New-Item -ItemType Directory -Force -Path $logDir | Out-Null

Write-Host "=========================================" -ForegroundColor Green
Write-Host "AgileOS Log Export" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Exporting logs to: $logDir" -ForegroundColor Cyan
Write-Host ""

# Export container logs
$containers = @("agileos-db", "agileos-nats")

foreach ($container in $containers) {
    Write-Host "Exporting logs from $container..." -ForegroundColor Cyan
    
    $logFile = Join-Path $logDir "$container.log"
    docker logs --tail $Lines $container > $logFile 2>&1
    
    if (Test-Path $logFile) {
        $size = (Get-Item $logFile).Length
        Write-Host "  Saved: $logFile ($size bytes)" -ForegroundColor Green
    } else {
        Write-Host "  Failed to export logs from $container" -ForegroundColor Red
    }
}

# Export container stats
Write-Host ""
Write-Host "Exporting container stats..." -ForegroundColor Cyan
$statsFile = Join-Path $logDir "container-stats.txt"
docker stats --no-stream --format "table {{.Name}}`t{{.CPUPerc}}`t{{.MemUsage}}`t{{.MemPerc}}`t{{.NetIO}}`t{{.BlockIO}}" > $statsFile

# Export container inspect
foreach ($container in $containers) {
    Write-Host "Exporting inspect data from $container..." -ForegroundColor Cyan
    $inspectFile = Join-Path $logDir "$container-inspect.json"
    docker inspect $container | Out-File -FilePath $inspectFile -Encoding UTF8
}

# Create summary
$summaryFile = Join-Path $logDir "summary.txt"
@"
AgileOS Log Export Summary
==========================
Timestamp: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
Export Directory: $logDir

Containers Exported:
$(foreach ($c in $containers) { "  - $c" })

Files Created:
$(Get-ChildItem $logDir | ForEach-Object { "  - $($_.Name) ($($_.Length) bytes)" })

Total Size: $((Get-ChildItem $logDir | Measure-Object -Property Length -Sum).Sum) bytes
"@ | Out-File -FilePath $summaryFile -Encoding UTF8

Write-Host ""
Write-Host "=========================================" -ForegroundColor Green
Write-Host "Export Complete!" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Summary saved to: $summaryFile" -ForegroundColor Cyan
Write-Host ""
Write-Host "To upload to Azure App Insights:" -ForegroundColor Yellow
Write-Host "  az monitor app-insights component create --app agileos --location eastus --resource-group agileos-rg" -ForegroundColor White
Write-Host "  # Then configure Application Insights instrumentation key in backend" -ForegroundColor White
Write-Host ""
