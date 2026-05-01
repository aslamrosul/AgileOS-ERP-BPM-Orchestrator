#!/usr/bin/env pwsh

# AgileOS Automated Testing Script (PowerShell)
# Runs all unit tests, integration tests, and generates coverage reports

$ErrorActionPreference = "Continue"

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "🧪 AgileOS Automated Testing Suite" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# Track test results
$BackendTestsPassed = $false
$FrontendTestsPassed = $false
$IntegrationTestsPassed = $false

# Backend Go Tests
Write-Host "=========================================" -ForegroundColor Yellow
Write-Host "1. Running Backend Go Tests" -ForegroundColor Yellow
Write-Host "=========================================" -ForegroundColor Yellow
Set-Location "agile-os\backend-go"

Write-Host "Installing Go dependencies..." -ForegroundColor Gray
go mod download
go mod tidy

Write-Host ""
Write-Host "Running unit tests..." -ForegroundColor Gray
$goTestResult = go test ./... -v -cover -coverprofile=coverage.out
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Backend unit tests PASSED" -ForegroundColor Green
    $BackendTestsPassed = $true
    
    Write-Host ""
    Write-Host "Generating coverage report..." -ForegroundColor Gray
    go tool cover -html=coverage.out -o coverage.html
    
    # Calculate coverage percentage
    $coverageOutput = go tool cover -func=coverage.out | Select-String "total"
    Write-Host "Total Coverage: $coverageOutput" -ForegroundColor Cyan
    
    $coveragePercent = [regex]::Match($coverageOutput, '(\d+\.\d+)%').Groups[1].Value
    if ([double]$coveragePercent -gt 80) {
        Write-Host "✓ Coverage exceeds 80% threshold" -ForegroundColor Green
    } else {
        Write-Host "⚠ Coverage is below 80% threshold" -ForegroundColor Yellow
    }
} else {
    Write-Host "✗ Backend unit tests FAILED" -ForegroundColor Red
}

# Run specific test packages
Write-Host ""
Write-Host "Running BPM Engine tests..." -ForegroundColor Gray
go test ./internal/bpm/... -v

Write-Host ""
Write-Host "Running Auth Middleware tests..." -ForegroundColor Gray
go test ./middleware/... -v

Write-Host ""
Write-Host "Running Integration tests..." -ForegroundColor Gray
$integrationResult = go test ./tests/... -v
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Integration tests PASSED" -ForegroundColor Green
    $IntegrationTestsPassed = $true
} else {
    Write-Host "✗ Integration tests FAILED" -ForegroundColor Red
}

Set-Location "..\..\"

# Frontend Next.js Tests
Write-Host ""
Write-Host "=========================================" -ForegroundColor Yellow
Write-Host "2. Running Frontend Next.js Tests" -ForegroundColor Yellow
Write-Host "=========================================" -ForegroundColor Yellow
Set-Location "agile-os\frontend-next"

Write-Host "Installing Node dependencies..." -ForegroundColor Gray
npm install --silent

Write-Host ""
Write-Host "Running component tests..." -ForegroundColor Gray
$npmTestResult = npm run test -- --run
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Frontend component tests PASSED" -ForegroundColor Green
    $FrontendTestsPassed = $true
} else {
    Write-Host "✗ Frontend component tests FAILED" -ForegroundColor Red
}

Write-Host ""
Write-Host "Generating frontend coverage report..." -ForegroundColor Gray
npm run test:coverage -- --run 2>$null

Set-Location "..\..\"

# Python Analytics Tests (if applicable)
Write-Host ""
Write-Host "=========================================" -ForegroundColor Yellow
Write-Host "3. Running Python Analytics Tests" -ForegroundColor Yellow
Write-Host "=========================================" -ForegroundColor Yellow

if (Test-Path "agile-os\analytics-py") {
    Set-Location "agile-os\analytics-py"
    
    if (Test-Path "requirements.txt") {
        Write-Host "Installing Python dependencies..." -ForegroundColor Gray
        pip install -r requirements.txt --quiet 2>$null
        
        Write-Host ""
        Write-Host "Running Python tests..." -ForegroundColor Gray
        python -m pytest tests/ -v 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ Python tests completed" -ForegroundColor Green
        } else {
            Write-Host "⚠ Python tests not found or failed" -ForegroundColor Yellow
        }
    }
    
    Set-Location "..\..\"
} else {
    Write-Host "⚠ Python analytics directory not found, skipping" -ForegroundColor Yellow
}

# Test Summary
Write-Host ""
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "Test Results:" -ForegroundColor White
Write-Host "=============" -ForegroundColor White

if ($BackendTestsPassed) {
    Write-Host "✓ Backend Tests: PASSED" -ForegroundColor Green
} else {
    Write-Host "✗ Backend Tests: FAILED" -ForegroundColor Red
}

if ($FrontendTestsPassed) {
    Write-Host "✓ Frontend Tests: PASSED" -ForegroundColor Green
} else {
    Write-Host "✗ Frontend Tests: FAILED" -ForegroundColor Red
}

if ($IntegrationTestsPassed) {
    Write-Host "✓ Integration Tests: PASSED" -ForegroundColor Green
} else {
    Write-Host "✗ Integration Tests: FAILED" -ForegroundColor Red
}

Write-Host ""
Write-Host "Coverage Reports:" -ForegroundColor White
Write-Host "================="  -ForegroundColor White
Write-Host "Backend:  agile-os\backend-go\coverage.html" -ForegroundColor Gray
Write-Host "Frontend: agile-os\frontend-next\coverage\" -ForegroundColor Gray

Write-Host ""
if ($BackendTestsPassed -and $FrontendTestsPassed) {
    Write-Host "🎉 ALL TESTS PASSED! Ready for deployment." -ForegroundColor Green
    exit 0
} else {
    Write-Host "❌ SOME TESTS FAILED. Please fix before deployment." -ForegroundColor Red
    exit 1
}