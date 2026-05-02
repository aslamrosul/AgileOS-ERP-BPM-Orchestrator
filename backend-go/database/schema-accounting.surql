-- ============================================
-- ACCOUNTING MODULE SCHEMA - SurrealDB v1.4
-- AgileOS Enterprise ERP
-- ============================================

USE NS agileos;
USE DB main;

-- ============================================
-- TABLE: ACCOUNT (Chart of Accounts)
-- ============================================
DEFINE TABLE IF NOT EXISTS account SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE $auth.role IN ['admin', 'manager', 'finance']
        FOR create, update WHERE $auth.role IN ['admin', 'finance']
        FOR delete WHERE $auth.role = 'admin';

DEFINE FIELD IF NOT EXISTS account_code ON account TYPE string
    ASSERT $value != NONE AND string::len($value) >= 2;
DEFINE FIELD IF NOT EXISTS account_name ON account TYPE string
    ASSERT $value != NONE AND string::len($value) >= 3;
DEFINE FIELD IF NOT EXISTS account_type ON account TYPE string
    ASSERT $value IN ['asset', 'liability', 'equity', 'revenue', 'expense'];
DEFINE FIELD IF NOT EXISTS parent_account ON account TYPE option<string>;
DEFINE FIELD IF NOT EXISTS level ON account TYPE int DEFAULT 1
    ASSERT $value >= 1 AND $value <= 5;
DEFINE FIELD IF NOT EXISTS is_active ON account TYPE bool DEFAULT true;
DEFINE FIELD IF NOT EXISTS currency ON account TYPE string DEFAULT 'IDR';
DEFINE FIELD IF NOT EXISTS opening_balance ON account TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS current_balance ON account TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS is_control_account ON account TYPE bool DEFAULT false;
DEFINE FIELD IF NOT EXISTS allow_posting ON account TYPE bool DEFAULT true;
DEFINE FIELD IF NOT EXISTS created_at ON account TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS updated_at ON account TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS created_by ON account TYPE string ASSERT $value != NONE;

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_account_code ON account FIELDS account_code UNIQUE;
DEFINE INDEX IF NOT EXISTS idx_account_type ON account FIELDS account_type;
DEFINE INDEX IF NOT EXISTS idx_account_parent ON account FIELDS parent_account;
DEFINE INDEX IF NOT EXISTS idx_account_active ON account FIELDS is_active;

-- ============================================
-- TABLE: JOURNAL_ENTRY (General Ledger Header)
-- ============================================
DEFINE TABLE IF NOT EXISTS journal_entry SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE $auth.role IN ['admin', 'manager', 'finance']
        FOR create WHERE $auth.role IN ['admin', 'finance']
        FOR update WHERE $auth.role IN ['admin', 'finance'] AND status = 'draft'
        FOR delete WHERE $auth.role = 'admin' AND status = 'draft';

DEFINE FIELD IF NOT EXISTS entry_number ON journal_entry TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS entry_date ON journal_entry TYPE datetime
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS entry_type ON journal_entry TYPE string
    ASSERT $value IN ['manual', 'auto', 'opening', 'closing', 'adjustment']
    DEFAULT 'manual';
DEFINE FIELD IF NOT EXISTS description ON journal_entry TYPE string;
DEFINE FIELD IF NOT EXISTS reference ON journal_entry TYPE option<string>;
DEFINE FIELD IF NOT EXISTS status ON journal_entry TYPE string
    ASSERT $value IN ['draft', 'posted', 'reversed']
    DEFAULT 'draft';
DEFINE FIELD IF NOT EXISTS total_debit ON journal_entry TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS total_credit ON journal_entry TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS posted_by ON journal_entry TYPE option<string>;
DEFINE FIELD IF NOT EXISTS posted_at ON journal_entry TYPE option<datetime>;
DEFINE FIELD IF NOT EXISTS reversed_by ON journal_entry TYPE option<string>;
DEFINE FIELD IF NOT EXISTS reversed_at ON journal_entry TYPE option<datetime>;
DEFINE FIELD IF NOT EXISTS created_by ON journal_entry TYPE string ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS created_at ON journal_entry TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS updated_at ON journal_entry TYPE datetime DEFAULT time::now();

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_journal_entry_number ON journal_entry FIELDS entry_number UNIQUE;
DEFINE INDEX IF NOT EXISTS idx_journal_entry_date ON journal_entry FIELDS entry_date;
DEFINE INDEX IF NOT EXISTS idx_journal_entry_status ON journal_entry FIELDS status;
DEFINE INDEX IF NOT EXISTS idx_journal_entry_type ON journal_entry FIELDS entry_type;

-- ============================================
-- TABLE: JOURNAL_LINE (General Ledger Lines)
-- ============================================
DEFINE TABLE IF NOT EXISTS journal_line SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE $auth.role IN ['admin', 'manager', 'finance']
        FOR create, update, delete WHERE $auth.role IN ['admin', 'finance'];

DEFINE FIELD IF NOT EXISTS journal_entry_id ON journal_line TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS line_number ON journal_line TYPE int
    ASSERT $value > 0;
DEFINE FIELD IF NOT EXISTS account_id ON journal_line TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS account_code ON journal_line TYPE string;
DEFINE FIELD IF NOT EXISTS account_name ON journal_line TYPE string;
DEFINE FIELD IF NOT EXISTS debit ON journal_line TYPE number DEFAULT 0
    ASSERT $value >= 0;
DEFINE FIELD IF NOT EXISTS credit ON journal_line TYPE number DEFAULT 0
    ASSERT $value >= 0;
DEFINE FIELD IF NOT EXISTS description ON journal_line TYPE option<string>;
DEFINE FIELD IF NOT EXISTS cost_center ON journal_line TYPE option<string>;
DEFINE FIELD IF NOT EXISTS project_id ON journal_line TYPE option<string>;
DEFINE FIELD IF NOT EXISTS department_id ON journal_line TYPE option<string>;
DEFINE FIELD IF NOT EXISTS created_at ON journal_line TYPE datetime DEFAULT time::now();

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_journal_line_entry ON journal_line FIELDS journal_entry_id;
DEFINE INDEX IF NOT EXISTS idx_journal_line_account ON journal_line FIELDS account_id;

-- ============================================
-- TABLE: VENDOR (Supplier/Vendor Master)
-- ============================================
DEFINE TABLE IF NOT EXISTS vendor SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE $auth.role IN ['admin', 'manager', 'finance', 'procurement']
        FOR create, update WHERE $auth.role IN ['admin', 'finance', 'procurement']
        FOR delete WHERE $auth.role = 'admin';

DEFINE FIELD IF NOT EXISTS vendor_code ON vendor TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS vendor_name ON vendor TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS vendor_type ON vendor TYPE string
    ASSERT $value IN ['supplier', 'contractor', 'service_provider']
    DEFAULT 'supplier';
DEFINE FIELD IF NOT EXISTS contact_person ON vendor TYPE option<string>;
DEFINE FIELD IF NOT EXISTS email ON vendor TYPE option<string>;
DEFINE FIELD IF NOT EXISTS phone ON vendor TYPE option<string>;
DEFINE FIELD IF NOT EXISTS address ON vendor TYPE option<string>;
DEFINE FIELD IF NOT EXISTS tax_id ON vendor TYPE option<string>;
DEFINE FIELD IF NOT EXISTS payment_terms ON vendor TYPE int DEFAULT 30;
DEFINE FIELD IF NOT EXISTS credit_limit ON vendor TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS current_balance ON vendor TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS is_active ON vendor TYPE bool DEFAULT true;
DEFINE FIELD IF NOT EXISTS created_at ON vendor TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS updated_at ON vendor TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS created_by ON vendor TYPE string ASSERT $value != NONE;

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_vendor_code ON vendor FIELDS vendor_code UNIQUE;
DEFINE INDEX IF NOT EXISTS idx_vendor_name ON vendor FIELDS vendor_name;
DEFINE INDEX IF NOT EXISTS idx_vendor_active ON vendor FIELDS is_active;

-- ============================================
-- TABLE: CUSTOMER (Customer Master)
-- ============================================
DEFINE TABLE IF NOT EXISTS customer SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE $auth.role IN ['admin', 'manager', 'finance']
        FOR create, update WHERE $auth.role IN ['admin', 'finance']
        FOR delete WHERE $auth.role = 'admin';

DEFINE FIELD IF NOT EXISTS customer_code ON customer TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS customer_name ON customer TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS customer_type ON customer TYPE string
    ASSERT $value IN ['individual', 'corporate', 'government']
    DEFAULT 'individual';
DEFINE FIELD IF NOT EXISTS contact_person ON customer TYPE option<string>;
DEFINE FIELD IF NOT EXISTS email ON customer TYPE option<string>;
DEFINE FIELD IF NOT EXISTS phone ON customer TYPE option<string>;
DEFINE FIELD IF NOT EXISTS address ON customer TYPE option<string>;
DEFINE FIELD IF NOT EXISTS tax_id ON customer TYPE option<string>;
DEFINE FIELD IF NOT EXISTS payment_terms ON customer TYPE int DEFAULT 30;
DEFINE FIELD IF NOT EXISTS credit_limit ON customer TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS current_balance ON customer TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS is_active ON customer TYPE bool DEFAULT true;
DEFINE FIELD IF NOT EXISTS created_at ON customer TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS updated_at ON customer TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS created_by ON customer TYPE string ASSERT $value != NONE;

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_customer_code ON customer FIELDS customer_code UNIQUE;
DEFINE INDEX IF NOT EXISTS idx_customer_name ON customer FIELDS customer_name;
DEFINE INDEX IF NOT EXISTS idx_customer_active ON customer FIELDS is_active;

-- ============================================
-- TABLE: PURCHASE_INVOICE (Account Payable)
-- ============================================
DEFINE TABLE IF NOT EXISTS purchase_invoice SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE $auth.role IN ['admin', 'manager', 'finance', 'procurement']
        FOR create, update WHERE $auth.role IN ['admin', 'finance', 'procurement']
        FOR delete WHERE $auth.role = 'admin' AND status = 'draft';

DEFINE FIELD IF NOT EXISTS invoice_number ON purchase_invoice TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS vendor_id ON purchase_invoice TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS vendor_name ON purchase_invoice TYPE string;
DEFINE FIELD IF NOT EXISTS invoice_date ON purchase_invoice TYPE datetime
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS due_date ON purchase_invoice TYPE datetime
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS total_amount ON purchase_invoice TYPE number
    ASSERT $value > 0;
DEFINE FIELD IF NOT EXISTS tax_amount ON purchase_invoice TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS discount_amount ON purchase_invoice TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS paid_amount ON purchase_invoice TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS status ON purchase_invoice TYPE string
    ASSERT $value IN ['draft', 'submitted', 'approved', 'paid', 'cancelled']
    DEFAULT 'draft';
DEFINE FIELD IF NOT EXISTS payment_status ON purchase_invoice TYPE string
    ASSERT $value IN ['unpaid', 'partial', 'paid']
    DEFAULT 'unpaid';
DEFINE FIELD IF NOT EXISTS description ON purchase_invoice TYPE option<string>;
DEFINE FIELD IF NOT EXISTS reference ON purchase_invoice TYPE option<string>;
DEFINE FIELD IF NOT EXISTS created_by ON purchase_invoice TYPE string ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS created_at ON purchase_invoice TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS updated_at ON purchase_invoice TYPE datetime DEFAULT time::now();

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_purchase_invoice_number ON purchase_invoice FIELDS invoice_number UNIQUE;
DEFINE INDEX IF NOT EXISTS idx_purchase_invoice_vendor ON purchase_invoice FIELDS vendor_id;
DEFINE INDEX IF NOT EXISTS idx_purchase_invoice_status ON purchase_invoice FIELDS status;
DEFINE INDEX IF NOT EXISTS idx_purchase_invoice_payment_status ON purchase_invoice FIELDS payment_status;
DEFINE INDEX IF NOT EXISTS idx_purchase_invoice_due_date ON purchase_invoice FIELDS due_date;

-- ============================================
-- TABLE: SALES_INVOICE (Account Receivable)
-- ============================================
DEFINE TABLE IF NOT EXISTS sales_invoice SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE $auth.role IN ['admin', 'manager', 'finance']
        FOR create, update WHERE $auth.role IN ['admin', 'finance']
        FOR delete WHERE $auth.role = 'admin' AND status = 'draft';

DEFINE FIELD IF NOT EXISTS invoice_number ON sales_invoice TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS customer_id ON sales_invoice TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS customer_name ON sales_invoice TYPE string;
DEFINE FIELD IF NOT EXISTS invoice_date ON sales_invoice TYPE datetime
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS due_date ON sales_invoice TYPE datetime
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS total_amount ON sales_invoice TYPE number
    ASSERT $value > 0;
DEFINE FIELD IF NOT EXISTS tax_amount ON sales_invoice TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS discount_amount ON sales_invoice TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS received_amount ON sales_invoice TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS status ON sales_invoice TYPE string
    ASSERT $value IN ['draft', 'submitted', 'approved', 'paid', 'cancelled']
    DEFAULT 'draft';
DEFINE FIELD IF NOT EXISTS payment_status ON sales_invoice TYPE string
    ASSERT $value IN ['unpaid', 'partial', 'paid', 'overdue']
    DEFAULT 'unpaid';
DEFINE FIELD IF NOT EXISTS description ON sales_invoice TYPE option<string>;
DEFINE FIELD IF NOT EXISTS reference ON sales_invoice TYPE option<string>;
DEFINE FIELD IF NOT EXISTS created_by ON sales_invoice TYPE string ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS created_at ON sales_invoice TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS updated_at ON sales_invoice TYPE datetime DEFAULT time::now();

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_sales_invoice_number ON sales_invoice FIELDS invoice_number UNIQUE;
DEFINE INDEX IF NOT EXISTS idx_sales_invoice_customer ON sales_invoice FIELDS customer_id;
DEFINE INDEX IF NOT EXISTS idx_sales_invoice_status ON sales_invoice FIELDS status;
DEFINE INDEX IF NOT EXISTS idx_sales_invoice_payment_status ON sales_invoice FIELDS payment_status;
DEFINE INDEX IF NOT EXISTS idx_sales_invoice_due_date ON sales_invoice FIELDS due_date;

-- ============================================
-- TABLE: PAYMENT (Payment Transactions)
-- ============================================
DEFINE TABLE IF NOT EXISTS payment SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE $auth.role IN ['admin', 'manager', 'finance']
        FOR create, update WHERE $auth.role IN ['admin', 'finance']
        FOR delete WHERE $auth.role = 'admin' AND status = 'draft';

DEFINE FIELD IF NOT EXISTS payment_number ON payment TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS payment_type ON payment TYPE string
    ASSERT $value IN ['vendor_payment', 'customer_receipt'];
DEFINE FIELD IF NOT EXISTS party_id ON payment TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS party_name ON payment TYPE string;
DEFINE FIELD IF NOT EXISTS payment_date ON payment TYPE datetime
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS payment_method ON payment TYPE string
    ASSERT $value IN ['cash', 'bank_transfer', 'check', 'credit_card']
    DEFAULT 'bank_transfer';
DEFINE FIELD IF NOT EXISTS amount ON payment TYPE number
    ASSERT $value > 0;
DEFINE FIELD IF NOT EXISTS bank_account ON payment TYPE option<string>;
DEFINE FIELD IF NOT EXISTS reference_number ON payment TYPE option<string>;
DEFINE FIELD IF NOT EXISTS status ON payment TYPE string
    ASSERT $value IN ['draft', 'submitted', 'cleared', 'cancelled']
    DEFAULT 'draft';
DEFINE FIELD IF NOT EXISTS description ON payment TYPE option<string>;
DEFINE FIELD IF NOT EXISTS created_by ON payment TYPE string ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS created_at ON payment TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS updated_at ON payment TYPE datetime DEFAULT time::now();

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_payment_number ON payment FIELDS payment_number UNIQUE;
DEFINE INDEX IF NOT EXISTS idx_payment_type ON payment FIELDS payment_type;
DEFINE INDEX IF NOT EXISTS idx_payment_party ON payment FIELDS party_id;
DEFINE INDEX IF NOT EXISTS idx_payment_date ON payment FIELDS payment_date;
DEFINE INDEX IF NOT EXISTS idx_payment_status ON payment FIELDS status;

-- ============================================
-- TABLE: BUDGET (Budget Management)
-- ============================================
DEFINE TABLE IF NOT EXISTS budget SCHEMAFULL TYPE NORMAL
    PERMISSIONS
        FOR select WHERE $auth.role IN ['admin', 'manager', 'finance']
        FOR create, update WHERE $auth.role IN ['admin', 'finance']
        FOR delete WHERE $auth.role = 'admin';

DEFINE FIELD IF NOT EXISTS budget_name ON budget TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS fiscal_year ON budget TYPE int
    ASSERT $value >= 2020 AND $value <= 2100;
DEFINE FIELD IF NOT EXISTS account_id ON budget TYPE string
    ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS account_code ON budget TYPE string;
DEFINE FIELD IF NOT EXISTS account_name ON budget TYPE string;
DEFINE FIELD IF NOT EXISTS department ON budget TYPE option<string>;
DEFINE FIELD IF NOT EXISTS cost_center ON budget TYPE option<string>;
DEFINE FIELD IF NOT EXISTS period_type ON budget TYPE string
    ASSERT $value IN ['monthly', 'quarterly', 'yearly']
    DEFAULT 'monthly';
DEFINE FIELD IF NOT EXISTS budget_amount ON budget TYPE number
    ASSERT $value >= 0;
DEFINE FIELD IF NOT EXISTS actual_amount ON budget TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS variance ON budget TYPE number DEFAULT 0;
DEFINE FIELD IF NOT EXISTS status ON budget TYPE string
    ASSERT $value IN ['draft', 'approved', 'active', 'closed']
    DEFAULT 'draft';
DEFINE FIELD IF NOT EXISTS created_by ON budget TYPE string ASSERT $value != NONE;
DEFINE FIELD IF NOT EXISTS created_at ON budget TYPE datetime DEFAULT time::now();
DEFINE FIELD IF NOT EXISTS updated_at ON budget TYPE datetime DEFAULT time::now();

-- Indexes
DEFINE INDEX IF NOT EXISTS idx_budget_fiscal_year ON budget FIELDS fiscal_year;
DEFINE INDEX IF NOT EXISTS idx_budget_account ON budget FIELDS account_id;
DEFINE INDEX IF NOT EXISTS idx_budget_status ON budget FIELDS status;

-- ============================================
-- EVENTS - Audit Trail for Accounting
-- ============================================

-- Event: Log account changes
DEFINE EVENT IF NOT EXISTS evt_account_audit ON account
    WHEN $event IN ["CREATE", "UPDATE", "DELETE"]
    THEN {
        CREATE audit_trails SET
            actor_id = $auth.id ?? "system",
            action = $event,
            resource_type = "account",
            resource_id = $value.id ?? "unknown",
            old_value = $before,
            new_value = $after,
            timestamp = time::now();
    };

-- Event: Log journal entry changes
DEFINE EVENT IF NOT EXISTS evt_journal_audit ON journal_entry
    WHEN $event IN ["CREATE", "UPDATE", "DELETE"]
    THEN {
        CREATE audit_trails SET
            actor_id = $auth.id ?? "system",
            action = $event,
            resource_type = "journal_entry",
            resource_id = $value.id ?? "unknown",
            old_value = $before,
            new_value = $after,
            timestamp = time::now();
    };

-- Event: Log invoice changes
DEFINE EVENT IF NOT EXISTS evt_invoice_audit ON purchase_invoice
    WHEN $event IN ["CREATE", "UPDATE", "DELETE"]
    THEN {
        CREATE audit_trails SET
            actor_id = $auth.id ?? "system",
            action = $event,
            resource_type = "purchase_invoice",
            resource_id = $value.id ?? "unknown",
            old_value = $before,
            new_value = $after,
            timestamp = time::now();
    };

-- ============================================
-- FUNCTIONS - Accounting Business Logic
-- ============================================

-- Function: Get account balance
DEFINE FUNCTION IF NOT EXISTS fn::get_account_balance($account_id: string) {
    LET $account = (SELECT current_balance FROM account WHERE id = $account_id LIMIT 1);
    RETURN IF array::len($account) > 0 THEN $account[0].current_balance ELSE 0 END;
};

-- Function: Check if journal entry is balanced
DEFINE FUNCTION IF NOT EXISTS fn::is_journal_balanced($journal_id: string) {
    LET $journal = (SELECT total_debit, total_credit FROM journal_entry WHERE id = $journal_id LIMIT 1);
    RETURN IF array::len($journal) > 0 THEN $journal[0].total_debit = $journal[0].total_credit ELSE false END;
};

-- Function: Get vendor outstanding balance
DEFINE FUNCTION IF NOT EXISTS fn::vendor_outstanding($vendor_id: string) {
    LET $invoices = (SELECT VALUE math::sum(total_amount - paid_amount) 
                     FROM purchase_invoice 
                     WHERE vendor_id = $vendor_id AND payment_status != 'paid' 
                     GROUP ALL);
    RETURN IF array::len($invoices) > 0 THEN $invoices[0] ELSE 0 END;
};

-- Function: Get customer outstanding balance
DEFINE FUNCTION IF NOT EXISTS fn::customer_outstanding($customer_id: string) {
    LET $invoices = (SELECT VALUE math::sum(total_amount - received_amount) 
                     FROM sales_invoice 
                     WHERE customer_id = $customer_id AND payment_status != 'paid' 
                     GROUP ALL);
    RETURN IF array::len($invoices) > 0 THEN $invoices[0] ELSE 0 END;
};

-- ============================================
-- FULL-TEXT SEARCH INDEXES
-- ============================================

-- Full-text search on account name
DEFINE INDEX IF NOT EXISTS idx_account_fts ON account FIELDS account_name, account_code
    SEARCH ANALYZER simple_analyzer BM25;

-- Full-text search on vendor name
DEFINE INDEX IF NOT EXISTS idx_vendor_fts ON vendor FIELDS vendor_name
    SEARCH ANALYZER simple_analyzer BM25;

-- Full-text search on customer name
DEFINE INDEX IF NOT EXISTS idx_customer_fts ON customer FIELDS customer_name
    SEARCH ANALYZER simple_analyzer BM25;

-- ============================================
-- VERIFICATION
-- ============================================

INFO FOR TABLE account;
INFO FOR TABLE journal_entry;
INFO FOR TABLE journal_line;
INFO FOR TABLE vendor;
INFO FOR TABLE customer;
INFO FOR TABLE purchase_invoice;
INFO FOR TABLE sales_invoice;
INFO FOR TABLE payment;
INFO FOR TABLE budget;
