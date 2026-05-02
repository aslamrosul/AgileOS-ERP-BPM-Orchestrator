#!/usr/bin/env pwsh
# Apply Accounting Module Schema to SurrealDB

Write-Host "🚀 Applying Accounting Module Schema to SurrealDB..." -ForegroundColor Cyan

# Configuration
$SURREAL_URL = "http://localhost:8000"
$SURREAL_USER = "root"
$SURREAL_PASS = "root"
$NAMESPACE = "agileos"
$DATABASE = "main"
$SCHEMA_FILE = "../database/schema-accounting.surql"

# Check if SurrealDB is running
Write-Host "📡 Checking SurrealDB connection..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$SURREAL_URL/health" -Method GET -ErrorAction Stop
    Write-Host "✓ SurrealDB is running" -ForegroundColor Green
} catch {
    Write-Host "✗ SurrealDB is not running!" -ForegroundColor Red
    Write-Host "Please start SurrealDB first with: docker-compose up -d agileos-db" -ForegroundColor Yellow
    exit 1
}

# Check if schema file exists
if (-not (Test-Path $SCHEMA_FILE)) {
    Write-Host "✗ Schema file not found: $SCHEMA_FILE" -ForegroundColor Red
    exit 1
}

Write-Host "✓ Schema file found: $SCHEMA_FILE" -ForegroundColor Green

# Read schema file
$schemaContent = Get-Content $SCHEMA_FILE -Raw

# Apply schema using SurrealDB SQL endpoint
Write-Host "📝 Applying accounting schema..." -ForegroundColor Yellow

$headers = @{
    "Accept" = "application/json"
    "NS" = $NAMESPACE
    "DB" = $DATABASE
}

$auth = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("${SURREAL_USER}:${SURREAL_PASS}"))
$headers["Authorization"] = "Basic $auth"

try {
    $response = Invoke-RestMethod -Uri "$SURREAL_URL/sql" -Method POST -Headers $headers -Body $schemaContent -ContentType "text/plain"
    
    Write-Host "✓ Accounting schema applied successfully!" -ForegroundColor Green
    
    # Display results
    Write-Host "`n📊 Schema Application Results:" -ForegroundColor Cyan
    Write-Host "================================" -ForegroundColor Cyan
    
    if ($response) {
        $response | ForEach-Object {
            if ($_.status -eq "OK") {
                Write-Host "✓ " -NoNewline -ForegroundColor Green
            } else {
                Write-Host "✗ " -NoNewline -ForegroundColor Red
            }
            Write-Host $_.result
        }
    }
    
    Write-Host "`n✅ Accounting Module Schema Applied!" -ForegroundColor Green
    Write-Host "`n📋 Created Tables:" -ForegroundColor Cyan
    Write-Host "  - account (Chart of Accounts)" -ForegroundColor White
    Write-Host "  - journal_entry (General Ledger Header)" -ForegroundColor White
    Write-Host "  - journal_line (General Ledger Lines)" -ForegroundColor White
    Write-Host "  - vendor (Vendor Master)" -ForegroundColor White
    Write-Host "  - customer (Customer Master)" -ForegroundColor White
    Write-Host "  - purchase_invoice (Account Payable)" -ForegroundColor White
    Write-Host "  - sales_invoice (Account Receivable)" -ForegroundColor White
    Write-Host "  - payment (Payment Transactions)" -ForegroundColor White
    Write-Host "  - budget (Budget Management)" -ForegroundColor White
    
    Write-Host "`n🔧 Next Steps:" -ForegroundColor Cyan
    Write-Host "  1. Restart backend: docker-compose restart agileos-backend" -ForegroundColor Yellow
    Write-Host "  2. Test API endpoints: http://localhost:8080/swagger/index.html" -ForegroundColor Yellow
    Write-Host "  3. Create sample accounts: Run seed-accounting.ps1" -ForegroundColor Yellow
    
} catch {
    Write-Host "✗ Failed to apply schema!" -ForegroundColor Red
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
}
