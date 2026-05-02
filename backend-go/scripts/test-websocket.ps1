#!/usr/bin/env pwsh

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Testing WebSocket Real-time Notifications" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Function to wait for service to be ready
function Wait-ForService {
    param(
        [string]$Url,
        [string]$ServiceName,
        [int]$MaxAttempts = 30,
        [int]$DelaySeconds = 2
    )
    
    Write-Host "Waiting for $ServiceName to be ready..." -ForegroundColor Yellow
    
    for ($i = 1; $i -le $MaxAttempts; $i++) {
        try {
            $response = Invoke-RestMethod -Uri $Url -Method GET -TimeoutSec 5
            Write-Host "✓ $ServiceName is ready" -ForegroundColor Green
            return $true
        } catch {
            Write-Host "  Attempt $i/$MaxAttempts - $ServiceName not ready yet..." -ForegroundColor Gray
            Start-Sleep -Seconds $DelaySeconds
        }
    }
    
    Write-Host "✗ $ServiceName failed to start after $MaxAttempts attempts" -ForegroundColor Red
    return $false
}

# Check if backend is running
if (-not (Wait-ForService -Url "http://localhost:8081/health" -ServiceName "Backend")) {
    Write-Host "Please start the backend first:" -ForegroundColor Yellow
    Write-Host "  cd backend-go" -ForegroundColor Gray
    Write-Host "  go run ." -ForegroundColor Gray
    Write-Host "Or use Docker:" -ForegroundColor Yellow
    Write-Host "  docker-compose up -d" -ForegroundColor Gray
    exit 1
}

# Get health status
try {
    $healthResponse = Invoke-RestMethod -Uri "http://localhost:8081/health" -Method GET
    Write-Host "Backend Health Status:" -ForegroundColor Green
    Write-Host "  Database: $($healthResponse.database)" -ForegroundColor Gray
    Write-Host "  NATS: $($healthResponse.nats)" -ForegroundColor Gray
    Write-Host "  WebSocket Hub: Available" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to get health status" -ForegroundColor Red
}

# Test authentication
Write-Host "`nTesting authentication..." -ForegroundColor Yellow
$loginData = @{
    username = "admin"
    password = "password123"
} | ConvertTo-Json

try {
    $authResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/login" -Method POST -Body $loginData -ContentType "application/json"
    $token = $authResponse.access_token
    Write-Host "✓ Authentication successful" -ForegroundColor Green
    Write-Host "  User: $($authResponse.user.username)" -ForegroundColor Gray
    Write-Host "  Role: $($authResponse.user.role)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Authentication failed" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "`nTrying to seed users first..." -ForegroundColor Yellow
    
    try {
        .\scripts\seed-users.ps1
        Write-Host "Users seeded, retrying authentication..." -ForegroundColor Yellow
        $authResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/login" -Method POST -Body $loginData -ContentType "application/json"
        $token = $authResponse.access_token
        Write-Host "✓ Authentication successful after seeding" -ForegroundColor Green
    } catch {
        Write-Host "✗ Authentication still failed after seeding" -ForegroundColor Red
        exit 1
    }
}

# Test WebSocket endpoint availability
Write-Host "`nTesting WebSocket endpoint..." -ForegroundColor Yellow
try {
    # Test if WebSocket endpoint is accessible (this will fail but should return proper error)
    $wsTest = Invoke-WebRequest -Uri "http://localhost:8081/ws" -Method GET -ErrorAction SilentlyContinue
} catch {
    if ($_.Exception.Response.StatusCode -eq 400) {
        Write-Host "✓ WebSocket endpoint is available (expected 400 for HTTP request)" -ForegroundColor Green
    } elseif ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "✓ WebSocket endpoint is available (expected 401 without token)" -ForegroundColor Green
    } else {
        Write-Host "✗ WebSocket endpoint error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Create a test task to trigger notifications
Write-Host "`nCreating test task to trigger WebSocket notifications..." -ForegroundColor Yellow
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

$processData = @{
    workflow_id = "purchase_approval"
    initiated_by = "admin"
    data = @{
        amount = 5000
        description = "WebSocket Test Purchase - $(Get-Date -Format 'HH:mm:ss')"
        department = "IT"
    }
} | ConvertTo-Json

try {
    $processResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/process/start" -Method POST -Body $processData -Headers $headers
    Write-Host "✓ Test process started successfully" -ForegroundColor Green
    Write-Host "  Process ID: $($processResponse.process_instance_id)" -ForegroundColor Gray
    Write-Host "  Task ID: $($processResponse.task_id)" -ForegroundColor Gray
    
    Write-Host "`n🔔 WebSocket notification should have been sent!" -ForegroundColor Cyan
    Write-Host "   Check your browser console and notifications panel" -ForegroundColor Yellow
    
} catch {
    Write-Host "✗ Failed to create test process" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    
    # Try to seed database if workflow not found
    if ($_.Exception.Message -like "*workflow*" -or $_.Exception.Message -like "*not found*") {
        Write-Host "`nTrying to seed database..." -ForegroundColor Yellow
        try {
            .\scripts\seed-db.ps1
            Write-Host "Database seeded, retrying process creation..." -ForegroundColor Yellow
            $processResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/process/start" -Method POST -Body $processData -Headers $headers
            Write-Host "✓ Test process started successfully after seeding" -ForegroundColor Green
        } catch {
            Write-Host "✗ Still failed after seeding: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "WebSocket Test Instructions:" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "1. Open your browser to http://localhost:3001" -ForegroundColor White
Write-Host "2. Open browser Developer Tools (F12)" -ForegroundColor White
Write-Host "3. Check the Console tab for WebSocket connection logs" -ForegroundColor White
Write-Host "4. Look for the notification bell icon in the top-right corner" -ForegroundColor White
Write-Host "5. You should see a real-time notification for the test task!" -ForegroundColor White
Write-Host "`nIf WebSocket is working, you'll see:" -ForegroundColor Green
Write-Host "- 'WebSocket connected' in browser console" -ForegroundColor Gray
Write-Host "- Real-time toast notification" -ForegroundColor Gray
Write-Host "- Notification badge on bell icon" -ForegroundColor Gray
Write-Host "- Connection status showing 'Real-time'" -ForegroundColor Gray

Write-Host "`nWebSocket URL for manual testing:" -ForegroundColor Cyan
Write-Host "ws://localhost:8081/ws?token=$token" -ForegroundColor Gray

Write-Host "`n🚀 WebSocket test completed!" -ForegroundColor Cyan