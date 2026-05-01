package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ============================================
// CHART OF ACCOUNTS (COA)
// ============================================

// Account represents a financial account in the Chart of Accounts
type Account struct {
	ID             string          `json:"id,omitempty"`
	AccountCode    string          `json:"account_code"`    // e.g., "1-1000", "2-2100"
	AccountName    string          `json:"account_name"`    // e.g., "Cash in Bank", "Accounts Payable"
	AccountType    AccountType     `json:"account_type"`    // asset, liability, equity, revenue, expense
	ParentAccount  string          `json:"parent_account"`  // For hierarchical COA (parent account ID)
	Level          int             `json:"level"`           // 1, 2, 3, 4, 5 (hierarchy depth)
	IsActive       bool            `json:"is_active"`       // Active/Inactive
	Currency       string          `json:"currency"`        // IDR, USD, EUR, etc.
	OpeningBalance decimal.Decimal `json:"opening_balance"` // Opening balance
	CurrentBalance decimal.Decimal `json:"current_balance"` // Current balance (calculated)
	IsControlAccount bool          `json:"is_control_account"` // Control account (has children)
	AllowPosting   bool            `json:"allow_posting"`   // Allow direct posting (false for control accounts)
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	CreatedBy      string          `json:"created_by"`
}

// AccountType defines the type of account
type AccountType string

const (
	AccountTypeAsset     AccountType = "asset"
	AccountTypeLiability AccountType = "liability"
	AccountTypeEquity    AccountType = "equity"
	AccountTypeRevenue   AccountType = "revenue"
	AccountTypeExpense   AccountType = "expense"
)

// AccountTree represents hierarchical account structure
type AccountTree struct {
	Account  Account        `json:"account"`
	Children []AccountTree  `json:"children,omitempty"`
}

// ============================================
// GENERAL LEDGER (GL)
// ============================================

// JournalEntry represents a journal entry header
type JournalEntry struct {
	ID          string          `json:"id,omitempty"`
	EntryNumber string          `json:"entry_number"` // Auto-generated: JE-2026-0001
	EntryDate   time.Time       `json:"entry_date"`
	EntryType   JournalType     `json:"entry_type"`
	Description string          `json:"description"`
	Reference   string          `json:"reference"`   // External reference (invoice, payment, etc.)
	Status      JournalStatus   `json:"status"`      // draft, posted, reversed
	Lines       []JournalLine   `json:"lines"`       // Journal lines (debit/credit)
	TotalDebit  decimal.Decimal `json:"total_debit"` // Total debit amount
	TotalCredit decimal.Decimal `json:"total_credit"` // Total credit amount
	PostedBy    string          `json:"posted_by"`
	PostedAt    *time.Time      `json:"posted_at"`
	ReversedBy  string          `json:"reversed_by"`
	ReversedAt  *time.Time      `json:"reversed_at"`
	CreatedBy   string          `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// JournalType defines the type of journal entry
type JournalType string

const (
	JournalTypeManual     JournalType = "manual"
	JournalTypeAuto       JournalType = "auto"
	JournalTypeOpening    JournalType = "opening"
	JournalTypeClosing    JournalType = "closing"
	JournalTypeAdjustment JournalType = "adjustment"
)

// JournalStatus defines the status of journal entry
type JournalStatus string

const (
	JournalStatusDraft    JournalStatus = "draft"
	JournalStatusPosted   JournalStatus = "posted"
	JournalStatusReversed JournalStatus = "reversed"
)

// JournalLine represents a journal entry line (debit/credit)
type JournalLine struct {
	ID             string          `json:"id,omitempty"`
	JournalEntryID string          `json:"journal_entry_id"`
	LineNumber     int             `json:"line_number"`     // Line sequence
	AccountID      string          `json:"account_id"`      // Account ID
	AccountCode    string          `json:"account_code"`    // Account code (for display)
	AccountName    string          `json:"account_name"`    // Account name (for display)
	Debit          decimal.Decimal `json:"debit"`           // Debit amount
	Credit         decimal.Decimal `json:"credit"`          // Credit amount
	Description    string          `json:"description"`     // Line description
	CostCenter     string          `json:"cost_center"`     // Cost center (optional)
	ProjectID      string          `json:"project_id"`      // Project ID (optional)
	DepartmentID   string          `json:"department_id"`   // Department ID (optional)
	CreatedAt      time.Time       `json:"created_at"`
}

// ============================================
// ACCOUNT PAYABLE (AP)
// ============================================

// Vendor represents a vendor/supplier
type Vendor struct {
	ID             string          `json:"id,omitempty"`
	VendorCode     string          `json:"vendor_code"`     // Auto-generated: VEN-0001
	VendorName     string          `json:"vendor_name"`
	VendorType     VendorType      `json:"vendor_type"`     // supplier, contractor, service_provider
	ContactPerson  string          `json:"contact_person"`
	Email          string          `json:"email"`
	Phone          string          `json:"phone"`
	Address        string          `json:"address"`
	TaxID          string          `json:"tax_id"`          // NPWP (Indonesia)
	PaymentTerms   int             `json:"payment_terms"`   // Payment terms in days (e.g., 30)
	CreditLimit    decimal.Decimal `json:"credit_limit"`    // Credit limit
	CurrentBalance decimal.Decimal `json:"current_balance"` // Current outstanding balance
	IsActive       bool            `json:"is_active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	CreatedBy      string          `json:"created_by"`
}

// VendorType defines the type of vendor
type VendorType string

const (
	VendorTypeSupplier        VendorType = "supplier"
	VendorTypeContractor      VendorType = "contractor"
	VendorTypeServiceProvider VendorType = "service_provider"
)

// PurchaseInvoice represents a purchase invoice (AP)
type PurchaseInvoice struct {
	ID             string          `json:"id,omitempty"`
	InvoiceNumber  string          `json:"invoice_number"`  // Auto-generated: PI-2026-0001
	VendorID       string          `json:"vendor_id"`
	VendorName     string          `json:"vendor_name"`     // For display
	InvoiceDate    time.Time       `json:"invoice_date"`
	DueDate        time.Time       `json:"due_date"`
	TotalAmount    decimal.Decimal `json:"total_amount"`    // Total invoice amount
	TaxAmount      decimal.Decimal `json:"tax_amount"`      // Tax amount (PPN)
	DiscountAmount decimal.Decimal `json:"discount_amount"` // Discount amount
	PaidAmount     decimal.Decimal `json:"paid_amount"`     // Amount paid
	Status         InvoiceStatus   `json:"status"`          // draft, submitted, approved, paid, cancelled
	PaymentStatus  PaymentStatus   `json:"payment_status"`  // unpaid, partial, paid
	Description    string          `json:"description"`
	Reference      string          `json:"reference"`       // PO number, etc.
	CreatedBy      string          `json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// InvoiceStatus defines the status of invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSubmitted InvoiceStatus = "submitted"
	InvoiceStatusApproved  InvoiceStatus = "approved"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

// PaymentStatus defines the payment status
type PaymentStatus string

const (
	PaymentStatusUnpaid  PaymentStatus = "unpaid"
	PaymentStatusPartial PaymentStatus = "partial"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusOverdue PaymentStatus = "overdue"
)

// ============================================
// ACCOUNT RECEIVABLE (AR)
// ============================================

// Customer represents a customer
type Customer struct {
	ID             string          `json:"id,omitempty"`
	CustomerCode   string          `json:"customer_code"`   // Auto-generated: CUS-0001
	CustomerName   string          `json:"customer_name"`
	CustomerType   CustomerType    `json:"customer_type"`   // individual, corporate, government
	ContactPerson  string          `json:"contact_person"`
	Email          string          `json:"email"`
	Phone          string          `json:"phone"`
	Address        string          `json:"address"`
	TaxID          string          `json:"tax_id"`          // NPWP (Indonesia)
	PaymentTerms   int             `json:"payment_terms"`   // Payment terms in days
	CreditLimit    decimal.Decimal `json:"credit_limit"`    // Credit limit
	CurrentBalance decimal.Decimal `json:"current_balance"` // Current outstanding balance
	IsActive       bool            `json:"is_active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	CreatedBy      string          `json:"created_by"`
}

// CustomerType defines the type of customer
type CustomerType string

const (
	CustomerTypeIndividual CustomerType = "individual"
	CustomerTypeCorporate  CustomerType = "corporate"
	CustomerTypeGovernment CustomerType = "government"
)

// SalesInvoice represents a sales invoice (AR)
type SalesInvoice struct {
	ID              string          `json:"id,omitempty"`
	InvoiceNumber   string          `json:"invoice_number"`   // Auto-generated: SI-2026-0001
	CustomerID      string          `json:"customer_id"`
	CustomerName    string          `json:"customer_name"`    // For display
	InvoiceDate     time.Time       `json:"invoice_date"`
	DueDate         time.Time       `json:"due_date"`
	TotalAmount     decimal.Decimal `json:"total_amount"`     // Total invoice amount
	TaxAmount       decimal.Decimal `json:"tax_amount"`       // Tax amount (PPN)
	DiscountAmount  decimal.Decimal `json:"discount_amount"`  // Discount amount
	ReceivedAmount  decimal.Decimal `json:"received_amount"`  // Amount received
	Status          InvoiceStatus   `json:"status"`           // draft, submitted, approved, paid, cancelled
	PaymentStatus   PaymentStatus   `json:"payment_status"`   // unpaid, partial, paid, overdue
	Description     string          `json:"description"`
	Reference       string          `json:"reference"`        // SO number, etc.
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ============================================
// PAYMENT
// ============================================

// Payment represents a payment transaction (vendor payment or customer receipt)
type Payment struct {
	ID              string          `json:"id,omitempty"`
	PaymentNumber   string          `json:"payment_number"`   // Auto-generated: PAY-2026-0001
	PaymentType     PaymentType     `json:"payment_type"`     // vendor_payment, customer_receipt
	PartyID         string          `json:"party_id"`         // Vendor ID or Customer ID
	PartyName       string          `json:"party_name"`       // For display
	PaymentDate     time.Time       `json:"payment_date"`
	PaymentMethod   PaymentMethod   `json:"payment_method"`   // cash, bank_transfer, check, credit_card
	Amount          decimal.Decimal `json:"amount"`           // Payment amount
	BankAccount     string          `json:"bank_account"`     // Bank account (if bank_transfer)
	ReferenceNumber string          `json:"reference_number"` // Check number, transfer reference, etc.
	Status          PaymentStatusEnum `json:"status"`         // draft, submitted, cleared, cancelled
	Description     string          `json:"description"`
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// PaymentType defines the type of payment
type PaymentType string

const (
	PaymentTypeVendorPayment   PaymentType = "vendor_payment"
	PaymentTypeCustomerReceipt PaymentType = "customer_receipt"
)

// PaymentMethod defines the payment method
type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodCheck        PaymentMethod = "check"
	PaymentMethodCreditCard   PaymentMethod = "credit_card"
)

// PaymentStatusEnum defines the payment status
type PaymentStatusEnum string

const (
	PaymentStatusEnumDraft     PaymentStatusEnum = "draft"
	PaymentStatusEnumSubmitted PaymentStatusEnum = "submitted"
	PaymentStatusEnumCleared   PaymentStatusEnum = "cleared"
	PaymentStatusEnumCancelled PaymentStatusEnum = "cancelled"
)

// ============================================
// BUDGET
// ============================================

// Budget represents a budget plan
type Budget struct {
	ID           string          `json:"id,omitempty"`
	BudgetName   string          `json:"budget_name"`
	FiscalYear   int             `json:"fiscal_year"`     // e.g., 2026
	AccountID    string          `json:"account_id"`      // Account ID
	AccountCode  string          `json:"account_code"`    // For display
	AccountName  string          `json:"account_name"`    // For display
	Department   string          `json:"department"`      // Department (optional)
	CostCenter   string          `json:"cost_center"`     // Cost center (optional)
	PeriodType   PeriodType      `json:"period_type"`     // monthly, quarterly, yearly
	BudgetAmount decimal.Decimal `json:"budget_amount"`   // Budget amount
	ActualAmount decimal.Decimal `json:"actual_amount"`   // Actual amount (calculated)
	Variance     decimal.Decimal `json:"variance"`        // Variance (budget - actual)
	Status       BudgetStatus    `json:"status"`          // draft, approved, active, closed
	CreatedBy    string          `json:"created_by"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// PeriodType defines the budget period type
type PeriodType string

const (
	PeriodTypeMonthly   PeriodType = "monthly"
	PeriodTypeQuarterly PeriodType = "quarterly"
	PeriodTypeYearly    PeriodType = "yearly"
)

// BudgetStatus defines the budget status
type BudgetStatus string

const (
	BudgetStatusDraft    BudgetStatus = "draft"
	BudgetStatusApproved BudgetStatus = "approved"
	BudgetStatusActive   BudgetStatus = "active"
	BudgetStatusClosed   BudgetStatus = "closed"
)

// ============================================
// FINANCIAL REPORTS
// ============================================

// BalanceSheetReport represents a balance sheet report
type BalanceSheetReport struct {
	AsOfDate         time.Time                `json:"as_of_date"`
	Assets           BalanceSheetSection      `json:"assets"`
	Liabilities      BalanceSheetSection      `json:"liabilities"`
	Equity           BalanceSheetSection      `json:"equity"`
	TotalAssets      decimal.Decimal          `json:"total_assets"`
	TotalLiabilities decimal.Decimal          `json:"total_liabilities"`
	TotalEquity      decimal.Decimal          `json:"total_equity"`
}

// BalanceSheetSection represents a section in balance sheet
type BalanceSheetSection struct {
	Accounts []AccountBalance `json:"accounts"`
	Total    decimal.Decimal  `json:"total"`
}

// AccountBalance represents account balance
type AccountBalance struct {
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	Balance     decimal.Decimal `json:"balance"`
}

// ProfitLossReport represents a profit & loss report
type ProfitLossReport struct {
	FromDate       time.Time       `json:"from_date"`
	ToDate         time.Time       `json:"to_date"`
	Revenue        []AccountBalance `json:"revenue"`
	Expenses       []AccountBalance `json:"expenses"`
	TotalRevenue   decimal.Decimal `json:"total_revenue"`
	TotalExpenses  decimal.Decimal `json:"total_expenses"`
	GrossProfit    decimal.Decimal `json:"gross_profit"`
	NetProfit      decimal.Decimal `json:"net_profit"`
}

// CashFlowReport represents a cash flow statement
type CashFlowReport struct {
	FromDate                time.Time       `json:"from_date"`
	ToDate                  time.Time       `json:"to_date"`
	
	// Operating Activities
	OperatingActivities     []CashFlowItem  `json:"operating_activities"`
	NetCashFromOperating    decimal.Decimal `json:"net_cash_from_operating"`
	
	// Investing Activities
	InvestingActivities     []CashFlowItem  `json:"investing_activities"`
	NetCashFromInvesting    decimal.Decimal `json:"net_cash_from_investing"`
	
	// Financing Activities
	FinancingActivities     []CashFlowItem  `json:"financing_activities"`
	NetCashFromFinancing    decimal.Decimal `json:"net_cash_from_financing"`
	
	// Summary
	NetIncreaseInCash       decimal.Decimal `json:"net_increase_in_cash"`
	CashAtBeginning         decimal.Decimal `json:"cash_at_beginning"`
	CashAtEnd               decimal.Decimal `json:"cash_at_end"`
}

// CashFlowItem represents a line item in cash flow statement
type CashFlowItem struct {
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
}

// TrialBalanceReport represents a trial balance report
type TrialBalanceReport struct {
	AsOfDate    time.Time            `json:"as_of_date"`
	Accounts    []TrialBalanceAccount `json:"accounts"`
	TotalDebit  decimal.Decimal      `json:"total_debit"`
	TotalCredit decimal.Decimal      `json:"total_credit"`
	IsBalanced  bool                 `json:"is_balanced"`
}

// TrialBalanceAccount represents an account in trial balance
type TrialBalanceAccount struct {
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	AccountType AccountType     `json:"account_type"`
	Debit       decimal.Decimal `json:"debit"`
	Credit      decimal.Decimal `json:"credit"`
}

// GeneralLedgerReport represents a general ledger report
type GeneralLedgerReport struct {
	FromDate    time.Time              `json:"from_date"`
	ToDate      time.Time              `json:"to_date"`
	AccountID   string                 `json:"account_id,omitempty"`
	AccountCode string                 `json:"account_code,omitempty"`
	AccountName string                 `json:"account_name,omitempty"`
	Entries     []GeneralLedgerEntry   `json:"entries"`
	OpeningBalance decimal.Decimal     `json:"opening_balance"`
	TotalDebit  decimal.Decimal        `json:"total_debit"`
	TotalCredit decimal.Decimal        `json:"total_credit"`
	ClosingBalance decimal.Decimal     `json:"closing_balance"`
}

// GeneralLedgerEntry represents an entry in general ledger
type GeneralLedgerEntry struct {
	Date            time.Time       `json:"date"`
	JournalNumber   string          `json:"journal_number"`
	Description     string          `json:"description"`
	Reference       string          `json:"reference,omitempty"`
	Debit           decimal.Decimal `json:"debit"`
	Credit          decimal.Decimal `json:"credit"`
	Balance         decimal.Decimal `json:"balance"`
}

// AgingReport represents accounts receivable/payable aging report
type AgingReport struct {
	AsOfDate    time.Time       `json:"as_of_date"`
	ReportType  string          `json:"report_type"`  // receivable, payable
	Items       []AgingItem     `json:"items"`
	TotalCurrent decimal.Decimal `json:"total_current"`
	Total1to30   decimal.Decimal `json:"total_1_to_30"`
	Total31to60  decimal.Decimal `json:"total_31_to_60"`
	Total61to90  decimal.Decimal `json:"total_61_to_90"`
	TotalOver90  decimal.Decimal `json:"total_over_90"`
	GrandTotal   decimal.Decimal `json:"grand_total"`
}

// AgingItem represents an item in aging report
type AgingItem struct {
	PartyID     string          `json:"party_id"`      // Customer or Vendor ID
	PartyName   string          `json:"party_name"`
	InvoiceNumber string        `json:"invoice_number"`
	InvoiceDate time.Time       `json:"invoice_date"`
	DueDate     time.Time       `json:"due_date"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	PaidAmount  decimal.Decimal `json:"paid_amount"`
	Outstanding decimal.Decimal `json:"outstanding"`
	DaysOverdue int             `json:"days_overdue"`
	Current     decimal.Decimal `json:"current"`       // Not yet due
	Days1to30   decimal.Decimal `json:"days_1_to_30"`
	Days31to60  decimal.Decimal `json:"days_31_to_60"`
	Days61to90  decimal.Decimal `json:"days_61_to_90"`
	DaysOver90  decimal.Decimal `json:"days_over_90"`
}
