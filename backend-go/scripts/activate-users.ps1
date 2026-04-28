# Activate all users in the database

$surrealUrl = "http://localhost:8000/sql"
$auth = "Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("root:root"))

Write-Host "Activating all users..." -ForegroundColor Cyan

$query = "UPDATE user SET is_active = true"
$body = @{ query = $query } | ConvertTo-Json

try {
    $result = Invoke-RestMethod -Uri $surrealUrl -Method POST -Body $body -ContentType "application/json" -Headers @{ "NS" = "agileos"; "DB" = "main"; "Authorization" = $auth }
    Write-Host "Users activated successfully!" -ForegroundColor Green
    
    # Verify
    $query2 = "SELECT username, is_active FROM user"
    $body2 = @{ query = $query2 } | ConvertTo-Json
    $result2 = Invoke-RestMethod -Uri $surrealUrl -Method POST -Body $body2 -ContentType "application/json" -Headers @{ "NS" = "agileos"; "DB" = "main"; "Authorization" = $auth }
    
    Write-Host "`nUser Status:" -ForegroundColor Cyan
    $result2[0].result | ForEach-Object {
        Write-Host "  $($_.username): is_active = $($_.is_active)" -ForegroundColor White
    }
} catch {
    Write-Host "Failed to activate users: $($_.Exception.Message)" -ForegroundColor Red
}
