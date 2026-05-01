package handlers

import (
"fmt"
"net/http"
"time"

"agileos-backend/logger"
"agileos-backend/models"

"github.com/gin-gonic/gin"
"github.com/shopspring/decimal"
)

// ============================================
// CHART OF ACCOUNTS (COA) HANDLERS
// ============================================

// CreateAccount creates a new account in the chart of accounts
func (h *AccountingHandler) CreateAccount(c *gin.Context) {
var account models.Account
if err := c.ShouldBindJSON(&account); err != nil {
logger.LogError("Failed to bind account data", err, nil)
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

userID, exists := c.Get("user_id")
if !exists {
c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
return
}

account.CreatedBy = userID.(string)
account.CreatedAt = time.Now()
account.UpdatedAt = time.Now()
account.IsActive = true
account.CurrentBalance = decimal.Zero

query := `CREATE account CONTENT {
account_code: $account_code,
account_name: $account_name,
account_type: $account_type,
parent_account: $parent_account,
level: $level,
is_active: $is_active,
currency: $currency,
opening_balance: $opening_balance,
current_balance: $current_balance,
is_control_account: $is_control_account,
allow_posting: $allow_posting,
created_by: $created_by,
created_at: $created_at,
updated_at: $updated_at
}`

params := map[string]interface{}{
"account_code":       account.AccountCode,
"account_name":       account.AccountName,
"account_type":       account.AccountType,
"parent_account":     account.ParentAccount,
"level":              account.Level,
"is_active":          account.IsActive,
"currency":           account.Currency,
"opening_balance":    account.OpeningBalance,
"current_balance":    account.CurrentBalance,
"is_control_account": account.IsControlAccount,
"allow_posting":      account.AllowPosting,
"created_by":         account.CreatedBy,
"created_at":         account.CreatedAt,
"updated_at":         account.UpdatedAt,
}

result, err := h.db.QuerySlice(query, params)
if err != nil {
logger.LogError("Failed to create account", err, nil)
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
return
}

logger.Log.Info().
Str("account_code", account.AccountCode).
Str("account_name", account.AccountName).
Msg("Account created successfully")

if len(result) > 0 {
c.JSON(http.StatusCreated, result[0])
} else {
c.JSON(http.StatusCreated, gin.H{"message": "Account created successfully"})
}
}

// GetAccounts retrieves all accounts
func (h *AccountingHandler) GetAccounts(c *gin.Context) {
	query := `SELECT * FROM account WHERE is_active = true ORDER BY account_code`
	
	result, err := h.db.QuerySlice(query, nil)
	if err != nil {
		logger.LogError("Failed to get accounts", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get accounts"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAccountTree retrieves hierarchical account tree
func (h *AccountingHandler) GetAccountTree(c *gin.Context) {
	query := `SELECT * FROM account WHERE is_active = true ORDER BY account_code`
	
	result, err := h.db.QuerySlice(query, nil)
	if err != nil {
		logger.LogError("Failed to get account tree", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get account tree"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAccount retrieves a single account by ID
func (h *AccountingHandler) GetAccount(c *gin.Context) {
	id := c.Param("id")
	
	query := fmt.Sprintf(`SELECT * FROM account:%s`, id)
	
	result, err := h.db.QuerySlice(query, nil)
	if err != nil {
		logger.LogError("Failed to get account", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get account"})
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, result[0])
}

// UpdateAccount updates an existing account
func (h *AccountingHandler) UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	
	var account models.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		logger.LogError("Failed to bind account data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account.UpdatedAt = time.Now()

	query := fmt.Sprintf(`UPDATE account:%s MERGE {
		account_name: $account_name,
		account_type: $account_type,
		parent_account: $parent_account,
		level: $level,
		is_active: $is_active,
		currency: $currency,
		is_control_account: $is_control_account,
		allow_posting: $allow_posting,
		updated_at: $updated_at
	}`, id)

	params := map[string]interface{}{
		"account_name":       account.AccountName,
		"account_type":       account.AccountType,
		"parent_account":     account.ParentAccount,
		"level":              account.Level,
		"is_active":          account.IsActive,
		"currency":           account.Currency,
		"is_control_account": account.IsControlAccount,
		"allow_posting":      account.AllowPosting,
		"updated_at":         account.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to update account", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}

	logger.Log.Info().Str("account_id", id).Msg("Account updated successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Account updated successfully"})
	}
}

// DeleteAccount deletes an account
func (h *AccountingHandler) DeleteAccount(c *gin.Context) {
	id := c.Param("id")
	
	query := fmt.Sprintf(`DELETE account:%s`, id)
	
	_, err := h.db.QuerySlice(query, nil)
	if err != nil {
		logger.LogError("Failed to delete account", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	logger.Log.Info().Str("account_id", id).Msg("Account deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

// ============================================
// JOURNAL ENTRY HANDLERS
// ============================================

// CreateJournalEntry creates a new journal entry
func (h *AccountingHandler) CreateJournalEntry(c *gin.Context) {
	var entry models.JournalEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		logger.LogError("Failed to bind journal entry data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	entry.CreatedBy = userID.(string)
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()
	entry.Status = models.JournalStatusDraft

	// Calculate totals
	entry.TotalDebit = decimal.Zero
	entry.TotalCredit = decimal.Zero
	for _, line := range entry.Lines {
		entry.TotalDebit = entry.TotalDebit.Add(line.Debit)
		entry.TotalCredit = entry.TotalCredit.Add(line.Credit)
	}

	// Validate balanced entry
	if !entry.TotalDebit.Equal(entry.TotalCredit) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Journal entry is not balanced"})
		return
	}

	query := `CREATE journal_entry CONTENT {
		entry_number: $entry_number,
		entry_date: $entry_date,
		entry_type: $entry_type,
		description: $description,
		reference: $reference,
		status: $status,
		total_debit: $total_debit,
		total_credit: $total_credit,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"entry_number": entry.EntryNumber,
		"entry_date":   entry.EntryDate,
		"entry_type":   entry.EntryType,
		"description":  entry.Description,
		"reference":    entry.Reference,
		"status":       entry.Status,
		"total_debit":  entry.TotalDebit,
		"total_credit": entry.TotalCredit,
		"created_by":   entry.CreatedBy,
		"created_at":   entry.CreatedAt,
		"updated_at":   entry.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create journal entry", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create journal entry"})
		return
	}

	logger.Log.Info().
		Str("entry_number", entry.EntryNumber).
		Msg("Journal entry created successfully")

	if len(result) > 0 {
		c.JSON(http.StatusCreated, result[0])
	} else {
		c.JSON(http.StatusCreated, gin.H{"message": "Journal entry created successfully"})
	}
}

// GetJournalEntries retrieves all journal entries
func (h *AccountingHandler) GetJournalEntries(c *gin.Context) {
	query := `SELECT * FROM journal_entry ORDER BY entry_date DESC`
	
	result, err := h.db.QuerySlice(query, nil)
	if err != nil {
		logger.LogError("Failed to get journal entries", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get journal entries"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetJournalEntry retrieves a single journal entry by ID
func (h *AccountingHandler) GetJournalEntry(c *gin.Context) {
	id := c.Param("id")
	
	query := fmt.Sprintf(`SELECT * FROM journal_entry:%s`, id)
	
	result, err := h.db.QuerySlice(query, nil)
	if err != nil {
		logger.LogError("Failed to get journal entry", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get journal entry"})
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Journal entry not found"})
		return
	}

	c.JSON(http.StatusOK, result[0])
}

// PostJournalEntry posts a journal entry
func (h *AccountingHandler) PostJournalEntry(c *gin.Context) {
	id := c.Param("id")
	
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	now := time.Now()
	query := fmt.Sprintf(`UPDATE journal_entry:%s MERGE {
		status: $status,
		posted_by: $posted_by,
		posted_at: $posted_at
	}`, id)

	params := map[string]interface{}{
		"status":    models.JournalStatusPosted,
		"posted_by": userID.(string),
		"posted_at": now,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to post journal entry", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to post journal entry"})
		return
	}

	logger.Log.Info().Str("entry_id", id).Msg("Journal entry posted successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Journal entry posted successfully"})
	}
}

// ReverseJournalEntry reverses a journal entry
func (h *AccountingHandler) ReverseJournalEntry(c *gin.Context) {
	id := c.Param("id")
	
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	now := time.Now()
	query := fmt.Sprintf(`UPDATE journal_entry:%s MERGE {
		status: $status,
		reversed_by: $reversed_by,
		reversed_at: $reversed_at
	}`, id)

	params := map[string]interface{}{
		"status":      models.JournalStatusReversed,
		"reversed_by": userID.(string),
		"reversed_at": now,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to reverse journal entry", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reverse journal entry"})
		return
	}

	logger.Log.Info().Str("entry_id", id).Msg("Journal entry reversed successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Journal entry reversed successfully"})
	}
}

// DeleteJournalEntry deletes a journal entry
func (h *AccountingHandler) DeleteJournalEntry(c *gin.Context) {
	id := c.Param("id")
	
	query := fmt.Sprintf(`DELETE journal_entry:%s`, id)
	
	_, err := h.db.QuerySlice(query, nil)
	if err != nil {
		logger.LogError("Failed to delete journal entry", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete journal entry"})
		return
	}

	logger.Log.Info().Str("entry_id", id).Msg("Journal entry deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Journal entry deleted successfully"})
}

// ============================================
// VENDOR HANDLERS (ACCOUNT PAYABLE)
// ============================================

// CreateVendor creates a new vendor
func (h *AccountingHandler) CreateVendor(c *gin.Context) {
	var vendor models.Vendor
	if err := c.ShouldBindJSON(&vendor); err != nil {
		logger.LogError("Failed to bind vendor data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	vendor.CreatedBy = userID.(string)
	vendor.CreatedAt = time.Now()
	vendor.UpdatedAt = time.Now()
	vendor.IsActive = true
	vendor.CurrentBalance = decimal.Zero

	query := `CREATE vendor CONTENT {
		vendor_code: $vendor_code,
		vendor_name: $vendor_name,
		vendor_type: $vendor_type,
		contact_person: $contact_person,
		email: $email,
		phone: $phone,
		address: $address,
		tax_id: $tax_id,
		payment_terms: $payment_terms,
		credit_limit: $credit_limit,
		current_balance: $current_balance,
		is_active: $is_active,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"vendor_code":     vendor.VendorCode,
		"vendor_name":     vendor.VendorName,
		"vendor_type":     vendor.VendorType,
		"contact_person":  vendor.ContactPerson,
		"email":           vendor.Email,
		"phone":           vendor.Phone,
		"address":         vendor.Address,
		"tax_id":          vendor.TaxID,
		"payment_terms":   vendor.PaymentTerms,
		"credit_limit":    vendor.CreditLimit,
		"current_balance": vendor.CurrentBalance,
		"is_active":       vendor.IsActive,
		"created_by":      vendor.CreatedBy,
		"created_at":      vendor.CreatedAt,
		"updated_at":      vendor.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create vendor", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vendor"})
		return
	}

	logger.Log.Info().
		Str("vendor_code", vendor.VendorCode).
		Str("vendor_name", vendor.VendorName).
		Msg("Vendor created successfully")

	if len(result) > 0 {
		c.JSON(http.StatusCreated, result[0])
	} else {
		c.JSON(http.StatusCreated, gin.H{"message": "Vendor created successfully"})
	}
}

// GetVendors retrieves all vendors
func (h *AccountingHandler) GetVendors(c *gin.Context) {
	query := `SELECT * FROM vendor WHERE is_active = true ORDER BY vendor_name`
	
	result, err := h.db.QuerySlice(query, nil)
	if err != nil {
		logger.LogError("Failed to get vendors", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vendors"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetVendor retrieves a single vendor by ID
func (h *AccountingHandler) GetVendor(c *gin.Context) {
	id := c.Param("id")
	
	query := fmt.Sprintf(`SELECT * FROM vendor:%s`, id)
	
	result, err := h.db.QuerySlice(query, nil)
	if err != nil {
		logger.LogError("Failed to get vendor", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vendor"})
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vendor not found"})
		return
	}

	c.JSON(http.StatusOK, result[0])
}

// UpdateVendor updates an existing vendor
func (h *AccountingHandler) UpdateVendor(c *gin.Context) {
	id := c.Param("id")
	
	var vendor models.Vendor
	if err := c.ShouldBindJSON(&vendor); err != nil {
		logger.LogError("Failed to bind vendor data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vendor.UpdatedAt = time.Now()

	query := fmt.Sprintf(`UPDATE vendor:%s MERGE {
		vendor_name: $vendor_name,
		vendor_type: $vendor_type,
		contact_person: $contact_person,
		email: $email,
		phone: $phone,
		address: $address,
		tax_id: $tax_id,
		payment_terms: $payment_terms,
		credit_limit: $credit_limit,
		is_active: $is_active,
		updated_at: $updated_at
	}`, id)

	params := map[string]interface{}{
		"vendor_name":    vendor.VendorName,
		"vendor_type":    vendor.VendorType,
		"contact_person": vendor.ContactPerson,
		"email":          vendor.Email,
		"phone":          vendor.Phone,
		"address":        vendor.Address,
		"tax_id":         vendor.TaxID,
		"payment_terms":  vendor.PaymentTerms,
		"credit_limit":   vendor.CreditLimit,
		"is_active":      vendor.IsActive,
		"updated_at":     vendor.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to update vendor", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vendor"})
		return
	}

	logger.Log.Info().Str("vendor_id", id).Msg("Vendor updated successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Vendor updated successfully"})
	}
}

// DeleteVendor deletes a vendor
func (h *AccountingHandler) DeleteVendor(c *gin.Context) {
	id := c.Param("id")
	
	query := fmt.Sprintf(`DELETE vendor:%s`, id)
	
	_, err := h.db.QuerySlice(query, nil)
	if err != nil {
		logger.LogError("Failed to delete vendor", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vendor"})
		return
	}

	logger.Log.Info().Str("vendor_id", id).Msg("Vendor deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Vendor deleted successfully"})
}

// ============================================
// PURCHASE INVOICE HANDLERS (ACCOUNT PAYABLE)
// ============================================

// CreatePurchaseInvoice creates a new purchase invoice
func (h *AccountingHandler) CreatePurchaseInvoice(c *gin.Context) {
	var invoice models.PurchaseInvoice
	if err := c.ShouldBindJSON(&invoice); err != nil {
		logger.LogError("Failed to bind purchase invoice data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	invoice.CreatedBy = userID.(string)
	invoice.CreatedAt = time.Now()
	invoice.UpdatedAt = time.Now()
	invoice.Status = models.InvoiceStatusDraft
	invoice.PaymentStatus = models.PaymentStatusUnpaid
	invoice.PaidAmount = decimal.Zero

	query := `CREATE purchase_invoice CONTENT {
		invoice_number: $invoice_number,
		vendor_id: $vendor_id,
		vendor_name: $vendor_name,
		invoice_date: $invoice_date,
		due_date: $due_date,
		total_amount: $total_amount,
		tax_amount: $tax_amount,
		discount_amount: $discount_amount,
		paid_amount: $paid_amount,
		status: $status,
		payment_status: $payment_status,
		description: $description,
		reference: $reference,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"invoice_number":  invoice.InvoiceNumber,
		"vendor_id":       invoice.VendorID,
		"vendor_name":     invoice.VendorName,
		"invoice_date":    invoice.InvoiceDate,
		"due_date":        invoice.DueDate,
		"total_amount":    invoice.TotalAmount,
		"tax_amount":      invoice.TaxAmount,
		"discount_amount": invoice.DiscountAmount,
		"paid_amount":     invoice.PaidAmount,
		"status":          invoice.Status,
		"payment_status":  invoice.PaymentStatus,
		"description":     invoice.Description,
		"reference":       invoice.Reference,
		"created_by":      invoice.CreatedBy,
		"created_at":      invoice.CreatedAt,
		"updated_at":      invoice.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create purchase invoice", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase invoice"})
		return
	}

	logger.Log.Info().
		Str("invoice_number", invoice.InvoiceNumber).
		Str("vendor_name", invoice.VendorName).
		Msg("Purchase invoice created successfully")

	if len(result) > 0 {
		c.JSON(http.StatusCreated, result[0])
	} else {
		c.JSON(http.StatusCreated, gin.H{"message": "Purchase invoice created successfully"})
	}
}

// GetPurchaseInvoices retrieves all purchase invoices
func (h *AccountingHandler) GetPurchaseInvoices(c *gin.Context) {
	query := `SELECT * FROM purchase_invoice ORDER BY invoice_date DESC`
	
	result, err := h.db.QuerySlice(query, nil)
	if err != nil {
		logger.LogError("Failed to get purchase invoices", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get purchase invoices"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPurchaseInvoice retrieves a single purchase invoice by ID
func (h *AccountingHandler) GetPurchaseInvoice(c *gin.Context) {
	id := c.Param("id")
	
	query := fmt.Sprintf(`SELECT * FROM purchase_invoice:%s`, id)
	
	result, err := h.db.QuerySlice(query, nil)
	if err != nil {
		logger.LogError("Failed to get purchase invoice", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get purchase invoice"})
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase invoice not found"})
		return
	}

	c.JSON(http.StatusOK, result[0])
}

// UpdatePurchaseInvoice updates an existing purchase invoice
func (h *AccountingHandler) UpdatePurchaseInvoice(c *gin.Context) {
	id := c.Param("id")
	
	var invoice models.PurchaseInvoice
	if err := c.ShouldBindJSON(&invoice); err != nil {
		logger.LogError("Failed to bind purchase invoice data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice.UpdatedAt = time.Now()

	query := fmt.Sprintf(`UPDATE purchase_invoice:%s MERGE {
		vendor_id: $vendor_id,
		vendor_name: $vendor_name,
		invoice_date: $invoice_date,
		due_date: $due_date,
		total_amount: $total_amount,
		tax_amount: $tax_amount,
		discount_amount: $discount_amount,
		description: $description,
		reference: $reference,
		updated_at: $updated_at
	}`, id)

	params := map[string]interface{}{
		"vendor_id":       invoice.VendorID,
		"vendor_name":     invoice.VendorName,
		"invoice_date":    invoice.InvoiceDate,
		"due_date":        invoice.DueDate,
		"total_amount":    invoice.TotalAmount,
		"tax_amount":      invoice.TaxAmount,
		"discount_amount": invoice.DiscountAmount,
		"description":     invoice.Description,
		"reference":       invoice.Reference,
		"updated_at":      invoice.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to update purchase invoice", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update purchase invoice"})
		return
	}

	logger.Log.Info().Str("invoice_id", id).Msg("Purchase invoice updated successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Purchase invoice updated successfully"})
	}
}

// DeletePurchaseInvoice deletes a purchase invoice
func (h *AccountingHandler) DeletePurchaseInvoice(c *gin.Context) {
	id := c.Param("id")
	
	query := fmt.Sprintf(`DELETE purchase_invoice:%s`, id)
	
	_, err := h.db.QuerySlice(query, nil)
	if err != nil {
		logger.LogError("Failed to delete purchase invoice", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete purchase invoice"})
		return
	}

	logger.Log.Info().Str("invoice_id", id).Msg("Purchase invoice deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Purchase invoice deleted successfully"})
}

// ApprovePurchaseInvoice approves a purchase invoice
func (h *AccountingHandler) ApprovePurchaseInvoice(c *gin.Context) {
	id := c.Param("id")
	
	query := fmt.Sprintf(`UPDATE purchase_invoice:%s MERGE {
		status: $status,
		updated_at: $updated_at
	}`, id)

	params := map[string]interface{}{
		"status":     models.InvoiceStatusApproved,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to approve purchase invoice", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve purchase invoice"})
		return
	}

	logger.Log.Info().Str("invoice_id", id).Msg("Purchase invoice approved successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Purchase invoice approved successfully"})
	}
}

// CancelPurchaseInvoice cancels a purchase invoice
func (h *AccountingHandler) CancelPurchaseInvoice(c *gin.Context) {
	id := c.Param("id")
	
	query := fmt.Sprintf(`UPDATE purchase_invoice:%s MERGE {
		status: $status,
		updated_at: $updated_at
	}`, id)

	params := map[string]interface{}{
		"status":     models.InvoiceStatusCancelled,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to cancel purchase invoice", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel purchase invoice"})
		return
	}

	logger.Log.Info().Str("invoice_id", id).Msg("Purchase invoice cancelled successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Purchase invoice cancelled successfully"})
	}
}
