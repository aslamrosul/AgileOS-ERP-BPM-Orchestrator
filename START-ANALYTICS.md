# Quick Start - Analytics Dashboard

Ikuti langkah-langkah ini untuk menjalankan Analytics Dashboard:

## 1. Start Docker Containers

```powershell
cd agile-os
docker-compose up -d
```

Tunggu 10 detik untuk containers siap.

## 2. Start Backend

Buka terminal baru dan jalankan:

```powershell
cd agile-os/backend-go
$env:SURREAL_URL="ws://localhost:8002/rpc"
$env:NATS_URL="nats://localhost:4223"
$env:PORT="8081"
$env:JWT_SECRET="your-super-secret-jwt-key-change-in-production-min-32-chars"
go run main.go
```

Tunggu sampai muncul: `🚀 AgileOS Engine running on port 8081`

## 3. Create Admin User

Buka terminal baru dan jalankan:

```powershell
$body = @{ username = "admin"; password = "password123"; email = "admin@agileos.com"; full_name = "Admin User"; role = "admin" } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/register" -Method POST -Body $body -ContentType "application/json"
```

Jika error "already exists", jalankan ini untuk activate user:

```powershell
$auth = "Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("root:root"))
$query = "UPDATE user SET is_active = true"
$body = @{ query = $query } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8002/sql" -Method POST -Body $body -ContentType "application/json" -Headers @{ "NS" = "agileos"; "DB" = "main"; "Authorization" = $auth }
```

## 4. Login dan Get Token

```powershell
$loginBody = @{ username = "admin"; password = "password123" } | ConvertTo-Json
$response = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
$token = $response.access_token
Write-Host "Token: $token"
```

## 5. Generate Test Data (Optional)

```powershell
# Create some test tasks
for ($i = 1; $i -le 10; $i++) {
    $taskData = @{
        step_id = "step:approval_$i"
        step_name = "Approval Step $i"
        assigned_to = if ($i % 3 -eq 0) { "Finance" } elseif ($i % 2 -eq 0) { "Manager" } else { "Employee" }
        status = if ($i % 4 -eq 0) { "completed" } else { "pending" }
        process_instance_id = "process:test_$i"
    } | ConvertTo-Json
    
    # Insert directly to database
    $auth = "Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("root:root"))
    $query = "CREATE task_instance CONTENT $taskData"
    $body = @{ query = $query } | ConvertTo-Json
    Invoke-RestMethod -Uri "http://localhost:8002/sql" -Method POST -Body $body -ContentType "application/json" -Headers @{ "NS" = "agileos"; "DB" = "main"; "Authorization" = $auth } | Out-Null
    Write-Host "Task $i created"
}
```

## 6. Start Frontend

Buka terminal baru:

```powershell
cd agile-os/frontend-next
npm run dev
```

## 7. Access Analytics Dashboard

Buka browser:
```
http://localhost:3001/analytics
```

## Login Credentials

- Username: `admin`
- Password: `password123`

## Troubleshooting

### Backend tidak bisa connect ke database
- Pastikan port mapping benar: 8002 untuk SurrealDB, 4223 untuk NATS
- Check dengan: `docker ps`

### User deactivated
```powershell
$auth = "Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("root:root"))
$query = "UPDATE user SET is_active = true"
$body = @{ query = $query } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8002/sql" -Method POST -Body $body -ContentType "application/json" -Headers @{ "NS" = "agileos"; "DB" = "main"; "Authorization" = $auth }
```

### Frontend error
- Pastikan Recharts sudah terinstall: `cd frontend-next && npm install`
- Check `NEXT_PUBLIC_API_URL` di `.env` atau hardcoded di code

## API Endpoints untuk Testing

```powershell
# Get analytics overview
Invoke-RestMethod -Uri "http://localhost:8081/api/v1/analytics/overview" -Headers @{ "Authorization" = "Bearer $token" }

# Get department metrics
Invoke-RestMethod -Uri "http://localhost:8081/api/v1/analytics/departments" -Headers @{ "Authorization" = "Bearer $token" }

# Get insights
Invoke-RestMethod -Uri "http://localhost:8081/api/v1/analytics/insights" -Headers @{ "Authorization" = "Bearer $token" }
```
