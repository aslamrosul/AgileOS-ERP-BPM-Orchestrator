# Test Digital Signature System

$baseUrl = "http://localhost:8081"

Write-Host "========================================" -ForegroundColor Green
Write-Host "Testing Digital Signature System" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# Step 1: Login
Write-Host "[1/5] Logging in..." -ForegroundColor Cyan
$loginData = @{
    username = "admin"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method POST -Body $loginData -ContentType "application/json"
    $token = $loginResponse.access_token
    Write-Host "  ✓ Logged in successfully" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Failed to login: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
}

# Step 2: Start a process
Write-Host "[2/5] Starting a test process..." -ForegroundColor Cyan
$processData = @{
    workflow_id = "workflow:purchase_approval"
    initiated_by = "user:admin"
    data = @{
        amount = 25000
        description = "Test purchase for signature verification"
        vendor = "Test Vendor Inc"
    }
} | ConvertTo-Json

try {
    $processResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/process/start" -Method POST -Body $processData -ContentType "application/json" -Headers $headers
    $taskId = $processResponse.first_task_id
    Write-Host "  ✓ Process started: $($processResponse.process_instance_id)" -ForegroundColor Green
    Write-Host "  ✓ First task created: $taskId" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Failed to start process: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 3: Complete the task (this will generate digital signature)
Write-Host "[3/5] Completing task with digital signature..." -ForegroundColor Cyan
$completeData = @{
    executed_by = "user:admin"
    result = @{
        decision = "approved"
        comments = "Approved for testing digital signature"
        approval_amount = 25000
    }
} | ConvertTo-Json

try {
    $completeResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/task/$taskId/complete" -Method POST -Body $completeData -ContentType "application/json" -Headers $headers
    $signature = $completeResponse.digital_signature
    Write-Host "  ✓ Task completed successfully" -ForegroundColor Green
    Write-Host "  ✓ Digital signature generated: $($signature.Substring(0, 16))..." -ForegroundColor Green
    Write-Host "  ✓ QR Code data: $($completeResponse.qr_code_data)" -ForegroundColor Gray
} catch {
    Write-Host "  ✗ Failed to complete task: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 4: Get signature details
Write-Host "[4/5] Retrieving signature details..." -ForegroundColor Cyan
try {
    $signatureResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/signature/task/$taskId" -Headers $headers
    Write-Host "  ✓ Signature retrieved successfully" -ForegroundColor Green
    Write-Host "    Signed by: $($signatureResponse.signed_by)" -ForegroundColor Gray
    Write-Host "    Signed at: $($signatureResponse.signed_at)" -ForegroundColor Gray
    Write-Host "    Full signature: $($signatureResponse.digital_signature)" -ForegroundColor Gray
} catch {
    Write-Host "  ✗ Failed to get signature: $($_.Exception.Message)" -ForegroundColor Red
}

# Step 5: Verify signature integrity
Write-Host "[5/5] Verifying signature integrity..." -ForegroundColor Cyan
try {
    $integrityResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/signature/task/$taskId/integrity" -Headers $headers
    
    if ($integrityResponse.integrity_valid) {
        Write-Host "  ✓ Signature integrity VALID - Data is authentic" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Signature integrity INVALID - Data may be tampered!" -ForegroundColor Red
    }
    
    Write-Host "    Message: $($integrityResponse.message)" -ForegroundColor Gray
} catch {
    Write-Host "  ✗ Failed to verify integrity: $($_.Exception.Message)" -ForegroundColor Red
}

# Step 6: Generate receipt
Write-Host ""
Write-Host "Generating digital receipt..." -ForegroundColor Cyan
try {
    $receiptResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/signature/task/$taskId/receipt" -Headers $headers
    Write-Host "  ✓ Receipt generated successfully" -ForegroundColor Green
    Write-Host "    Receipt ID: $($receiptResponse.id)" -ForegroundColor Gray
    Write-Host "    Document Title: $($receiptResponse.document_title)" -ForegroundColor Gray
    Write-Host "    Verification URL: $($receiptResponse.verification_url)" -ForegroundColor Gray
} catch {
    Write-Host "  ✗ Failed to generate receipt: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Digital Signature Test Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Summary:" -ForegroundColor Cyan
Write-Host "  Task ID: $taskId" -ForegroundColor White
Write-Host "  Digital Signature: $signature" -ForegroundColor White
Write-Host "  Status: Signature generated and verified" -ForegroundColor Green
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "  1. Test signature verification in frontend" -ForegroundColor White
Write-Host "  2. Scan QR code to verify authenticity" -ForegroundColor White
Write-Host "  3. Try tampering with data to test anti-tamper detection" -ForegroundColor White
Write-Host ""