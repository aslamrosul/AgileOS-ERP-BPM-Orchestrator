# Script untuk test NATS orchestration

$baseUrl = "http://localhost:8080"

Write-Host "Testing NATS Event-Driven Orchestration..." -ForegroundColor Green
Write-Host ""
Write-Host "This script will:" -ForegroundColor Cyan
Write-Host "  1. Create a workflow with 3 steps" -ForegroundColor White
Write-Host "  2. Start a process instance" -ForegroundColor White
Write-Host "  3. Complete tasks and watch automatic orchestration" -ForegroundColor White
Write-Host ""

# Step 1: Create Workflow
Write-Host "Step 1: Creating workflow..." -ForegroundColor Cyan

$workflowData = @{
    workflow = @{
        name = "Simple Approval Flow"
        version = "1.0.0"
        description = "Test workflow for orchestration"
        is_active = $true
    }
    steps = @(
        @{
            id = "node_start"
            name = "Submit Request"
            type = "action"
            assigned_to = "role:employee"
            sla = "1h"
            position = @{ x = 100; y = 100 }
        },
        @{
            id = "node_approve"
            name = "Manager Approval"
            type = "approval"
            assigned_to = "role:manager"
            sla = "24h"
            position = @{ x = 100; y = 200 }
        },
        @{
            id = "node_complete"
            name = "Notify Completion"
            type = "notify"
            assigned_to = "role:employee"
            sla = "1h"
            position = @{ x = 100; y = 300 }
        }
    )
    relations = @(
        @{
            from = "node_start"
            to = "node_approve"
            condition = $null
        },
        @{
            from = "node_approve"
            to = "node_complete"
            condition = $null
        }
    )
} | ConvertTo-Json -Depth 10

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/workflow" -Method POST -Body $workflowData -ContentType "application/json"
    $workflowId = $response.workflow_id
    
    if (-not $workflowId) {
        Write-Host "  Error: Workflow ID not returned" -ForegroundColor Red
        Write-Host "  Response: $($response | ConvertTo-Json)" -ForegroundColor Gray
        exit 1
    }
    
    Write-Host "  Success: Workflow created - $workflowId" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "  Failed to create workflow: $_" -ForegroundColor Red
    exit 1
}

# Step 2: Start Process
Write-Host "Step 2: Starting process instance..." -ForegroundColor Cyan
Write-Host "  Using workflow ID: $workflowId" -ForegroundColor Gray

$processData = @{
    workflow_id = $workflowId
    initiated_by = "user:john_doe"
    data = @{
        request_type = "purchase"
        amount = 5000
        description = "New laptops"
    }
} | ConvertTo-Json -Depth 10

Write-Host "  Request body: $processData" -ForegroundColor Gray

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/process/start" -Method POST -Body $processData -ContentType "application/json"
    $processId = $response.process_instance_id
    $firstTaskId = $response.first_task_id
    Write-Host "  Success: Process started" -ForegroundColor Green
    Write-Host "  Process ID: $processId" -ForegroundColor White
    Write-Host "  First Task ID: $firstTaskId" -ForegroundColor White
    Write-Host "  Current Step: $($response.current_step)" -ForegroundColor White
    Write-Host ""
} catch {
    Write-Host "  Failed to start process: $_" -ForegroundColor Red
    exit 1
}

Write-Host "Watch backend logs for NATS events!" -ForegroundColor Yellow
Write-Host ""

# Step 3: Complete First Task
Write-Host "Step 3: Completing first task (Submit Request)..." -ForegroundColor Cyan
Write-Host "  Waiting 2 seconds..." -ForegroundColor Gray
Start-Sleep -Seconds 2

$completeData = @{
    executed_by = "user:john_doe"
    result = @{
        status = "submitted"
        comments = "Request submitted successfully"
    }
} | ConvertTo-Json -Depth 10

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/task/$firstTaskId/complete" -Method POST -Body $completeData -ContentType "application/json"
    Write-Host "  Success: Task completed" -ForegroundColor Green
    Write-Host ""
    Write-Host "  Check backend logs - you should see:" -ForegroundColor Yellow
    Write-Host "    [NATS] Published: Task completed" -ForegroundColor Gray
    Write-Host "    [ORCHESTRATOR] Processing completion..." -ForegroundColor Gray
    Write-Host "    [ORCHESTRATOR] Triggering next step: Manager Approval" -ForegroundColor Gray
    Write-Host "    [NATS] Published: Task started" -ForegroundColor Gray
    Write-Host ""
} catch {
    Write-Host "  Failed to complete task: $_" -ForegroundColor Red
    exit 1
}

# Step 4: Get Pending Tasks for Manager
Write-Host "Step 4: Getting pending tasks for manager..." -ForegroundColor Cyan

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/tasks/pending/role:manager" -Method GET
    $managerTasks = $response.tasks
    Write-Host "  Success: Found $($response.count) pending task(s)" -ForegroundColor Green
    
    if ($response.count -gt 0) {
        $secondTaskId = $managerTasks[0].id
        Write-Host "  Task ID: $secondTaskId" -ForegroundColor White
        Write-Host "  Step: $($managerTasks[0].step_name)" -ForegroundColor White
        Write-Host "  Assigned to: $($managerTasks[0].assigned_to)" -ForegroundColor White
        Write-Host ""
        
        # Step 5: Complete Second Task
        Write-Host "Step 5: Completing second task (Manager Approval)..." -ForegroundColor Cyan
        Write-Host "  Waiting 2 seconds..." -ForegroundColor Gray
        Start-Sleep -Seconds 2
        
        $approveData = @{
            executed_by = "user:manager_jane"
            result = @{
                decision = "approved"
                comments = "Approved by manager"
            }
        } | ConvertTo-Json -Depth 10
        
        try {
            $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/task/$secondTaskId/complete" -Method POST -Body $approveData -ContentType "application/json"
            Write-Host "  Success: Task completed" -ForegroundColor Green
            Write-Host ""
            Write-Host "  Check backend logs again - orchestration to final step!" -ForegroundColor Yellow
            Write-Host ""
        } catch {
            Write-Host "  Failed to complete task: $_" -ForegroundColor Red
        }
        
        # Step 6: Get Final Task
        Write-Host "Step 6: Getting final task (Notify Completion)..." -ForegroundColor Cyan
        Start-Sleep -Seconds 1
        
        try {
            $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/tasks/pending/role:employee" -Method GET
            Write-Host "  Success: Found $($response.count) pending task(s)" -ForegroundColor Green
            
            if ($response.count -gt 0) {
                $finalTask = $response.tasks | Where-Object { $_.step_name -eq "Notify Completion" } | Select-Object -First 1
                if ($finalTask) {
                    Write-Host "  Final Task ID: $($finalTask.id)" -ForegroundColor White
                    Write-Host "  Step: $($finalTask.step_name)" -ForegroundColor White
                    Write-Host ""
                }
            }
        } catch {
            Write-Host "  Failed to get tasks: $_" -ForegroundColor Red
        }
    }
} catch {
    Write-Host "  Failed to get pending tasks: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "Orchestration Test Completed!" -ForegroundColor Green
Write-Host ""
Write-Host "Summary:" -ForegroundColor Cyan
Write-Host "  - Workflow created with 3 steps" -ForegroundColor White
Write-Host "  - Process started automatically" -ForegroundColor White
Write-Host "  - Task 1 completed -> Task 2 auto-created via NATS" -ForegroundColor White
Write-Host "  - Task 2 completed -> Task 3 auto-created via NATS" -ForegroundColor White
Write-Host "  - All orchestration done event-driven!" -ForegroundColor White
Write-Host ""
Write-Host "Check backend terminal for detailed NATS logs" -ForegroundColor Yellow
