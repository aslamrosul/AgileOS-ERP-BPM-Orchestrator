# Remove unnecessary type assertions since QuerySlice already returns []interface{}

Write-Host "=== Removing Unnecessary Type Assertions ===" -ForegroundColor Cyan

$files = Get-ChildItem -Path "handlers" -Filter "*.go" -Recurse

foreach ($file in $files) {
    $content = Get-Content $file.FullName -Raw
    $original = $content
    
    # Pattern 1: customersRaw, ok := ... -> customers, err :=
    $content = $content -replace '(\w+)Raw, ok := h\.db\.QuerySlice', '$1, err := h.db.QuerySlice'
    
    # Pattern 2: if customers, ok := customersRaw.([]interface{}); ok && len(customers) > 0
    # -> if len(customers) > 0
    $content = $content -replace 'if (\w+), ok := (\w+)Raw\.\(\[\]interface\{\}\); ok && len\(\$1\) > 0', 'if len($1) > 0'
    
    # Pattern 3: if customers, ok := customersRaw.([]interface{}); ok {
    # -> if len(customers) > 0 {
    $content = $content -replace 'if (\w+), ok := (\w+)Raw\.\(\[\]interface\{\}\); ok \{', 'if len($1) > 0 {'
    
    # Pattern 4: Remove lines with just type assertion
    $content = $content -replace '\s+(\w+), ok := (\w+)Raw\.\(\[\]interface\{\}\)\s+if !ok \|\| len\(\$1\) == 0 \{', "`n`tif len(`$1) == 0 {"
    
    if ($content -ne $original) {
        Set-Content $file.FullName $content -NoNewline
        Write-Host "Fixed: $($file.Name)" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "Done! Running go build..." -ForegroundColor Yellow
go build -o agileos.exe 2>&1 | Select-Object -First 30
