# AgileOS Load Test Runner
# Runs performance monitoring and load testing simultaneously

param(
    [string]$TestType = "k6",  # k6 or bash
    [int]$Duration = 1200,     # 20 minutes
    [string]$BaseURL = "http://localhost:8081"
)

$ErrorActionPreference = "Stop"

Write-Host "🚀 AgileOS Load Test Runner" -ForegroundColor Cyan
Write-Host "============================" -ForegroundColor Cyan
Write-Host "Test Type: $TestType"
Write-Host "Duration: $Duration seconds"
Write-Host "Target: $BaseURL"
Write-Host ""

# Check if system is ready
Write-Host "Checking system readiness..." -ForegroundColor Yellow

# Check if Docker is running
$dockerRunning = docker ps 2>$null
if (-not $dockerRunning) {
    Write-Host "❌ Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

# Check if AgileOS containers are running
$agileosContainers = docker ps --filter "name=agileos" --format "{{.Names}}"
if (-not $agileosContainers) {
    Write-Host "❌ AgileOS containers are not running. Please run: docker-compose up -d" -ForegroundColor Red
    exit 1
}

Write-Host "✓ Docker is running" -ForegroundColor Green
Write-Host "✓ AgileOS containers are running" -ForegroundColor Green

# Check if backend is accessible
try {
    $health = Invoke-WebRequest -Uri "$BaseURL/health" -UseBasicParsing -TimeoutSec 5
    if ($health.StatusCode -eq 200) {
        Write-Host "✓ Backend is accessible" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Backend is not accessible at $BaseURL" -ForegroundColor Red
    Write-Host "   Please verify the backend is running and the URL is correct." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "System is ready for load testing!" -ForegroundColor Green
Write-Host ""

# Create results directory
$resultsDir = "load-test-results"
if (-not (Test-Path $resultsDir)) {
    New-Item -ItemType Directory -Path $resultsDir | Out-Null
}

$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"

# Start performance monitoring in background
Write-Host "Starting performance monitoring..." -ForegroundColor Yellow
$monitorJob = Start-Job -ScriptBlock {
    param($Duration, $ResultsDir, $Timestamp)
    Set-Location $using:PWD
    .\scripts\performance-monitor.ps1 -Duration $Duration -Interval 5 -OutputFile "$ResultsDir\performance-$Timestamp.csv"
} -ArgumentList $Duration, $resultsDir, $timestamp

Write-Host "✓ Performance monitoring started (Job ID: $($monitorJob.Id))" -ForegroundColor Green
Start-Sleep -Seconds 2

# Run load test
Write-Host ""
Write-Host "Starting load test..." -ForegroundColor Yellow
Write-Host ""

if ($TestType -eq "k6") {
    # Check if k6 is installed
    $k6Installed = Get-Command k6 -ErrorAction SilentlyContinue
    if (-not $k6Installed) {
        Write-Host "❌ k6 is not installed. Install with: choco install k6" -ForegroundColor Red
        Write-Host "   Or use bash test: .\RUN-LOAD-TEST.ps1 -TestType bash" -ForegroundColor Yellow
        Stop-Job -Job $monitorJob
        Remove-Job -Job $monitorJob
        exit 1
    }
    
    $env:BASE_URL = $BaseURL
    k6 run stress-test.js
    
} elseif ($TestType -eq "bash") {
    # Check if WSL or Git Bash is available
    $bashAvailable = Get-Command bash -ErrorAction SilentlyContinue
    if (-not $bashAvailable) {
        Write-Host "❌ Bash is not available. Install Git Bash or WSL." -ForegroundColor Red
        Write-Host "   Or use k6 test: .\RUN-LOAD-TEST.ps1 -TestType k6" -ForegroundColor Yellow
        Stop-Job -Job $monitorJob
        Remove-Job -Job $monitorJob
        exit 1
    }
    
    $env:BASE_URL = $BaseURL
    $env:CONCURRENT_USERS = "100"
    $env:TOTAL_REQUESTS = "5000"
    bash load-test.sh
    
} else {
    Write-Host "❌ Invalid test type: $TestType. Use 'k6' or 'bash'." -ForegroundColor Red
    Stop-Job -Job $monitorJob
    Remove-Job -Job $monitorJob
    exit 1
}

# Wait for monitoring to complete
Write-Host ""
Write-Host "Waiting for performance monitoring to complete..." -ForegroundColor Yellow
Wait-Job -Job $monitorJob | Out-Null
Receive-Job -Job $monitorJob
Remove-Job -Job $monitorJob

# Generate summary report
Write-Host ""
Write-Host "📊 Generating Summary Report" -ForegroundColor Cyan
Write-Host "=============================" -ForegroundColor Cyan

$reportFile = "$resultsDir\summary-$timestamp.txt"
$report = @"
AgileOS Load Test Summary
=========================
Date: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
Test Type: $TestType
Duration: $Duration seconds
Target: $BaseURL

Results Location:
- Performance Data: $resultsDir\performance-$timestamp.csv
- Load Test Results: $resultsDir\

Next Steps:
1. Review performance metrics in CSV file
2. Check k6 results for detailed statistics
3. Compare with baseline in PERFORMANCE-OPTIMIZATION.md
4. Identify bottlenecks and optimize

For detailed analysis, see: PERFORMANCE-OPTIMIZATION.md
"@

$report | Out-File -FilePath $reportFile -Encoding UTF8
Write-Host $report

Write-Host ""
Write-Host "✓ Load test completed successfully!" -ForegroundColor Green
Write-Host "Summary saved to: $reportFile" -ForegroundColor Cyan
