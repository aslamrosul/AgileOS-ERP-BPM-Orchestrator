# AgileOS Backup Verification Script
# Verifies backup integrity and tests restore capability

param(
    [string]$BackupFile,
    [string]$BackupDir = ".\backups",
    [switch]$TestRestore,
    [switch]$All
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Info { Write-Host $args -ForegroundColor Cyan }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }
function Write-Error { Write-Host $args -ForegroundColor Red }

Write-Info "========================================="
Write-Info "  AgileOS Backup Verification"
Write-Info "========================================="
Write-Info ""

# Function to verify a single backup file
function Test-BackupFile {
    param([string]$FilePath)
    
    $fileName = Split-Path $FilePath -Leaf
    Write-Info "Verifying: $fileName"
    
    $results = @{
        file = $fileName
        exists = $false
        readable = $false
        valid_gzip = $false
        size_mb = 0
        age_hours = 0
        status = "FAIL"
    }
    
    # Check if file exists
    if (-not (Test-Path $FilePath)) {
        Write-Error "  ✗ File not found"
        return $results
    }
    $results.exists = $true
    
    # Check if file is readable
    try {
        $file = Get-Item $FilePath
        $results.size_mb = [math]::Round($file.Length / 1MB, 2)
        $results.age_hours = [math]::Round(((Get-Date) - $file.LastWriteTime).TotalHours, 1)
        $results.readable = $true
        Write-Success "  ✓ File exists ($($results.size_mb) MB, $($results.age_hours)h old)"
    } catch {
        Write-Error "  ✗ Cannot read file: $_"
        return $results
    }
    
    # Check if file size is reasonable (> 1KB)
    if ($file.Length -lt 1KB) {
        Write-Error "  ✗ File too small (possible corruption)"
        return $results
    }
    Write-Success "  ✓ File size is reasonable"
    
    # Verify Gzip integrity
    if ($FilePath -match "\.gz$") {
        try {
            $stream = [System.IO.File]::OpenRead($FilePath)
            $gzip = New-Object System.IO.Compression.GzipStream($stream, [System.IO.Compression.CompressionMode]::Decompress)
            
            # Try to read first few bytes
            $buffer = New-Object byte[] 1024
            $bytesRead = $gzip.Read($buffer, 0, 1024)
            
            $gzip.Close()
            $stream.Close()
            
            if ($bytesRead -gt 0) {
                $results.valid_gzip = $true
                Write-Success "  ✓ Gzip compression is valid"
            } else {
                Write-Error "  ✗ Gzip file is empty"
                return $results
            }
        } catch {
            Write-Error "  ✗ Gzip integrity check failed: $_"
            return $results
        }
    } else {
        # Not compressed, assume valid
        $results.valid_gzip = $true
        Write-Info "  ℹ File is not compressed"
    }
    
    # Check if file contains SurrealDB data
    try {
        if ($FilePath -match "\.gz$") {
            # Decompress first 1KB to check content
            $stream = [System.IO.File]::OpenRead($FilePath)
            $gzip = New-Object System.IO.Compression.GzipStream($stream, [System.IO.Compression.CompressionMode]::Decompress)
            $reader = New-Object System.IO.StreamReader($gzip)
            
            $firstLine = $reader.ReadLine()
            
            $reader.Close()
            $gzip.Close()
            $stream.Close()
            
            if ($firstLine -match "DEFINE|CREATE|INSERT|UPDATE") {
                Write-Success "  ✓ Contains valid SurrealDB statements"
            } else {
                Write-Warning "  ⚠ File content may not be valid SurrealDB export"
            }
        }
    } catch {
        Write-Warning "  ⚠ Could not verify file content: $_"
    }
    
    $results.status = "PASS"
    Write-Success "  ✓ Backup verification PASSED"
    Write-Info ""
    
    return $results
}

# Function to test restore capability
function Test-RestoreCapability {
    param([string]$FilePath)
    
    Write-Info "Testing restore capability for: $(Split-Path $FilePath -Leaf)"
    Write-Info ""
    
    # Check if Docker is available
    try {
        docker --version | Out-Null
        Write-Success "✓ Docker is available"
    } catch {
        Write-Error "✗ Docker is not available"
        return $false
    }
    
    # Create test container
    $testContainerName = "agileos-db-test-$(Get-Random -Maximum 9999)"
    Write-Info "Creating test container: $testContainerName"
    
    try {
        docker run -d --name $testContainerName `
            -p 8003:8000 `
            surrealdb/surrealdb:v1.4.0 start --user root --pass root 2>&1 | Out-Null
        
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to create test container"
        }
        
        Write-Success "✓ Test container created"
        
        # Wait for container to be ready
        Write-Info "Waiting for SurrealDB to start..."
        Start-Sleep -Seconds 10
        
        # Test connectivity
        $maxRetries = 5
        $retryCount = 0
        $connected = $false
        
        while ($retryCount -lt $maxRetries -and -not $connected) {
            try {
                $health = Invoke-WebRequest -Uri "http://localhost:8003/health" -Method GET -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
                $connected = $true
                Write-Success "✓ Test database is accessible"
            } catch {
                $retryCount++
                if ($retryCount -lt $maxRetries) {
                    Write-Info "  Retrying... ($retryCount/$maxRetries)"
                    Start-Sleep -Seconds 5
                }
            }
        }
        
        if (-not $connected) {
            throw "Could not connect to test database"
        }
        
        # Attempt restore
        Write-Info "Attempting restore..."
        
        $restoreScript = Join-Path $PSScriptRoot "restore-db.ps1"
        if (-not (Test-Path $restoreScript)) {
            throw "Restore script not found: $restoreScript"
        }
        
        & $restoreScript `
            -BackupFile $FilePath `
            -ContainerName $testContainerName `
            -SurrealUrl "http://localhost:8003" `
            -Force 2>&1 | Out-Null
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "✓ Restore test PASSED"
            $result = $true
        } else {
            Write-Error "✗ Restore test FAILED"
            $result = $false
        }
        
    } catch {
        Write-Error "✗ Restore test failed: $_"
        $result = $false
    } finally {
        # Cleanup test container
        Write-Info "Cleaning up test container..."
        docker stop $testContainerName 2>&1 | Out-Null
        docker rm $testContainerName 2>&1 | Out-Null
        Write-Success "✓ Test container removed"
    }
    
    Write-Info ""
    return $result
}

# Main verification logic
$verificationResults = @()

if ($BackupFile) {
    # Verify single backup file
    if (-not (Test-Path $BackupFile)) {
        Write-Error "Backup file not found: $BackupFile"
        exit 1
    }
    
    $result = Test-BackupFile -FilePath $BackupFile
    $verificationResults += $result
    
    if ($TestRestore -and $result.status -eq "PASS") {
        $restoreResult = Test-RestoreCapability -FilePath $BackupFile
        $result.restore_test = $restoreResult
    }
    
} elseif ($All) {
    # Verify all backups in directory
    Write-Info "Verifying all backups in: $BackupDir"
    Write-Info ""
    
    $backupFiles = Get-ChildItem -Path $BackupDir -Filter "backup_*.surql*" | 
                   Sort-Object LastWriteTime -Descending
    
    if ($backupFiles.Count -eq 0) {
        Write-Warning "No backup files found in $BackupDir"
        exit 0
    }
    
    Write-Info "Found $($backupFiles.Count) backup file(s)"
    Write-Info ""
    
    foreach ($file in $backupFiles) {
        $result = Test-BackupFile -FilePath $file.FullName
        $verificationResults += $result
    }
    
    # Test restore on latest backup only
    if ($TestRestore -and $backupFiles.Count -gt 0) {
        Write-Info "Testing restore capability on latest backup..."
        Write-Info ""
        $restoreResult = Test-RestoreCapability -FilePath $backupFiles[0].FullName
    }
    
} else {
    # Verify latest backup by default
    Write-Info "Verifying latest backup in: $BackupDir"
    Write-Info ""
    
    $latestBackup = Get-ChildItem -Path $BackupDir -Filter "backup_*.surql*" | 
                    Sort-Object LastWriteTime -Descending | 
                    Select-Object -First 1
    
    if (-not $latestBackup) {
        Write-Warning "No backup files found in $BackupDir"
        exit 0
    }
    
    $result = Test-BackupFile -FilePath $latestBackup.FullName
    $verificationResults += $result
    
    if ($TestRestore -and $result.status -eq "PASS") {
        $restoreResult = Test-RestoreCapability -FilePath $latestBackup.FullName
        $result.restore_test = $restoreResult
    }
}

# Summary report
Write-Info "========================================="
Write-Info "  Verification Summary"
Write-Info "========================================="
Write-Info ""

$passCount = ($verificationResults | Where-Object { $_.status -eq "PASS" }).Count
$failCount = ($verificationResults | Where-Object { $_.status -eq "FAIL" }).Count

Write-Info "Total backups verified: $($verificationResults.Count)"
Write-Success "Passed: $passCount"
if ($failCount -gt 0) {
    Write-Error "Failed: $failCount"
}
Write-Info ""

# Detailed results table
if ($verificationResults.Count -gt 0) {
    Write-Info "Detailed Results:"
    Write-Info ""
    
    $verificationResults | ForEach-Object {
        $statusColor = if ($_.status -eq "PASS") { "Green" } else { "Red" }
        Write-Host "  File: $($_.file)" -ForegroundColor Cyan
        Write-Host "    Status: $($_.status)" -ForegroundColor $statusColor
        Write-Host "    Size: $($_.size_mb) MB"
        Write-Host "    Age: $($_.age_hours) hours"
        if ($_.PSObject.Properties.Name -contains "restore_test") {
            $restoreStatus = if ($_.restore_test) { "PASS" } else { "FAIL" }
            $restoreColor = if ($_.restore_test) { "Green" } else { "Red" }
            Write-Host "    Restore Test: $restoreStatus" -ForegroundColor $restoreColor
        }
        Write-Host ""
    }
}

# Health recommendations
Write-Info "Health Recommendations:"
Write-Info ""

$latestBackup = $verificationResults | Sort-Object { $_.age_hours } | Select-Object -First 1
if ($latestBackup.age_hours -gt 36) {
    Write-Warning "⚠ Latest backup is over 36 hours old - run backup soon!"
} else {
    Write-Success "✓ Backup schedule is healthy"
}

$totalSize = ($verificationResults | Measure-Object -Property size_mb -Sum).Sum
if ($totalSize -gt 1000) {
    Write-Warning "⚠ Total backup storage is over 1GB - consider cleanup"
} else {
    Write-Success "✓ Backup storage usage is reasonable"
}

if ($failCount -eq 0) {
    Write-Success "✓ All backups are healthy"
} else {
    Write-Error "✗ Some backups have issues - investigate failed backups"
}

Write-Info ""
Write-Info "Verification completed at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
Write-Info ""

# Exit with appropriate code
if ($failCount -gt 0) {
    exit 1
} else {
    exit 0
}

