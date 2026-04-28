# Seed users via API registration endpoint

$baseUrl = "http://localhost:8081"

Write-Host "========================================" -ForegroundColor Green
Write-Host "Seeding Users" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

$users = @(
    @{ username = "admin"; password = "password123"; email = "admin@agileos.com"; full_name = "Admin User"; role = "admin" },
    @{ username = "manager"; password = "password123"; email = "manager@agileos.com"; full_name = "Manager User"; role = "manager" },
    @{ username = "employee"; password = "password123"; email = "employee@agileos.com"; full_name = "Employee User"; role = "employee" },
    @{ username = "finance"; password = "password123"; email = "finance@agileos.com"; full_name = "Finance User"; role = "finance" },
    @{ username = "procurement"; password = "password123"; email = "procurement@agileos.com"; full_name = "Procurement User"; role = "procurement" }
)

foreach ($user in $users) {
    $username = $user.username
    Write-Host "Creating user: $username..." -ForegroundColor Cyan
    
    $userData = $user | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/register" -Method POST -Body $userData -ContentType "application/json"
        Write-Host "  User created: $username" -ForegroundColor Green
    } catch {
        if ($_.Exception.Response.StatusCode -eq 409) {
            Write-Host "  User already exists: $username" -ForegroundColor Yellow
        } else {
            Write-Host "  Failed to create user: $username" -ForegroundColor Red
            Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "User seeding completed!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Test credentials for all users:" -ForegroundColor Cyan
Write-Host "  Password: password123" -ForegroundColor White
Write-Host ""
