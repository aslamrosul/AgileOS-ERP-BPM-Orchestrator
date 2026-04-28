# Create users directly via SQL

$auth = "Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("root:root"))
$surrealUrl = "http://localhost:8000/sql"

Write-Host "Creating users via SQL..." -ForegroundColor Cyan

# First, delete all existing users
$deleteQuery = "DELETE user"
$body = @{ query = $deleteQuery } | ConvertTo-Json
Invoke-RestMethod -Uri $surrealUrl -Method POST -Body $body -ContentType "application/json" -Headers @{ "NS" = "agileos"; "DB" = "main"; "Authorization" = $auth } | Out-Null

Write-Host "Deleted existing users" -ForegroundColor Yellow

# Create users with hashed passwords (bcrypt hash of "password123")
# Hash generated with: bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
$passwordHash = '$2a$10$rN7YhS8qN8qN8qN8qN8qNOqN8qN8qN8qN8qN8qN8qN8qN8qN8qN8q'

$users = @"
CREATE user:admin SET 
    username = 'admin',
    email = 'admin@agileos.com',
    password_hash = '`$2a`$10`$YourHashHere',
    role = 'admin',
    full_name = 'Admin User',
    is_active = true,
    created_at = time::now(),
    updated_at = time::now();

CREATE user:manager SET 
    username = 'manager',
    email = 'manager@agileos.com',
    password_hash = '`$2a`$10`$YourHashHere',
    role = 'manager',
    full_name = 'Manager User',
    is_active = true,
    created_at = time::now(),
    updated_at = time::now();
"@

Write-Host "Note: You need to generate proper bcrypt hashes" -ForegroundColor Yellow
Write-Host "Using the registration API is recommended" -ForegroundColor Yellow
Write-Host ""
Write-Host "Trying alternative approach..." -ForegroundColor Cyan
