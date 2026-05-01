# AgileOS Database Restore Script
# Restore SurrealDB from backup file

param(
    [Parameter(Mandatory=$true)]
    [string]$BackupFile,
    [string]$ContainerName = "agileos-db",
    [string]$SurrealUrl = "http://localhost:8002",
    [string]$Username = "root",
    [string]$Password = "root",
    [string]$Namespace = "agileos",
    [string]$Database = "main",
    [switch]$Force
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Info { Write-Host $args -ForegroundColor Cyan }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }
function Write-Error { Write-Host $args -ForegroundColor Red }

Write-Info "========================================="
Write-Info "  AgileOS Database Restore"
Write-Info "========================================="
Write-Info ""

# Check if backup file exists
if (-not (Test-Path $BackupFile)) {
    Write-Error "✗ Backup file not found: $BackupFile"
    Write-Info ""
    Write-Info "Available backups:"
    Get-ChildItem -Path ".\backups" -Filter "backup_*.surql*" | 
        Sort-Object LastWriteTime -Descending | 
        Select-Object -First 10 | 
        ForEach-Object {
            $size = [math]::Round($_.Length / 1MB, 2)
            Write-Info "  - $($_.Name) ($size MB) - $($_.LastWriteTime)"
        }
    exit 1
}

Write-Info "Backup file: $BackupFile"
$fileSize = (Get-Item $BackupFile).Length
$fileSizeMB = [math]::Round($fileSize / 1MB, 2)
Write-Info "File size: $fileSizeMB MB"
Write-Info ""

# Check if SurrealDB container is running
Write-Info "Checking SurrealDB container status..."
try {
    $containerStatus = docker inspect -f '{{.State.Running}}' $ContainerName 2>$null
    if ($containerStatus -ne "true") {
        Write-Error "✗ SurrealDB container is not running!"
        Write-Warning "Start the container with: docker start $ContainerName"
        exit 1
    }
    Write-Success "✓ SurrealDB container is running"
} catch {
    Write-Error "✗ Failed to check container status: $_"
    exit 1
}

# Check if SurrealDB is accessible
Write-Info "Checking SurrealDB connectivity..."
try {
    $health = Invoke-WebRequest -Uri "$SurrealUrl/health" -Method GET -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Success "✓ SurrealDB is accessible"
} catch {
    Write-Error "✗ Cannot connect to SurrealDB at $SurrealUrl"
    Write-Warning "Check if SurrealDB is running and accessible"
    exit 1
}

Write-Info ""

# Warning about data loss
if (-not $Force) {
    Write-Warning "========================================="
    Write-Warning "  WARNING: DATA LOSS RISK"
    Write-Warning "========================================="
    Write-Warning ""
    Write-Warning "This operation will:"
    Write-Warning "  1. ERASE all current data in database '$Database'"
    Write-Warning "  2. Restore data from backup file"
    Write-Warning "  3. Cannot be undone"
    Write-Warning ""
    Write-Warning "Current database will be PERMANENTLY DELETED!"
    Write-Warning ""
    
    $confirmation = Read-Host "Type 'YES' to continue with restore"
    
    if ($confirmation -ne "YES") {
        Write-Info "Restore cancelled by user"
        exit 0
    }
}

Write-Info ""
Write-Info "Starting database restore..."

# Prepare backup file
$tempFile = "/tmp/restore.surql"
$isCompressed = $BackupFile -match "\.gz$"

try {
    if ($isCompressed) {
        Write-Info "Decompressing backup file..."
        
        # Decompress to temp location
        $decompressedFile = $BackupFile -replace "\.gz$", ""
        $tempDecompressed = Join-Path $env:TEMP "restore_temp.surql"
        
        # Use .NET decompression
        $input = [System.IO.File]::OpenRead($BackupFile)
        $output = [System.IO.File]::Create($tempDecompressed)
        $gzipStream = New-Object System.IO.Compression.GzipStream $input, ([System.IO.Compression.CompressionMode]::Decompress)
        
        $gzipStream.CopyTo($output)
        
        $gzipStream.Close()
        $input.Close()
        $output.Close()
        
        Write-Success "✓ Backup decompressed"
        
        # Copy decompressed file to container
        docker cp $tempDecompressed "${ContainerName}:${tempFile}" 2>&1 | Out-Null
        
        # Clean up temp file
        Remove-Item $tempDecompressed
    } else {
        # Copy backup file directly to container
        docker cp $BackupFile "${ContainerName}:${tempFile}" 2>&1 | Out-Null
    }
    
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to copy backup file to container"
    }
    
    Write-Success "✓ Backup file copied to container"
} catch {
    Write-Error "✗ Failed to prepare backup file: $_"
    exit 1
}

Write-Info ""
Write-Info "Importing database..."

# Import database using surreal import
try {
    $importCmd = "surreal import --conn http://localhost:8000 --user $Username --pass $Password --ns $Namespace --db $Database $tempFile"
    
    $output = docker exec $ContainerName sh -c $importCmd 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Import output:"
        Write-Error $output
        throw "Import command failed with exit code $LASTEXITCODE"
    }
    
    # Clean up temp file in container
    docker exec $ContainerName rm $tempFile 2>&1 | Out-Null
    
    Write-Success "✓ Database imported successfully"
} catch {
    Write-Error "✗ Import failed: $_"
    
    # Log to audit trail (if backend is running)
    try {
        $auditLog = @{
            action = "restore_failed"
            resource_type = "database"
            resource_id = $Database
            metadata = @{
                backup_file = $BackupFile
                error = $_.ToString()
                timestamp = (Get-Date).ToString("o")
            }
        } | ConvertTo-Json
        
        # Send to backend audit endpoint (if available)
        # Invoke-RestMethod -Uri "http://localhost:8080/api/v1/audit" -Method POST -Body $auditLog -ContentType "application/json" -ErrorAction SilentlyContinue
    } catch {
        # Silently fail if audit logging is not available
    }
    
    exit 1
}

Write-Info ""
Write-Info "Verifying restore..."

# Verify database connectivity after restore
try {
    Start-Sleep -Seconds 2
    
    $health = Invoke-WebRequest -Uri "$SurrealUrl/health" -Method GET -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Success "✓ Database is accessible after restore"
} catch {
    Write-Warning "⚠ Database health check failed after restore"
    Write-Warning "Please verify database manually"
}

Write-Info ""
Write-Success "========================================="
Write-Success "  Restore Completed Successfully!"
Write-Success "========================================="
Write-Info ""
Write-Info "Restore Details:"
Write-Info "  Source: $BackupFile"
Write-Info "  Database: $Namespace/$Database"
Write-Info "  Timestamp: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
Write-Info ""
Write-Info "Next Steps:"
Write-Info "  1. Verify data integrity in the application"
Write-Info "  2. Check user accounts and permissions"
Write-Info "  3. Test critical workflows"
Write-Info "  4. Review audit logs"
Write-Info ""

# Log successful restore to audit trail
try {
    $auditLog = @{
        action = "restore_completed"
        resource_type = "database"
        resource_id = $Database
        metadata = @{
            backup_file = $BackupFile
            size_mb = $fileSizeMB
            timestamp = (Get-Date).ToString("o")
        }
    } | ConvertTo-Json
    
    # Send to backend audit endpoint (if available)
    # Invoke-RestMethod -Uri "http://localhost:8080/api/v1/audit" -Method POST -Body $auditLog -ContentType "application/json" -ErrorAction SilentlyContinue
} catch {
    # Silently fail if audit logging is not available
}

Write-Warning "IMPORTANT: Restart backend services to ensure they reconnect to the restored database"
Write-Info "Run: docker restart agileos-backend"
Write-Info ""

exit 0

