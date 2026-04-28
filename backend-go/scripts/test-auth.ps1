# AgileOS BPM - Authentication Testing Script

$baseUrl = "http://localhost:8081"

Write-Host "========================================" -ForegroundColor Green
Write-Host "AgileOS BPM - Authentication Test" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# Test 1: Register new user
Write-Host "Test 1: Register new user..." -ForegroundColor Cyan
$registerData = @{
    username = "testuser"
    email = "test@agileos.com"
    password = "password123"
    full_name = "Test User"
    department = "Testing"
    role = "employee"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/register" -Method POST -Body $registerData -ContentType "application/json"
    Write-Host "  Success: User registered" -ForegroundColor Green
    Write-Host "  Username: $($response.user.username)" -ForegroundColor Gray
    Write-Host "  Role: $($response.user.role)" -ForegroundColor Gray
    $token = $response.access_token
    Write-Host ""
} catch {
    Write-Host "  Note: User might already exist (this is OK)" -ForegroundColor Yellow
    Write-Host ""
}

# Test 2: Login
Write-Host "Test 2: Login with credentials..." -ForegroundColor Cyan
$loginData = @{
    username = "admin"
    password = "password123"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method POST -Body $loginData -ContentType "application/json"
    Write-Host "  Success: Login successful" -ForegroundColor Green
    Write-Host "  User: $($response.user.full_name)" -ForegroundColor Gray
    Write-Host "  Role: $($response.user.role)" -ForegroundColor Gray
    $token = $response.access_token
    Write-Host ""
} catch {
    Write-Host "  Failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "  Make sure to seed users first!" -ForegroundColor Yellow
    Write-Host ""
    exit 1
}

# Test 3: Access protected endpoint
Write-Host "Test 3: Access protected endpoint..." -ForegroundColor Cyan
$headers = @{
    "Authorization" = "Bearer $token"
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/profile" -Method GET -Headers $headers
    Write-Host "  Success: Profile retrieved" -ForegroundColor Green
    Write-Host "  Username: $($response.username)" -ForegroundColor Gray
    Write-Host "  Email: $($response.email)" -ForegroundColor Gray
    Write-Host ""
} catch {
    Write-Host "  Failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
}

# Test 4: Access without token (should fail)
Write-Host "Test 4: Access without token (should fail)..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/profile" -Method GET
    Write-Host "  Unexpected: Request succeeded without token!" -ForegroundColor Red
    Write-Host ""
} catch {
    Write-Host "  Success: Unauthorized access blocked" -ForegroundColor Green
    Write-Host ""
}

# Test 5: Access admin endpoint (role-based)
Write-Host "Test 5: Access admin endpoint..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/users" -Method GET -Headers $headers
    Write-Host "  Success: Admin access granted" -ForegroundColor Green
    Write-Host "  Users count: $($response.count)" -ForegroundColor Gray
    Write-Host ""
} catch {
    Write-Host "  Failed: $($_.Exception.Message)" -ForegroundColor Yellow
    Write-Host "  Note: This is expected if not logged in as admin" -ForegroundColor Gray
    Write-Host ""
}

Write-Host "========================================" -ForegroundColor Green
Write-Host "Authentication tests completed!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Default test credentials:" -ForegroundColor Cyan
Write-Host "  Admin:    admin / password123" -ForegroundColor White
Write-Host "  Manager:  manager / password123" -ForegroundColor White
Write-Host "  Employee: employee / password123" -ForegroundColor White
Write-Host ""
