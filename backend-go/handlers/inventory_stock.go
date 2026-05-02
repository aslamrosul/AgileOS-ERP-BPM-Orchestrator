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
// STOCK MANAGEMENT HANDLERS
// ============================================

// GetStocks retrieves all stock records with filters
func (h *InventoryHandler) GetStocks(c *gin.Context) {
	productID := c.Query("product_id")
	warehouseID := c.Query("warehouse_id")

	query := "SELECT * FROM stock"
	params := make(map[string]interface{})

	var conditions []string
	if productID != "" {
		conditions = append(conditions, "product_id = $product_id")
		params["product_id"] = productID
	}
	if warehouseID != "" {
		conditions = append(conditions, "warehouse_id = $warehouse_id")
		params["warehouse_id"] = warehouseID
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY product_code ASC"

	stocks, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get stocks", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stocks"})
		return
	}

	c.JSON(http.StatusOK, stocks)
}

// GetStock retrieves a stock record by ID
func (h *InventoryHandler) GetStock(c *gin.Context) {
	stockID := c.Param("id")

	stocks, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": stockID},
	)
	if err != nil || len(stocks) == 0 {
		logger.LogError("Stock not found", err, map[string]interface{}{"stock_id": stockID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock not found"})
		return
	}

	c.JSON(http.StatusOK, stocks[0])
}

// CreateStockMovement creates a new stock movement
func (h *InventoryHandler) CreateStockMovement(c *gin.Context) {
	var movement models.StockMovement
	if err := c.ShouldBindJSON(&movement); err != nil {
		logger.LogError("Failed to bind stock movement data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate movement number
	year := time.Now().Year()
	movements, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT movement_number FROM stock_movement WHERE movement_number LIKE 'SM-%d-%%' ORDER BY movement_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last movement number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate movement number"})
		return
	}

	movementNumber := fmt.Sprintf("SM-%d-0001", year)
	if len(movements) > 0 {
		lastNumber := movements[0].(map[string]interface{})["movement_number"].(string)
		var lastNum int
		fmt.Sscanf(lastNumber, fmt.Sprintf("SM-%d-%%d", year), &lastNum)
		movementNumber = fmt.Sprintf("SM-%d-%04d", year, lastNum+1)
	}

	movement.MovementNumber = movementNumber
	movement.CreatedBy = userID.(string)
	movement.CreatedAt = time.Now()
	movement.MovementDate = time.Now()
	movement.TotalCost = movement.UnitCost.Mul(movement.Quantity)

	query := `CREATE stock_movement CONTENT {
		movement_number: $movement_number,
		movement_type: $movement_type,
		product_id: $product_id,
		product_code: $product_code,
		product_name: $product_name,
		warehouse_id: $warehouse_id,
		warehouse_name: $warehouse_name,
		quantity: $quantity,
		unit_of_measure: $unit_of_measure,
		unit_cost: $unit_cost,
		total_cost: $total_cost,
		reference_type: $reference_type,
		reference_id: $reference_id,
		reference_number: $reference_number,
		notes: $notes,
		movement_date: $movement_date,
		created_by: $created_by,
		created_at: $created_at
	}`

	params := map[string]interface{}{
		"movement_number":  movement.MovementNumber,
		"movement_type":    movement.MovementType,
		"product_id":       movement.ProductID,
		"product_code":     movement.ProductCode,
		"product_name":     movement.ProductName,
		"warehouse_id":     movement.WarehouseID,
		"warehouse_name":   movement.WarehouseName,
		"quantity":         movement.Quantity,
		"unit_of_measure":  movement.UnitOfMeasure,
		"unit_cost":        movement.UnitCost,
		"total_cost":       movement.TotalCost,
		"reference_type":   movement.ReferenceType,
		"reference_id":     movement.ReferenceID,
		"reference_number": movement.ReferenceNumber,
		"notes":            movement.Notes,
		"movement_date":    movement.MovementDate,
		"created_by":       movement.CreatedBy,
		"created_at":       movement.CreatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create stock movement", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stock movement"})
		return
	}

	// Update stock quantity
	h.updateStockQuantity(movement.ProductID, movement.WarehouseID, movement.Quantity, movement.MovementType)

	logger.Log.Info().
		Str("movement_number", movement.MovementNumber).
		Str("movement_type", string(movement.MovementType)).
		Msg("Stock movement created successfully")

	c.JSON(http.StatusCreated, result[0])
}

// GetStockMovements retrieves stock movements with filters
func (h *InventoryHandler) GetStockMovements(c *gin.Context) {
	productID := c.Query("product_id")
	warehouseID := c.Query("warehouse_id")
	movementType := c.Query("movement_type")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	query := "SELECT * FROM stock_movement"
	params := make(map[string]interface{})

	var conditions []string
	if productID != "" {
		conditions = append(conditions, "product_id = $product_id")
		params["product_id"] = productID
	}
	if warehouseID != "" {
		conditions = append(conditions, "warehouse_id = $warehouse_id")
		params["warehouse_id"] = warehouseID
	}
	if movementType != "" {
		conditions = append(conditions, "movement_type = $movement_type")
		params["movement_type"] = movementType
	}
	if fromDate != "" {
		conditions = append(conditions, "movement_date >= $from_date")
		params["from_date"] = fromDate
	}
	if toDate != "" {
		conditions = append(conditions, "movement_date <= $to_date")
		params["to_date"] = toDate
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY movement_date DESC, movement_number DESC"

	movements, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get stock movements", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stock movements"})
		return
	}

	c.JSON(http.StatusOK, movements)
}

// CreateStockAdjustment creates a stock adjustment
func (h *InventoryHandler) CreateStockAdjustment(c *gin.Context) {
	var adjustment models.StockAdjustment
	if err := c.ShouldBindJSON(&adjustment); err != nil {
		logger.LogError("Failed to bind stock adjustment data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate adjustment number
	year := time.Now().Year()
	adjustments, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT adjustment_number FROM stock_adjustment WHERE adjustment_number LIKE 'SA-%d-%%' ORDER BY adjustment_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last adjustment number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate adjustment number"})
		return
	}

	adjustmentNumber := fmt.Sprintf("SA-%d-0001", year)
	if len(adjustments) > 0 {
		lastNumber := adjustments[0].(map[string]interface{})["adjustment_number"].(string)
		var lastNum int
		fmt.Sscanf(lastNumber, fmt.Sprintf("SA-%d-%%d", year), &lastNum)
		adjustmentNumber = fmt.Sprintf("SA-%d-%04d", year, lastNum+1)
	}

	adjustment.AdjustmentNumber = adjustmentNumber
	adjustment.AdjustmentQty = adjustment.NewQuantity.Sub(adjustment.OldQuantity)
	adjustment.CreatedBy = userID.(string)
	adjustment.CreatedAt = time.Now()
	adjustment.AdjustmentDate = time.Now()

	query := `CREATE stock_adjustment CONTENT {
		adjustment_number: $adjustment_number,
		product_id: $product_id,
		product_code: $product_code,
		product_name: $product_name,
		warehouse_id: $warehouse_id,
		warehouse_name: $warehouse_name,
		old_quantity: $old_quantity,
		new_quantity: $new_quantity,
		adjustment_qty: $adjustment_qty,
		reason: $reason,
		notes: $notes,
		adjustment_date: $adjustment_date,
		created_by: $created_by,
		created_at: $created_at
	}`

	params := map[string]interface{}{
		"adjustment_number": adjustment.AdjustmentNumber,
		"product_id":        adjustment.ProductID,
		"product_code":      adjustment.ProductCode,
		"product_name":      adjustment.ProductName,
		"warehouse_id":      adjustment.WarehouseID,
		"warehouse_name":    adjustment.WarehouseName,
		"old_quantity":      adjustment.OldQuantity,
		"new_quantity":      adjustment.NewQuantity,
		"adjustment_qty":    adjustment.AdjustmentQty,
		"reason":            adjustment.Reason,
		"notes":             adjustment.Notes,
		"adjustment_date":   adjustment.AdjustmentDate,
		"created_by":        adjustment.CreatedBy,
		"created_at":        adjustment.CreatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create stock adjustment", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stock adjustment"})
		return
	}

	logger.Log.Info().
		Str("adjustment_number", adjustment.AdjustmentNumber).
		Msg("Stock adjustment created successfully")

	c.JSON(http.StatusCreated, result[0])
}

// GetStockAdjustments retrieves stock adjustments with filters
func (h *InventoryHandler) GetStockAdjustments(c *gin.Context) {
	productID := c.Query("product_id")
	warehouseID := c.Query("warehouse_id")

	query := "SELECT * FROM stock_adjustment"
	params := make(map[string]interface{})

	var conditions []string
	if productID != "" {
		conditions = append(conditions, "product_id = $product_id")
		params["product_id"] = productID
	}
	if warehouseID != "" {
		conditions = append(conditions, "warehouse_id = $warehouse_id")
		params["warehouse_id"] = warehouseID
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY adjustment_date DESC"

	adjustments, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get stock adjustments", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stock adjustments"})
		return
	}

	c.JSON(http.StatusOK, adjustments)
}

// GetLowStockProducts retrieves products with low stock levels
func (h *InventoryHandler) GetLowStockProducts(c *gin.Context) {
	warehouseID := c.Query("warehouse_id")

	query := `SELECT s.*, p.reorder_level 
		FROM stock s 
		JOIN product p ON s.product_id = p.id 
		WHERE s.quantity_available <= p.reorder_level`
	
	params := make(map[string]interface{})

	if warehouseID != "" {
		query += " AND s.warehouse_id = $warehouse_id"
		params["warehouse_id"] = warehouseID
	}

	query += " ORDER BY s.quantity_available ASC"

	products, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get low stock products", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve low stock products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// Helper function to update stock quantity
func (h *InventoryHandler) updateStockQuantity(productID, warehouseID string, quantity decimal.Decimal, movementType models.MovementType) error {
	// Get current stock
	stocks, err := h.db.QuerySlice(
		"SELECT * FROM stock WHERE product_id = $product_id AND warehouse_id = $warehouse_id",
		map[string]interface{}{
			"product_id":   productID,
			"warehouse_id": warehouseID,
		},
	)

	var currentQty decimal.Decimal
	var stockID string

	if err == nil && len(stocks) > 0 {
		stock := stocks[0].(map[string]interface{})
		stockID = stock["id"].(string)
		currentQty = decimal.NewFromFloat(stock["quantity_on_hand"].(float64))
	} else {
		currentQty = decimal.Zero
	}

	// Calculate new quantity
	var newQty decimal.Decimal
	switch movementType {
	case models.MovementTypeIn:
		newQty = currentQty.Add(quantity)
	case models.MovementTypeOut:
		newQty = currentQty.Sub(quantity)
	case models.MovementTypeAdjustment:
		newQty = quantity
	default:
		newQty = currentQty
	}

	// Update or create stock record
	if stockID != "" {
		_, err = h.db.QuerySlice(
			`UPDATE $id SET 
				quantity_on_hand = $quantity_on_hand,
				quantity_available = $quantity_available,
				last_stock_date = $last_stock_date,
				updated_at = $updated_at`,
			map[string]interface{}{
				"id":                 stockID,
				"quantity_on_hand":   newQty,
				"quantity_available": newQty,
				"last_stock_date":    time.Now(),
				"updated_at":         time.Now(),
			},
		)
	}

	return err
}
