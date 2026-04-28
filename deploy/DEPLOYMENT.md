# AgileOS BPM Platform - Deployment Guide

## Overview

This guide covers deploying AgileOS BPM Platform to Microsoft Azure using Docker containers.

## Prerequisites

- Azure account with active subscription
- Azure CLI installed
- Docker and Docker Compose installed
- SSH key pair generated (`ssh-keygen`)
- Terraform (optional, for infrastructure provisioning)

## Deployment Options

### Option 1: Manual Deployment to Azure VM

#### Step 1: Provision Infrastructure

**Using Terraform (Recommended):**
```bash
cd deploy/azure/terraform
terraform init
terraform plan
terraform apply
```

**Or manually create:**
- Resource Group
- Virtual Network
- VM (Ubuntu 22.04, Standard_B2s or higher)
- Network Security Group (allow ports 80, 443, 22)
- Public IP address

#### Step 2: Configure Environment

1. Copy environment template:
```bash
cp .env.example .env
```

2. Edit `.env` and fill in your values:
```bash
# Database credentials
SURREAL_USER=root
SURREAL_PASS=your_secure_password

# Azure details
AZURE_VM_IP=your_vm_public_ip
AZURE_VM_ADMIN=azureuser
ACR_NAME=your_acr_name
```

#### Step 3: Deploy

**On Windows:**
```powershell
.\deploy\azure\deploy.ps1
```

**On Linux/Mac:**
```bash
chmod +x deploy/azure/deploy.sh
./deploy/azure/deploy.sh
```

The script will:
1. Login to Azure
2. Build Docker images
3. Push to Azure Container Registry
4. Deploy to VM
5. Run health checks

#### Step 4: Verify Deployment

```bash
# Check application health
curl http://YOUR_VM_IP/health

# Check backend API
curl http://YOUR_VM_IP/api/v1/health

# Access frontend
open http://YOUR_VM_IP
```

### Option 2: Azure Container Instances (ACI)

For simpler deployment without managing VMs:

```bash
# Create container group
az container create \
  --resource-group agileos-rg \
  --name agileos-app \
  --image your_acr.azurecr.io/agileos-backend:latest \
  --dns-name-label agileos-app \
  --ports 80 443
```

### Option 3: Azure Kubernetes Service (AKS)

For production-grade scalability:

```bash
# Create AKS cluster
az aks create \
  --resource-group agileos-rg \
  --name agileos-aks \
  --node-count 2 \
  --enable-addons monitoring \
  --generate-ssh-keys

# Deploy using Helm or kubectl
kubectl apply -f deploy/kubernetes/
```

## Local Testing

Test the production setup locally before deploying:

```bash
# Build and run with production compose
docker-compose -f docker-compose.prod.yml up --build

# Access application
open http://localhost
```

## SSL/HTTPS Setup

### Using Let's Encrypt (Free)

1. Install Certbot on VM:
```bash
ssh azureuser@YOUR_VM_IP
sudo apt-get install certbot
```

2. Generate certificate:
```bash
sudo certbot certonly --standalone -d your-domain.com
```

3. Copy certificates:
```bash
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem ~/agileos/deploy/nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem ~/agileos/deploy/nginx/ssl/key.pem
```

4. Update nginx.conf to enable HTTPS server block

5. Restart containers:
```bash
docker-compose -f docker-compose.prod.yml restart agileos-nginx
```

## Monitoring & Logging

### View Container Logs

```bash
# All containers
docker-compose -f docker-compose.prod.yml logs -f

# Specific service
docker-compose -f docker-compose.prod.yml logs -f agileos-backend
```

### Health Checks

```bash
# Overall health
curl http://YOUR_VM_IP/health

# Backend health
curl http://YOUR_VM_IP/api/v1/health

# Database health
curl http://YOUR_VM_IP:8000/health

# NATS health
curl http://YOUR_VM_IP:8222/healthz
```

### Azure Monitor Integration

Add to docker-compose.prod.yml:

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## Scaling

### Horizontal Scaling

Update docker-compose.prod.yml:

```yaml
agileos-backend:
  deploy:
    replicas: 3
```

### Vertical Scaling

Resize Azure VM:

```bash
az vm resize \
  --resource-group agileos-rg \
  --name agileos-vm \
  --size Standard_B4ms
```

## Backup & Recovery

### Database Backup

```bash
# Backup SurrealDB
docker exec agileos-db-prod surreal export \
  --conn http://localhost:8000 \
  --user root --pass root \
  --ns agileos --db main \
  backup.surql

# Copy to local
docker cp agileos-db-prod:/backup.surql ./backups/
```

### Restore Database

```bash
# Copy backup to container
docker cp ./backups/backup.surql agileos-db-prod:/backup.surql

# Restore
docker exec agileos-db-prod surreal import \
  --conn http://localhost:8000 \
  --user root --pass root \
  --ns agileos --db main \
  /backup.surql
```

## Troubleshooting

### Container won't start

```bash
# Check logs
docker-compose -f docker-compose.prod.yml logs agileos-backend

# Check container status
docker-compose -f docker-compose.prod.yml ps

# Restart specific service
docker-compose -f docker-compose.prod.yml restart agileos-backend
```

### Network issues

```bash
# Check network
docker network inspect agileos-prod-network

# Test connectivity between containers
docker exec agileos-backend ping agileos-db
```

### Performance issues

```bash
# Check resource usage
docker stats

# Check VM resources
az vm show -d \
  --resource-group agileos-rg \
  --name agileos-vm \
  --query "{Name:name, PowerState:powerState, Size:hardwareProfile.vmSize}"
```

## Security Best Practices

1. **Change default passwords** in `.env`
2. **Enable firewall** on Azure NSG
3. **Use HTTPS** with valid SSL certificates
4. **Regular updates**: `docker-compose pull && docker-compose up -d`
5. **Backup regularly**: Automate database backups
6. **Monitor logs**: Set up Azure Monitor alerts
7. **Limit SSH access**: Use Azure Bastion or VPN

## Cost Optimization

- Use **Azure Reserved Instances** for 1-3 year commitments (up to 72% savings)
- Enable **auto-shutdown** for non-production VMs
- Use **Azure Spot VMs** for dev/test environments
- Monitor costs with **Azure Cost Management**

## Support

For issues or questions:
- Check logs: `docker-compose logs`
- Review documentation: `/docs`
- Open issue on GitHub

---

**Last Updated**: April 2026
**Version**: 1.0.0
