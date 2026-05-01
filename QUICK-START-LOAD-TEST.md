# Quick Start: Load Testing AgileOS

## Prerequisites

### Option 1: k6 (Recommended)
```bash
# Windows
choco install k6

# Mac
brew install k6

# Linux
sudo apt-get install k6
```

### Option 2: Apache Bench (Alternative)
```bash
# Linux/WSL
sudo apt-get install apache2-utils curl jq

# Mac
brew install apache2-utils curl jq
```

## Quick Start (Automated)

### 1. Start AgileOS
```bash
cd agile-os
docker-compose up -d
```

### 2. Run Load Test
```powershell
# Automated test with monitoring
.\RUN-LOAD-TEST.ps1

# Or with custom settings
.\RUN-LOAD-TEST.ps1 -TestType k6 -Duration 600 -BaseURL http://localhost:8081
```

## Manual Testing

### k6 Test
```bash
# Default test (50-200 users, 17 minutes)
k6 run stress-test.js

# Quick test (50 users, 2 minutes)
k6 run --vus 50 --duration 2m stress-test.js

# Custom base URL
BASE_URL=http://your-server:8081 k6 run stress-test.js
```

### Bash Test
```bash
# Default test
./load-test.sh

# Custom configuration
CONCURRENT_USERS=100 TOTAL_REQUESTS=5000 ./load-test.sh
```

## Performance Monitoring

### Start Monitoring
```powershell
# Monitor for 20 minutes
.\scripts\performance-monitor.ps1 -Duration 1200 -Interval 5
```

### Combined Test + Monitoring
```powershell
# Terminal 1: Start monitoring
.\scripts\performance-monitor.ps1 -Duration 1200

# Terminal 2: Run load test
k6 run stress-test.js
```

## Apply Performance Optimizations

### 1. Install Go Dependencies
```bash
cd backend-go
go get github.com/patrickmn/go-cache
go mod tidy
```

### 2. Apply Database Indexes
```bash
surreal import --conn http://localhost:8002 --user root --pass root \
  --ns agileos --db main backend-go/database/create-indexes.surql
```

### 3. Restart Services
```bash
docker-compose restart
```

## View Results

### k6 Results
```bash
# View JSON results
cat load-test-results-*.json

# View summary in terminal (already displayed)
```

### Bash Results
```bash
# View detailed report
cat load-test-results/load-test-*.txt

# View TSV data
cat load-test-results/*.tsv
```

### Performance Monitoring
```powershell
# View CSV data
Import-Csv performance-report-*.csv | Format-Table

# View in Excel
start performance-report-*.csv
```

## Expected Results

### Before Optimization
- Requests/sec: ~120 RPS
- P95 Response Time: 1800ms
- Error Rate: 8%

### After Optimization
- Requests/sec: 450+ RPS
- P95 Response Time: 420ms
- Error Rate: <2%

## Troubleshooting

### "k6 not found"
```bash
# Install k6
choco install k6  # Windows
brew install k6   # Mac
```

### "Backend not accessible"
```bash
# Check if containers are running
docker ps

# Check backend logs
docker logs agileos-backend

# Restart services
docker-compose restart
```

### "High error rate"
```bash
# Check resource limits
docker stats

# Increase Docker memory
# Docker Desktop > Settings > Resources > Memory: 4GB+

# Reduce concurrent users
k6 run --vus 50 stress-test.js
```

## Performance Tuning

### Increase NATS Workers
```go
// backend-go/messaging/nats.go
workerPool := make(chan struct{}, 100) // Increase from 50
```

### Adjust Cache TTL
```go
// Shorter TTL for frequently changing data
cache.SetWithExpiration("analytics", data, 1*time.Minute)

// Longer TTL for static data
cache.SetWithExpiration("workflows", data, 10*time.Minute)
```

### Increase Docker Resources
```yaml
# docker-compose.yml
services:
  agileos-backend:
    deploy:
      resources:
        limits:
          cpus: '4.0'
          memory: 4G
```

## Next Steps

1. Run baseline test
2. Apply optimizations
3. Re-run test
4. Compare results
5. Deploy to Azure
6. Monitor production

## Documentation

- Full Guide: `PERFORMANCE-OPTIMIZATION.md`
- Task Summary: `TASK-15-PERFORMANCE-COMPLETED.md`
- Troubleshooting: See "Troubleshooting" section in `PERFORMANCE-OPTIMIZATION.md`

---

**Quick Commands**:
```bash
# Start system
docker-compose up -d

# Run test
.\RUN-LOAD-TEST.ps1

# View results
cat load-test-results-*.json

# Stop system
docker-compose down
```
