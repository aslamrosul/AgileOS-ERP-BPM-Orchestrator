package handlers

import (
	"fmt"
	"net/http"
	"time"

	"agileos-backend/database"
	"agileos-backend/logger"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
)

// InventoryHandler handles inventory operations
type InventoryHandler struct {
	db *database.SurrealDB
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(db *database.SurrealDB) *InventoryHandler {
	return &InventoryHandler{db: db}
}

// ============================================
// PRODUCT MANAGEMENT HANDLERS
// ============================================

// CreateProduct creates a new product
func (h *InventoryHandler) CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		logger.LogError("Failed to bind product data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate product code
	products, err := h.db.QuerySlice(
		"SELECT product_code FROM product ORDER BY product_code DESC LIMIT 1",
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last product code", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate product code"})
		return
	}

	productCode := "PRD-0001"
	if len(products) > 0 {
		lastCode := products[0].(map[string]interface{})["product_code"].(string)
		var lastNum int
		fmt.Sscanf(lastCode, "PRD-%d", &lastNum)
		productCode = fmt.Sprintf("PRD-%04d", lastNum+1)
	}

	product.ProductCode = productCode
	product.CreatedBy = userID.(string)
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	product.IsActive = true

	query := `CREATE product CONTENT {
		product_code: $product_code,
		product_name: $product_name,
		description: $description,
		category: $category,
		sub_category: $sub_category,
		unit_of_measure: $unit_of_measure,
		product_type: $product_type,
		cost_price: $cost_price,
		selling_price: $selling_price,
		currency: $currency,
		track_inventory: $track_inventory,
		min_stock_level: $min_stock_level,
		max_stock_level: $max_stock_level,
		reorder_level: $reorder_level,
		reorder_quantity: $reorder_quantity,
		weight: $weight,
		weight_unit: $weight_unit,
		volume: $volume,
		volume_unit: $volume_unit,
		barcode: $barcode,
		sku: $sku,
		is_active: $is_active,
		is_saleable: $is_saleable,
		is_purchaseable: $is_purchaseable,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"product_code":     product.ProductCode,
		"product_name":     product.ProductName,
		"description":      product.Description,
		"category":         product.Category,
		"sub_category":     product.SubCategory,
		"unit_of_measure":  product.UnitOfMeasure,
		"product_type":     product.ProductType,
		"cost_price":       product.CostPrice,
		"selling_price":    product.SellingPrice,
		"currency":         product.Currency,
		"track_inventory":  product.TrackInventory,
		"min_stock_level":  product.MinStockLevel,
		"max_stock_level":  product.MaxStockLevel,
		"reorder_level":    product.ReorderLevel,
		"reorder_quantity": product.ReorderQuantity,
		"weight":           product.Weight,
		"weight_unit":      product.WeightUnit,
		"volume":           product.Volume,
		"volume_unit":      product.VolumeUnit,
		"barcode":          product.Barcode,
		"sku":              product.SKU,
		"is_active":        product.IsActive,
		"is_saleable":      product.IsSaleable,
		"is_purchaseable":  product.IsPurchaseable,
		"created_by":       product.CreatedBy,
		"created_at":       product.CreatedAt,
		"updated_at":       product.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create product", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	logger.Log.Info().
		Str("product_code", product.ProductCode).
		Str("product_name", product.ProductName).
		Msg("Product created successfully")

	c.JSON(http.StatusCreated, result[0])
}

// GetProducts retrieves all products with filters
func (h *InventoryHandler) GetProducts(c *gin.Context) {
	category := c.Query("category")
	productType := c.Query("product_type")
	isActive := c.Query("is_active")

	query := "SELECT * FROM product"
	params := make(map[string]interface{})

	var conditions []string
	if category != "" {
		conditions = append(conditions, "category = $category")
		params["category"] = category
	}
	if productType != "" {
		conditions = append(conditions, "product_type = $product_type")
		params["product_type"] = productType
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

	query += " ORDER BY product_code ASC"

	products, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get products", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetProduct retrieves a product by ID
func (h *InventoryHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")

	products, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": productID},
	)
	if err != nil || len(products) == 0 {
		logger.LogError("Product not found", err, map[string]interface{}{"product_id": productID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, products[0])
}

// UpdateProduct updates an existing product
func (h *InventoryHandler) UpdateProduct(c *gin.Context) {
	productID := c.Param("id")

	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		logger.LogError("Failed to bind product data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.UpdatedAt = time.Now()

	query := `UPDATE $id SET
		product_name = $product_name,
		description = $description,
		category = $category,
		sub_category = $sub_category,
		unit_of_measure = $unit_of_measure,
		product_type = $product_type,
		cost_price = $cost_price,
		selling_price = $selling_price,
		track_inventory = $track_inventory,
		min_stock_level = $min_stock_level,
		max_stock_level = $max_stock_level,
		reorder_level = $reorder_level,
		reorder_quantity = $reorder_quantity,
		weight = $weight,
		weight_unit = $weight_unit,
		volume = $volume,
		volume_unit = $volume_unit,
		barcode = $barcode,
		sku = $sku,
		is_active = $is_active,
		is_saleable = $is_saleable,
		is_purchaseable = $is_purchaseable,
		updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":               productID,
		"product_name":     product.ProductName,
		"description":      product.Description,
		"category":         product.Category,
		"sub_category":     product.SubCategory,
		"unit_of_measure":  product.UnitOfMeasure,
		"product_type":     product.ProductType,
		"cost_price":       product.CostPrice,
		"selling_price":    product.SellingPrice,
		"track_inventory":  product.TrackInventory,
		"min_stock_level":  product.MinStockLevel,
		"max_stock_level":  product.MaxStockLevel,
		"reorder_level":    product.ReorderLevel,
		"reorder_quantity": product.ReorderQuantity,
		"weight":           product.Weight,
		"weight_unit":      product.WeightUnit,
		"volume":           product.Volume,
		"volume_unit":      product.VolumeUnit,
		"barcode":          product.Barcode,
		"sku":              product.SKU,
		"is_active":        product.IsActive,
		"is_saleable":      product.IsSaleable,
		"is_purchaseable":  product.IsPurchaseable,
		"updated_at":       product.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to update product", err, map[string]interface{}{"product_id": productID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	logger.Log.Info().
		Str("product_id", productID).
		Str("product_name", product.ProductName).
		Msg("Product updated successfully")

	c.JSON(http.StatusOK, result[0])
}

// DeleteProduct soft deletes a product
func (h *InventoryHandler) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")

	query := `UPDATE $id SET is_active = false, updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         productID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to delete product", err, map[string]interface{}{"product_id": productID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	logger.Log.Info().
		Str("product_id", productID).
		Msg("Product deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
