# Generate test data for analytics dashboard

$baseUrl = "http://localhost:8081"

Write-Host "========================================" -ForegroundColor Green
Write-Host "Generating Test Data for Analytics" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# Login as admin
Write-Host "Logging in..." -ForegroundColor Cyan
$loginData = @{
    username = "admin"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method POST -Body $loginData -ContentType "application/json"
    $token = $loginResponse.access_token
    Write-Host "  Logged in successfully" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "  Failed to login. Make sure backend is running and users are seeded." -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
}

# Generate 10 test processes
Write-Host "Creating test processes..." -ForegroundColor Cyan

for ($i = 1; $i -le 10; $i++) {
    $processData = @{
        workflow_id = "workflow:purchase_approval"
        initiated_by = "user:testuser$i"
        data = @{
            amount = Get-Random -Minimum 1000 -Maximum 50000
            description = "Test purchase request #$i"
            quantity = Get-Random -Minimum 1 -Maximum 10
        }
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/process/start" -Method POST -Body $processData -ContentType "application/json" -Headers $headers
        Write-Host "  Process $i created: $($response.first_task_id)" -ForegroundColor Gray
        
        # Randomly complete some tasks
        if ((Get-Random -Minimum 1 -Maximum 10) -gt 3) {
            Start-Sleep -Milliseconds 500
            
            $completeData = @{
                executed_by = "user:manager"
                result = @{
                    decision = if ((Get-Random -Minimum 1 -Maximum 10) -gt 2) { "approved" } else { "rejected" }
                    comments = "Test completion"
                }
            } | ConvertTo-Json

            try {
                Invoke-RestMethod -Uri "$baseUrl/api/v1/task/$($response.first_task_id)/complete" -Method POST -Body $completeData -ContentType "application/json" -Headers $headers | Out-Null
                Write-Host "    Task completed" -ForegroundColor Green
            } catch {
                Write-Host "    Task completion failed (expected)" -ForegroundColor Yellow
            }
        }
        
    } catch {
        Write-Host "  Failed to create process $i" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Test data generation completed!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Now you can:" -ForegroundColor Cyan
Write-Host "  1. Open analytics dashboard: http://localhost:3001/analytics" -ForegroundColor White
Write-Host "  2. Or test API: curl -H 'Authorization: Bearer $token' http://localhost:8081/api/v1/analytics/overview" -ForegroundColor White
Write-Host ""
