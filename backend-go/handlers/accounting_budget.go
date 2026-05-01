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
// BUDGET MANAGEMENT HANDLERS
// ============================================

// CreateBudget creates a new budget
func (h *AccountingHandler) CreateBudget(c *gin.Context) {
	var budget models.Budget
	if err := c.ShouldBindJSON(&budget); err != nil {
		logger.LogError("Failed to bind budget data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate budget name if not provided
	if budget.BudgetName == "" {
		budget.BudgetName = fmt.Sprintf("Budget %d - %s", budget.FiscalYear, budget.Department)
	}

	budget.CreatedBy = userID.(string)
	budget.CreatedAt = time.Now()
	budget.UpdatedAt = time.Now()
	budget.Status = "draft"
	budget.ActualAmount = decimal.Zero
	budget.Variance = budget.BudgetAmount

	query := `CREATE budget CONTENT {
		budget_name: $budget_name,
		fiscal_year: $fiscal_year,
		account_id: $account_id,
		account_code: $account_code,
		account_name: $account_name,
		department: $department,
		cost_center: $cost_center,
		period_type: $period_type,
		budget_amount: $budget_amount,
		actual_amount: $actual_amount,
		variance: $variance,
		status: $status,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"budget_name":   budget.BudgetName,
		"fiscal_year":   budget.FiscalYear,
		"account_id":    budget.AccountID,
		"account_code":  budget.AccountCode,
		"account_name":  budget.AccountName,
		"department":    budget.Department,
		"cost_center":   budget.CostCenter,
		"period_type":   budget.PeriodType,
		"budget_amount": budget.BudgetAmount,
		"actual_amount": budget.ActualAmount,
		"variance":      budget.Variance,
		"status":        budget.Status,
		"created_by":    budget.CreatedBy,
		"created_at":    budget.CreatedAt,
		"updated_at":    budget.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create budget", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create budget"})
		return
	}

	logger.Log.Info().
		Str("budget_name", budget.BudgetName).
		Int("fiscal_year", budget.FiscalYear).
		Msg("Budget created successfully")

	if len(result) > 0 {
		c.JSON(http.StatusCreated, result[0])
	} else {
		c.JSON(http.StatusCreated, gin.H{"message": "Budget created successfully"})
	}
}

// GetBudgets retrieves all budgets with filters
func (h *AccountingHandler) GetBudgets(c *gin.Context) {
	fiscalYear := c.Query("fiscal_year")
	department := c.Query("department")
	status := c.Query("status")

	query := "SELECT * FROM budget"
	params := make(map[string]interface{})

	var conditions []string
	if fiscalYear != "" {
		conditions = append(conditions, "fiscal_year = $fiscal_year")
		params["fiscal_year"] = fiscalYear
	}
	if department != "" {
		conditions = append(conditions, "department = $department")
		params["department"] = department
	}
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY fiscal_year DESC, budget_code DESC"

	budgets, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get budgets", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve budgets"})
		return
	}

	c.JSON(http.StatusOK, budgets)
}

// GetBudget retrieves a budget by ID
func (h *AccountingHandler) GetBudget(c *gin.Context) {
	budgetID := c.Param("id")

	budgets, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": budgetID},
	)
	if err != nil {
		logger.LogError("Budget not found", err, map[string]interface{}{"budget_id": budgetID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	if len(budgets) > 0 {
		c.JSON(http.StatusOK, budgets[0])
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
	}
}

// UpdateBudget updates an existing budget
func (h *AccountingHandler) UpdateBudget(c *gin.Context) {
	budgetID := c.Param("id")

	var budget models.Budget
	if err := c.ShouldBindJSON(&budget); err != nil {
		logger.LogError("Failed to bind budget data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if budget is draft
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": budgetID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	existingBudget := existing[0].(map[string]interface{})
	if existingBudget["status"] != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft budgets can be updated"})
		return
	}

	budget.UpdatedAt = time.Now()

	query := `UPDATE $id SET
		budget_name = $budget_name,
		fiscal_year = $fiscal_year,
		account_id = $account_id,
		account_code = $account_code,
		account_name = $account_name,
		department = $department,
		cost_center = $cost_center,
		period_type = $period_type,
		budget_amount = $budget_amount,
		updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":            budgetID,
		"budget_name":   budget.BudgetName,
		"fiscal_year":   budget.FiscalYear,
		"account_id":    budget.AccountID,
		"account_code":  budget.AccountCode,
		"account_name":  budget.AccountName,
		"department":    budget.Department,
		"cost_center":   budget.CostCenter,
		"period_type":   budget.PeriodType,
		"budget_amount": budget.BudgetAmount,
		"updated_at":    budget.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to update budget", err, map[string]interface{}{"budget_id": budgetID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update budget"})
		return
	}

	logger.Log.Info().
		Str("budget_id", budgetID).
		Str("budget_name", budget.BudgetName).
		Msg("Budget updated successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Budget updated successfully"})
	}
}

// DeleteBudget deletes a draft budget
func (h *AccountingHandler) DeleteBudget(c *gin.Context) {
	budgetID := c.Param("id")

	// Check if budget is draft
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": budgetID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	existingBudget := existing[0].(map[string]interface{})
	if existingBudget["status"] != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft budgets can be deleted"})
		return
	}

	_, err = h.db.QuerySlice(
		"DELETE $id",
		map[string]interface{}{"id": budgetID},
	)
	if err != nil {
		logger.LogError("Failed to delete budget", err, map[string]interface{}{"budget_id": budgetID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete budget"})
		return
	}

	logger.Log.Info().
		Str("budget_id", budgetID).
		Msg("Budget deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Budget deleted successfully"})
}

// ApproveBudget approves a submitted budget
func (h *AccountingHandler) ApproveBudget(c *gin.Context) {
	budgetID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if budget is submitted
	existing, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": budgetID},
	)
	if err != nil || len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	existingBudget, ok := existing[0].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid budget data"})
		return
	}
	if existingBudget["status"] != "submitted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only submitted budgets can be approved"})
		return
	}

	query := `UPDATE $id SET 
		status = 'approved', 
		approved_by = $approved_by,
		approved_at = $approved_at,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":          budgetID,
		"approved_by": userID.(string),
		"approved_at": time.Now(),
		"updated_at":  time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to approve budget", err, map[string]interface{}{"budget_id": budgetID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve budget"})
		return
	}

	logger.Log.Info().
		Str("budget_id", budgetID).
		Msg("Budget approved successfully")

	if len(result) > 0 {
		c.JSON(http.StatusOK, result[0])
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Budget approved successfully"})
	}
}

// GetBudgetVariance retrieves budget variance analysis
func (h *AccountingHandler) GetBudgetVariance(c *gin.Context) {
	budgetID := c.Param("id")

	budgets, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": budgetID},
	)
	if err != nil {
		logger.LogError("Budget not found", err, map[string]interface{}{"budget_id": budgetID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	if len(budgets) > 0 {
		budget := budgets[0].(map[string]interface{})
		
		// Calculate variance
		budgetAmount := decimal.NewFromFloat(budget["budget_amount"].(float64))
		actualAmount := decimal.NewFromFloat(budget["actual_amount"].(float64))
		variance := budgetAmount.Sub(actualAmount)
		variancePercent := decimal.Zero
		if !budgetAmount.IsZero() {
			variancePercent = variance.Div(budgetAmount).Mul(decimal.NewFromInt(100))
		}

		response := gin.H{
			"budget_name":      budget["budget_name"],
			"fiscal_year":      budget["fiscal_year"],
			"budget_amount":    budgetAmount,
			"actual_amount":    actualAmount,
			"variance":         variance,
			"variance_percent": variancePercent,
			"status":           budget["status"],
		}

		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
	}
}
