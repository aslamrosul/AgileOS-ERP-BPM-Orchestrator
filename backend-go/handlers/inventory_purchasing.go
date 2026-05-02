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
// PURCHASE REQUISITION HANDLERS
// ============================================

// CreatePurchaseRequisition creates a new purchase requisition
func (h *InventoryHandler) CreatePurchaseRequisition(c *gin.Context) {
	var pr models.PurchaseRequisition
	if err := c.ShouldBindJSON(&pr); err != nil {
		logger.LogError("Failed to bind PR data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate PR number
	year := time.Now().Year()
	prs, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT pr_number FROM purchase_requisition WHERE pr_number LIKE 'PR-%d-%%' ORDER BY pr_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last PR number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PR number"})
		return
	}

	prNumber := fmt.Sprintf("PR-%d-0001", year)
	if len(prs) > 0 {
		lastNumber := prs[0].(map[string]interface{})["pr_number"].(string)
		var lastNum int
		fmt.Sscanf(lastNumber, fmt.Sprintf("PR-%d-%%d", year), &lastNum)
		prNumber = fmt.Sprintf("PR-%d-%04d", year, lastNum+1)
	}

	pr.PRNumber = prNumber
	pr.RequestedBy = userID.(string)
	pr.Status = models.PRStatusDraft
	pr.CreatedAt = time.Now()
	pr.UpdatedAt = time.Now()

	query := `CREATE purchase_requisition CONTENT {
		pr_number: $pr_number,
		request_date: $request_date,
		required_date: $required_date,
		department: $department,
		requested_by: $requested_by,
		requested_by_name: $requested_by_name,
		status: $status,
		notes: $notes,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"pr_number":           pr.PRNumber,
		"request_date":        pr.RequestDate,
		"required_date":       pr.RequiredDate,
		"department":          pr.Department,
		"requested_by":        pr.RequestedBy,
		"requested_by_name":   pr.RequestedByName,
		"status":              pr.Status,
		"notes":               pr.Notes,
		"created_at":          pr.CreatedAt,
		"updated_at":          pr.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create PR", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase requisition"})
		return
	}

	logger.Log.Info().Str("pr_number", pr.PRNumber).Msg("Purchase requisition created successfully")
	c.JSON(http.StatusCreated, result[0])
}

// GetPurchaseRequisitions retrieves all PRs with filters
func (h *InventoryHandler) GetPurchaseRequisitions(c *gin.Context) {
	status := c.Query("status")
	department := c.Query("department")

	query := "SELECT * FROM purchase_requisition"
	params := make(map[string]interface{})

	var conditions []string
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}
	if department != "" {
		conditions = append(conditions, "department = $department")
		params["department"] = department
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY request_date DESC"

	prs, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get PRs", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve purchase requisitions"})
		return
	}

	c.JSON(http.StatusOK, prs)
}

// ApprovePurchaseRequisition approves a PR
func (h *InventoryHandler) ApprovePurchaseRequisition(c *gin.Context) {
	prID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	query := `UPDATE $id SET 
		status = 'approved',
		approved_by = $approved_by,
		approved_at = $approved_at,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":          prID,
		"approved_by": userID.(string),
		"approved_at": time.Now(),
		"updated_at":  time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to approve PR", err, map[string]interface{}{"pr_id": prID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve purchase requisition"})
		return
	}

	logger.Log.Info().Str("pr_id", prID).Msg("Purchase requisition approved successfully")
	c.JSON(http.StatusOK, result[0])
}

// ============================================
// PURCHASE ORDER HANDLERS
// ============================================

// CreatePurchaseOrder creates a new purchase order
func (h *InventoryHandler) CreatePurchaseOrder(c *gin.Context) {
	var po models.PurchaseOrder
	if err := c.ShouldBindJSON(&po); err != nil {
		logger.LogError("Failed to bind PO data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate PO number
	year := time.Now().Year()
	pos, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT po_number FROM purchase_order WHERE po_number LIKE 'PO-%d-%%' ORDER BY po_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last PO number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PO number"})
		return
	}

	poNumber := fmt.Sprintf("PO-%d-0001", year)
	if len(pos) > 0 {
		lastNumber := pos[0].(map[string]interface{})["po_number"].(string)
		var lastNum int
		fmt.Sscanf(lastNumber, fmt.Sprintf("PO-%d-%%d", year), &lastNum)
		poNumber = fmt.Sprintf("PO-%d-%04d", year, lastNum+1)
	}

	po.PONumber = poNumber
	po.CreatedBy = userID.(string)
	po.Status = models.POStatusDraft
	po.CreatedAt = time.Now()
	po.UpdatedAt = time.Now()

	// Calculate totals
	po.SubTotal = decimal.Zero
	for _, line := range po.Lines {
		po.SubTotal = po.SubTotal.Add(line.LineTotal)
	}
	po.TotalAmount = po.SubTotal.Add(po.TaxAmount).Sub(po.DiscountAmount)

	query := `CREATE purchase_order CONTENT {
		po_number: $po_number,
		vendor_id: $vendor_id,
		vendor_name: $vendor_name,
		order_date: $order_date,
		expected_date: $expected_date,
		status: $status,
		sub_total: $sub_total,
		tax_amount: $tax_amount,
		discount_amount: $discount_amount,
		total_amount: $total_amount,
		payment_terms: $payment_terms,
		delivery_address: $delivery_address,
		notes: $notes,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"po_number":        po.PONumber,
		"vendor_id":        po.VendorID,
		"vendor_name":      po.VendorName,
		"order_date":       po.OrderDate,
		"expected_date":    po.ExpectedDate,
		"status":           po.Status,
		"sub_total":        po.SubTotal,
		"tax_amount":       po.TaxAmount,
		"discount_amount":  po.DiscountAmount,
		"total_amount":     po.TotalAmount,
		"payment_terms":    po.PaymentTerms,
		"delivery_address": po.DeliveryAddress,
		"notes":            po.Notes,
		"created_by":       po.CreatedBy,
		"created_at":       po.CreatedAt,
		"updated_at":       po.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create PO", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase order"})
		return
	}

	logger.Log.Info().Str("po_number", po.PONumber).Msg("Purchase order created successfully")
	c.JSON(http.StatusCreated, result[0])
}

// GetPurchaseOrders retrieves all POs with filters
func (h *InventoryHandler) GetPurchaseOrders(c *gin.Context) {
	status := c.Query("status")
	vendorID := c.Query("vendor_id")

	query := "SELECT * FROM purchase_order"
	params := make(map[string]interface{})

	var conditions []string
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}
	if vendorID != "" {
		conditions = append(conditions, "vendor_id = $vendor_id")
		params["vendor_id"] = vendorID
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY order_date DESC"

	pos, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get POs", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve purchase orders"})
		return
	}

	c.JSON(http.StatusOK, pos)
}

// GetPurchaseOrder retrieves a PO by ID
func (h *InventoryHandler) GetPurchaseOrder(c *gin.Context) {
	poID := c.Param("id")

	pos, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": poID},
	)
	if err != nil || len(pos) == 0 {
		logger.LogError("PO not found", err, map[string]interface{}{"po_id": poID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase order not found"})
		return
	}

	c.JSON(http.StatusOK, pos[0])
}

// ApprovePurchaseOrder approves a PO
func (h *InventoryHandler) ApprovePurchaseOrder(c *gin.Context) {
	poID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	query := `UPDATE $id SET 
		status = 'confirmed',
		approved_by = $approved_by,
		approved_at = $approved_at,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":          poID,
		"approved_by": userID.(string),
		"approved_at": time.Now(),
		"updated_at":  time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to approve PO", err, map[string]interface{}{"po_id": poID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve purchase order"})
		return
	}

	logger.Log.Info().Str("po_id", poID).Msg("Purchase order approved successfully")
	c.JSON(http.StatusOK, result[0])
}

// ============================================
// GOODS RECEIPT HANDLERS
// ============================================

// CreateGoodsReceipt creates a new goods receipt
func (h *InventoryHandler) CreateGoodsReceipt(c *gin.Context) {
	var gr models.GoodsReceipt
	if err := c.ShouldBindJSON(&gr); err != nil {
		logger.LogError("Failed to bind GR data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate GR number
	year := time.Now().Year()
	grs, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT gr_number FROM goods_receipt WHERE gr_number LIKE 'GR-%d-%%' ORDER BY gr_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last GR number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate GR number"})
		return
	}

	grNumber := fmt.Sprintf("GR-%d-0001", year)
	if len(grs) > 0 {
		lastNumber := grs[0].(map[string]interface{})["gr_number"].(string)
		var lastNum int
		fmt.Sscanf(lastNumber, fmt.Sprintf("GR-%d-%%d", year), &lastNum)
		grNumber = fmt.Sprintf("GR-%d-%04d", year, lastNum+1)
	}

	gr.GRNumber = grNumber
	gr.ReceivedBy = userID.(string)
	gr.Status = models.GRStatusDraft
	gr.CreatedAt = time.Now()
	gr.UpdatedAt = time.Now()

	query := `CREATE goods_receipt CONTENT {
		gr_number: $gr_number,
		po_number: $po_number,
		po_id: $po_id,
		vendor_id: $vendor_id,
		vendor_name: $vendor_name,
		receipt_date: $receipt_date,
		warehouse_id: $warehouse_id,
		warehouse_name: $warehouse_name,
		status: $status,
		notes: $notes,
		received_by: $received_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"gr_number":      gr.GRNumber,
		"po_number":      gr.PONumber,
		"po_id":          gr.POID,
		"vendor_id":      gr.VendorID,
		"vendor_name":    gr.VendorName,
		"receipt_date":   gr.ReceiptDate,
		"warehouse_id":   gr.WarehouseID,
		"warehouse_name": gr.WarehouseName,
		"status":         gr.Status,
		"notes":          gr.Notes,
		"received_by":    gr.ReceivedBy,
		"created_at":     gr.CreatedAt,
		"updated_at":     gr.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create GR", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create goods receipt"})
		return
	}

	logger.Log.Info().Str("gr_number", gr.GRNumber).Msg("Goods receipt created successfully")
	c.JSON(http.StatusCreated, result[0])
}

// GetGoodsReceipts retrieves all GRs with filters
func (h *InventoryHandler) GetGoodsReceipts(c *gin.Context) {
	status := c.Query("status")
	poID := c.Query("po_id")

	query := "SELECT * FROM goods_receipt"
	params := make(map[string]interface{})

	var conditions []string
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}
	if poID != "" {
		conditions = append(conditions, "po_id = $po_id")
		params["po_id"] = poID
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY receipt_date DESC"

	grs, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get GRs", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve goods receipts"})
		return
	}

	c.JSON(http.StatusOK, grs)
}

// ConfirmGoodsReceipt confirms a GR and updates stock
func (h *InventoryHandler) ConfirmGoodsReceipt(c *gin.Context) {
	grID := c.Param("id")

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
		"id":           grID,
		"confirmed_by": userID.(string),
		"confirmed_at": time.Now(),
		"updated_at":   time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to confirm GR", err, map[string]interface{}{"gr_id": grID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm goods receipt"})
		return
	}

	logger.Log.Info().Str("gr_id", grID).Msg("Goods receipt confirmed successfully")
	c.JSON(http.StatusOK, result[0])
}
