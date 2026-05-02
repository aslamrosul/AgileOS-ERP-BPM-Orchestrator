# Simple fix: Remove Raw suffix and type assertions

Write-Host "=== Fixing Variable Names and Type Assertions ===" -ForegroundColor Cyan

$files = Get-ChildItem -Path "handlers" -Filter "*.go"

foreach ($file in $files) {
    $content = Get-Content $file.FullName -Raw
    $original = $content
    
    # Fix 1: customersRaw -> customers (variable names)
    $content = $content -replace '(\w+)Raw, err := h\.db\.QuerySlice', '$1, err := h.db.QuerySlice'
    
    # Fix 2: Remove type assertion checks like: if customers, ok := customersRaw.([]interface{}); ok && len(customers) > 0
    $content = $content -replace 'if (\w+), ok := \1Raw\.\(\[\]interface\{\}\); ok && len\(\$1\) > 0 \{', 'if len($1) > 0 {'
    
    # Fix 3: Remove type assertion checks like: if customers, ok := customersRaw.([]interface{}); ok {
    $content = $content -replace 'if (\w+), ok := \1Raw\.\(\[\]interface\{\}\); ok \{', 'if len($1) > 0 {'
    
    # Fix 4: Remove standalone type assertions
    $content = $content -replace '\n\s+(\w+), ok := \1Raw\.\(\[\]interface\{\}\)\s+if !ok \|\| len\(\$1\) == 0 \{', "`n`tif len(`$1) == 0 {"
    
    # Fix 5: existing, ok := existingRaw.([]interface{}) -> just use existing
    $content = $content -replace '\n\s+(\w+), ok := \1Raw\.\(\[\]interface\{\}\)\s+if !ok', "`n`tif len(`$1) == 0"
    
    if ($content -ne $original) {
        Set-Content $file.FullName $content -NoNewline
        Write-Host "  Fixed: $($file.Name)" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "Compiling..." -ForegroundColor Yellow
go build -o agileos.exe 2>&1 | Select-Object -First 40
