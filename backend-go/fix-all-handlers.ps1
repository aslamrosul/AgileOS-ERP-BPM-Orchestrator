# Fix All Type Assertions in Handler Files
# This script replaces h.db.Query with h.db.QuerySlice for cleaner code

Write-Host "=== Fixing Type Assertions in All Handler Files ===" -ForegroundColor Cyan
Write-Host ""

$handlerFiles = @(
    "handlers/accounting_ar.go",
    "handlers/accounting_budget.go", 
    "handlers/accounting_reports.go",
    "handlers/hrm.go",
    "handlers/hrm_payroll.go",
    "handlers/hrm_attendance.go",
    "handlers/hrm_leave.go",
    "handlers/inventory.go",
    "handlers/inventory_stock.go",
    "handlers/inventory_warehouse.go",
    "handlers/inventory_purchasing.go",
    "handlers/crm.go",
    "handlers/crm_lead.go",
    "handlers/crm_opportunity.go",
    "handlers/crm_sales.go",
    "handlers/manufacturing.go",
    "handlers/manufacturing_planning.go"
)

$totalFixed = 0

foreach ($file in $handlerFiles) {
    if (Test-Path $file) {
        Write-Host "Processing: $file" -ForegroundColor Yellow
        
        $content = Get-Content $file -Raw
        $originalContent = $content
        
        # Replace h.db.Query with h.db.QuerySlice
        $content = $content -replace 'h\.db\.Query\(', 'h.db.QuerySlice('
        
        if ($content -ne $originalContent) {
            Set-Content $file $content -NoNewline
            $changes = ([regex]::Matches($originalContent, 'h\.db\.Query\(')).Count
            $totalFixed += $changes
            Write-Host "  Fixed $changes Query calls" -ForegroundColor Green
        } else {
            Write-Host "  No changes needed" -ForegroundColor Gray
        }
    } else {
        Write-Host "  File not found: $file" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "=== Summary ===" -ForegroundColor Cyan
Write-Host "Total Query calls fixed: $totalFixed" -ForegroundColor Green
Write-Host ""
Write-Host "Now running go build..." -ForegroundColor Yellow
go build -o agileos.exe 2>&1 | Select-Object -First 20
