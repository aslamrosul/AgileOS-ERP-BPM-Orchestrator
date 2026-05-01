package handlers

import (
	"fmt"
	"net/http"
	"time"

	"agileos-backend/logger"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
)

// ============================================
// WAREHOUSE MANAGEMENT HANDLERS
// ============================================

// CreateWarehouse creates a new warehouse
func (h *InventoryHandler) CreateWarehouse(c *gin.Context) {
	var warehouse models.Warehouse
	if err := c.ShouldBindJSON(&warehouse); err != nil {
		logger.LogError("Failed to bind warehouse data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate warehouse code
	warehouses, err := h.db.QuerySlice(
		"SELECT warehouse_code FROM warehouse ORDER BY warehouse_code DESC LIMIT 1",
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last warehouse code", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate warehouse code"})
		return
	}

	warehouseCode := "WH-001"
	if len(warehouses) > 0 {
		lastCode := warehouses[0].(map[string]interface{})["warehouse_code"].(string)
		var lastNum int
		fmt.Sscanf(lastCode, "WH-%d", &lastNum)
		warehouseCode = fmt.Sprintf("WH-%03d", lastNum+1)
	}

	warehouse.WarehouseCode = warehouseCode
	warehouse.CreatedBy = userID.(string)
	warehouse.CreatedAt = time.Now()
	warehouse.UpdatedAt = time.Now()
	warehouse.IsActive = true

	query := `CREATE warehouse CONTENT {
		warehouse_code: $warehouse_code,
		warehouse_name: $warehouse_name,
		description: $description,
		warehouse_type: $warehouse_type,
		address: $address,
		city: $city,
		state: $state,
		country: $country,
		postal_code: $postal_code,
		phone: $phone,
		email: $email,
		manager_id: $manager_id,
		manager_name: $manager_name,
		is_active: $is_active,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"warehouse_code": warehouse.WarehouseCode,
		"warehouse_name": warehouse.WarehouseName,
		"description":    warehouse.Description,
		"warehouse_type": warehouse.WarehouseType,
		"address":        warehouse.Address,
		"city":           warehouse.City,
		"state":          warehouse.State,
		"country":        warehouse.Country,
		"postal_code":    warehouse.PostalCode,
		"phone":          warehouse.Phone,
		"email":          warehouse.Email,
		"manager_id":     warehouse.ManagerID,
		"manager_name":   warehouse.ManagerName,
		"is_active":      warehouse.IsActive,
		"created_by":     warehouse.CreatedBy,
		"created_at":     warehouse.CreatedAt,
		"updated_at":     warehouse.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create warehouse", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create warehouse"})
		return
	}

	logger.Log.Info().
		Str("warehouse_code", warehouse.WarehouseCode).
		Str("warehouse_name", warehouse.WarehouseName).
		Msg("Warehouse created successfully")

	c.JSON(http.StatusCreated, result[0])
}

// GetWarehouses retrieves all warehouses with filters
func (h *InventoryHandler) GetWarehouses(c *gin.Context) {
	warehouseType := c.Query("warehouse_type")
	isActive := c.Query("is_active")

	query := "SELECT * FROM warehouse"
	params := make(map[string]interface{})

	var conditions []string
	if warehouseType != "" {
		conditions = append(conditions, "warehouse_type = $warehouse_type")
		params["warehouse_type"] = warehouseType
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

	query += " ORDER BY warehouse_code ASC"

	warehouses, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get warehouses", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve warehouses"})
		return
	}

	c.JSON(http.StatusOK, warehouses)
}

// GetWarehouse retrieves a warehouse by ID
func (h *InventoryHandler) GetWarehouse(c *gin.Context) {
	warehouseID := c.Param("id")

	warehouses, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": warehouseID},
	)
	if err != nil || len(warehouses) == 0 {
		logger.LogError("Warehouse not found", err, map[string]interface{}{"warehouse_id": warehouseID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse not found"})
		return
	}

	c.JSON(http.StatusOK, warehouses[0])
}

// UpdateWarehouse updates an existing warehouse
func (h *InventoryHandler) UpdateWarehouse(c *gin.Context) {
	warehouseID := c.Param("id")

	var warehouse models.Warehouse
	if err := c.ShouldBindJSON(&warehouse); err != nil {
		logger.LogError("Failed to bind warehouse data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	warehouse.UpdatedAt = time.Now()

	query := `UPDATE $id SET
		warehouse_name = $warehouse_name,
		description = $description,
		warehouse_type = $warehouse_type,
		address = $address,
		city = $city,
		state = $state,
		country = $country,
		postal_code = $postal_code,
		phone = $phone,
		email = $email,
		manager_id = $manager_id,
		manager_name = $manager_name,
		is_active = $is_active,
		updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":             warehouseID,
		"warehouse_name": warehouse.WarehouseName,
		"description":    warehouse.Description,
		"warehouse_type": warehouse.WarehouseType,
		"address":        warehouse.Address,
		"city":           warehouse.City,
		"state":          warehouse.State,
		"country":        warehouse.Country,
		"postal_code":    warehouse.PostalCode,
		"phone":          warehouse.Phone,
		"email":          warehouse.Email,
		"manager_id":     warehouse.ManagerID,
		"manager_name":   warehouse.ManagerName,
		"is_active":      warehouse.IsActive,
		"updated_at":     warehouse.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to update warehouse", err, map[string]interface{}{"warehouse_id": warehouseID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update warehouse"})
		return
	}

	logger.Log.Info().
		Str("warehouse_id", warehouseID).
		Str("warehouse_name", warehouse.WarehouseName).
		Msg("Warehouse updated successfully")

	c.JSON(http.StatusOK, result[0])
}

// DeleteWarehouse soft deletes a warehouse
func (h *InventoryHandler) DeleteWarehouse(c *gin.Context) {
	warehouseID := c.Param("id")

	query := `UPDATE $id SET is_active = false, updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         warehouseID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to delete warehouse", err, map[string]interface{}{"warehouse_id": warehouseID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete warehouse"})
		return
	}

	logger.Log.Info().
		Str("warehouse_id", warehouseID).
		Msg("Warehouse deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Warehouse deleted successfully"})
}
