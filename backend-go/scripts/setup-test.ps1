# Setup task for mobile testing
$baseUrl = "http://localhost:8081"

Write-Host "Setting up Manager Task for Mobile Testing" -ForegroundColor Green
Write-Host ""

# Start process
Write-Host "Starting process..." -ForegroundColor Cyan
$body = '{"workflow_id":"workflow:purchase_approval","initiated_by":"user:john","data":{"amount":15000,"description":"Laptops"}}'

try {
    $response = Invoke-WebRequest -Uri "$baseUrl/api/v1/process/start" -Method POST -Body $body -ContentType "application/json" -UseBasicParsing
    $json = $response.Content | ConvertFrom-Json
    
    Write-Host "Process started" -ForegroundColor Green
    Write-Host "Task ID: $($json.first_task_id)" -ForegroundColor White
    Write-Host ""
    
    # Complete first task
    Write-Host "Completing employee task..." -ForegroundColor Cyan
    $taskId = $json.first_task_id
    $completeBody = '{"executed_by":"user:john","result":{"decision":"submitted"}}'
    
    $response2 = Invoke-WebRequest -Uri "$baseUrl/api/v1/task/$taskId/complete" -Method POST -Body $completeBody -ContentType "application/json" -UseBasicParsing
    
    Write-Host "Task completed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Manager task created - open mobile app!" -ForegroundColor Cyan
    
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host ""

# Check pending tasks
Write-Host "Checking pending tasks..." -ForegroundColor Cyan
$tasks = Invoke-WebRequest -Uri "$baseUrl/api/v1/tasks/pending/role:manager" -Method GET -UseBasicParsing
$tasksJson = $tasks.Content | ConvertFrom-Json

Write-Host "Found $($tasksJson.count) pending tasks" -ForegroundColor Yellow
