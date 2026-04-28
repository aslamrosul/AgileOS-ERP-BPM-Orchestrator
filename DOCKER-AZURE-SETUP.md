# AgileOS BPM - Docker & Azure Setup Complete ✅

## What's Been Created

### 1. Production Docker Setup

#### Frontend Dockerfile (`frontend-next/Dockerfile`)
- ✅ Multi-stage build (builder + runner)
- ✅ Node 20 Alpine (minimal size)
- ✅ Non-root user for security
- ✅ Health check included
- ✅ Optimized for production

#### Production Docker Compose (`docker-compose.prod.yml`)
- ✅ All 5 services orchestrated:
  - SurrealDB (database)
  - NATS (message broker)
  - Backend (Go API)
  - Frontend (Next.js)
  - Nginx (reverse proxy)
- ✅ Health checks for all services
- ✅ Persistent volumes for data
- ✅ Internal network isolation
- ✅ Environment variable support

#### Nginx Reverse Proxy (`deploy/nginx/nginx.conf`)
- ✅ Routes `/api/*` to backend
- ✅ Routes `/` to frontend
- ✅ Rate limiting configured
- ✅ Gzip compression enabled
- ✅ Static file caching
- ✅ HTTPS ready (commented out)

### 2. Environment Configuration

#### Environment Template (`.env.example`)
- ✅ Database credentials
- ✅ Azure configuration
- ✅ Docker registry settings
- ✅ SSL/TLS paths
- ✅ Monitoring options

### 3. Azure Deployment

#### PowerShell Script (`deploy/azure/deploy.ps1`)
- ✅ Azure login automation
- ✅ Build Docker images
- ✅ Push to Azure Container Registry
- ✅ Deploy to Azure VM
- ✅ Health check verification
- ✅ Skip flags for flexibility

#### Bash Script (`deploy/azure/deploy.sh`)
- ✅ Same features as PowerShell
- ✅ Linux/Mac compatible

#### Terraform IaC (`deploy/azure/terraform/main.tf`)
- ✅ Complete Azure infrastructure:
  - Resource Group
  - Virtual Network
  - Network Security Group
  - Public IP
  - Virtual Machine (Ubuntu 22.04)
  - Azure Container Registry
- ✅ Auto-install Docker on VM
- ✅ Security rules configured
- ✅ Outputs for easy access

### 4. Documentation

#### Deployment Guide (`deploy/DEPLOYMENT.md`)
- ✅ Step-by-step instructions
- ✅ Multiple deployment options
- ✅ SSL/HTTPS setup guide
- ✅ Monitoring & logging
- ✅ Backup & recovery
- ✅ Troubleshooting tips
- ✅ Security best practices

#### Quick Start (`deploy/quickstart.ps1`)
- ✅ One-command local testing
- ✅ Automatic health checks
- ✅ Service status display

### 5. Docker Optimization

#### .dockerignore Files
- ✅ Backend: Excludes tests, docs, dev files
- ✅ Frontend: Excludes node_modules, .next, tests

#### .gitignore
- ✅ Environment files
- ✅ SSL certificates
- ✅ Terraform state
- ✅ Build artifacts

---

## Quick Start Guide

### Local Testing (Production Mode)

```powershell
# 1. Setup environment
cp .env.example .env
# Edit .env with your values

# 2. Run quick start
.\deploy\quickstart.ps1

# 3. Access application
# Frontend: http://localhost
# Backend:  http://localhost/api
# Database: http://localhost:8000
```

### Deploy to Azure

```powershell
# 1. Configure environment
cp .env.example .env
# Fill in Azure details

# 2. Provision infrastructure (optional)
cd deploy/azure/terraform
terraform init
terraform apply

# 3. Deploy application
cd ../../..
.\deploy\azure\deploy.ps1

# 4. Access your app
# http://YOUR_VM_IP
```

---

## Architecture Overview

```
Internet
    ↓
[Nginx :80/443]
    ↓
    ├─→ /api/* → [Backend :8080] ←→ [SurrealDB :8000]
    │                ↓
    │           [NATS :4222]
    │
    └─→ /* → [Frontend :3000]
```

---

## File Structure

```
agile-os/
├── docker-compose.prod.yml          # Production orchestration
├── .env.example                     # Environment template
├── .gitignore                       # Git ignore rules
│
├── backend-go/
│   ├── Dockerfile                   # Backend container
│   └── .dockerignore               # Docker ignore
│
├── frontend-next/
│   ├── Dockerfile                   # Frontend container (NEW)
│   ├── .dockerignore               # Docker ignore (NEW)
│   └── next.config.mjs             # Updated for standalone
│
└── deploy/
    ├── DEPLOYMENT.md               # Full deployment guide
    ├── quickstart.ps1              # Quick local test
    │
    ├── nginx/
    │   └── nginx.conf              # Reverse proxy config
    │
    └── azure/
        ├── deploy.ps1              # Windows deployment
        ├── deploy.sh               # Linux/Mac deployment
        │
        └── terraform/
            └── main.tf             # Infrastructure as Code
```

---

## Environment Variables Reference

### Required for Production

```bash
# Database
SURREAL_USER=root
SURREAL_PASS=your_secure_password

# Azure
AZURE_VM_IP=your_vm_public_ip
ACR_NAME=your_acr_name
ACR_LOGIN_SERVER=your_acr.azurecr.io
```

### Optional

```bash
# Docker Registry (if not using ACR)
DOCKER_REGISTRY=docker.io
DOCKER_USERNAME=your_username
DOCKER_PASSWORD=your_password

# SSL/TLS
SSL_CERT_PATH=./deploy/nginx/ssl/cert.pem
SSL_KEY_PATH=./deploy/nginx/ssl/key.pem
```

---

## Next Steps

### 1. Local Testing
```powershell
.\deploy\quickstart.ps1
```

### 2. Azure Deployment
```powershell
# Provision infrastructure
cd deploy/azure/terraform
terraform apply

# Deploy application
cd ../../..
.\deploy\azure\deploy.ps1
```

### 3. Setup SSL
```bash
# On Azure VM
sudo certbot certonly --standalone -d your-domain.com
```

### 4. Configure DNS
Point your domain to Azure VM public IP

### 5. Enable Monitoring
- Azure Monitor
- Application Insights
- Log Analytics

---

## Cost Estimate (Azure)

### Minimal Setup
- VM: Standard_B2s (~$30/month)
- ACR: Basic (~$5/month)
- Storage: ~$2/month
- **Total: ~$37/month**

### Production Setup
- VM: Standard_B4ms (~$120/month)
- ACR: Standard (~$20/month)
- Load Balancer (~$20/month)
- Storage: ~$10/month
- **Total: ~$170/month**

---

## Support & Troubleshooting

### View Logs
```powershell
docker-compose -f docker-compose.prod.yml logs -f
```

### Restart Services
```powershell
docker-compose -f docker-compose.prod.yml restart
```

### Check Health
```powershell
curl http://localhost/health
```

### Full Documentation
See `deploy/DEPLOYMENT.md` for complete guide

---

**Status**: ✅ Production Ready
**Last Updated**: April 2026
**Version**: 1.0.0
