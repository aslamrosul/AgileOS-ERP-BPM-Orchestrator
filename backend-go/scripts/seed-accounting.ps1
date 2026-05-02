#!/usr/bin/env pwsh
# Seed Sample Accounting Data (Chart of Accounts)

Write-Host "🌱 Seeding Sample Accounting Data..." -ForegroundColor Cyan

# Configuration
$SURREAL_URL = "http://localhost:8000"
$SURREAL_USER = "root"
$SURREAL_PASS = "root"
$NAMESPACE = "agileos"
$DATABASE = "main"

# Sample Chart of Accounts (Indonesian Standard)
$sampleAccounts = @"
-- ============================================
-- SAMPLE CHART OF ACCOUNTS (COA)
-- Indonesian Accounting Standard
-- ============================================

USE NS agileos;
USE DB main;

-- ASSETS (1-xxxx)
CREATE account CONTENT {
    account_code: '1-0000',
    account_name: 'ASSETS',
    account_type: 'asset',
    parent_account: NONE,
    level: 1,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: true,
    allow_posting: false,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '1-1000',
    account_name: 'Current Assets',
    account_type: 'asset',
    parent_account: '1-0000',
    level: 2,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: true,
    allow_posting: false,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '1-1100',
    account_name: 'Cash and Bank',
    account_type: 'asset',
    parent_account: '1-1000',
    level: 3,
    is_active: true,
    currency: 'IDR',
    opening_balance: 100000000,
    current_balance: 100000000,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '1-1110',
    account_name: 'Petty Cash',
    account_type: 'asset',
    parent_account: '1-1000',
    level: 3,
    is_active: true,
    currency: 'IDR',
    opening_balance: 5000000,
    current_balance: 5000000,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '1-1200',
    account_name: 'Accounts Receivable',
    account_type: 'asset',
    parent_account: '1-1000',
    level: 3,
    is_active: true,
    currency: 'IDR',
    opening_balance: 50000000,
    current_balance: 50000000,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '1-1300',
    account_name: 'Inventory',
    account_type: 'asset',
    parent_account: '1-1000',
    level: 3,
    is_active: true,
    currency: 'IDR',
    opening_balance: 75000000,
    current_balance: 75000000,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

-- LIABILITIES (2-xxxx)
CREATE account CONTENT {
    account_code: '2-0000',
    account_name: 'LIABILITIES',
    account_type: 'liability',
    parent_account: NONE,
    level: 1,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: true,
    allow_posting: false,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '2-1000',
    account_name: 'Current Liabilities',
    account_type: 'liability',
    parent_account: '2-0000',
    level: 2,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: true,
    allow_posting: false,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '2-1100',
    account_name: 'Accounts Payable',
    account_type: 'liability',
    parent_account: '2-1000',
    level: 3,
    is_active: true,
    currency: 'IDR',
    opening_balance: 30000000,
    current_balance: 30000000,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '2-1200',
    account_name: 'Tax Payable',
    account_type: 'liability',
    parent_account: '2-1000',
    level: 3,
    is_active: true,
    currency: 'IDR',
    opening_balance: 10000000,
    current_balance: 10000000,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

-- EQUITY (3-xxxx)
CREATE account CONTENT {
    account_code: '3-0000',
    account_name: 'EQUITY',
    account_type: 'equity',
    parent_account: NONE,
    level: 1,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: true,
    allow_posting: false,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '3-1000',
    account_name: 'Share Capital',
    account_type: 'equity',
    parent_account: '3-0000',
    level: 2,
    is_active: true,
    currency: 'IDR',
    opening_balance: 200000000,
    current_balance: 200000000,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '3-2000',
    account_name: 'Retained Earnings',
    account_type: 'equity',
    parent_account: '3-0000',
    level: 2,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

-- REVENUE (4-xxxx)
CREATE account CONTENT {
    account_code: '4-0000',
    account_name: 'REVENUE',
    account_type: 'revenue',
    parent_account: NONE,
    level: 1,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: true,
    allow_posting: false,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '4-1000',
    account_name: 'Sales Revenue',
    account_type: 'revenue',
    parent_account: '4-0000',
    level: 2,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '4-2000',
    account_name: 'Service Revenue',
    account_type: 'revenue',
    parent_account: '4-0000',
    level: 2,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

-- EXPENSES (5-xxxx)
CREATE account CONTENT {
    account_code: '5-0000',
    account_name: 'EXPENSES',
    account_type: 'expense',
    parent_account: NONE,
    level: 1,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: true,
    allow_posting: false,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '5-1000',
    account_name: 'Cost of Goods Sold',
    account_type: 'expense',
    parent_account: '5-0000',
    level: 2,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '5-2000',
    account_name: 'Salary Expense',
    account_type: 'expense',
    parent_account: '5-0000',
    level: 2,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '5-3000',
    account_name: 'Rent Expense',
    account_type: 'expense',
    parent_account: '5-0000',
    level: 2,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};

CREATE account CONTENT {
    account_code: '5-4000',
    account_name: 'Utilities Expense',
    account_type: 'expense',
    parent_account: '5-0000',
    level: 2,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    current_balance: 0,
    is_control_account: false,
    allow_posting: true,
    created_by: 'system',
    created_at: time::now(),
    updated_at: time::now()
};
"@

# Apply seed data
Write-Host "📝 Creating sample Chart of Accounts..." -ForegroundColor Yellow

$headers = @{
    "Accept" = "application/json"
    "NS" = $NAMESPACE
    "DB" = $DATABASE
}

$auth = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("${SURREAL_USER}:${SURREAL_PASS}"))
$headers["Authorization"] = "Basic $auth"

try {
    $response = Invoke-RestMethod -Uri "$SURREAL_URL/sql" -Method POST -Headers $headers -Body $sampleAccounts -ContentType "text/plain"
    
    Write-Host "✓ Sample accounts created successfully!" -ForegroundColor Green
    
    Write-Host "`n📊 Created Accounts:" -ForegroundColor Cyan
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host "ASSETS (1-xxxx):" -ForegroundColor Yellow
    Write-Host "  1-0000 ASSETS (Control)" -ForegroundColor White
    Write-Host "  1-1000 Current Assets (Control)" -ForegroundColor White
    Write-Host "  1-1100 Cash and Bank (IDR 100,000,000)" -ForegroundColor Green
    Write-Host "  1-1110 Petty Cash (IDR 5,000,000)" -ForegroundColor Green
    Write-Host "  1-1200 Accounts Receivable (IDR 50,000,000)" -ForegroundColor Green
    Write-Host "  1-1300 Inventory (IDR 75,000,000)" -ForegroundColor Green
    
    Write-Host "`nLIABILITIES (2-xxxx):" -ForegroundColor Yellow
    Write-Host "  2-0000 LIABILITIES (Control)" -ForegroundColor White
    Write-Host "  2-1000 Current Liabilities (Control)" -ForegroundColor White
    Write-Host "  2-1100 Accounts Payable (IDR 30,000,000)" -ForegroundColor Red
    Write-Host "  2-1200 Tax Payable (IDR 10,000,000)" -ForegroundColor Red
    
    Write-Host "`nEQUITY (3-xxxx):" -ForegroundColor Yellow
    Write-Host "  3-0000 EQUITY (Control)" -ForegroundColor White
    Write-Host "  3-1000 Share Capital (IDR 200,000,000)" -ForegroundColor Cyan
    Write-Host "  3-2000 Retained Earnings (IDR 0)" -ForegroundColor Cyan
    
    Write-Host "`nREVENUE (4-xxxx):" -ForegroundColor Yellow
    Write-Host "  4-0000 REVENUE (Control)" -ForegroundColor White
    Write-Host "  4-1000 Sales Revenue" -ForegroundColor Green
    Write-Host "  4-2000 Service Revenue" -ForegroundColor Green
    
    Write-Host "`nEXPENSES (5-xxxx):" -ForegroundColor Yellow
    Write-Host "  5-0000 EXPENSES (Control)" -ForegroundColor White
    Write-Host "  5-1000 Cost of Goods Sold" -ForegroundColor Red
    Write-Host "  5-2000 Salary Expense" -ForegroundColor Red
    Write-Host "  5-3000 Rent Expense" -ForegroundColor Red
    Write-Host "  5-4000 Utilities Expense" -ForegroundColor Red
    
    Write-Host "`n✅ Sample Accounting Data Seeded!" -ForegroundColor Green
    Write-Host "`n🔧 Next Steps:" -ForegroundColor Cyan
    Write-Host "  1. Test API: GET http://localhost:8080/api/v1/accounting/accounts" -ForegroundColor Yellow
    Write-Host "  2. View in Swagger: http://localhost:8080/swagger/index.html" -ForegroundColor Yellow
    Write-Host "  3. Create frontend pages for Chart of Accounts" -ForegroundColor Yellow
    
} catch {
    Write-Host "✗ Failed to seed data!" -ForegroundColor Red
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
}
