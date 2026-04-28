# Quick test for orchestration fix
$baseUrl = "http://localhost:8081"

Write-Host "🧪 Quick Orchestration Test" -ForegroundColor Green
Write-Host ""

# Start process
Write-Host "1. Starting process..." -ForegroundColor Cyan
$processData = @{
    workflow_id = "workflow:purchase_approval"
    initiated_by = "user:test"
    data = @{ amount = 1000 }
} | ConvertTo-Json

try {
    $startResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/process/start" -Method POST -Body $processData -ContentType "application/json"
    $taskId = $startResponse.first_task_id
    Write-Host "   ✓ Process started, Task ID: $taskId" -ForegroundColor Green
    
    # Complete task
    Write-Host ""
    Write-Host "2. Completing task..." -ForegroundColor Cyan
    $completeData = @{
        executed_by = "user:test"
        result = @{ decision = "approved" }
    } | ConvertTo-Json
    
    $completeResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/task/$taskId/complete" -Method POST -Body $completeData -ContentType "application/json"
    Write-Host "   ✓ Task completed successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "✅ TEST PASSED - Orchestration fix works!" -ForegroundColor Green
    
} catch {
    Write-Host "   ✗ FAILED: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "   Details: $($_.ErrorDetails.Message)" -ForegroundColor Red
    }
    Write-Host ""
    Write-Host "❌ TEST FAILED" -ForegroundColor Red
}
