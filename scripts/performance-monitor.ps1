# AgileOS Performance Monitoring Script
# Monitors system resources during load testing
# Usage: .\scripts\performance-monitor.ps1 -Duration 300 -Interval 5

param(
    [int]$Duration = 300,      # Total monitoring duration in seconds (default: 5 minutes)
    [int]$Interval = 5,        # Sampling interval in seconds (default: 5 seconds)
    [string]$OutputFile = "performance-report-$(Get-Date -Format 'yyyyMMdd_HHmmss').csv"
)

$ErrorActionPreference = "Stop"

Write-Host "🔍 AgileOS Performance Monitor" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Duration: $Duration seconds"
Write-Host "Interval: $Interval seconds"
Write-Host "Output: $OutputFile"
Write-Host ""

# Initialize CSV file
$csvHeader = "Timestamp,CPU_Percent,Memory_MB,Memory_Percent,Docker_CPU,Docker_Memory_MB"
$csvHeader | Out-File -FilePath $OutputFile -Encoding UTF8

# Calculate number of samples
$samples = [math]::Floor($Duration / $Interval)
$currentSample = 0

Write-Host "Starting monitoring... Press Ctrl+C to stop" -ForegroundColor Yellow
Write-Host ""

try {
    while ($currentSample -lt $samples) {
        $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
        
        # Get system CPU usage
        $cpuPercent = (Get-Counter '\Processor(_Total)\% Processor Time').CounterSamples.CookedValue
        $cpuPercent = [math]::Round($cpuPercent, 2)
        
        # Get system memory usage
        $os = Get-CimInstance Win32_OperatingSystem
        $totalMemoryMB = [math]::Round($os.TotalVisibleMemorySize / 1KB, 2)
        $freeMemoryMB = [math]::Round($os.FreePhysicalMemory / 1KB, 2)
        $usedMemoryMB = $totalMemoryMB - $freeMemoryMB
        $memoryPercent = [math]::Round(($usedMemoryMB / $totalMemoryMB) * 100, 2)

        # Get Docker container stats
        $dockerStats = docker stats --no-stream --format "{{.Container}},{{.CPUPerc}},{{.MemUsage}}" 2>$null
        
        $dockerCPU = "N/A"
        $dockerMemory = "N/A"
        
        if ($dockerStats) {
            $agileosContainers = $dockerStats | Where-Object { $_ -match "agileos" }
            if ($agileosContainers) {
                $totalDockerCPU = 0
                $totalDockerMemory = 0
                $containerCount = 0
                
                foreach ($line in $agileosContainers) {
                    $parts = $line -split ","
                    if ($parts.Length -ge 3) {
                        $cpu = $parts[1] -replace '%', '' -replace ' ', ''
                        $mem = $parts[2] -split '/' | Select-Object -First 1
                        $mem = $mem -replace 'MiB', '' -replace 'GiB', '' -replace ' ', ''
                        
                        if ($cpu -match '^\d+\.?\d*$') {
                            $totalDockerCPU += [double]$cpu
                        }
                        if ($mem -match '^\d+\.?\d*$') {
                            $totalDockerMemory += [double]$mem
                        }
                        $containerCount++
                    }
                }
                
                if ($containerCount -gt 0) {
                    $dockerCPU = [math]::Round($totalDockerCPU, 2)
                    $dockerMemory = [math]::Round($totalDockerMemory, 2)
                }
            }
        }
        
        # Write to CSV
        $csvLine = "$timestamp,$cpuPercent,$usedMemoryMB,$memoryPercent,$dockerCPU,$dockerMemory"
        $csvLine | Out-File -FilePath $OutputFile -Append -Encoding UTF8
        
        # Display current stats
        Write-Host "[$timestamp] CPU: $cpuPercent% | Memory: $usedMemoryMB MB ($memoryPercent%) | Docker CPU: $dockerCPU% | Docker Mem: $dockerMemory MB" -ForegroundColor Green
        
        $currentSample++
        
        if ($currentSample -lt $samples) {
            Start-Sleep -Seconds $Interval
        }
    }

    Write-Host ""
    Write-Host "✓ Monitoring completed" -ForegroundColor Green
    Write-Host "Results saved to: $OutputFile" -ForegroundColor Cyan
    
    # Generate summary
    Write-Host ""
    Write-Host "📊 Performance Summary" -ForegroundColor Cyan
    Write-Host "======================" -ForegroundColor Cyan
    
    $data = Import-Csv -Path $OutputFile
    
    $avgCPU = ($data | Measure-Object -Property CPU_Percent -Average).Average
    $maxCPU = ($data | Measure-Object -Property CPU_Percent -Maximum).Maximum
    $avgMemory = ($data | Measure-Object -Property Memory_MB -Average).Average
    $maxMemory = ($data | Measure-Object -Property Memory_MB -Maximum).Maximum
    
    Write-Host "System CPU:"
    Write-Host "  Average: $([math]::Round($avgCPU, 2))%"
    Write-Host "  Peak: $([math]::Round($maxCPU, 2))%"
    Write-Host ""
    Write-Host "System Memory:"
    Write-Host "  Average: $([math]::Round($avgMemory, 2)) MB"
    Write-Host "  Peak: $([math]::Round($maxMemory, 2)) MB"
    Write-Host ""
    
    # Performance assessment
    if ($maxCPU -gt 90) {
        Write-Host "⚠️  WARNING: CPU usage exceeded 90%. Consider scaling resources." -ForegroundColor Yellow
    } elseif ($maxCPU -gt 70) {
        Write-Host "⚡ CPU usage is moderate. System is handling load well." -ForegroundColor Yellow
    } else {
        Write-Host "✓ CPU usage is healthy. System has capacity for more load." -ForegroundColor Green
    }
    
    if ($maxMemory -gt ($totalMemoryMB * 0.9)) {
        Write-Host "⚠️  WARNING: Memory usage exceeded 90%. Risk of OOM errors." -ForegroundColor Yellow
    } elseif ($maxMemory -gt ($totalMemoryMB * 0.7)) {
        Write-Host "⚡ Memory usage is moderate. Monitor for memory leaks." -ForegroundColor Yellow
    } else {
        Write-Host "✓ Memory usage is healthy. System has sufficient RAM." -ForegroundColor Green
    }
    
} catch {
    Write-Host "❌ Error during monitoring: $_" -ForegroundColor Red
    exit 1
}
