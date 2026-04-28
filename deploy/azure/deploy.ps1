# AgileOS BPM Platform - Azure Deployment Script (PowerShell)
# This script automates the deployment process to Azure from Windows

param(
    [switch]$SkipBuild,
    [switch]$SkipPush,
    [switch]$SkipDeploy
)

$ErrorActionPreference = "Stop"

# Load environment variables from .env file
if (Test-Path ".env") {
    Get-Content ".env" | ForEach-Object {
        if ($_ -match '^([^#][^=]+)=(.*)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            Set-Item -Path "env:$name" -Value $value
        }
    }
} else {
    Write-Host "Error: .env file not found. Copy .env.example to .env and configure it." -ForegroundColor Red
    exit 1
}

Write-Host "========================================" -ForegroundColor Green
Write-Host "AgileOS BPM - Azure Deployment" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# Step 1: Login to Azure
Write-Host "Step 1: Logging in to Azure..." -ForegroundColor Yellow
az login --use-device-code

# Step 2: Set subscription
Write-Host "Step 2: Setting Azure subscription..." -ForegroundColor Yellow
az account set --subscription $env:AZURE_SUBSCRIPTION_ID

# Step 3: Login to Azure Container Registry
Write-Host "Step 3: Logging in to Azure Container Registry..." -ForegroundColor Yellow
az acr login --name $env:ACR_NAME

# Step 4: Build Docker images
if (-not $SkipBuild) {
    Write-Host "Step 4: Building Docker images..." -ForegroundColor Yellow
    
    # Build backend
    Write-Host "Building backend image..." -ForegroundColor Cyan
    docker build -t "$env:ACR_LOGIN_SERVER/agileos-backend:$env:IMAGE_TAG" ./backend-go
    
    # Build frontend
    Write-Host "Building frontend image..." -ForegroundColor Cyan
    docker build -t "$env:ACR_LOGIN_SERVER/agileos-frontend:$env:IMAGE_TAG" ./frontend-next
    
    Write-Host "Images built successfully" -ForegroundColor Green
} else {
    Write-Host "Step 4: Skipping build (--SkipBuild flag)" -ForegroundColor Yellow
}

# Step 5: Push images to ACR
if (-not $SkipPush) {
    Write-Host "Step 5: Pushing images to Azure Container Registry..." -ForegroundColor Yellow
    
    docker push "$env:ACR_LOGIN_SERVER/agileos-backend:$env:IMAGE_TAG"
    docker push "$env:ACR_LOGIN_SERVER/agileos-frontend:$env:IMAGE_TAG"
    
    Write-Host "Images pushed successfully" -ForegroundColor Green
} else {
    Write-Host "Step 5: Skipping push (--SkipPush flag)" -ForegroundColor Yellow
}

# Step 6: Deploy to Azure VM
if (-not $SkipDeploy -and $env:AZURE_VM_IP) {
    Write-Host "Step 6: Deploying to Azure VM..." -ForegroundColor Yellow
    
    # Copy files to VM
    Write-Host "Copying files to VM..." -ForegroundColor Cyan
    scp docker-compose.prod.yml "$env:AZURE_VM_ADMIN@$env:AZURE_VM_IP`:~/agileos/"
    scp .env "$env:AZURE_VM_ADMIN@$env:AZURE_VM_IP`:~/agileos/"
    scp -r deploy/nginx "$env:AZURE_VM_ADMIN@$env:AZURE_VM_IP`:~/agileos/deploy/"
    
    # Deploy on VM
    Write-Host "Deploying containers on VM..." -ForegroundColor Cyan
    ssh "$env:AZURE_VM_ADMIN@$env:AZURE_VM_IP" @"
        cd ~/agileos
        az acr login --name $env:ACR_NAME
        docker-compose -f docker-compose.prod.yml pull
        docker-compose -f docker-compose.prod.yml down
        docker-compose -f docker-compose.prod.yml up -d
        docker-compose -f docker-compose.prod.yml ps
"@
    
    Write-Host "Deployed to Azure VM successfully" -ForegroundColor Green
    Write-Host "Application URL: http://$env:AZURE_VM_IP" -ForegroundColor Green
} elseif (-not $env:AZURE_VM_IP) {
    Write-Host "Step 6: Skipping VM deployment (AZURE_VM_IP not configured)" -ForegroundColor Yellow
} else {
    Write-Host "Step 6: Skipping deployment (--SkipDeploy flag)" -ForegroundColor Yellow
}

# Step 7: Health check
Write-Host "Step 7: Running health check..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

if ($env:AZURE_VM_IP) {
    $healthUrl = "http://$env:AZURE_VM_IP/health"
    
    try {
        $response = Invoke-WebRequest -Uri $healthUrl -UseBasicParsing
        if ($response.StatusCode -eq 200) {
            Write-Host "Health check passed" -ForegroundColor Green
        }
    } catch {
        Write-Host "Health check failed" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Deployment completed successfully!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:"
Write-Host "1. Configure DNS to point to $env:AZURE_VM_IP"
Write-Host "2. Setup SSL certificates for HTTPS"
Write-Host "3. Configure monitoring and alerts"
Write-Host ""
