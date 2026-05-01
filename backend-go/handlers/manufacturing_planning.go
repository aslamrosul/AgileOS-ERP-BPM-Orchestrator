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
// PRODUCTION PLANNING HANDLERS
// ============================================

// CreateProductionPlan creates a new production plan
func (h *ManufacturingHandler) CreateProductionPlan(c *gin.Context) {
	var plan models.ProductionPlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		logger.LogError("Failed to bind production plan data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate plan number
	year := time.Now().Year()
	plans, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT plan_number FROM production_plan WHERE plan_number LIKE 'PP-%d-%%' ORDER BY plan_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last plan number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate plan number"})
		return
	}

	planNumber := fmt.Sprintf("PP-%d-0001", year)
	if len(plans) > 0 {
		lastNumber := plans[0].(map[string]interface{})["plan_number"].(string)
		var lastNum int
		fmt.Sscanf(lastNumber, fmt.Sprintf("PP-%d-%%d", year), &lastNum)
		planNumber = fmt.Sprintf("PP-%d-%04d", year, lastNum+1)
	}

	plan.PlanNumber = planNumber
	plan.CreatedBy = userID.(string)
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()
	plan.Status = models.PlanStatusDraft
	plan.TotalQuantity = decimal.Zero
	plan.TotalCost = decimal.Zero

	query := `CREATE production_plan CONTENT {
		plan_number: $plan_number,
		plan_name: $plan_name,
		plan_type: $plan_type,
		start_date: $start_date,
		end_date: $end_date,
		status: $status,
		total_quantity: $total_quantity,
		total_cost: $total_cost,
		notes: $notes,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"plan_number":    plan.PlanNumber,
		"plan_name":      plan.PlanName,
		"plan_type":      plan.PlanType,
		"start_date":     plan.StartDate,
		"end_date":       plan.EndDate,
		"status":         plan.Status,
		"total_quantity": plan.TotalQuantity,
		"total_cost":     plan.TotalCost,
		"notes":          plan.Notes,
		"created_by":     plan.CreatedBy,
		"created_at":     plan.CreatedAt,
		"updated_at":     plan.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create production plan", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create production plan"})
		return
	}

	logger.Log.Info().Str("plan_number", plan.PlanNumber).Msg("Production plan created successfully")
	c.JSON(http.StatusCreated, result[0])
}

// GetProductionPlans retrieves all production plans with filters
func (h *ManufacturingHandler) GetProductionPlans(c *gin.Context) {
	status := c.Query("status")
	planType := c.Query("plan_type")

	query := "SELECT * FROM production_plan"
	params := make(map[string]interface{})

	var conditions []string
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}
	if planType != "" {
		conditions = append(conditions, "plan_type = $plan_type")
		params["plan_type"] = planType
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY start_date DESC"

	plans, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get production plans", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve production plans"})
		return
	}

	c.JSON(http.StatusOK, plans)
}

// GetProductionPlan retrieves a production plan by ID
func (h *ManufacturingHandler) GetProductionPlan(c *gin.Context) {
	planID := c.Param("id")

	plans, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": planID},
	)
	if err != nil || len(plans) == 0 {
		logger.LogError("Production plan not found", err, map[string]interface{}{"plan_id": planID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Production plan not found"})
		return
	}

	c.JSON(http.StatusOK, plans[0])
}

// ApproveProductionPlan approves a production plan
func (h *ManufacturingHandler) ApproveProductionPlan(c *gin.Context) {
	planID := c.Param("id")

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
		"id":          planID,
		"approved_by": userID.(string),
		"approved_at": time.Now(),
		"updated_at":  time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to approve production plan", err, map[string]interface{}{"plan_id": planID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve production plan"})
		return
	}

	logger.Log.Info().Str("plan_id", planID).Msg("Production plan approved successfully")
	c.JSON(http.StatusOK, result[0])
}

// ============================================
// PRODUCTION ORDER HANDLERS
// ============================================

// CreateProductionOrder creates a new production order
func (h *ManufacturingHandler) CreateProductionOrder(c *gin.Context) {
	var order models.ProductionOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		logger.LogError("Failed to bind production order data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate order number
	year := time.Now().Year()
	orders, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT order_number FROM production_order WHERE order_number LIKE 'MO-%d-%%' ORDER BY order_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last order number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate order number"})
		return
	}

	orderNumber := fmt.Sprintf("MO-%d-0001", year)
	if len(orders) > 0 {
		lastNumber := orders[0].(map[string]interface{})["order_number"].(string)
		var lastNum int
		fmt.Sscanf(lastNumber, fmt.Sprintf("MO-%d-%%d", year), &lastNum)
		orderNumber = fmt.Sprintf("MO-%d-%04d", year, lastNum+1)
	}

	order.OrderNumber = orderNumber
	order.CreatedBy = userID.(string)
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.Status = models.MOStatusDraft
	order.QuantityProduced = decimal.Zero

	query := `CREATE production_order CONTENT {
		order_number: $order_number,
		product_id: $product_id,
		product_code: $product_code,
		product_name: $product_name,
		bom_id: $bom_id,
		bom_code: $bom_code,
		quantity: $quantity,
		quantity_produced: $quantity_produced,
		unit_of_measure: $unit_of_measure,
		scheduled_start: $scheduled_start,
		scheduled_end: $scheduled_end,
		status: $status,
		priority: $priority,
		source_document: $source_document,
		source_document_id: $source_document_id,
		warehouse_id: $warehouse_id,
		warehouse_name: $warehouse_name,
		notes: $notes,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"order_number":       order.OrderNumber,
		"product_id":         order.ProductID,
		"product_code":       order.ProductCode,
		"product_name":       order.ProductName,
		"bom_id":             order.BOMID,
		"bom_code":           order.BOMCode,
		"quantity":           order.Quantity,
		"quantity_produced":  order.QuantityProduced,
		"unit_of_measure":    order.UnitOfMeasure,
		"scheduled_start":    order.ScheduledStart,
		"scheduled_end":      order.ScheduledEnd,
		"status":             order.Status,
		"priority":           order.Priority,
		"source_document":    order.SourceDocument,
		"source_document_id": order.SourceDocumentID,
		"warehouse_id":       order.WarehouseID,
		"warehouse_name":     order.WarehouseName,
		"notes":              order.Notes,
		"created_by":         order.CreatedBy,
		"created_at":         order.CreatedAt,
		"updated_at":         order.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create production order", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create production order"})
		return
	}

	logger.Log.Info().Str("order_number", order.OrderNumber).Msg("Production order created successfully")
	c.JSON(http.StatusCreated, result[0])
}

// GetProductionOrders retrieves all production orders with filters
func (h *ManufacturingHandler) GetProductionOrders(c *gin.Context) {
	status := c.Query("status")
	productID := c.Query("product_id")

	query := "SELECT * FROM production_order"
	params := make(map[string]interface{})

	var conditions []string
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}
	if productID != "" {
		conditions = append(conditions, "product_id = $product_id")
		params["product_id"] = productID
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY scheduled_start DESC"

	orders, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get production orders", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve production orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetProductionOrder retrieves a production order by ID
func (h *ManufacturingHandler) GetProductionOrder(c *gin.Context) {
	orderID := c.Param("id")

	orders, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": orderID},
	)
	if err != nil || len(orders) == 0 {
		logger.LogError("Production order not found", err, map[string]interface{}{"order_id": orderID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Production order not found"})
		return
	}

	c.JSON(http.StatusOK, orders[0])
}

// StartProductionOrder starts a production order
func (h *ManufacturingHandler) StartProductionOrder(c *gin.Context) {
	orderID := c.Param("id")

	query := `UPDATE $id SET 
		status = 'in_progress',
		actual_start = $actual_start,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":           orderID,
		"actual_start": time.Now(),
		"updated_at":   time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to start production order", err, map[string]interface{}{"order_id": orderID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start production order"})
		return
	}

	logger.Log.Info().Str("order_id", orderID).Msg("Production order started successfully")
	c.JSON(http.StatusOK, result[0])
}

// CompleteProductionOrder completes a production order
func (h *ManufacturingHandler) CompleteProductionOrder(c *gin.Context) {
	orderID := c.Param("id")

	var req struct {
		QuantityProduced decimal.Decimal `json:"quantity_produced"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.LogError("Failed to bind completion data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `UPDATE $id SET 
		status = 'done',
		quantity_produced = $quantity_produced,
		actual_end = $actual_end,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":                 orderID,
		"quantity_produced":  req.QuantityProduced,
		"actual_end":         time.Now(),
		"updated_at":         time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to complete production order", err, map[string]interface{}{"order_id": orderID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete production order"})
		return
	}

	logger.Log.Info().Str("order_id", orderID).Msg("Production order completed successfully")
	c.JSON(http.StatusOK, result[0])
}

// CancelProductionOrder cancels a production order
func (h *ManufacturingHandler) CancelProductionOrder(c *gin.Context) {
	orderID := c.Param("id")

	query := `UPDATE $id SET status = 'cancelled', updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         orderID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to cancel production order", err, map[string]interface{}{"order_id": orderID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel production order"})
		return
	}

	logger.Log.Info().Str("order_id", orderID).Msg("Production order cancelled successfully")
	c.JSON(http.StatusOK, result[0])
}

// GetProductionSchedule retrieves production schedule
func (h *ManufacturingHandler) GetProductionSchedule(c *gin.Context) {
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	query := "SELECT * FROM production_order WHERE status IN ('confirmed', 'in_progress')"
	params := make(map[string]interface{})

	if fromDate != "" && toDate != "" {
		query += " AND scheduled_start >= $from_date AND scheduled_end <= $to_date"
		params["from_date"] = fromDate
		params["to_date"] = toDate
	}

	query += " ORDER BY scheduled_start ASC"

	schedule, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get production schedule", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve production schedule"})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// GetProductionCapacity retrieves production capacity analysis
func (h *ManufacturingHandler) GetProductionCapacity(c *gin.Context) {
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	query := `SELECT 
		COUNT(*) as total_orders,
		SUM(quantity) as total_planned,
		SUM(quantity_produced) as total_produced,
		AVG(quantity_produced / quantity * 100) as efficiency_rate
		FROM production_order 
		WHERE status = 'done'`
	
	params := make(map[string]interface{})

	if fromDate != "" && toDate != "" {
		query += " AND actual_end >= $from_date AND actual_end <= $to_date"
		params["from_date"] = fromDate
		params["to_date"] = toDate
	}

	capacity, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get production capacity", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve production capacity"})
		return
	}

	c.JSON(http.StatusOK, capacity)
}
