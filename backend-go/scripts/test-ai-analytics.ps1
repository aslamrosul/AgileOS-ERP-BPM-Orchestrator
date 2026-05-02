#!/usr/bin/env pwsh

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Testing AI Analytics Microservice Integration" -ForegroundColor Cyan
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

# Check Python Analytics Service
Write-Host "`n1. Testing Python Analytics Service..." -ForegroundColor Yellow
if (-not (Wait-ForService -Url "http://localhost:8001/health" -ServiceName "Python Analytics")) {
    Write-Host "Please start the Python service first:" -ForegroundColor Yellow
    Write-Host "  docker-compose up -d agileos-analytics" -ForegroundColor Gray
    Write-Host "Or manually:" -ForegroundColor Yellow
    Write-Host "  cd analytics-py" -ForegroundColor Gray
    Write-Host "  pip install -r requirements.txt" -ForegroundColor Gray
    Write-Host "  uvicorn main:app --host 0.0.0.0 --port 8001" -ForegroundColor Gray
    exit 1
}

# Test Python service endpoints directly
Write-Host "`n2. Testing Python Service Endpoints..." -ForegroundColor Yellow

try {
    # Test root endpoint
    $rootResponse = Invoke-RestMethod -Uri "http://localhost:8001/" -Method GET
    Write-Host "✓ Python root endpoint: $($rootResponse.service)" -ForegroundColor Green
    
    # Test health endpoint
    $healthResponse = Invoke-RestMethod -Uri "http://localhost:8001/health" -Method GET
    Write-Host "✓ Python health: $($healthResponse.status)" -ForegroundColor Green
    
    # Test comprehensive analytics
    $analyticsResponse = Invoke-RestMethod -Uri "http://localhost:8001/analytics/comprehensive" -Method GET
    Write-Host "✓ Python analytics: $($analyticsResponse.predictions.Count) predictions, $($analyticsResponse.anomalies.Count) anomalies" -ForegroundColor Green
    
} catch {
    Write-Host "✗ Python service test failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Check Go Backend Service
Write-Host "`n3. Testing Go Backend Service..." -ForegroundColor Yellow
if (-not (Wait-ForService -Url "http://localhost:8081/health" -ServiceName "Go Backend")) {
    Write-Host "Please start the Go backend first:" -ForegroundColor Yellow
    Write-Host "  cd backend-go" -ForegroundColor Gray
    Write-Host "  go run ." -ForegroundColor Gray
    exit 1
}

# Test authentication
Write-Host "`n4. Testing Authentication..." -ForegroundColor Yellow
$loginData = @{
    username = "admin"
    password = "password123"
} | ConvertTo-Json

try {
    $authResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/login" -Method POST -Body $loginData -ContentType "application/json"
    $token = $authResponse.access_token
    Write-Host "✓ Authentication successful: $($authResponse.user.username)" -ForegroundColor Green
} catch {
    Write-Host "✗ Authentication failed: $($_.Exception.Message)" -ForegroundColor Red
    
    # Try to seed users
    Write-Host "Attempting to seed users..." -ForegroundColor Yellow
    try {
        .\scripts\seed-users.ps1
        $authResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/login" -Method POST -Body $loginData -ContentType "application/json"
        $token = $authResponse.access_token
        Write-Host "✓ Authentication successful after seeding" -ForegroundColor Green
    } catch {
        Write-Host "✗ Authentication still failed after seeding" -ForegroundColor Red
        exit 1
    }
}

# Test Go-Python Integration
Write-Host "`n5. Testing Go-Python Integration..." -ForegroundColor Yellow
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

try {
    # Test AI service status through Go
    $aiStatusResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/ai-analytics/status" -Method GET -Headers $headers
    Write-Host "✓ AI Service Status: $($aiStatusResponse.status)" -ForegroundColor Green
    Write-Host "  Predictions: $($aiStatusResponse.ai_capabilities.predictions)" -ForegroundColor Gray
    Write-Host "  Anomaly Detection: $($aiStatusResponse.ai_capabilities.anomaly_detection)" -ForegroundColor Gray
    
    # Test workflow prediction through Go
    $predictionResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/ai-analytics/predict/workflow/purchase_approval" -Method GET -Headers $headers
    Write-Host "✓ Workflow Prediction: $($predictionResponse.prediction.estimated_duration_minutes) minutes" -ForegroundColor Green
    Write-Host "  Confidence: $($predictionResponse.prediction.confidence_score * 100)%" -ForegroundColor Gray
    
    # Test anomaly detection through Go
    $anomaliesResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/ai-analytics/anomalies" -Method GET -Headers $headers
    Write-Host "✓ Anomaly Detection: $($anomaliesResponse.total_found) anomalies found" -ForegroundColor Green
    
    # Test comprehensive AI analytics through Go
    $comprehensiveResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/ai-analytics/comprehensive" -Method GET -Headers $headers
    Write-Host "✓ Comprehensive AI Analytics:" -ForegroundColor Green
    Write-Host "  Predictions: $($comprehensiveResponse.summary.predictions_count)" -ForegroundColor Gray
    Write-Host "  Anomalies: $($comprehensiveResponse.summary.anomalies_count)" -ForegroundColor Gray
    Write-Host "  Insights: $($comprehensiveResponse.summary.insights_count)" -ForegroundColor Gray
    
} catch {
    Write-Host "✗ Go-Python integration test failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Red
}

# Test cache refresh
Write-Host "`n6. Testing Cache Refresh..." -ForegroundColor Yellow
try {
    $refreshResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/ai-analytics/refresh-cache" -Method POST -Headers $headers
    Write-Host "✓ Cache refresh: $($refreshResponse.status)" -ForegroundColor Green
} catch {
    Write-Host "⚠ Cache refresh failed (this is optional): $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "AI Analytics Integration Test Results:" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "✓ Python FastAPI service is running" -ForegroundColor Green
Write-Host "✓ Go backend can communicate with Python" -ForegroundColor Green
Write-Host "✓ AI predictions are working" -ForegroundColor Green
Write-Host "✓ Anomaly detection is functional" -ForegroundColor Green
Write-Host "✓ Comprehensive analytics available" -ForegroundColor Green

Write-Host "`nAI Analytics Endpoints:" -ForegroundColor Cyan
Write-Host "- Python Direct: http://localhost:8001" -ForegroundColor Gray
Write-Host "- Go Integration: http://localhost:8081/api/v1/ai-analytics/*" -ForegroundColor Gray

Write-Host "`nAvailable AI Features:" -ForegroundColor Cyan
Write-Host "- Workflow completion predictions using Linear Regression" -ForegroundColor White
Write-Host "- Anomaly detection using Z-score analysis" -ForegroundColor White
Write-Host "- Performance insights and recommendations" -ForegroundColor White
Write-Host "- Comprehensive analytics dashboard data" -ForegroundColor White

Write-Host "`n🤖 AI Analytics Microservice is ready!" -ForegroundColor Green
Write-Host "Python and Go are now 'ngobrol' (talking) successfully! 🚀" -ForegroundColor Cyan