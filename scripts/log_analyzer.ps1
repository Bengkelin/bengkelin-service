# Bengkelin Service Log Analyzer
# This script analyzes application logs for performance bottlenecks

param(
    [string]$LogFile = "",
    [int]$SlowThresholdMs = 500,
    [switch]$RealTime
)

Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "       BENGKELIN SERVICE - LOG ANALYZER" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan

# Performance tracking
$endpointStats = @{}
$slowRequests = @()
$errorCount = 0
$totalRequests = 0

function Analyze-LogEntry {
    param([string]$line)
    
    # Check for slow request warnings
    if ($line -match "Slow HTTP request detected") {
        if ($line -match '"method":"(\w+)"') { $method = $matches[1] }
        if ($line -match '"path":"([^"]+)"') { $path = $matches[1] }
        if ($line -match '"duration_ms":(\d+)') { $duration = [int]$matches[1] }
        if ($line -match '"request_id":"([^"]+)"') { $requestId = $matches[1] }
        
        $slowRequest = [PSCustomObject]@{
            Method = $method
            Path = $path
            DurationMs = $duration
            RequestId = $requestId
            Timestamp = Get-Date
        }
        
        Write-Host "`n[SLOW REQUEST] " -ForegroundColor Red -NoNewline
        Write-Host "$method $path - ${duration}ms" -ForegroundColor Yellow
        
        return $slowRequest
    }
    
    # Track HTTP completions
    if ($line -match "HTTP request completed") {
        $script:totalRequests++
        
        if ($line -match '"method":"(\w+)"') { $method = $matches[1] }
        if ($line -match '"path":"([^"]+)"') { $path = $matches[1] }
        if ($line -match '"duration_ms":(\d+)') { $duration = [int]$matches[1] }
        if ($line -match '"status_code":(\d+)') { $statusCode = [int]$matches[1] }
        
        $key = "$method $path"
        
        if (-not $endpointStats.ContainsKey($key)) {
            $endpointStats[$key] = @{
                Count = 0
                TotalDuration = 0
                MinDuration = $duration
                MaxDuration = $duration
                ErrorCount = 0
            }
        }
        
        $stats = $endpointStats[$key]
        $stats.Count++
        $stats.TotalDuration += $duration
        
        if ($duration -lt $stats.MinDuration) { $stats.MinDuration = $duration }
        if ($duration -gt $stats.MaxDuration) { $stats.MaxDuration = $duration }
        if ($statusCode -ge 400) { $stats.ErrorCount++ }
    }
    
    # Track errors
    if ($line -match '"level":"ERROR"' -or $line -match '\[31mERROR\[0m') {
        $script:errorCount++
        if ($line -match "Failed to connect") {
            Write-Host "[ERROR] Connection issue detected: " -ForegroundColor Red -NoNewline
            Write-Host ($line -replace '.*\"error\":\"([^\"]+)\".*','$1') -ForegroundColor Yellow
        }
    }
}

function Show-PerformanceReport {
    Write-Host "`n============================================================" -ForegroundColor Green
    Write-Host "           PERFORMANCE ANALYSIS REPORT" -ForegroundColor Green
    Write-Host "============================================================" -ForegroundColor Green
    
    Write-Host "`nTotal Requests Processed: $totalRequests" -ForegroundColor Cyan
    Write-Host "Total Errors: $errorCount" -ForegroundColor $(if ($errorCount -gt 0) { "Red" } else { "Green" })
    
    if ($endpointStats.Count -gt 0) {
        Write-Host "`n--- Endpoint Performance ---" -ForegroundColor Yellow
        
        $sortedEndpoints = $endpointStats.GetEnumerator() | Sort-Object { $_.Value.MaxDuration } -Descending
        
        foreach ($endpoint in $sortedEndpoints) {
            $key = $endpoint.Key
            $stats = $endpoint.Value
            $avgDuration = [math]::Round($stats.TotalDuration / $stats.Count, 2)
            
            $color = "Green"
            if ($stats.MaxDuration -gt 1000) { $color = "Red" }
            elseif ($stats.MaxDuration -gt 500) { $color = "Yellow" }
            
            Write-Host "`n$key" -ForegroundColor $color
            Write-Host "  Requests: $($stats.Count) | Avg: ${avgDuration}ms | Min: $($stats.MinDuration)ms | Max: $($stats.MaxDuration)ms | Errors: $($stats.ErrorCount)"
        }
    }
    
    # Critical issues
    Write-Host "`n--- Critical Issues ---" -ForegroundColor Red
    $criticalEndpoints = $endpointStats.GetEnumerator() | Where-Object { $_.Value.MaxDuration -gt 1000 }
    if ($criticalEndpoints) {
        Write-Host "Endpoints with >1000ms response time:" -ForegroundColor Red
        foreach ($ep in $criticalEndpoints) {
            Write-Host "  ⚠ $($ep.Key) - Max: $($ep.Value.MaxDuration)ms" -ForegroundColor Yellow
        }
    } else {
        Write-Host "  ✓ No critical latency issues detected" -ForegroundColor Green
    }
    
    Write-Host "`n============================================================" -ForegroundColor Green
}

# Real-time monitoring
if ($RealTime) {
    Write-Host "`nStarting real-time log monitoring..." -ForegroundColor Cyan
    Write-Host "Press Ctrl+C to stop`n" -ForegroundColor Gray
    
    while ($true) {
        # This would typically tail the log file
        Start-Sleep -Seconds 5
        Show-PerformanceReport
    }
}

# Manual analysis of provided log
Write-Host "`nAnalyzing log data..." -ForegroundColor Cyan