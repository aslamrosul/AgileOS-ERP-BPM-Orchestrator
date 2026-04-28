# AgileOS - Quick Start Guide

Panduan cepat untuk menjalankan AgileOS BPM Platform secara lokal.

## Prerequisites

- Docker Desktop (running)
- Go 1.22+
- Node.js 18+ & npm
- PowerShell (Windows)

## 🚀 Start in 5 Minutes

### 1. Start Infrastructure (Docker)

```bash
cd agile-os
docker-compose up -d
```

Wait 10 seconds, then verify:
```bash
docker-compose ps
```

Expected: Both `agileos-db` and `agileos-nats` are running.

### 2. Start Backend (Go)

```bash
cd backend-go
.\run-local.ps1
```

Expected output:
```
✓ Connected to SurrealDB: agileos/main
✓ Connected to NATS
🚀 AgileOS Engine running on port 8080
```

Test: `curl http://localhost:8080/health`

### 3. Start Frontend (Next.js)

Open NEW terminal:

```bash
cd frontend-next
npm install
npm run dev
```

Expected output:
```
- ready started server on 0.0.0.0:3000
- Local: http://localhost:3000
```

### 4. Open Browser

Navigate to: **http://localhost:3000**

Click "Open Workflow Builder" button.

## 🎨 Using Workflow Builder

### Create Your First Workflow

1. **Drag Nodes**: From left sidebar, drag nodes to canvas
   - Start node (gray)
   - Action nodes (green)
   - Approval nodes (blue)
   - End node (gray)

2. **Connect Nodes**: Click and drag from bottom handle of one node to top handle of another

3. **Edit Workflow Name**: Click "Untitled Workflow" at top to rename

4. **Save**: Click "Save" button (top right)

### Example: Simple Approval Workflow

```
[Start] 
   ↓
[Submit Request] (Action)
   ↓
[Manager Approval] (Approval)
   ↓
[End]
```

Steps:
1. Drag "Start" node to canvas
2. Drag "Action" node below it
3. Drag "Approval" node below that
4. Drag "End" node at bottom
5. Connect them in sequence
6. Click "Save"

## 📊 Architecture Overview

```
┌─────────────────────────────────────┐
│  Frontend (Next.js)                 │
│  http://localhost:3000              │
└─────────────────────────────────────┘
              ↓ REST API
┌─────────────────────────────────────┐
│  Backend (Go + Gin)                 │
│  http://localhost:8080              │
└─────────────────────────────────────┘
              ↓
┌──────────────────┐  ┌──────────────┐
│  SurrealDB       │  │    NATS      │
│  Port 8000       │  │  Port 4222   │
└──────────────────┘  └──────────────┘
```

## 🔍 Verify Everything Works

### Test Backend API

```bash
# Health check
curl http://localhost:8080/health

# Create workflow via API
curl -X POST http://localhost:8080/api/v1/workflow \
  -H "Content-Type: application/json" \
  -d '{
    "workflow": {
      "name": "Test Workflow",
      "version": "1.0.0",
      "is_active": true
    },
    "steps": [],
    "relations": []
  }'
```

### Check Database

1. Open SurrealDB Dashboard: http://localhost:8000
2. Login: `root` / `root`
3. Select namespace: `agileos`, database: `main`
4. Run query:
```sql
SELECT * FROM workflow;
```

## 🛠️ Troubleshooting

### Docker containers not starting

```bash
docker-compose down
docker-compose up -d
docker-compose logs -f
```

### Backend cannot connect to database

Check environment variables:
```bash
$env:SURREAL_URL="ws://localhost:8000/rpc"
$env:NATS_URL="nats://localhost:4222"
```

### Frontend cannot connect to backend

1. Check backend is running: `curl http://localhost:8080/health`
2. Check `.env.local` in frontend-next:
```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Port already in use

- Port 3000 (Frontend): Change with `npm run dev -- -p 3001`
- Port 8080 (Backend): Set `$env:PORT="8081"` before running
- Port 8000 (SurrealDB): Edit `docker-compose.yml`

### npm install fails

```bash
cd frontend-next
rm -rf node_modules package-lock.json
npm install
```

## 📝 Next Steps

After successfully running the platform:

1. **Seed Sample Data**
   ```bash
   cd backend-go
   .\scripts\seed-db.ps1
   ```

2. **Explore Sample Workflow**
   - Open SurrealDB Dashboard
   - Run queries from `database/seed.surql`

3. **Build Complex Workflows**
   - Use Decision nodes for branching
   - Add Notify nodes for alerts
   - Export/Import workflows as JSON

4. **API Integration**
   - Check `frontend-next/lib/api.ts` for API examples
   - Build custom integrations

## 🎯 Success Checklist

- [ ] Docker containers running
- [ ] Backend health check returns 200
- [ ] Frontend loads at localhost:3000
- [ ] Can drag nodes to canvas
- [ ] Can connect nodes
- [ ] Save workflow succeeds
- [ ] Workflow appears in database

## 📚 Documentation

- Backend Architecture: `backend-go/ARCHITECTURE.md`
- Database Guide: `backend-go/database/README.md`
- Frontend Guide: `frontend-next/README.md`
- Verification: `VERIFICATION.md`

## 🚦 Stopping Services

```bash
# Stop frontend: Ctrl+C in terminal

# Stop backend: Ctrl+C in terminal

# Stop Docker:
docker-compose down
```

## 💡 Tips

1. Keep all 3 terminals open (Docker logs, Backend, Frontend)
2. Use browser DevTools Network tab to debug API calls
3. Check backend logs for detailed error messages
4. Export workflows regularly as backup
5. Use SurrealDB Dashboard to inspect data

Happy building! 🎉
