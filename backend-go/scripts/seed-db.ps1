# Script untuk seed database dengan sample workflow
# Pastikan SurrealDB sudah running dan schema sudah di-apply

param(
    [string]$SurrealUrl = "http://localhost:8002",
    [string]$Username = "root",
    [string]$Password = "root",
    [string]$Namespace = "agileos",
    [string]$Database = "main"
)

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "Seeding AgileOS Database..." -ForegroundColor Green
Write-Host ""

# Check if SurrealDB is running
Write-Host "Checking SurrealDB connection..." -ForegroundColor Cyan
try {
    $health = Invoke-WebRequest -Uri "$SurrealUrl/health" -Method GET -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "   [OK] SurrealDB is running at $SurrealUrl" -ForegroundColor Green
} catch {
    Write-Host "   [ERROR] SurrealDB is not running at $SurrealUrl" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please start SurrealDB first:" -ForegroundColor Yellow
    Write-Host "   docker-compose up -d agileos-db" -ForegroundColor White
    Write-Host ""
    Write-Host "Or run fresh start script:" -ForegroundColor Yellow
    Write-Host "   .\FRESH-START-DB.ps1" -ForegroundColor White
    Write-Host ""
    exit 1
}
Write-Host ""

# Check if seed file exists
$seedFile = Join-Path $PSScriptRoot "..\database\seed.surql"
if (-not (Test-Path $seedFile)) {
    Write-Host "[ERROR] Seed file not found: $seedFile" -ForegroundColor Red
    exit 1
}

Write-Host "Seed file found: $seedFile" -ForegroundColor Cyan
Write-Host ""

# Method 1: Try using docker exec (recommended)
Write-Host "Method 1: Using Docker Exec..." -ForegroundColor Cyan
$seedSuccess = $false
try {
    # Check if container is running
    $containerRunning = docker ps --filter "name=agileos-db" --format "{{.Names}}" 2>$null
    
    if ($containerRunning -eq "agileos-db") {
        Write-Host "   [OK] Container 'agileos-db' is running" -ForegroundColor Green
        
        # Copy seed file to container
        Write-Host "   Copying seed file to container..." -ForegroundColor Gray
        docker cp $seedFile agileos-db:/tmp/seed.surql
        
        # Execute seed
        Write-Host "   Executing seed script..." -ForegroundColor Gray
        $output = docker exec agileos-db surreal sql `
            --endpoint http://localhost:8000 `
            --username $Username `
            --password $Password `
            --namespace $Namespace `
            --database $Database `
            --file /tmp/seed.surql 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "   [OK] Seed executed successfully via Docker!" -ForegroundColor Green
            $seedSuccess = $true
        } else {
            Write-Host "   [WARN] Docker exec completed with warnings" -ForegroundColor Yellow
            Write-Host "   Trying alternative method..." -ForegroundColor Gray
            $seedSuccess = $false
        }
    } else {
        Write-Host "   [WARN] Container not found, trying HTTP API..." -ForegroundColor Yellow
        $seedSuccess = $false
    }
} catch {
    Write-Host "   [WARN] Docker method failed: $_" -ForegroundColor Yellow
    Write-Host "   Trying alternative method..." -ForegroundColor Gray
    $seedSuccess = $false
}
Write-Host ""

# Method 2: HTTP API fallback
if (-not $seedSuccess) {
    Write-Host "Method 2: Using HTTP API..." -ForegroundColor Cyan
    try {
        # Read seed content
        $seedContent = Get-Content $seedFile -Raw
        
        # Prepare headers
        $auth = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("${Username}:${Password}"))
        $headers = @{
            "Accept" = "application/json"
            "NS" = $Namespace
            "DB" = $Database
            "Authorization" = "Basic $auth"
        }
        
        # Execute seed
        Write-Host "   Executing seed script via HTTP..." -ForegroundColor Gray
        $result = Invoke-RestMethod -Uri "$SurrealUrl/sql" `
            -Method POST `
            -Headers $headers `
            -Body $seedContent `
            -ContentType "text/plain" `
            -UseBasicParsing
        
        Write-Host "   [OK] Seed executed successfully via HTTP!" -ForegroundColor Green
        $seedSuccess = $true
    } catch {
        Write-Host "   [ERROR] HTTP method failed: $_" -ForegroundColor Red
        Write-Host ""
        Write-Host "Please try manual seeding:" -ForegroundColor Yellow
        Write-Host "   1. Open: $SurrealUrl" -ForegroundColor White
        Write-Host "   2. Login with: $Username / $Password" -ForegroundColor White
        Write-Host "   3. Use namespace: $Namespace, database: $Database" -ForegroundColor White
        Write-Host "   4. Copy content from: $seedFile" -ForegroundColor White
        Write-Host "   5. Execute the queries" -ForegroundColor White
        Write-Host ""
        exit 1
    }
}
Write-Host ""

# Verify seeded data
if ($seedSuccess) {
    Write-Host "Verifying seeded data..." -ForegroundColor Cyan
    try {
        $auth = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("${Username}:${Password}"))
        $headers = @{
            "Accept" = "application/json"
            "NS" = $Namespace
            "DB" = $Database
            "Authorization" = "Basic $auth"
        }
        
        # Query workflows
        $workflowResult = Invoke-RestMethod -Uri "$SurrealUrl/sql" `
            -Method POST `
            -Headers $headers `
            -Body "SELECT * FROM workflow;" `
            -ContentType "text/plain" `
            -UseBasicParsing
        
        # Query steps
        $stepResult = Invoke-RestMethod -Uri "$SurrealUrl/sql" `
            -Method POST `
            -Headers $headers `
            -Body "SELECT count() AS total FROM step;" `
            -ContentType "text/plain" `
            -UseBasicParsing
        
        # Query process instances
        $processResult = Invoke-RestMethod -Uri "$SurrealUrl/sql" `
            -Method POST `
            -Headers $headers `
            -Body "SELECT * FROM process_instance;" `
            -ContentType "text/plain" `
            -UseBasicParsing
        
        Write-Host ""
        Write-Host "Seeded Data Summary:" -ForegroundColor Green
        Write-Host "========================================" -ForegroundColor Gray
        
        if ($workflowResult[0].result.Count -gt 0) {
            $workflow = $workflowResult[0].result[0]
            Write-Host "   Workflow:" -ForegroundColor Cyan
            Write-Host "      ID: $($workflow.id)" -ForegroundColor White
            Write-Host "      Name: $($workflow.name)" -ForegroundColor White
            Write-Host "      Description: $($workflow.description)" -ForegroundColor White
        }
        
        Write-Host ""
        Write-Host "   Steps: $($stepResult[0].result[0].total)" -ForegroundColor Cyan
        
        if ($processResult[0].result.Count -gt 0) {
            $process = $processResult[0].result[0]
            Write-Host ""
            Write-Host "   Process Instance:" -ForegroundColor Cyan
            Write-Host "      ID: $($process.id)" -ForegroundColor White
            Write-Host "      Status: $($process.status)" -ForegroundColor White
            Write-Host "      Current Step: $($process.current_step_id)" -ForegroundColor White
        }
        
        Write-Host "========================================" -ForegroundColor Gray
        Write-Host ""
        Write-Host "[SUCCESS] Database seeded successfully!" -ForegroundColor Green
        
    } catch {
        Write-Host "   [WARN] Could not verify data: $_" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host "   1. Verify data in SurrealDB" -ForegroundColor White
Write-Host "   2. Test graph traversal" -ForegroundColor White
Write-Host "   3. Start the backend: cd backend-go; go run main.go" -ForegroundColor White
Write-Host ""
