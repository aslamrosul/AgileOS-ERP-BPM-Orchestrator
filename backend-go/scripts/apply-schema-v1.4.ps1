# Apply SurrealDB v1.4 Schema to AgileOS
# This script applies the new schema with TYPE RELATION and IF NOT EXISTS features

param(
    [string]$SurrealURL = "http://localhost:8002",
    [string]$Username = "root",
    [string]$Password = "root",
    [string]$Namespace = "agileos",
    [string]$Database = "main"
)

$ErrorActionPreference = "Stop"

Write-Host "AgileOS Schema v1.4 Deployment" -ForegroundColor Cyan
Write-Host "===================================" -ForegroundColor Cyan
Write-Host "Target: $SurrealURL"
Write-Host "Namespace: $Namespace"
Write-Host "Database: $Database"
Write-Host ""

# Check if SurrealDB is accessible
Write-Host "Checking SurrealDB connectivity..." -ForegroundColor Yellow
try {
    $health = Invoke-WebRequest -Uri "$SurrealURL/health" -UseBasicParsing -TimeoutSec 5
    if ($health.StatusCode -eq 200) {
        Write-Host "[OK] SurrealDB is accessible" -ForegroundColor Green
    }
} catch {
    Write-Host "[ERROR] Cannot connect to SurrealDB at $SurrealURL" -ForegroundColor Red
    Write-Host "        Please ensure SurrealDB is running" -ForegroundColor Red
    exit 1
}

# Function to execute SurrealQL file
function Invoke-SurrealQLFile {
    param(
        [string]$FilePath,
        [string]$Description
    )
    
    Write-Host ""
    Write-Host "$Description" -ForegroundColor Yellow
    Write-Host "File: $FilePath"
    
    if (-not (Test-Path $FilePath)) {
        Write-Host "[ERROR] File not found: $FilePath" -ForegroundColor Red
        return $false
    }
    
    $content = Get-Content -Path $FilePath -Raw
    
    try {
        $base64Auth = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("${Username}:${Password}"))
        $response = Invoke-RestMethod -Uri "$SurrealURL/sql" `
            -Method POST `
            -Headers @{
                "Accept" = "application/json"
                "Authorization" = "Basic $base64Auth"
                "NS" = $Namespace
                "DB" = $Database
            } `
            -Body $content `
            -ContentType "application/json"
        
        Write-Host "[OK] Successfully applied" -ForegroundColor Green
        return $true
    } catch {
        Write-Host "[ERROR] Error applying schema: $_" -ForegroundColor Red
        return $false
    }
}

# Apply schema
$schemaPath = Join-Path $PSScriptRoot "..\database\schema-v1.4.surql"
$schemaApplied = Invoke-SurrealQLFile -FilePath $schemaPath -Description "Applying Schema v1.4 (TYPE RELATION, IF NOT EXISTS)"

if (-not $schemaApplied) {
    Write-Host ""
    Write-Host "[ERROR] Schema application failed. Aborting." -ForegroundColor Red
    exit 1
}

# Ask if user wants to seed data
Write-Host ""
$seedData = Read-Host "Do you want to seed sample data? (y/n)"

if ($seedData -eq "y" -or $seedData -eq "Y") {
    $seedPath = Join-Path $PSScriptRoot "..\database\seed-v1.4.surql"
    $seedApplied = Invoke-SurrealQLFile -FilePath $seedPath -Description "Seeding Sample Data"
    
    if ($seedApplied) {
        Write-Host ""
        Write-Host "[OK] Sample data seeded successfully" -ForegroundColor Green
    }
}

# Verify schema
Write-Host ""
Write-Host "Verifying Schema..." -ForegroundColor Yellow

try {
    $base64Auth = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("${Username}:${Password}"))
    
    $verifyQuery = 'INFO FOR DB;'
    
    $response = Invoke-RestMethod -Uri "$SurrealURL/sql" `
        -Method POST `
        -Headers @{
            "Accept" = "application/json"
            "Authorization" = "Basic $base64Auth"
            "NS" = $Namespace
            "DB" = $Database
        } `
        -Body $verifyQuery `
        -ContentType "application/json"
    
    Write-Host "[OK] Schema verification completed" -ForegroundColor Green
    
    # Count tables - using single quotes to avoid parsing issues
    $countQuery = 'SELECT count() FROM user; SELECT count() FROM workflow; SELECT count() FROM step; SELECT count() FROM next; SELECT count() FROM process_instance; SELECT count() FROM task_instance; SELECT count() FROM audit_trails;'
    
    $counts = Invoke-RestMethod -Uri "$SurrealURL/sql" `
        -Method POST `
        -Headers @{
            "Accept" = "application/json"
            "Authorization" = "Basic $base64Auth"
            "NS" = $Namespace
            "DB" = $Database
        } `
        -Body $countQuery `
        -ContentType "application/json"
    
    Write-Host ""
    Write-Host "Database Statistics:" -ForegroundColor Cyan
    
    if ($counts -and $counts.Count -ge 7) {
        try {
            Write-Host "   Users: $($counts[0].result[0].count)"
            Write-Host "   Workflows: $($counts[1].result[0].count)"
            Write-Host "   Steps: $($counts[2].result[0].count)"
            Write-Host "   Graph Edges (next): $($counts[3].result[0].count)"
            Write-Host "   Process Instances: $($counts[4].result[0].count)"
            Write-Host "   Task Instances: $($counts[5].result[0].count)"
            Write-Host "   Audit Trails: $($counts[6].result[0].count)"
        } catch {
            Write-Host "   Statistics not available yet (tables may be empty)" -ForegroundColor Yellow
        }
    } else {
        Write-Host "   Statistics not available yet (tables may be empty)" -ForegroundColor Yellow
    }
    
} catch {
    Write-Host "[WARNING] Verification warning: $_" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "[SUCCESS] Schema v1.4 Deployment Completed!" -ForegroundColor Green
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host "1. Restart your backend: docker-compose restart agileos-backend"
Write-Host "2. Test the new schema features"
Write-Host "3. Check audit trails: SELECT * FROM audit_trails;"
Write-Host "4. Test graph traversal: SELECT ->next->step.* FROM step:manager_approval;"
Write-Host ""
Write-Host "New Features in v1.4:" -ForegroundColor Yellow
Write-Host "[+] TYPE RELATION for graph edges (next table)"
Write-Host "[+] TYPE NORMAL for regular tables"
Write-Host "[+] IF NOT EXISTS for all DEFINE statements"
Write-Host "[+] Automated audit trail events"
Write-Host "[+] Custom functions (is_admin, can_approve, etc.)"
Write-Host "[+] Full-text search indexes"
Write-Host "[+] Improved permissions model"
