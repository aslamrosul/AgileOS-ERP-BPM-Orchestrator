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
// ACCOUNT RECEIVABLE (AR) - CUSTOMER HANDLERS
// ============================================

// CreateCustomer creates a new customer
func (h *AccountingHandler) CreateCustomer(c *gin.Context) {
	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		logger.LogError("Failed to bind customer data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate customer code
	customers, err := h.db.QuerySlice(
		"SELECT customer_code FROM customer ORDER BY customer_code DESC LIMIT 1",
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last customer code", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate customer code"})
		return
	}

	customerCode := "CUS-0001"
	if len(customers) > 0 {
		if customerMap, ok := customers[0].(map[string]interface{}); ok {
			if lastCode, ok := customerMap["customer_code"].(string); ok {
				var lastNum int
				fmt.Sscanf(lastCode, "CUS-%d", &lastNum)
				customerCode = fmt.Sprintf("CUS-%04d", lastNum+1)
			}
		}
	}

	customer.CustomerCode = customerCode
	customer.CreatedBy = userID.(string)
	customer.CreatedAt = time.Now()
	customer.UpdatedAt = time.Now()
	customer.IsActive = true
	customer.CurrentBalance = decimal.Zero

	query := `CREATE customer CONTENT {
		customer_code: $customer_code,
		customer_name: $customer_name,
		customer_type: $customer_type,
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
		"customer_code":    customer.CustomerCode,
		"customer_name":    customer.CustomerName,
		"customer_type":    customer.CustomerType,
		"contact_person":   customer.ContactPerson,
		"email":            customer.Email,
		"phone":            customer.Phone,
		"address":          customer.Address,
		"tax_id":           customer.TaxID,
		"payment_terms":    customer.PaymentTerms,
		"credit_limit":     customer.CreditLimit,
		"current_balance":  customer.CurrentBalance,
		"is_active":        customer.IsActive,
		"created_by":       customer.CreatedBy,
		"created_at":       customer.CreatedAt,
		"updated_at":       customer.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create customer", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		return
	}

	logger.Log.Info().
		Str("customer_code", customer.CustomerCode).
		Str("customer_name", customer.CustomerName).
		Msg("Customer created successfully")

	if len(result) > 0 {
		c.JSON(http.StatusCreated, result[0])
	} else {
		c.JSON(http.StatusCreated, gin.H{"message": "Customer created successfully"})
	}
}

// GetCustomers retrieves all customers with filters
func (h *AccountingHandler) GetCustomers(c *gin.Context) {
	customerType := c.Query("customer_type")
	isActive := c.Query("is_active")

	query := "SELECT * FROM customer"
	params := make(map[string]interface{})

	var conditions []string
	if customerType != "" {
		conditions = append(conditions, "customer_type = $customer_type")
		params["customer_type"] = customerType
	}
	if isActive != "" {
		conditions = append(conditions, "is_active = $is_active")
		params["is_active"] = isActive == "true"
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY customer_code ASC"

	customers, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get customers", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve customers"})
		return
	}

	c.JSON(http.StatusOK, customers)
}

// GetCustomer retrieves a customer by ID
func (h *AccountingHandler) GetCustomer(c *gin.Context) {
	customerID := c.Param("id")

	customers, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": customerID},
	)
	if err != nil || len(customers) == 0 {
		logger.LogError("Customer not found", err, map[string]interface{}{"customer_id": customerID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(http.StatusOK, customers[0])
}

// UpdateCustomer updates an existing customer
func (h *AccountingHandler) UpdateCustomer(c *gin.Context) {
	customerID := c.Param("id")

	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		logger.LogError("Failed to bind customer data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer.UpdatedAt = time.Now()

	query := `UPDATE $id SET
		customer_name = $customer_name,
		customer_type = $customer_type,
		contact_person = $contact_person,
		email = $email,
		phone = $phone,
		address = $address,
		tax_id = $tax_id,
		payment_terms = $payment_terms,
		credit_limit = $credit_limit,
		is_active = $is_active,
		updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":              customerID,
		"customer_name":   customer.CustomerName,
		"customer_type":   customer.CustomerType,
		"contact_person":  customer.ContactPerson,
		"email":           customer.Email,
		"phone":           customer.Phone,
		"address":         customer.Address,
		"tax_id":          customer.TaxID,
		"payment_terms":   customer.PaymentTerms,
		"credit_limit":    customer.CreditLimit,
		"is_active":       customer.IsActive,
		"updated_at":      customer.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to update customer", err, map[string]interface{}{"customer_id": customerID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		return
	}

	logger.Log.Info().
		Str("customer_id", customerID).
		Str("customer_name", customer.CustomerName).
		Msg("Customer updated successfully")

	c.JSON(http.StatusOK, result[0])
}

// DeleteCustomer soft deletes a customer
func (h *AccountingHandler) DeleteCustomer(c *gin.Context) {
	customerID := c.Param("id")

	query := `UPDATE $id SET is_active = false, updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         customerID,
		"updated_at": time.Now(),
	}

	_, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to delete customer", err, map[string]interface{}{"customer_id": customerID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer"})
		return
	}

	logger.Log.Info().
		Str("customer_id", customerID).
		Msg("Customer deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

// ============================================
// ACCOUNT RECEIVABLE (AR) - SALES INVOICE HANDLERS
// ============================================

// CreateSalesInvoice creates a new sales invoice
func (h *AccountingHandler) CreateSalesInvoice(c *gin.Context) {
	var invoice models.SalesInvoice
	if err := c.ShouldBindJSON(&invoice); err != nil {
		logger.LogError("Failed to bind invoice data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate invoice number
	year := time.Now().Year()
	invoices, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT invoice_number FROM sales_invoice WHERE invoice_number LIKE 'SI-%d-%%' ORDER BY invoice_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last invoice number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate invoice number"})
		return
	}

	invoiceNumber := fmt.Sprintf("SI-%d-0001", year)
	if len(invoices) > 0 {
		if invoiceMap, ok := invoices[0].(map[string]interface{}); ok {
			if lastNumber, ok := invoiceMap["invoice_number"].(string); ok {
				var lastNum int
				fmt.Sscanf(lastNumber, fmt.Sprintf("SI-%d-%%d", year), &lastNum)
				invoiceNumber = fmt.Sprintf("SI-%d-%04d", year, lastNum+1)
			}
		}
	}

	invoice.InvoiceNumber = invoiceNumber
	invoice.CreatedBy = userID.(string)
	invoice.CreatedAt = time.Now()
	invoice.UpdatedAt = time.Now()
	invoice.ReceivedAmount = decimal.Zero

	if invoice.Status == "draft" {
		invoice.PaymentStatus = "unpaid"
	}

	query := `CREATE sales_invoice CONTENT {
		invoice_number: $invoice_number,
		customer_id: $customer_id,
		customer_name: $customer_name,
		invoice_date: $invoice_date,
		due_date: $due_date,
		total_amount: $total_amount,
		tax_amount: $tax_amount,
		discount_amount: $discount_amount,
		received_amount: $received_amount,
		status: $status,
		payment_status: $payment_status,
		description: $description,
		reference: $reference,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"invoice_number":   invoice.InvoiceNumber,
		"customer_id":      invoice.CustomerID,
		"customer_name":    invoice.CustomerName,
		"invoice_date":     invoice.InvoiceDate,
		"due_date":         invoice.DueDate,
		"total_amount":     invoice.TotalAmount,
		"tax_amount":       invoice.TaxAmount,
		"discount_amount":  invoice.DiscountAmount,
		"received_amount":  invoice.ReceivedAmount,
		"status":           invoice.Status,
		"payment_status":   invoice.PaymentStatus,
		"description":      invoice.Description,
		"reference":        invoice.Reference,
		"created_by":       invoice.CreatedBy,
		"created_at":       invoice.CreatedAt,
		"updated_at":       invoice.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create sales invoice", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sales invoice"})
		return
	}

	logger.Log.Info().
		Str("invoice_number", invoice.InvoiceNumber).
		Str("customer_name", invoice.CustomerName).
		Msg("Sales invoice created successfully")

	if len(result) > 0 {
		c.JSON(http.StatusCreated, result[0])
	} else {
		c.JSON(http.StatusCreated, gin.H{"message": "Sales invoice created successfully"})
	}
}

// GetSalesInvoices retrieves all sales invoices with filters
func (h *AccountingHandler) GetSalesInvoices(c *gin.Context) {
	status := c.Query("status")
	paymentStatus := c.Query("payment_status")
	customerID := c.Query("customer_id")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	query := "SELECT * FROM sales_invoice"
	params := make(map[string]interface{})

	var conditions []string
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}
	if paymentStatus != "" {
		conditions = append(conditions, "payment_status = $payment_status")
		params["payment_status"] = paymentStatus
	}
	if customerID != "" {
		conditions = append(conditions, "customer_id = $customer_id")
		params["customer_id"] = customerID
	}
	if fromDate != "" {
		conditions = append(conditions, "invoice_date >= $from_date")
		params["from_date"] = fromDate
	}
	if toDate != "" {
		conditions = append(conditions, "invoice_date <= $to_date")
		params["to_date"] = toDate
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY invoice_date DESC, invoice_number DESC"

	invoices, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get sales invoices", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sales invoices"})
		return
	}

	c.JSON(http.StatusOK, invoices)
}

// GetSalesInvoice retrieves a sales invoice by ID
func (h *AccountingHandler) GetSalesInvoice(c *gin.Context) {
	invoiceID := c.Param("id")

	invoices, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": invoiceID},
	)
	if err != nil {
		logger.LogError("Sales invoice not found", err, map[string]interface{}{"invoice_id": invoiceID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Sales invoice not found"})
		return
	}

	if len(invoices) > 0 {
		c.JSON(http.StatusOK, invoices[0])
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sales invoice not found"})
	}
}

// UpdateSalesInvoice updates a sales invoice (draft only)
func (h *AccountingHandler) UpdateSalesInvoice(c *gin.Context) {
	invoiceID := c.Param("id")

	var invoice models.SalesInvoice
	if err := c.ShouldBindJSON(&invoice); err != nil {
		logger.LogError("Failed to bind invoice data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if invoice is draft
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": invoiceID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sales invoice not found"})
		return
	}

	existingInvoice, ok := existing[0].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid invoice data"})
		return
	}
	if existingInvoice["status"] != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft invoices can be updated"})
		return
	}

	invoice.UpdatedAt = time.Now()

	query := `UPDATE $id SET
		customer_id = $customer_id,
		customer_name = $customer_name,
		invoice_date = $invoice_date,
		due_date = $due_date,
		total_amount = $total_amount,
		tax_amount = $tax_amount,
		discount_amount = $discount_amount,
		status = $status,
		description = $description,
		reference = $reference,
		updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":               invoiceID,
		"customer_id":      invoice.CustomerID,
		"customer_name":    invoice.CustomerName,
		"invoice_date":     invoice.InvoiceDate,
		"due_date":         invoice.DueDate,
		"total_amount":     invoice.TotalAmount,
		"tax_amount":       invoice.TaxAmount,
		"discount_amount":  invoice.DiscountAmount,
		"status":           invoice.Status,
		"description":      invoice.Description,
		"reference":        invoice.Reference,
		"updated_at":       invoice.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to update sales invoice", err, map[string]interface{}{"invoice_id": invoiceID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sales invoice"})
		return
	}

	logger.Log.Info().
		Str("invoice_id", invoiceID).
		Msg("Sales invoice updated successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Sales invoice updated successfully"})
	}
}

// DeleteSalesInvoice deletes a draft sales invoice
func (h *AccountingHandler) DeleteSalesInvoice(c *gin.Context) {
	invoiceID := c.Param("id")

	// Check if invoice is draft
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": invoiceID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sales invoice not found"})
		return
	}

	existingInvoice, ok := existing[0].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid invoice data"})
		return
	}
	if existingInvoice["status"] != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft invoices can be deleted"})
		return
	}

	_, err = h.db.QuerySlice(
		"DELETE $id",
		map[string]interface{}{"id": invoiceID},
	)
	if err != nil {
		logger.LogError("Failed to delete sales invoice", err, map[string]interface{}{"invoice_id": invoiceID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sales invoice"})
		return
	}

	logger.Log.Info().
		Str("invoice_id", invoiceID).
		Msg("Sales invoice deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Sales invoice deleted successfully"})
}

// ApproveSalesInvoice approves a submitted sales invoice
func (h *AccountingHandler) ApproveSalesInvoice(c *gin.Context) {
	invoiceID := c.Param("id")

	// Check if invoice is submitted
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": invoiceID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sales invoice not found"})
		return
	}

	existingInvoice, ok := existing[0].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid invoice data"})
		return
	}
	if existingInvoice["status"] != "submitted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only submitted invoices can be approved"})
		return
	}

	query := `UPDATE $id SET status = 'approved', updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         invoiceID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to approve sales invoice", err, map[string]interface{}{"invoice_id": invoiceID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve sales invoice"})
		return
	}

	logger.Log.Info().
		Str("invoice_id", invoiceID).
		Msg("Sales invoice approved successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Sales invoice approved successfully"})
	}
}

// CancelSalesInvoice cancels a sales invoice
func (h *AccountingHandler) CancelSalesInvoice(c *gin.Context) {
	invoiceID := c.Param("id")

	// Check if invoice exists
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": invoiceID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sales invoice not found"})
		return
	}

	existingInvoice, ok := existing[0].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid invoice data"})
		return
	}
	if existingInvoice["status"] == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice is already cancelled"})
		return
	}
	if existingInvoice["status"] == "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paid invoices cannot be cancelled"})
		return
	}

	query := `UPDATE $id SET status = 'cancelled', payment_status = 'unpaid', updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         invoiceID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to cancel sales invoice", err, map[string]interface{}{"invoice_id": invoiceID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel sales invoice"})
		return
	}

	logger.Log.Info().
		Str("invoice_id", invoiceID).
		Msg("Sales invoice cancelled successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Sales invoice cancelled successfully"})
	}
}
