#!/bin/bash

# AgileOS BPM Platform - Azure Deployment Script
# This script automates the deployment process to Azure

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
else
    echo -e "${RED}Error: .env file not found. Copy .env.example to .env and configure it.${NC}"
    exit 1
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}AgileOS BPM - Azure Deployment${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Step 1: Login to Azure
echo -e "${YELLOW}Step 1: Logging in to Azure...${NC}"
az login --use-device-code

# Step 2: Set subscription
echo -e "${YELLOW}Step 2: Setting Azure subscription...${NC}"
az account set --subscription "${AZURE_SUBSCRIPTION_ID}"

# Step 3: Login to Azure Container Registry
echo -e "${YELLOW}Step 3: Logging in to Azure Container Registry...${NC}"
az acr login --name "${ACR_NAME}"

# Step 4: Build Docker images
echo -e "${YELLOW}Step 4: Building Docker images...${NC}"

# Build backend
echo "Building backend image..."
docker build -t "${ACR_LOGIN_SERVER}/agileos-backend:${IMAGE_TAG}" ./backend-go

# Build frontend
echo "Building frontend image..."
docker build -t "${ACR_LOGIN_SERVER}/agileos-frontend:${IMAGE_TAG}" ./frontend-next

echo -e "${GREEN}✓ Images built successfully${NC}"

# Step 5: Push images to ACR
echo -e "${YELLOW}Step 5: Pushing images to Azure Container Registry...${NC}"

docker push "${ACR_LOGIN_SERVER}/agileos-backend:${IMAGE_TAG}"
docker push "${ACR_LOGIN_SERVER}/agileos-frontend:${IMAGE_TAG}"

echo -e "${GREEN}✓ Images pushed successfully${NC}"

# Step 6: Deploy to Azure VM (if configured)
if [ ! -z "${AZURE_VM_IP}" ]; then
    echo -e "${YELLOW}Step 6: Deploying to Azure VM...${NC}"
    
    # Copy docker-compose file to VM
    scp docker-compose.prod.yml "${AZURE_VM_ADMIN}@${AZURE_VM_IP}:~/agileos/"
    scp .env "${AZURE_VM_ADMIN}@${AZURE_VM_IP}:~/agileos/"
    
    # SSH to VM and deploy
    ssh "${AZURE_VM_ADMIN}@${AZURE_VM_IP}" << 'ENDSSH'
        cd ~/agileos
        
        # Login to ACR from VM
        az acr login --name ${ACR_NAME}
        
        # Pull latest images
        docker-compose -f docker-compose.prod.yml pull
        
        # Stop old containers
        docker-compose -f docker-compose.prod.yml down
        
        # Start new containers
        docker-compose -f docker-compose.prod.yml up -d
        
        # Show status
        docker-compose -f docker-compose.prod.yml ps
ENDSSH
    
    echo -e "${GREEN}✓ Deployed to Azure VM successfully${NC}"
    echo -e "${GREEN}Application URL: http://${AZURE_VM_IP}${NC}"
else
    echo -e "${YELLOW}Step 6: Skipping VM deployment (AZURE_VM_IP not configured)${NC}"
fi

# Step 7: Health check
echo -e "${YELLOW}Step 7: Running health check...${NC}"
sleep 10

if [ ! -z "${AZURE_VM_IP}" ]; then
    HEALTH_URL="http://${AZURE_VM_IP}/health"
    
    if curl -f -s "${HEALTH_URL}" > /dev/null; then
        echo -e "${GREEN}✓ Health check passed${NC}"
    else
        echo -e "${RED}✗ Health check failed${NC}"
        exit 1
    fi
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Deployment completed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Next steps:"
echo "1. Configure DNS to point to ${AZURE_VM_IP}"
echo "2. Setup SSL certificates for HTTPS"
echo "3. Configure monitoring and alerts"
echo ""
