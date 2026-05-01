# AgileOS BPM - Monitoring & Logging Strategy

## Overview

AgileOS implements enterprise-grade monitoring and logging for production readiness, compliance, and operational excellence.

---

## Structured Logging with Zerolog

### Features

- **Multi-level logging**: DEBUG, INFO, WARN, ERROR, FATAL
- **Structured JSON output**: Machine-readable logs for analysis
- **Pretty console output**: Human-readable for development
- **File logging**: Persistent logs for audit trails
- **Context-rich**: Every log includes timestamp, caller, and custom fields

### Log Types

#### 1. Audit Logs
Tracks all business-critical actions for compliance:
```go
logger.LogAudit("workflow_updated", userID, "workflow:purchase_approval", map[string]interface{}{
    "changes": []string{"added_step", "updated_approval_threshold"},
    "version": "2.0",
})
```

#### 2. Security Logs
Monitors authentication and authorization events:
```go
logger.LogSecurity("failed_login", username, ipAddress, map[string]interface{}{
    "reason": "invalid_password",
    "attempts": 3,
})
```

#### 3. BPM Logs
Tracks workflow execution and task completion:
```go
logger.LogBPM("task_completed", workflowID, processID, map[string]interface{}{
    "task_id": taskID,
    "assigned_to": "role:manager",
    "duration_ms": 1500,
})
```

#### 4. Performance Logs
Monitors system performance metrics:
```go
logger.LogPerformance("database_query", duration, map[string]interface{}{
    "query_type": "SELECT",
    "table": "task_instance",
    "rows_returned": 150,
})
```

### Configuration

Environment variables:
```bash
LOG_LEVEL=info              # debug, info, warn, error
LOG_TO_FILE=true            # Enable file logging
LOG_FILE_PATH=./logs/agileos.log
```

---

## Health Check System

### Endpoints

#### 1. `/health` - Comprehensive Health Check
Returns detailed health status of all dependencies:

```json
{
  "status": "healthy",
  "timestamp": "2026-04-28T10:30:00Z",
  "uptime_seconds": 3600,
  "version": "1.0.0",
  "database": {
    "status": "up",
    "response_time_ms": 5,
    "message": "Database is healthy"
  },
  "message_broker": {
    "status": "up",
    "response_time_ms": 2,
    "message": "NATS is healthy",
    "details": {
      "servers": "nats://localhost:4223"
    }
  }
}
```

**Status Codes:**
- `200 OK`: All systems healthy
- `200 OK` + `status: degraded`: Some systems slow but operational
- `503 Service Unavailable`: Critical systems down

#### 2. `/health/live` - Liveness Probe
Lightweight check for Kubernetes/Azure:
```json
{
  "status": "alive",
  "timestamp": 1777378870
}
```

#### 3. `/health/ready` - Readiness Probe
Checks if service can accept traffic:
```json
{
  "status": "ready",
  "database": "connected",
  "message_broker": "connected"
}
```

### Azure Load Balancer Integration

Configure health probes in Azure:
```bash
# Liveness probe
Path: /health/live
Interval: 10s
Timeout: 5s

# Readiness probe
Path: /health/ready
Interval: 15s
Timeout: 5s
Unhealthy threshold: 3
```

---

## Container Resource Limits

### Docker Compose Configuration

Resource limits prevent container resource exhaustion:

```yaml
services:
  agileos-db:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 128M

  agileos-nats:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 64M
```

### Recommended Limits for Azure VM

**For Standard_B2s (2 vCPU, 4GB RAM):**
- SurrealDB: 1 CPU, 512MB RAM
- NATS: 0.5 CPU, 256MB RAM
- Backend Go: 0.5 CPU, 512MB RAM
- Frontend: 0.5 CPU, 256MB RAM
- Reserve: 0.5 CPU, 2.5GB RAM (OS + overhead)

---

## Monitoring Tools

### 1. Real-Time Container Monitor

**PowerShell (Windows):**
```powershell
.\scripts\monitor.ps1
```

**Bash (Linux/Mac):**
```bash
chmod +x scripts/monitor.sh
./scripts/monitor.sh
```

**Output:**
```
=========================================
AgileOS Container Resource Monitor
Time: 2026-04-28 19:30:00
=========================================

Container Resource Usage:
NAME          CPU %    MEM USAGE / LIMIT    MEM %    NET I/O
agileos-db    2.5%     128MB / 512MB        25%      1.2MB / 800KB
agileos-nats  0.8%     64MB / 256MB         25%      500KB / 300KB

=========================================
Container Health Status
=========================================
agileos-db: running (health: healthy)
agileos-nats: running (health: healthy)

=========================================
Backend Health Check
=========================================
Backend Status: healthy
  Database: up (5ms)
  NATS: up (2ms)
  Uptime: 3600s
```

### 2. Log Export for Azure

Export logs for Azure App Insights or analysis:

```powershell
.\scripts\export-logs.ps1 -OutputDir ".\logs" -Lines 1000
```

Creates timestamped directory with:
- Container logs (agileos-db.log, agileos-nats.log)
- Container stats (container-stats.txt)
- Container inspect data (JSON)
- Summary report

---

## Audit Trail & Compliance

### Database Audit Logs

All critical actions are logged to `audit_log` table:

```sql
SELECT * FROM audit_log 
WHERE type = 'security' 
  AND timestamp > time::now() - 24h
ORDER BY timestamp DESC;
```

### Workflow History

Every workflow change is versioned in `workflow_history`:

```sql
SELECT * FROM workflow_history 
WHERE workflow_id = 'workflow:purchase_approval'
ORDER BY changed_at DESC;
```

**Use Cases:**
- Compliance audits (SOX, GDPR)
- Forensic analysis
- Change tracking
- Rollback capability

---

## Azure App Insights Integration

### Setup

1. **Create App Insights Resource:**
```bash
az monitor app-insights component create \
  --app agileos-insights \
  --location eastus \
  --resource-group agileos-rg \
  --application-type web
```

2. **Get Instrumentation Key:**
```bash
az monitor app-insights component show \
  --app agileos-insights \
  --resource-group agileos-rg \
  --query instrumentationKey
```

3. **Configure Backend:**
```bash
export APPINSIGHTS_INSTRUMENTATIONKEY="your-key-here"
```

### Log Streaming

Stream logs to Azure:
```bash
# From Docker containers
docker logs -f agileos-backend | \
  az monitor app-insights events show \
    --app agileos-insights \
    --type trace
```

---

## Performance Monitoring

### Metrics Tracked

1. **HTTP Request Metrics**
   - Request duration
   - Status codes
   - Endpoint paths
   - Client IPs

2. **Database Metrics**
   - Query response times
   - Connection pool status
   - Failed queries

3. **BPM Metrics**
   - Workflow execution times
   - Task completion rates
   - Bottleneck detection

### Performance Logs

Automatically logged for slow operations:
```json
{
  "level": "info",
  "type": "performance",
  "operation": "workflow_execution",
  "duration_ms": 1500,
  "workflow_id": "workflow:purchase_approval",
  "process_id": "process:12345"
}
```

---

## Alerting Strategy

### Critical Alerts

1. **Service Down**
   - Health check returns 503
   - Action: Restart service, notify on-call

2. **High Error Rate**
   - >5% requests return 5xx
   - Action: Investigate logs, scale resources

3. **Resource Exhaustion**
   - CPU >80% for 5 minutes
   - Memory >90% for 5 minutes
   - Action: Scale up, optimize queries

4. **Security Events**
   - Multiple failed logins
   - Unauthorized access attempts
   - Action: Block IP, notify security team

### Azure Monitor Alerts

```bash
# Create alert rule
az monitor metrics alert create \
  --name high-cpu-alert \
  --resource-group agileos-rg \
  --scopes /subscriptions/{sub-id}/resourceGroups/agileos-rg/providers/Microsoft.Compute/virtualMachines/agileos-vm \
  --condition "avg Percentage CPU > 80" \
  --window-size 5m \
  --evaluation-frequency 1m \
  --action email admin@company.com
```

---

## Best Practices

### Development
- Use `LOG_LEVEL=debug` for detailed logs
- Monitor console output for immediate feedback
- Use `/health` endpoint to verify dependencies

### Staging
- Use `LOG_LEVEL=info`
- Enable file logging
- Test health check integration with load balancer
- Verify resource limits don't cause OOM

### Production
- Use `LOG_LEVEL=warn` or `error`
- Enable file logging with rotation
- Configure Azure App Insights
- Set up alerting rules
- Regular log analysis for security events
- Monthly audit log reviews

### Security
- Never log sensitive data (passwords, tokens)
- Sanitize user input in logs
- Restrict log file access (chmod 600)
- Encrypt logs at rest
- Regular log rotation and archival

---

## Troubleshooting

### High CPU Usage
```bash
# Check container stats
docker stats agileos-db agileos-nats

# Check backend logs
docker logs agileos-backend | grep "performance"

# Identify slow queries
grep "duration_ms" logs/agileos.log | sort -t: -k4 -n | tail -20
```

### Memory Leaks
```bash
# Monitor memory over time
watch -n 5 'docker stats --no-stream agileos-backend'

# Check for goroutine leaks
curl http://localhost:8081/debug/pprof/goroutine
```

### Failed Health Checks
```bash
# Test health endpoint
curl http://localhost:8081/health | jq

# Check database connectivity
curl http://localhost:8081/health/ready

# View detailed logs
docker logs agileos-backend --tail 100
```

---

## Interview Talking Points

### DevOps Expertise

> "I implemented enterprise-grade monitoring using Zerolog for structured logging, with separate audit trails for compliance. The system logs all security events, BPM operations, and performance metrics to both console and persistent storage."

### SRE Mindset

> "I designed comprehensive health checks with liveness and readiness probes for Azure Load Balancer integration. The system returns 503 when dependencies are down, ensuring traffic isn't routed to unhealthy instances."

### Production Readiness

> "I configured resource limits in Docker Compose to prevent resource exh