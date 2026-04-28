# AgileOS - Verification Guide

Panduan untuk memverifikasi bahwa sistem AgileOS BPM sudah berjalan dengan benar.

## Prerequisites

- Docker Desktop running
- Go 1.22+ installed
- PowerShell (Windows)

## Step-by-Step Verification

### 1. Verify Infrastructure (Docker)

```bash
cd agile-os
docker-compose ps
```

Expected output:
```
NAME           STATUS
agileos-db     Up (healthy)
agileos-nats   Up (healthy)
```

Jika tidak healthy, restart:
```bash
docker-compose down
docker-compose up -d
```

### 2. Verify SurrealDB

Test koneksi ke SurrealDB:

```bash
curl http://localhost:8000/health
```

Expected: HTTP 200 OK

Atau buka browser: `http://localhost:8000`
- Login: `root` / `root`
- Namespace: `agileos`
- Database: `main`

### 3. Verify NATS

Test NATS monitoring:

```bash
curl http://localhost:8222/healthz
```

Expected: `ok`

### 4. Build Backend

```bash
cd backend-go
go mod tidy
go build -o agileos-engine.exe .
```

Expected: No errors, `agileos-engine.exe` created

### 5. Run Backend

```bash
.\run-local.ps1
```

Expected output:
```
🚀 Starting AgileOS Engine (Local Development Mode)
   Database: ws://localhost:8000/rpc
   NATS: nats://localhost:4222
   Port: 8080

✓ Connected to SurrealDB: agileos/main
✓ Connected to NATS
🚀 AgileOS Engine running on port 8080
```

### 6. Test Health Endpoint

Di terminal baru:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "engine_running",
  "database": true,
  "message_broker": true,
  "timestamp": 1714320000
}
```

### 7. Seed Database

```bash
.\scripts\seed-db.ps1
```

Follow instructions untuk seed via SurrealDB Dashboard.

Copy content dari `database/seed.surql` dan execute di dashboard.

### 8. Verify Database Schema

Di SurrealDB Dashboard (`http://localhost:8000`), jalankan:

```sql
-- Check workflows
SELECT * FROM workflow;
```

Expected: 1 workflow (Purchase Request Approval)

```sql
-- Check steps
SELECT * FROM step WHERE workflow_id = "workflow:purchase_approval";
```

Expected: 6 steps (pr_submit, manager_review, finance_approval, procurement, completed, rejected)

```sql
-- Check graph relations
SELECT id, name, ->next->step.name AS next_steps 
FROM step 
WHERE workflow_id = "workflow:purchase_approval";
```

Expected: Each step shows its next steps

### 9. Test Graph Traversal

```sql
-- Get next steps from pr_submit
SELECT ->next->step.* AS next_steps FROM step:pr_submit;
```

Expected: Returns manager_review step

```sql
-- Get next steps from manager_review (conditional branching)
SELECT ->next->step.* AS next_steps FROM step:manager_review;
```

Expected: Returns both finance_approval and rejected steps

### 10. Verify Process Instance

```sql
-- Check sample process instance
SELECT * FROM process_instance:test_001;
```

Expected: 1 process instance with status "running"

## Troubleshooting

### SurrealDB tidak start

**Symptom**: `docker-compose ps` shows agileos-db as exited

**Solution**:
```bash
docker-compose logs agileos-db
docker-compose down
docker-compose up -d
```

### Backend tidak bisa connect ke SurrealDB

**Symptom**: `Failed to connect to SurrealDB: dial tcp`

**Solution**: Pastikan environment variables benar
```bash
$env:SURREAL_URL="ws://localhost:8000/rpc"
$env:NATS_URL="nats://localhost:4222"
```

### Port already in use

**Symptom**: `bind: address already in use`

**Solution**: 
- Port 8000: Stop aplikasi lain yang pakai port ini
- Port 8080: Ubah `$env:PORT="8081"` sebelum run backend
- Port 4222: Stop NATS lain yang running

### Go module errors

**Symptom**: `missing go.sum entry`

**Solution**:
```bash
cd backend-go
go mod tidy
go mod download
```

## Success Criteria

✅ Docker containers running dan healthy
✅ SurrealDB accessible di port 8000
✅ NATS accessible di port 4222
✅ Backend compile tanpa error
✅ Backend running di port 8080
✅ Health endpoint returns 200 OK
✅ Database seeded dengan sample workflow
✅ Graph relations verified
✅ Graph traversal working

## Next Steps

Setelah semua verification passed:

1. **Prompt 3**: Build REST API endpoints untuk workflow management
2. **Prompt 4**: Implement BPM orchestration engine
3. **Prompt 5**: Add authentication & authorization
4. **Prompt 6**: Build Next.js dashboard

## Architecture Verification

Verify arsitektur dengan diagram:

```
✓ Frontend (Next.js) - TBD
       ↓
✓ Backend (Go + Gin) - Port 8080
       ↓
✓ SurrealDB - Port 8000 (Graph Database)
✓ NATS - Port 4222 (Message Broker)
```

## Database Schema Verification

```
✓ workflow table - Workflow definitions
✓ step table - Workflow steps
✓ next relation - Graph edges between steps
✓ process_instance table - Running workflows
```

## Code Structure Verification

```
✓ models/workflow.go - Data models defined
✓ database/surreal.go - Repository pattern implemented
✓ database/seed.surql - Sample data ready
✓ main.go - Entry point with DB connection
✓ Dockerfile - Multi-stage build ready
✓ docker-compose.yml - Infrastructure orchestration
```

Semua komponen sudah siap untuk fase berikutnya! 🚀
