# Script untuk test API endpoints

$baseUrl = "http://localhost:8080"

Write-Host "Testing AgileOS API Endpoints..." -ForegroundColor Green
Write-Host ""

# Test 1: Health Check
Write-Host "1. Testing Health Endpoint..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/health" -Method GET
    Write-Host "   Success: Health check passed" -ForegroundColor Green
    Write-Host "   Status: $($response.status)" -ForegroundColor White
    Write-Host "   Database: $($response.database)" -ForegroundColor White
    Write-Host "   Message Broker: $($response.message_broker)" -ForegroundColor White
} catch {
    Write-Host "   Failed: Health check failed" -ForegroundColor Red
    Write-Host "   Error: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Test 2: Create Workflow
Write-Host "2. Testing Create Workflow..." -ForegroundColor Cyan

$workflowData = @{
    workflow = @{
        name = "Test Workflow from API"
        version = "1.0.0"
        description = "Created via PowerShell test script"
        is_active = $true
    }
    steps = @(
        @{
            id = "node_1"
            name = "Start Process"
            type = "action"
            assigned_to = "role:employee"
            sla = "1h"
            position = @{ x = 100; y = 100 }
        },
        @{
            id = "node_2"
            name = "Manager Review"
            type = "approval"
            assigned_to = "role:manager"
            sla = "24h"
            position = @{ x = 100; y = 200 }
        },
        @{
            id = "node_3"
            name = "Complete"
            type = "notify"
            assigned_to = "role:employee"
            sla = "1h"
            position = @{ x = 100; y = 300 }
        }
    )
    relations = @(
        @{
            from = "node_1"
            to = "node_2"
            condition = $null
        },
        @{
            from = "node_2"
            to = "node_3"
            condition = @{ decision = "approved" }
        }
    )
} | ConvertTo-Json -Depth 10

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/workflow" -Method POST -Body $workflowData -ContentType "application/json"
    Write-Host "   Success: Workflow created" -ForegroundColor Green
    Write-Host "   Workflow ID: $($response.workflow_id)" -ForegroundColor White
    Write-Host "   Steps Created: $($response.steps_created)" -ForegroundColor White
    Write-Host "   Relations Created: $($response.relations_created)" -ForegroundColor White
    
    $workflowId = $response.workflow_id
} catch {
    Write-Host "   Failed: Could not create workflow" -ForegroundColor Red
    Write-Host "   Error: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Test 3: Get Workflow
if ($workflowId) {
    Write-Host "3. Testing Get Workflow..." -ForegroundColor Cyan
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/workflow/$workflowId" -Method GET
        Write-Host "   Success: Workflow retrieved" -ForegroundColor Green
        Write-Host "   Name: $($response.workflow.name)" -ForegroundColor White
        Write-Host "   Version: $($response.workflow.version)" -ForegroundColor White
        Write-Host "   Steps Count: $($response.steps.Count)" -ForegroundColor White
    } catch {
        Write-Host "   Failed: Could not get workflow" -ForegroundColor Red
        Write-Host "   Error: $_" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "API Tests Completed!" -ForegroundColor Green
Write-Host ""
Write-Host "Next: Open frontend at http://localhost:3000/workflow" -ForegroundColor Yellow
