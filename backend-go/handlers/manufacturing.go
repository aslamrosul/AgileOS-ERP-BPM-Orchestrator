package handlers

import (
	"fmt"
	"net/http"
	"time"

	"agileos-backend/database"
	"agileos-backend/logger"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// ManufacturingHandler handles manufacturing operations
type ManufacturingHandler struct {
	db *database.SurrealDB
}

// NewManufacturingHandler creates a new manufacturing handler
func NewManufacturingHandler(db *database.SurrealDB) *ManufacturingHandler {
	return &ManufacturingHandler{db: db}
}

// ============================================
// BILL OF MATERIALS (BOM) HANDLERS
// ============================================

// CreateBOM creates a new BOM
func (h *ManufacturingHandler) CreateBOM(c *gin.Context) {
	var bom models.BOM
	if err := c.ShouldBindJSON(&bom); err != nil {
		logger.LogError("Failed to bind BOM data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate BOM code
	boms, err := h.db.QuerySlice(
		"SELECT bom_code FROM bom ORDER BY bom_code DESC LIMIT 1",
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last BOM code", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate BOM code"})
		return
	}

	bomCode := "BOM-0001"
	if len(boms) > 0 {
		lastCode := boms[0].(map[string]interface{})["bom_code"].(string)
		var lastNum int
		fmt.Sscanf(lastCode, "BOM-%d", &lastNum)
		bomCode = fmt.Sprintf("BOM-%04d", lastNum+1)
	}

	bom.BOMCode = bomCode
	bom.CreatedBy = userID.(string)
	bom.CreatedAt = time.Now()
	bom.UpdatedAt = time.Now()
	bom.IsActive = true
	bom.Version = 1

	// Calculate total cost
	bom.TotalCost = decimal.Zero
	for _, component := range bom.Components {
		bom.TotalCost = bom.TotalCost.Add(component.TotalCost)
	}
	bom.TotalProductionCost = bom.TotalCost.Add(bom.LaborCost).Add(bom.OverheadCost)

	query := `CREATE bom CONTENT {
		bom_code: $bom_code,
		bom_name: $bom_name,
		product_id: $product_id,
		product_code: $product_code,
		product_name: $product_name,
		bom_type: $bom_type,
		quantity: $quantity,
		unit_of_measure: $unit_of_measure,
		version: $version,
		is_active: $is_active,
		is_default: $is_default,
		total_cost: $total_cost,
		labor_cost: $labor_cost,
		overhead_cost: $overhead_cost,
		total_production_cost: $total_production_cost,
		notes: $notes,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"bom_code":              bom.BOMCode,
		"bom_name":              bom.BOMName,
		"product_id":            bom.ProductID,
		"product_code":          bom.ProductCode,
		"product_name":          bom.ProductName,
		"bom_type":              bom.BOMType,
		"quantity":              bom.Quantity,
		"unit_of_measure":       bom.UnitOfMeasure,
		"version":               bom.Version,
		"is_active":             bom.IsActive,
		"is_default":            bom.IsDefault,
		"total_cost":            bom.TotalCost,
		"labor_cost":            bom.LaborCost,
		"overhead_cost":         bom.OverheadCost,
		"total_production_cost": bom.TotalProductionCost,
		"notes":                 bom.Notes,
		"created_by":            bom.CreatedBy,
		"created_at":            bom.CreatedAt,
		"updated_at":            bom.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create BOM", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create BOM"})
		return
	}

	logger.Log.Info().Str("bom_code", bom.BOMCode).Msg("BOM created successfully")
	c.JSON(http.StatusCreated, result[0])
}

// GetBOMs retrieves all BOMs with filters
func (h *ManufacturingHandler) GetBOMs(c *gin.Context) {
	productID := c.Query("product_id")
	bomType := c.Query("bom_type")
	isActive := c.Query("is_active")

	query := "SELECT * FROM bom"
	params := make(map[string]interface{})

	var conditions []string
	if productID != "" {
		conditions = append(conditions, "product_id = $product_id")
		params["product_id"] = productID
	}
	if bomType != "" {
		conditions = append(conditions, "bom_type = $bom_type")
		params["bom_type"] = bomType
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

	query += " ORDER BY bom_code ASC"

	boms, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get BOMs", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve BOMs"})
		return
	}

	c.JSON(http.StatusOK, boms)
}

// GetBOM retrieves a BOM by ID
func (h *ManufacturingHandler) GetBOM(c *gin.Context) {
	bomID := c.Param("id")

	boms, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": bomID},
	)
	if err != nil || len(boms) == 0 {
		logger.LogError("BOM not found", err, map[string]interface{}{"bom_id": bomID})
		c.JSON(http.StatusNotFound, gin.H{"error": "BOM not found"})
		return
	}

	c.JSON(http.StatusOK, boms[0])
}

// UpdateBOM updates an existing BOM
func (h *ManufacturingHandler) UpdateBOM(c *gin.Context) {
	bomID := c.Param("id")

	var bom models.BOM
	if err := c.ShouldBindJSON(&bom); err != nil {
		logger.LogError("Failed to bind BOM data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bom.UpdatedAt = time.Now()

	// Recalculate total cost
	bom.TotalCost = decimal.Zero
	for _, component := range bom.Components {
		bom.TotalCost = bom.TotalCost.Add(component.TotalCost)
	}
	bom.TotalProductionCost = bom.TotalCost.Add(bom.LaborCost).Add(bom.OverheadCost)

	query := `UPDATE $id SET
		bom_name = $bom_name,
		bom_type = $bom_type,
		quantity = $quantity,
		unit_of_measure = $unit_of_measure,
		is_active = $is_active,
		is_default = $is_default,
		total_cost = $total_cost,
		labor_cost = $labor_cost,
		overhead_cost = $overhead_cost,
		total_production_cost = $total_production_cost,
		notes = $notes,
		updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":                    bomID,
		"bom_name":              bom.BOMName,
		"bom_type":              bom.BOMType,
		"quantity":              bom.Quantity,
		"unit_of_measure":       bom.UnitOfMeasure,
		"is_active":             bom.IsActive,
		"is_default":            bom.IsDefault,
		"total_cost":            bom.TotalCost,
		"labor_cost":            bom.LaborCost,
		"overhead_cost":         bom.OverheadCost,
		"total_production_cost": bom.TotalProductionCost,
		"notes":                 bom.Notes,
		"updated_at":            bom.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to update BOM", err, map[string]interface{}{"bom_id": bomID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update BOM"})
		return
	}

	logger.Log.Info().Str("bom_id", bomID).Msg("BOM updated successfully")
	c.JSON(http.StatusOK, result[0])
}

// DeleteBOM soft deletes a BOM
func (h *ManufacturingHandler) DeleteBOM(c *gin.Context) {
	bomID := c.Param("id")

	query := `UPDATE $id SET is_active = false, updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         bomID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to delete BOM", err, map[string]interface{}{"bom_id": bomID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete BOM"})
		return
	}

	logger.Log.Info().Str("bom_id", bomID).Msg("BOM deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "BOM deleted successfully"})
}

// SetDefaultBOM sets a BOM as default for a product
func (h *ManufacturingHandler) SetDefaultBOM(c *gin.Context) {
	bomID := c.Param("id")

	// Get BOM to find product_id
	boms, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": bomID},
	)
	if err != nil || len(boms) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "BOM not found"})
		return
	}

	bom := boms[0].(map[string]interface{})
	productID := bom["product_id"].(string)

	// Unset all other BOMs for this product
	_, err = h.db.QuerySlice(
		"UPDATE bom SET is_default = false WHERE product_id = $product_id",
		map[string]interface{}{"product_id": productID},
	)
	if err != nil {
		logger.LogError("Failed to unset default BOMs", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set default BOM"})
		return
	}

	// Set this BOM as default
	query := `UPDATE $id SET is_default = true, updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         bomID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to set default BOM", err, map[string]interface{}{"bom_id": bomID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set default BOM"})
		return
	}

	logger.Log.Info().Str("bom_id", bomID).Msg("BOM set as default successfully")
	c.JSON(http.StatusOK, result[0])
}

// CopyBOM creates a new version of BOM
func (h *ManufacturingHandler) CopyBOM(c *gin.Context) {
	bomID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get existing BOM
	boms, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": bomID},
	)
	if err != nil || len(boms) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "BOM not found"})
		return
	}

	existingBOM := boms[0].(map[string]interface{})

	// Auto-generate new BOM code
	allBOMs, err := h.db.QuerySlice(
		"SELECT bom_code FROM bom ORDER BY bom_code DESC LIMIT 1",
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last BOM code", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate BOM code"})
		return
	}

	bomCode := "BOM-0001"
	if len(allBOMs) > 0 {
		lastCode := allBOMs[0].(map[string]interface{})["bom_code"].(string)
		var lastNum int
		fmt.Sscanf(lastCode, "BOM-%d", &lastNum)
		bomCode = fmt.Sprintf("BOM-%04d", lastNum+1)
	}

	// Get current version and increment
	currentVersion := int(existingBOM["version"].(float64))
	newVersion := currentVersion + 1

	query := `CREATE bom CONTENT {
		bom_code: $bom_code,
		bom_name: $bom_name,
		product_id: $product_id,
		product_code: $product_code,
		product_name: $product_name,
		bom_type: $bom_type,
		quantity: $quantity,
		unit_of_measure: $unit_of_measure,
		version: $version,
		is_active: true,
		is_default: false,
		total_cost: $total_cost,
		labor_cost: $labor_cost,
		overhead_cost: $overhead_cost,
		total_production_cost: $total_production_cost,
		notes: $notes,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"bom_code":              bomCode,
		"bom_name":              existingBOM["bom_name"].(string) + " (v" + fmt.Sprintf("%d", newVersion) + ")",
		"product_id":            existingBOM["product_id"],
		"product_code":          existingBOM["product_code"],
		"product_name":          existingBOM["product_name"],
		"bom_type":              existingBOM["bom_type"],
		"quantity":              existingBOM["quantity"],
		"unit_of_measure":       existingBOM["unit_of_measure"],
		"version":               newVersion,
		"total_cost":            existingBOM["total_cost"],
		"labor_cost":            existingBOM["labor_cost"],
		"overhead_cost":         existingBOM["overhead_cost"],
		"total_production_cost": existingBOM["total_production_cost"],
		"notes":                 "Copied from " + existingBOM["bom_code"].(string),
		"created_by":            userID.(string),
		"created_at":            time.Now(),
		"updated_at":            time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to copy BOM", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy BOM"})
		return
	}

	logger.Log.Info().Str("new_bom_code", bomCode).Str("source_bom_id", bomID).Msg("BOM copied successfully")
	c.JSON(http.StatusCreated, result[0])
}

// GetBOMCostBreakdown retrieves cost breakdown for a BOM
func (h *ManufacturingHandler) GetBOMCostBreakdown(c *gin.Context) {
	bomID := c.Param("id")

	boms, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": bomID},
	)
	if err != nil || len(boms) == 0 {
		logger.LogError("BOM not found", err, map[string]interface{}{"bom_id": bomID})
		c.JSON(http.StatusNotFound, gin.H{"error": "BOM not found"})
		return
	}

	bom := boms[0].(map[string]interface{})

	breakdown := gin.H{
		"bom_code":              bom["bom_code"],
		"bom_name":              bom["bom_name"],
		"material_cost":         bom["total_cost"],
		"labor_cost":            bom["labor_cost"],
		"overhead_cost":         bom["overhead_cost"],
		"total_production_cost": bom["total_production_cost"],
	}

	c.JSON(http.StatusOK, breakdown)
}
