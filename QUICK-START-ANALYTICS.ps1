# Quick Start Script untuk Analytics Dashboard
# Script ini akan setup semua yang diperlukan

Write-Host "========================================" -ForegroundColor Green
Write-Host "AgileOS Analytics - Quick Start" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# Step 1: Stop semua yang jalan
Write-Host "[1/6] Stopping existing services..." -ForegroundColor Cyan
docker-compose down -v 2>$null
Start-Sleep -Seconds 2

# Step 2: Start Docker containers
Write-Host "[2/6] Starting Docker containers..." -ForegroundColor Cyan
docker-compose up -d
Start-Sleep -Seconds 10

# Step 3: Start Backend
Write-Host "[3/6] Starting Backend Go..." -ForegroundColor Cyan
$env:SURREAL_URL="ws://localhost:8002/rpc"
$env:NATS_URL="nats://localhost:4223"
$env:PORT="8081"
$env:JWT_SECRET="your-super-secret-jwt-key-change-in-production-min-32-chars"

Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PWD\backend-go'; go run main.go" -WindowStyle Minimized
Start-Sleep -Seconds 8

# Step 4: Create Admin User
Write-Host "[4/6] Creating admin user..." -ForegroundColor Cyan
$userData = @{
    username = "admin"
    password = "password123"
    email = "admin@agileos.com"
    full_name = "Admin User"
    role = "admin"
} | ConvertTo-Json

try {
    $registerResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/register" -Method POST -Body $userData -ContentType "application/json"
    Write-Host "  ✓ Admin user created" -ForegroundColor Green
    $token = $registerResponse.access_token
} catch {
    Write-Host "  ⚠ User might already exist, trying login..." -ForegroundColor Yellow
    
    # Activate user first
    $auth = "Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("root:root"))
    $query = "UPDATE user SET is_active = true WHERE username = 'admin'"
    $body = @{ query = $query } | ConvertTo-Json
    Invoke-RestMethod -Uri "http://localhost:8002/sql" -Method POST -Body $body -ContentType "application/json" -Headers @{ "NS" = "agileos"; "DB" = "main"; "Authorization" = $auth } | Out-Null
    
    # Try login
    $loginData = @{
        username = "admin"
        password = "password123"
    } | ConvertTo-Json
    
    try {
        $loginResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/login" -Method POST -Body $loginData -ContentType "application/json"
        $token = $loginResponse.access_token
        Write-Host "  ✓ Logged in successfully" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ Failed to login. Please check backend logs." -ForegroundColor Red
        exit 1
    }
}

# Step 5: Generate Test Data
Write-Host "[5/6] Generating test data..." -ForegroundColor Cyan

# First, create workflow if not exists
$workflowData = @{
    name = "Purchase Approval"
    description = "Standard purchase approval workflow"
    version = "1.0"
} | ConvertTo-Json

try {
    Invoke-RestMethod -Uri "http://localhost:8081/api/v1/workflow" -Method POST -Body $workflowData -ContentType "application/json" -Headers @{ "Authorization" = "Bearer $token" } | Out-Null
    Write-Host "  ✓ Workflow created" -ForegroundColor Green
} catch {
    Write-Host "  ⚠ Workflow might already exist" -ForegroundColor Yellow
}

# Create some test processes
for ($i = 1; $i -le 5; $i++) {
    $processData = @{
        workflow_id = "workflow:purchase_approval"
        initiated_by = "user:admin"
        data = @{
            amount = Get-Random -Minimum 1000 -Maximum 50000
            description = "Test purchase #$i"
        }
    } | ConvertTo-Json

    try {
        Invoke-RestMethod -Uri "http://localhost:8081/api/v1/process/start" -Method POST -Body $processData -ContentType "application/json" -Headers @{ "Authorization" = "Bearer $token" } | Out-Null
        Write-Host "  ✓ Process $i created" -ForegroundColor Gray
    } catch {
        Write-Host "  ⚠ Process $i failed" -ForegroundColor Yellow
    }
}

# Step 6: Start Frontend
Write-Host "[6/6] Starting Frontend..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PWD\frontend-next'; npm run dev" -WindowStyle Minimized
Start-Sleep -Seconds 5

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Setup Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Access Points:" -ForegroundColor Cyan
Write-Host "  Analytics Dashboard: http://localhost:3001/analytics" -ForegroundColor White
Write-Host "  Backend API: http://localhost:8081" -ForegroundColor White
Write-Host ""
Write-Host "Login Credentials:" -ForegroundColor Cyan
Write-Host "  Username: admin" -ForegroundColor White
Write-Host "  Password: password123" -ForegroundColor White
Write-Host ""
Write-Host "Press any key to open analytics dashboard..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
Start-Process "http://localhost:3001/analytics"
