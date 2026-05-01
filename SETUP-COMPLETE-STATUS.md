# ✅ AgileOS Setup Complete - Status Update

## 🎉 Yang Sudah Selesai

### 1. Database Setup ✅
- **SurrealDB v1.4.2** running di port 8002
- **Schema v1.4** sudah di-apply dengan semua fitur:
  - Graph relationships (TYPE RELATION)
  - Full-text search
  - Custom functions
  - Events & audit trails
  - Permissions berbasis role
- **Sample data** sudah di-seed:
  - 1 workflow: "Purchase Request Approval"
  - 6 steps dengan graph edges
  - 1 process instance untuk testing

### 2. Backend Go ✅
- **go.mod** sudah dibuat ulang
- **Build berhasil** tanpa error
- **Server running** di port 8080
- **Koneksi ke SurrealDB** berhasil (ws://localhost:8002/rpc)
- **Koneksi ke NATS** berhasil (nats://localhost:4222)
- **WebSocket Hub** aktif
- **Logs directory** sudah dibuat

### 3. Scripts & Automation ✅
- `FRESH-START-DB.ps1` - Fresh start database
- `backend-go/scripts/apply-schema-v1.4.ps1` - Apply schema
- `backend-go/scripts/seed-db.ps1` - Seed database (fixed, no emoji)
- `backend-go/run-local.ps1` - Run backend (updated port 8002)

## 🔧 Minor Issues (Non-blocking)

### Health Check Query
- Health endpoint mengembalikan "unhealthy" karena query `SELECT 1` tidak valid di SurrealDB
- **Fix sudah dibuat**: Ganti dengan `SELECT * FROM workflow LIMIT 1;`
- **Perlu restart backend** untuk apply fix

## 🚀 Cara Menjalankan

### 1. Start Database (Sudah Running)
```powershell
# Database sudah running di port 8002
docker ps | Select-String "agileos-db"
```

### 2. Start Backend
```powershell
cd backend-go

# Rebuild dengan fix health check
go build -o agileos-engine.exe .

# Run backend
.\run-local.ps1
```

Backend akan jalan di: http://localhost:8080

### 3. Test API

#### Health Check
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
```

#### Login (Get Token)
```powershell
$body = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"

$token = $response.token
```

#### Get Workflows (Authenticated)
```powershell
$headers = @{
    "Authorization" = "Bearer $token"
}

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/workflows" `
    -Method GET `
    -Headers $headers
```

## 📊 API Endpoints

### Public Endpoints
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/register` - Register
- `POST /api/v1/auth/refresh` - Refresh token

### Protected Endpoints (Require Authentication)
- `GET /api/v1/workflows` - List workflows
- `GET /api/v1/workflow/:id` - Get workflow detail
- `GET /api/v1/tasks/pending/:assignedTo` - Get pending tasks
- `POST /api/v1/task/:id/complete` - Complete task
- `POST /api/v1/process/start` - Start process (manager/admin)
- `POST /api/v1/workflow` - Create workflow (admin only)

### Analytics Endpoints (Manager/Admin)
- `GET /api/v1/analytics/overview` - Analytics overview
- `GET /api/v1/analytics/workflows` - Workflow efficiency
- `GET /api/v1/analytics/steps` - Step performance
- `GET /api/v1/analytics/departments` - Department metrics
- `GET /api/v1/analytics/summary` - Summary
- `GET /api/v1/analytics/insights` - AI insights

### WebSocket
- `ws://localhost:8080/ws` - Real-time notifications

## 🗄️ Database Info

**Connection:**
- URL: http://localhost:8002
- WebSocket: ws://localhost:8002/rpc
- Username: root
- Password: root
- Namespace: agileos
- Database: main

**Verify Data:**
```powershell
$auth = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("root:root"))
$headers = @{
    "Accept" = "application/json"
    "NS" = "agileos"
    "DB" = "main"
    "Authorization" = "Basic $auth"
}

# Get workflows
$result = Invoke-RestMethod -Uri "http://localhost:8002/sql" `
    -Method POST `
    -Headers $headers `
    -Body "SELECT * FROM workflow;" `
    -ContentType "text/plain"

$result[0].result
```

## 📝 Next Steps

### 1. Restart Backend (Apply Health Check Fix)
```powershell
# Stop current backend (Ctrl+C)
# Rebuild
cd backend-go
go build -o agileos-engine.exe .

# Run again
.\run-local.ps1
```

### 2. Create Test User
Perlu buat user di database untuk testing login:
```surql
-- Via SurrealDB CLI atau HTTP API
USE NS agileos DB main;

CREATE user:admin SET
    username = "admin",
    password = crypto::argon2::generate("admin123"),
    email = "admin@agileos.com",
    role = "admin",
    created_at = time::now();
```

### 3. Start Frontend
```powershell
cd frontend-next
npm install
npm run dev
```
Frontend akan jalan di: http://localhost:3000

### 4. Start Analytics Service (Optional)
```powershell
cd analytics-py
pip install -r requirements.txt
python main.py
```
Analytics akan jalan di: http://localhost:8001

## 🐛 Troubleshooting

### Backend tidak bisa connect ke database
```powershell
# Cek container running
docker ps | Select-String "agileos-db"

# Cek logs
docker logs agileos-db --tail 50

# Test health
Invoke-WebRequest -Uri "http://localhost:8002/health" -UseBasicParsing
```

### Port sudah digunakan
```powershell
# Cek port 8080
netstat -ano | findstr :8080

# Kill process jika perlu
taskkill /PID <PID> /F
```

### Database kosong
```powershell
# Re-seed database
.\backend-go\scripts\seed-db.ps1
```

## 📚 Documentation

- `DATABASE-SETUP-COMPLETE.md` - Database setup guide
- `DATABASE-SEED-GUIDE.md` - Seeding guide
- `SURREALDB-V1.4-SCHEMA-VERIFICATION.md` - Schema verification
- `backend-go/database/surreal_v14_migration.go` - Migration guide
- `README.md` - Main documentation
- `QUICKSTART.md` - Quick start guide

## ✅ Checklist

- [x] SurrealDB v1.4.2 running
- [x] Schema applied
- [x] Database seeded
- [x] go.mod created
- [x] Backend compiled
- [x] Backend running
- [x] Connected to SurrealDB
- [x] Connected to NATS
- [x] WebSocket Hub active
- [x] Health check fix prepared
- [ ] Backend restarted with fix
- [ ] Test user created
- [ ] API tested with authentication
- [ ] Frontend running
- [ ] End-to-end workflow tested

## 🎯 Status: READY FOR TESTING!

Backend sudah running dan siap untuk testing. Tinggal:
1. Restart backend untuk apply health check fix
2. Buat test user untuk login
3. Test API endpoints
4. Start frontend

Semua komponen utama sudah berfungsi! 🚀
