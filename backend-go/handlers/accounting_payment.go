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
// PAYMENT HANDLERS
// ============================================

// CreatePayment creates a new payment (vendor payment or customer receipt)
func (h *AccountingHandler) CreatePayment(c *gin.Context) {
	var payment models.Payment
	if err := c.ShouldBindJSON(&payment); err != nil {
		logger.LogError("Failed to bind payment data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate payment number
	year := time.Now().Year()
	payments, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT payment_number FROM payment WHERE payment_number LIKE 'PAY-%d-%%' ORDER BY payment_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last payment number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate payment number"})
		return
	}

	paymentNumber := fmt.Sprintf("PAY-%d-0001", year)
	if len(payments) > 0 {
		if paymentMap, ok := payments[0].(map[string]interface{}); ok {
			if lastNumber, ok := paymentMap["payment_number"].(string); ok {
				var lastNum int
				fmt.Sscanf(lastNumber, fmt.Sprintf("PAY-%d-%%d", year), &lastNum)
				paymentNumber = fmt.Sprintf("PAY-%d-%04d", year, lastNum+1)
			}
		}
	}

	payment.PaymentNumber = paymentNumber
	payment.CreatedBy = userID.(string)
	payment.CreatedAt = time.Now()
	payment.UpdatedAt = time.Now()

	if payment.Status == "" {
		payment.Status = models.PaymentStatusEnumDraft
	}

	query := `CREATE payment CONTENT {
		payment_number: $payment_number,
		payment_type: $payment_type,
		party_id: $party_id,
		party_name: $party_name,
		payment_date: $payment_date,
		payment_method: $payment_method,
		amount: $amount,
		bank_account: $bank_account,
		reference_number: $reference_number,
		status: $status,
		description: $description,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"payment_number":   payment.PaymentNumber,
		"payment_type":     payment.PaymentType,
		"party_id":         payment.PartyID,
		"party_name":       payment.PartyName,
		"payment_date":     payment.PaymentDate,
		"payment_method":   payment.PaymentMethod,
		"amount":           payment.Amount,
		"bank_account":     payment.BankAccount,
		"reference_number": payment.ReferenceNumber,
		"status":           payment.Status,
		"description":      payment.Description,
		"created_by":       payment.CreatedBy,
		"created_at":       payment.CreatedAt,
		"updated_at":       payment.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create payment", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
		return
	}

	logger.Log.Info().
		Str("payment_number", payment.PaymentNumber).
		Str("payment_type", string(payment.PaymentType)).
		Msg("Payment created successfully")

	if len(result) > 0 {
		c.JSON(http.StatusCreated, result[0])
	} else {
		c.JSON(http.StatusCreated, gin.H{"message": "Payment created successfully"})
	}
}

// GetPayments retrieves all payments with filters
func (h *AccountingHandler) GetPayments(c *gin.Context) {
	paymentType := c.Query("payment_type")
	status := c.Query("status")
	paymentMethod := c.Query("payment_method")
	partyID := c.Query("party_id")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	query := "SELECT * FROM payment"
	params := make(map[string]interface{})

	var conditions []string
	if paymentType != "" {
		conditions = append(conditions, "payment_type = $payment_type")
		params["payment_type"] = paymentType
	}
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}
	if paymentMethod != "" {
		conditions = append(conditions, "payment_method = $payment_method")
		params["payment_method"] = paymentMethod
	}
	if partyID != "" {
		conditions = append(conditions, "party_id = $party_id")
		params["party_id"] = partyID
	}
	if fromDate != "" {
		conditions = append(conditions, "payment_date >= $from_date")
		params["from_date"] = fromDate
	}
	if toDate != "" {
		conditions = append(conditions, "payment_date <= $to_date")
		params["to_date"] = toDate
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY payment_date DESC, payment_number DESC"

	payments, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get payments", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payments"})
		return
	}

	c.JSON(http.StatusOK, payments)
}

// GetPayment retrieves a payment by ID
func (h *AccountingHandler) GetPayment(c *gin.Context) {
	paymentID := c.Param("id")

	payments, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": paymentID},
	)
	if err != nil || len(payments) == 0 {
		logger.LogError("Payment not found", err, map[string]interface{}{"payment_id": paymentID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	c.JSON(http.StatusOK, payments[0])
}

// UpdatePayment updates a payment (draft only)
func (h *AccountingHandler) UpdatePayment(c *gin.Context) {
	paymentID := c.Param("id")

	var payment models.Payment
	if err := c.ShouldBindJSON(&payment); err != nil {
		logger.LogError("Failed to bind payment data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if payment is draft
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": paymentID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	existingPayment, ok := existing[0].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid payment data"})
		return
	}
	if existingPayment["status"] != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft payments can be updated"})
		return
	}

	payment.UpdatedAt = time.Now()

	query := `UPDATE $id SET
		payment_type = $payment_type,
		party_id = $party_id,
		party_name = $party_name,
		payment_date = $payment_date,
		payment_method = $payment_method,
		amount = $amount,
		bank_account = $bank_account,
		reference_number = $reference_number,
		status = $status,
		description = $description,
		updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":               paymentID,
		"payment_type":     payment.PaymentType,
		"party_id":         payment.PartyID,
		"party_name":       payment.PartyName,
		"payment_date":     payment.PaymentDate,
		"payment_method":   payment.PaymentMethod,
		"amount":           payment.Amount,
		"bank_account":     payment.BankAccount,
		"reference_number": payment.ReferenceNumber,
		"status":           payment.Status,
		"description":      payment.Description,
		"updated_at":       payment.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to update payment", err, map[string]interface{}{"payment_id": paymentID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment"})
		return
	}

	logger.Log.Info().
		Str("payment_id", paymentID).
		Msg("Payment updated successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Payment updated successfully"})
	}
}

// DeletePayment deletes a draft payment
func (h *AccountingHandler) DeletePayment(c *gin.Context) {
	paymentID := c.Param("id")

	// Check if payment is draft
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": paymentID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	existingPayment, ok := existing[0].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid payment data"})
		return
	}
	if existingPayment["status"] != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft payments can be deleted"})
		return
	}

	_, err = h.db.QuerySlice(
		"DELETE $id",
		map[string]interface{}{"id": paymentID},
	)
	if err != nil {
		logger.LogError("Failed to delete payment", err, map[string]interface{}{"payment_id": paymentID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete payment"})
		return
	}

	logger.Log.Info().
		Str("payment_id", paymentID).
		Msg("Payment deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Payment deleted successfully"})
}

// ClearPayment marks a payment as cleared
func (h *AccountingHandler) ClearPayment(c *gin.Context) {
	paymentID := c.Param("id")

	// Check if payment is submitted
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": paymentID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	existingPayment, ok := existing[0].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid payment data"})
		return
	}
	if existingPayment["status"] != "submitted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only submitted payments can be cleared"})
		return
	}

	query := `UPDATE $id SET status = 'cleared', updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         paymentID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to clear payment", err, map[string]interface{}{"payment_id": paymentID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear payment"})
		return
	}

	logger.Log.Info().
		Str("payment_id", paymentID).
		Msg("Payment cleared successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Payment cleared successfully"})
	}
}

// CancelPayment cancels a payment
func (h *AccountingHandler) CancelPayment(c *gin.Context) {
	paymentID := c.Param("id")

	// Check if payment exists
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": paymentID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	existingPayment, ok := existing[0].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid payment data"})
		return
	}
	if existingPayment["status"] == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment is already cancelled"})
		return
	}
	if existingPayment["status"] == "cleared" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cleared payments cannot be cancelled"})
		return
	}

	query := `UPDATE $id SET status = 'cancelled', updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         paymentID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to cancel payment", err, map[string]interface{}{"payment_id": paymentID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel payment"})
		return
	}

	logger.Log.Info().
		Str("payment_id", paymentID).
		Msg("Payment cancelled successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Payment cancelled successfully"})
	}
}
