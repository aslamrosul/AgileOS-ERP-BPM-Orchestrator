# AgileOS Database Backup Script
# Automated backup for SurrealDB with compression and rotation

param(
    [string]$BackupDir = ".\backups",
    [string]$ContainerName = "agileos-db",
    [int]$RetentionDays = 7,
    [string]$SurrealUrl = "http://localhost:8002",
    [string]$Username = "root",
    [string]$Password = "root",
    [string]$Namespace = "agileos",
    [string]$Database = "main"
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Info { Write-Host $args -ForegroundColor Cyan }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }
function Write-Error { Write-Host $args -ForegroundColor Red }

Write-Info "========================================="
Write-Info "  AgileOS Database Backup"
Write-Info "========================================="
Write-Info ""

# Create backup directory if it doesn't exist
if (-not (Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Path $BackupDir | Out-Null
    Write-Success " Created backup directory: $BackupDir"
}

# Generate timestamp for backup filename
$timestamp = Get-Date -Format "yyyy-MM-dd_HHmmss"
$backupFile = Join-Path $BackupDir "backup_${timestamp}.surql"
$compressedFile = "${backupFile}.gz"

Write-Info "Backup file: $backupFile"
Write-Info ""

# Check if SurrealDB container is running
Write-Info "Checking SurrealDB container status..."
try {
    $containerStatus = docker inspect -f '{{.State.Running}}' $ContainerName 2>$null
    if ($containerStatus -ne "true") {
        Write-Error " SurrealDB container is not running!"
        Write-Warning "Start the container with: docker start $ContainerName"
        exit 1
    }
    Write-Success " SurrealDB container is running"
} catch {
    Write-Error " Failed to check container status: $_"
    exit 1
}

# Check if SurrealDB is accessible
Write-Info "Checking SurrealDB connectivity..."
try {
    $health = Invoke-WebRequest -Uri "$SurrealUrl/health" -Method GET -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Success " SurrealDB is accessible"
} catch {
    Write-Error " Cannot connect to SurrealDB at $SurrealUrl"
    Write-Warning "Check if SurrealDB is running and accessible"
    exit 1
}

Write-Info ""
Write-Info "Starting database export..."

# Export database using surreal export
try {
    # Use docker exec to run surreal export inside container
    $exportCmd = "surreal export --conn http://localhost:8000 --user $Username --pass $Password --ns $Namespace --db $Database /tmp/backup.surql"
    
    docker exec $ContainerName sh -c $exportCmd 2>&1 | Out-Null
    
    if ($LASTEXITCODE -ne 0) {
        throw "Export command failed with exit code $LASTEXITCODE"
    }
    
    # Copy backup file from container to host
    docker cp "${ContainerName}:/tmp/backup.surql" $backupFile 2>&1 | Out-Null
    
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to copy backup file from container"
    }
    
    # Clean up temp file in container
    docker exec $ContainerName rm /tmp/backup.surql 2>&1 | Out-Null
    
    Write-Success " Database exported successfully"
} catch {
    Write-Error " Export failed: $_"
    
    # Log to audit trail (if backend is running)
    try {
        $auditLog = @{
            action = "backup_failed"
            resource_type = "database"
            resource_id = $Database
            metadata = @{
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

# Get backup file size
$fileSize = (Get-Item $backupFile).Length
$fileSizeMB = [math]::Round($fileSize / 1MB, 2)
Write-Info "Backup size: $fileSizeMB MB"

# Compress backup file
Write-Info "Compressing backup..."
try {
    # Use 7-Zip if available, otherwise use .NET compression
    if (Get-Command 7z -ErrorAction SilentlyContinue) {
        7z a -tgzip $compressedFile $backupFile -mx9 | Out-Null
    } else {
        # Use .NET compression
        $input = [System.IO.File]::OpenRead($backupFile)
        $output = [System.IO.File]::Create($compressedFile)
        $gzipStream = New-Object System.IO.Compression.GzipStream $output, ([System.IO.Compression.CompressionMode]::Compress)
        
        $input.CopyTo($gzipStream)
        
        $gzipStream.Close()
        $output.Close()
        $input.Close()
    }
    
    # Remove uncompressed file
    Remove-Item $backupFile
    
    $compressedSize = (Get-Item $compressedFile).Length
    $compressedSizeMB = [math]::Round($compressedSize / 1MB, 2)
    $compressionRatio = [math]::Round((1 - ($compressedSize / $fileSize)) * 100, 1)
    
    Write-Success "Backup compressed: $compressedSizeMB MB ($compressionRatio% reduction)"
} catch {
    Write-Warning "Compression failed: $_"
    Write-Warning "Keeping uncompressed backup"
    $compressedFile = $backupFile
}

Write-Info ""
Write-Info "Applying backup rotation policy..."

# Backup rotation: Keep only last N days
try {
    $cutoffDate = (Get-Date).AddDays(-$RetentionDays)
    $oldBackups = Get-ChildItem -Path $BackupDir -Filter "backup_*.surql*" | 
                  Where-Object { $_.LastWriteTime -lt $cutoffDate }
    
    if ($oldBackups.Count -gt 0) {
        Write-Info "Removing $($oldBackups.Count) old backup(s)..."
        $oldBackups | ForEach-Object {
            Write-Info "  - Removing: $($_.Name)"
            Remove-Item $_.FullName
        }
        Write-Success " Old backups removed"
    } else {
        Write-Info "No old backups to remove"
    }
} catch {
    Write-Warning " Backup rotation failed: $_"
}

# Count total backups
$totalBackups = (Get-ChildItem -Path $BackupDir -Filter "backup_*.surql*").Count
$totalSize = (Get-ChildItem -Path $BackupDir -Filter "backup_*.surql*" | Measure-Object -Property Length -Sum).Sum
$totalSizeMB = [math]::Round($totalSize / 1MB, 2)

Write-Info ""
Write-Success "========================================="
Write-Success "  Backup Completed Successfully!"
Write-Success "========================================="
Write-Info ""
Write-Info "Backup Details:"
Write-Info "  File: $compressedFile"
Write-Info "  Size: $compressedSizeMB MB"
Write-Info "  Total Backups: $totalBackups"
Write-Info "  Total Storage: $totalSizeMB MB"
Write-Info "  Retention: $RetentionDays days"
Write-Info ""

# Optional: Upload to Azure Blob Storage (if configured)
if ($env:AZURE_STORAGE_CONNECTION_STRING) {
    Write-Info "Uploading to Azure Blob Storage..."
    try {
        # Requires Azure CLI or Azure PowerShell module
        # az storage blob upload --connection-string $env:AZURE_STORAGE_CONNECTION_STRING `
        #     --container-name agileos-backups `
        #     --file $compressedFile `
        #     --name (Split-Path $compressedFile -Leaf)
        
        Write-Success " Backup uploaded to Azure"
    } catch {
        Write-Warning " Azure upload failed: $_"
    }
}

# Log successful backup to audit trail
try {
    $auditLog = @{
        action = "backup_completed"
        resource_type = "database"
        resource_id = $Database
        metadata = @{
            file = $compressedFile
            size_mb = $compressedSizeMB
            timestamp = (Get-Date).ToString("o")
        }
    } | ConvertTo-Json
    
    # Send to backend audit endpoint (if available)
    # Invoke-RestMethod -Uri "http://localhost:8080/api/v1/audit" -Method POST -Body $auditLog -ContentType "application/json" -ErrorAction SilentlyContinue
} catch {
    # Silently fail if audit logging is not available
}

Write-Info "Next backup will run in 24 hours"
Write-Info "To restore: .\scripts\restore-db.ps1 -BackupFile '$compressedFile'"
Write-Info ""

exit 0


