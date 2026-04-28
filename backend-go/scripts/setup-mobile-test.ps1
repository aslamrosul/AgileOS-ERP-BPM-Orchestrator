# Setup task for mobile testing
$baseUrl = "http://localhost:8081"

Write-Host "📱 Setting up Manager Task for Mobile Testing" -ForegroundColor Green
Write-Host ""

# Start process
Write-Host "Starting process..." -ForegroundColor Cyan
$body = '{"workflow_id":"workflow:purchase_approval","initiated_by":"user:john","data":{"amount":15000,"description":"Laptops for dev team"}}'

$response = Invoke-WebRequest -Uri "$baseUrl/api/v1/process/start" -Method POST -Body $body -ContentType "application/json" -UseBasicParsing
$json = $response.Content | ConvertFrom-Json

Write-Host "✓ Process started" -ForegroundColor Green
Write-Host "  Task ID: $($json.first_task_id)" -ForegroundColor White
Write-Host ""

# Complete first task to create manager task
Write-Host "Completing employee task..." -ForegroundColor Cyan
$taskId = $json.first_task_id
$completeBody = '{"executed_by":"user:john","result":{"decision":"submitted"}}'

$response2 = Invoke-WebRequest -Uri "$baseUrl/api/v1/task/$taskId/complete" -Method POST -Body $completeBody -ContentType "application/json" -UseBasicParsing

if ($response2.StatusCode -eq 200) {
    Write-Host "✓ Task completed" -ForegroundColor Green
    Write-Host ""
    Write-Host "🎯 Manager task created!" -ForegroundColor Cyan
    Write-Host "   Open mobile app to see it" -ForegroundColor White
} else {
    Write-Host "✗ Failed: $($response2.Content)" -ForegroundColor Red
}

Write-Host ""

# Check pending tasks
Write-Host "Checking pending tasks for manager..." -ForegroundColor Cyan
$tasks = Invoke-WebRequest -Uri "$baseUrl/api/v1/tasks/pending/role:manager" -Method GET -UseBasicParsing
$tasksJson = $tasks.Content | ConvertFrom-Json

Write-Host "📋 Found $($tasksJson.count) pending task(s)" -ForegroundColor Yellow
