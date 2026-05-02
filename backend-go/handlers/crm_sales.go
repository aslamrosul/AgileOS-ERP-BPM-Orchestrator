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
// QUOTATION HANDLERS
// ============================================

// CreateQuotation creates a new quotation
func (h *CRMHandler) CreateQuotation(c *gin.Context) {
	var quotation models.Quotation
	if err := c.ShouldBindJSON(&quotation); err != nil {
		logger.LogError("Failed to bind quotation data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate quotation number
	year := time.Now().Year()
	quotations, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT quotation_number FROM quotation WHERE quotation_number LIKE 'QUO-%d-%%' ORDER BY quotation_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last quotation number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate quotation number"})
		return
	}

	quotationNumber := fmt.Sprintf("QUO-%d-0001", year)
	if len(quotations) > 0 {
		lastNumber := quotations[0].(map[string]interface{})["quotation_number"].(string)
		var lastNum int
		fmt.Sscanf(lastNumber, fmt.Sprintf("QUO-%d-%%d", year), &lastNum)
		quotationNumber = fmt.Sprintf("QUO-%d-%04d", year, lastNum+1)
	}

	quotation.QuotationNumber = quotationNumber
	quotation.CreatedBy = userID.(string)
	quotation.CreatedAt = time.Now()
	quotation.UpdatedAt = time.Now()
	quotation.Status = models.QuotationStatusDraft

	// Calculate totals
	quotation.SubTotal = decimal.Zero
	for _, line := range quotation.Lines {
		quotation.SubTotal = quotation.SubTotal.Add(line.LineTotal)
	}
	quotation.TotalAmount = quotation.SubTotal.Add(quotation.TaxAmount).Sub(quotation.DiscountAmount)

	query := `CREATE quotation CONTENT {
		quotation_number: $quotation_number,
		customer_id: $customer_id,
		customer_name: $customer_name,
		contact_id: $contact_id,
		contact_name: $contact_name,
		quotation_date: $quotation_date,
		valid_until: $valid_until,
		status: $status,
		sub_total: $sub_total,
		tax_amount: $tax_amount,
		discount_amount: $discount_amount,
		total_amount: $total_amount,
		currency: $currency,
		payment_terms: $payment_terms,
		delivery_terms: $delivery_terms,
		notes: $notes,
		terms_conditions: $terms_conditions,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"quotation_number": quotation.QuotationNumber,
		"customer_id":      quotation.CustomerID,
		"customer_name":    quotation.CustomerName,
		"contact_id":       quotation.ContactID,
		"contact_name":     quotation.ContactName,
		"quotation_date":   quotation.QuotationDate,
		"valid_until":      quotation.ValidUntil,
		"status":           quotation.Status,
		"sub_total":        quotation.SubTotal,
		"tax_amount":       quotation.TaxAmount,
		"discount_amount":  quotation.DiscountAmount,
		"total_amount":     quotation.TotalAmount,
		"currency":         quotation.Currency,
		"payment_terms":    quotation.PaymentTerms,
		"delivery_terms":   quotation.DeliveryTerms,
		"notes":            quotation.Notes,
		"terms_conditions": quotation.TermsConditions,
		"created_by":       quotation.CreatedBy,
		"created_at":       quotation.CreatedAt,
		"updated_at":       quotation.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create quotation", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create quotation"})
		return
	}

	logger.Log.Info().Str("quotation_number", quotation.QuotationNumber).Msg("Quotation created successfully")
	c.JSON(http.StatusCreated, result[0])
}

// GetQuotations retrieves all quotations with filters
func (h *CRMHandler) GetQuotations(c *gin.Context) {
	status := c.Query("status")
	customerID := c.Query("customer_id")

	query := "SELECT * FROM quotation"
	params := make(map[string]interface{})

	var conditions []string
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}
	if customerID != "" {
		conditions = append(conditions, "customer_id = $customer_id")
		params["customer_id"] = customerID
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY quotation_date DESC"

	quotations, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get quotations", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve quotations"})
		return
	}

	c.JSON(http.StatusOK, quotations)
}

// GetQuotation retrieves a quotation by ID
func (h *CRMHandler) GetQuotation(c *gin.Context) {
	quotationID := c.Param("id")

	quotations, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": quotationID},
	)
	if err != nil || len(quotations) == 0 {
		logger.LogError("Quotation not found", err, map[string]interface{}{"quotation_id": quotationID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Quotation not found"})
		return
	}

	c.JSON(http.StatusOK, quotations[0])
}

// SendQuotation sends quotation to customer
func (h *CRMHandler) SendQuotation(c *gin.Context) {
	quotationID := c.Param("id")

	query := `UPDATE $id SET 
		status = 'sent',
		sent_at = $sent_at,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":         quotationID,
		"sent_at":    time.Now(),
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to send quotation", err, map[string]interface{}{"quotation_id": quotationID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send quotation"})
		return
	}

	logger.Log.Info().Str("quotation_id", quotationID).Msg("Quotation sent successfully")
	c.JSON(http.StatusOK, result[0])
}

// AcceptQuotation accepts a quotation
func (h *CRMHandler) AcceptQuotation(c *gin.Context) {
	quotationID := c.Param("id")

	query := `UPDATE $id SET 
		status = 'accepted',
		accepted_at = $accepted_at,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":          quotationID,
		"accepted_at": time.Now(),
		"updated_at":  time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to accept quotation", err, map[string]interface{}{"quotation_id": quotationID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept quotation"})
		return
	}

	logger.Log.Info().Str("quotation_id", quotationID).Msg("Quotation accepted successfully")
	c.JSON(http.StatusOK, result[0])
}

// ============================================
// SALES ORDER HANDLERS
// ============================================

// CreateSalesOrder creates a new sales order
func (h *CRMHandler) CreateSalesOrder(c *gin.Context) {
	var salesOrder models.SalesOrder
	if err := c.ShouldBindJSON(&salesOrder); err != nil {
		logger.LogError("Failed to bind sales order data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate sales order number
	year := time.Now().Year()
	orders, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT order_number FROM sales_order WHERE order_number LIKE 'SO-%d-%%' ORDER BY order_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last sales order number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate sales order number"})
		return
	}

	orderNumber := fmt.Sprintf("SO-%d-0001", year)
	if len(orders) > 0 {
		lastNumber := orders[0].(map[string]interface{})["order_number"].(string)
		var lastNum int
		fmt.Sscanf(lastNumber, fmt.Sprintf("SO-%d-%%d", year), &lastNum)
		orderNumber = fmt.Sprintf("SO-%d-%04d", year, lastNum+1)
	}

	salesOrder.OrderNumber = orderNumber
	salesOrder.CreatedBy = userID.(string)
	salesOrder.CreatedAt = time.Now()
	salesOrder.UpdatedAt = time.Now()
	salesOrder.Status = models.SOStatusDraft

	// Calculate totals
	salesOrder.SubTotal = decimal.Zero
	for _, line := range salesOrder.Lines {
		salesOrder.SubTotal = salesOrder.SubTotal.Add(line.LineTotal)
	}
	salesOrder.TotalAmount = salesOrder.SubTotal.Add(salesOrder.TaxAmount).Sub(salesOrder.DiscountAmount)

	query := `CREATE sales_order CONTENT {
		order_number: $order_number,
		quotation_id: $quotation_id,
		quotation_number: $quotation_number,
		customer_id: $customer_id,
		customer_name: $customer_name,
		contact_id: $contact_id,
		contact_name: $contact_name,
		order_date: $order_date,
		expected_delivery_date: $expected_delivery_date,
		status: $status,
		sub_total: $sub_total,
		tax_amount: $tax_amount,
		discount_amount: $discount_amount,
		total_amount: $total_amount,
		currency: $currency,
		payment_terms: $payment_terms,
		delivery_address: $delivery_address,
		delivery_notes: $delivery_notes,
		notes: $notes,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"order_number":           salesOrder.OrderNumber,
		"quotation_id":           salesOrder.QuotationID,
		"quotation_number":       salesOrder.QuotationNumber,
		"customer_id":            salesOrder.CustomerID,
		"customer_name":          salesOrder.CustomerName,
		"contact_id":             salesOrder.ContactID,
		"contact_name":           salesOrder.ContactName,
		"order_date":             salesOrder.OrderDate,
		"expected_delivery_date": salesOrder.ExpectedDeliveryDate,
		"status":                 salesOrder.Status,
		"sub_total":              salesOrder.SubTotal,
		"tax_amount":             salesOrder.TaxAmount,
		"discount_amount":        salesOrder.DiscountAmount,
		"total_amount":           salesOrder.TotalAmount,
		"currency":               salesOrder.Currency,
		"payment_terms":          salesOrder.PaymentTerms,
		"delivery_address":       salesOrder.DeliveryAddress,
		"delivery_notes":         salesOrder.DeliveryNotes,
		"notes":                  salesOrder.Notes,
		"created_by":             salesOrder.CreatedBy,
		"created_at":             salesOrder.CreatedAt,
		"updated_at":             salesOrder.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create sales order", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sales order"})
		return
	}

	logger.Log.Info().Str("order_number", salesOrder.OrderNumber).Msg("Sales order created successfully")
	c.JSON(http.StatusCreated, result[0])
}

// GetSalesOrders retrieves all sales orders with filters
func (h *CRMHandler) GetSalesOrders(c *gin.Context) {
	status := c.Query("status")
	customerID := c.Query("customer_id")

	query := "SELECT * FROM sales_order"
	params := make(map[string]interface{})

	var conditions []string
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}
	if customerID != "" {
		conditions = append(conditions, "customer_id = $customer_id")
		params["customer_id"] = customerID
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY order_date DESC"

	orders, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get sales orders", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sales orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetSalesOrder retrieves a sales order by ID
func (h *CRMHandler) GetSalesOrder(c *gin.Context) {
	orderID := c.Param("id")

	orders, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": orderID},
	)
	if err != nil || len(orders) == 0 {
		logger.LogError("Sales order not found", err, map[string]interface{}{"order_id": orderID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Sales order not found"})
		return
	}

	c.JSON(http.StatusOK, orders[0])
}

// ConfirmSalesOrder confirms a sales order
func (h *CRMHandler) ConfirmSalesOrder(c *gin.Context) {
	orderID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	query := `UPDATE $id SET 
		status = 'confirmed',
		confirmed_by = $confirmed_by,
		confirmed_at = $confirmed_at,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":           orderID,
		"confirmed_by": userID.(string),
		"confirmed_at": time.Now(),
		"updated_at":   time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to confirm sales order", err, map[string]interface{}{"order_id": orderID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm sales order"})
		return
	}

	logger.Log.Info().Str("order_id", orderID).Msg("Sales order confirmed successfully")
	c.JSON(http.StatusOK, result[0])
}

// DeliverSalesOrder marks sales order as delivered
func (h *CRMHandler) DeliverSalesOrder(c *gin.Context) {
	orderID := c.Param("id")

	query := `UPDATE $id SET 
		status = 'delivered',
		actual_delivery_date = $actual_delivery_date,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":                   orderID,
		"actual_delivery_date": time.Now(),
		"updated_at":           time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to deliver sales order", err, map[string]interface{}{"order_id": orderID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deliver sales order"})
		return
	}

	logger.Log.Info().Str("order_id", orderID).Msg("Sales order delivered successfully")
	c.JSON(http.StatusOK, result[0])
}

// CancelSalesOrder cancels a sales order
func (h *CRMHandler) CancelSalesOrder(c *gin.Context) {
	orderID := c.Param("id")

	query := `UPDATE $id SET status = 'cancelled', updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         orderID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to cancel sales order", err, map[string]interface{}{"order_id": orderID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel sales order"})
		return
	}

	logger.Log.Info().Str("order_id", orderID).Msg("Sales order cancelled successfully")
	c.JSON(http.StatusOK, result[0])
}
