# Create a task for manager to approve in mobile app
$baseUrl = "http://localhost:8081"

Write-Host "📱 Creating task for Manager approval..." -ForegroundColor Green

# Step 1: Start process
$processData = @{
    workflow_id = "workflow:purchase_approval"
    initiated_by = "user:john_doe"
    data = @{
        amount = 15000
        description = "New laptops for development team"
        quantity = 3
    }
} | ConvertTo-Json

$startResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/process/start" -Method POST -Body $processData -ContentType "application/json"
$firstTaskId = $startResponse.first_task_id

Write-Host "✓ Process started" -ForegroundColor Green
Write-Host "  First task ID: $firstTaskId" -ForegroundColor Gray

if (-not $firstTaskId) {
    Write-Host "✗ Failed to get task ID from response" -ForegroundColor Red
    Write-Host "  Response: $($startResponse | ConvertTo-Json)" -ForegroundColor Gray
    exit 1
}

# Step 2: Complete first task (employee submit) to trigger manager task
Start-Sleep -Seconds 1

$completeData = @{
    executed_by = "user:john_doe"
    result = @{
        decision = "submitted"
        comments = "Please review this purchase request"
    }
} | ConvertTo-Json

$completeResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/task/$firstTaskId/complete" -Method POST -Body $completeData -ContentType "application/json"

Write-Host "✓ First task completed" -ForegroundColor Green
Write-Host ""
Write-Host "🎯 Manager task should now be available!" -ForegroundColor Cyan
Write-Host "   Open mobile app and pull to refresh" -ForegroundColor White
Write-Host ""

# Step 3: Verify manager task exists
Start-Sleep -Seconds 1

$managerTasks = Invoke-RestMethod -Uri "$baseUrl/api/v1/tasks/pending/role:manager" -Method GET
Write-Host "📋 Pending tasks for role:manager: $($managerTasks.count)" -ForegroundColor Yellow

if ($managerTasks.count -gt 0) {
    Write-Host ""
    Write-Host "Tasks:" -ForegroundColor Cyan
    foreach ($task in $managerTasks.tasks) {
        Write-Host "  - ID: $($task.id)" -ForegroundColor White
        Write-Host "    Step: $($task.step_name)" -ForegroundColor Gray
        Write-Host "    Status: $($task.status)" -ForegroundColor Gray
        Write-Host ""
    }
}
