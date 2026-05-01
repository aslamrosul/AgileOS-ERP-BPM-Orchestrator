# Quick Start - Monitoring & Logging

## Setup Monitoring dalam 5 Menit

### 1. Start Services dengan Resource Limits

```powershell
cd agile-os
docker-compose up -d
```

Resource limits sudah dikonfigurasi:
- SurrealDB: Max 1 CPU, 512MB RAM
- NATS: Max 0.5 CPU, 256MB RAM

### 2. Start Backend dengan Logging

```powershell
cd backend-go

# Set environment variables
$env:SURREAL_URL="ws://localhost:8002/rpc"
$env:NATS_URL="nats://localhost:4223"
$env:PORT="8081"
$env:JWT_SECRET="your-super-secret-jwt-key-change-in-production-min-32-chars"
$env:LOG_LEVEL="info"
$env:LOG_TO_FILE="true"
$env:LOG_FILE_PATH="./logs/agileos.log"

# Create logs directory
New-Item -ItemType Directory -Force -Path logs

# Run backend
go run main.go
```

### 3. Test Health Checks

```powershell
# Comprehensive health check
Invoke-RestMethod -Uri "http://localhost:8081/health" | ConvertTo-Json

# Liveness probe
Invoke-RestMethod -Uri "http://localhost:8081/health/live"

# Readiness probe
Invoke-RestMethod -Uri "http://localhost:8081/health/ready"
```

### 4. Start Real-Time Monitor

Buka terminal baru:

```powershell
cd agile-os
.\scripts\monitor.ps1
```

Akan menampilkan:
- Container CPU & Memory usage
- Container health status
- Backend health check results
- Auto-refresh setiap 5 detik

### 5. View Logs

**Console logs** (pretty formatted):
```powershell
# Backend logs sudah tampil di console dengan warna
```

**File logs** (JSON structured):
```powershell
Get-Content backend-go\logs\agileos.log -Tail 20
```

**Container logs**:
```powershell
docker logs agileos-db --tail 50
docker logs agileos-nats --tail 50
```

### 6. Export Logs

```powershell
cd agile-os
.\scripts\export-logs.ps1 -OutputDir ".\logs" -Lines 1000
```

Creates timestamped folder dengan semua logs.

---

## Monitoring Dashboard

### Container Stats

```powershell
docker stats agileos-db agileos-nats
```

Output:
```
NAME          CPU %    MEM USAGE / LIMIT    MEM %
agileos-db    2.5%     128MB / 512MB        25%
agileos-nats  0.8%     64MB / 256MB         25%
```

### Health Status

```powershell
curl http://localhost:8081/health | jq
```

Output:
```json
{
  "status": "healthy",
  "database": {
    "status": "up",
    "response_time_ms": 5
  },
  "message_broker": {
    "status": "up",
    "response_time_ms": 2
  },
  "uptime_seconds": 3600
}
```

---

## Log Types & Examples

### 1. HTTP Request Logs

Setiap request otomatis di-log:
```json
{
  "level": "info",
  "method": "POST",
  "path": "/api/v1/auth/login",
  "status": 200,
  "duration_ms": 45,
  "ip": "::1",
  "message": "HTTP Request"
}
```

### 2. Audit Logs

Workflow changes:
```json
{
  "level": "info",
  "type": "audit",
  "action": "workflow_updated",
  "user_id": "user:admin",
  "resource": "workflow:purchase_approval",
  "message": "Audit log"
}
```

### 3. Security Logs

Failed logins:
```json
{
  "level": "warn",
  "type": "security",
  "event": "failed_login",
  "user_id": "admin",
  "ip_address": "192.168.1.100",
  "message": "Security event"
}
```

### 4. BPM Logs

Task completion:
```json
{
  "level": "info",
  "type": "bpm",
  "event": "task_completed",
  "workflow_id": "workflow:purchase_approval",
  "process_id": "process:12345",
  "message": "BPM event"
}
```

---

## Troubleshooting

### Backend tidak start

Check logs:
```powershell
# Lihat error di console
# Atau check file log
Get-Content backend-go\logs\agileos.log | Select-String "error"
```

### Health check returns 503

```powershell
# Check database
docker ps | Select-String "agileos-db"

# Check NATS
docker ps | Select-String "agileos-nats"

# Restart containers
docker-compose restart
```

### High CPU/Memory

```powershell
# Monitor real-time
.\scripts\monitor.ps1

# Check resource limits
docker inspect agileos-db | Select-String "Memory"
```

---

## Production Checklist

- [ ] Resource limits configured in docker-compose.yml
- [ ] LOG_LEVEL set to "warn" or "error"
- [ ] LOG_TO_FILE enabled
- [ ] Health checks tested
- [ ] Monitoring script tested
- [ ] Log export tested
- [ ] Azure App Insights configured (optional)
- [ ] Alert rules configured
- [ ] Log rotation configured

---

## Next Steps

1. **Analytics Dashboard**: See `START-ANALYTICS.md`
2. **Azure Deployment**: See `DOCKER-AZURE-SETUP.md`
3. **Security**: See `SECURITY.md`
4. **Full Documentation**: See `MONITORING-LOGGING.md`
