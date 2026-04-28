# Simple orchestration test using seeded workflow

$baseUrl = "http://localhost:8081"

Write-Host "Testing NATS Orchestration (Simple)" -ForegroundColor Green
Write-Host ""

# First, seed the database if not already done
Write-Host "Make sure you have seeded the database with:" -ForegroundColor Yellow
Write-Host "  - Open http://localhost:8000" -ForegroundColor White
Write-Host "  - Run queries from backend-go/database/seed.surql" -ForegroundColor White
Write-Host ""
Write-Host "Press Enter to continue..." -ForegroundColor Yellow
Read-Host

# Step 1: Start process with seeded workflow
Write-Host "Step 1: Starting process with seeded workflow..." -ForegroundColor Cyan

$processData = @{
    workflow_id = "workflow:purchase_approval"
    initiated_by = "user:john_doe"
    data = @{
        amount = 5000
        description = "Test purchase"
    }
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/process/start" -Method POST -Body $processData -ContentType "application/json"
    Write-Host "  Success: Process started" -ForegroundColor Green
    Write-Host "  Response: $($response | ConvertTo-Json)" -ForegroundColor Gray
    Write-Host ""
    
    # Get task ID from database query
    Write-Host "Step 2: Check SurrealDB for task instances..." -ForegroundColor Cyan
    Write-Host "  Open http://localhost:8000 and run:" -ForegroundColor White
    Write-Host "  SELECT * FROM task_instance ORDER BY created_at DESC LIMIT 1;" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Copy the task ID and paste here:" -ForegroundColor Yellow
    $taskId = Read-Host "  Task ID"
    
    if ($taskId) {
        Write-Host ""
        Write-Host "Step 3: Completing task..." -ForegroundColor Cyan
        Write-Host "  Watch backend logs for orchestration!" -ForegroundColor Yellow
        Write-Host ""
        
        $completeData = @{
            executed_by = "user:john_doe"
            result = @{
                decision = "approved"
                comments = "Test approval"
            }
        } | ConvertTo-Json
        
        try {
            $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/task/$taskId/complete" -Method POST -Body $completeData -ContentType "application/json"
            Write-Host "  Success: Task completed!" -ForegroundColor Green
            Write-Host ""
            Write-Host "Check backend logs - you should see:" -ForegroundColor Yellow
            Write-Host "  [NATS] Published: Task completed" -ForegroundColor Gray
            Write-Host "  [ORCHESTRATOR] Processing completion..." -ForegroundColor Gray
            Write-Host "  [ORCHESTRATOR] Triggering next step..." -ForegroundColor Gray
            Write-Host "  [NATS] Published: Task started" -ForegroundColor Gray
            Write-Host ""
            Write-Host "Step 4: Check for new task..." -ForegroundColor Cyan
            Write-Host "  Run in SurrealDB:" -ForegroundColor White
            Write-Host "  SELECT * FROM task_instance ORDER BY created_at DESC LIMIT 2;" -ForegroundColor Gray
            Write-Host ""
            Write-Host "  You should see 2 tasks - the completed one and a new pending one!" -ForegroundColor Green
        } catch {
            Write-Host "  Failed: $_" -ForegroundColor Red
        }
    }
} catch {
    Write-Host "  Failed to start process: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "Make sure:" -ForegroundColor Yellow
    Write-Host "  1. Backend is running" -ForegroundColor White
    Write-Host "  2. Database is seeded with purchase_approval workflow" -ForegroundColor White
}

Write-Host ""
Write-Host "Test completed!" -ForegroundColor Green
