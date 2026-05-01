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
// OPPORTUNITY MANAGEMENT HANDLERS
// ============================================

// CreateOpportunity creates a new opportunity
func (h *CRMHandler) CreateOpportunity(c *gin.Context) {
	var opportunity models.Opportunity
	if err := c.ShouldBindJSON(&opportunity); err != nil {
		logger.LogError("Failed to bind opportunity data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate opportunity number
	year := time.Now().Year()
	opportunities, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT opportunity_number FROM opportunity WHERE opportunity_number LIKE 'OPP-%d-%%' ORDER BY opportunity_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last opportunity number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate opportunity number"})
		return
	}

	opportunityNumber := fmt.Sprintf("OPP-%d-0001", year)
	if len(opportunities) > 0 {
		lastNumber := opportunities[0].(map[string]interface{})["opportunity_number"].(string)
		var lastNum int
		fmt.Sscanf(lastNumber, fmt.Sprintf("OPP-%d-%%d", year), &lastNum)
		opportunityNumber = fmt.Sprintf("OPP-%d-%04d", year, lastNum+1)
	}

	opportunity.OpportunityNumber = opportunityNumber
	opportunity.CreatedBy = userID.(string)
	opportunity.CreatedAt = time.Now()
	opportunity.UpdatedAt = time.Now()
	opportunity.Stage = models.OpportunityStageProspecting
	opportunity.Probability = 10
	opportunity.ActualRevenue = decimal.Zero

	query := `CREATE opportunity CONTENT {
		opportunity_number: $opportunity_number,
		opportunity_name: $opportunity_name,
		customer_id: $customer_id,
		customer_name: $customer_name,
		contact_id: $contact_id,
		contact_name: $contact_name,
		stage: $stage,
		probability: $probability,
		expected_revenue: $expected_revenue,
		actual_revenue: $actual_revenue,
		currency: $currency,
		expected_close_date: $expected_close_date,
		source: $source,
		description: $description,
		notes: $notes,
		assigned_to: $assigned_to,
		assigned_to_name: $assigned_to_name,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"opportunity_number": opportunity.OpportunityNumber,
		"opportunity_name":   opportunity.OpportunityName,
		"customer_id":        opportunity.CustomerID,
		"customer_name":      opportunity.CustomerName,
		"contact_id":         opportunity.ContactID,
		"contact_name":       opportunity.ContactName,
		"stage":              opportunity.Stage,
		"probability":        opportunity.Probability,
		"expected_revenue":   opportunity.ExpectedRevenue,
		"actual_revenue":     opportunity.ActualRevenue,
		"currency":           opportunity.Currency,
		"expected_close_date": opportunity.ExpectedCloseDate,
		"source":             opportunity.Source,
		"description":        opportunity.Description,
		"notes":              opportunity.Notes,
		"assigned_to":        opportunity.AssignedTo,
		"assigned_to_name":   opportunity.AssignedToName,
		"created_by":         opportunity.CreatedBy,
		"created_at":         opportunity.CreatedAt,
		"updated_at":         opportunity.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create opportunity", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create opportunity"})
		return
	}

	logger.Log.Info().Str("opportunity_number", opportunity.OpportunityNumber).Msg("Opportunity created successfully")
	c.JSON(http.StatusCreated, result[0])
}

// GetOpportunities retrieves all opportunities with filters
func (h *CRMHandler) GetOpportunities(c *gin.Context) {
	stage := c.Query("stage")
	assignedTo := c.Query("assigned_to")
	customerID := c.Query("customer_id")

	query := "SELECT * FROM opportunity"
	params := make(map[string]interface{})

	var conditions []string
	if stage != "" {
		conditions = append(conditions, "stage = $stage")
		params["stage"] = stage
	}
	if assignedTo != "" {
		conditions = append(conditions, "assigned_to = $assigned_to")
		params["assigned_to"] = assignedTo
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

	query += " ORDER BY expected_close_date ASC"

	opportunities, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get opportunities", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve opportunities"})
		return
	}

	c.JSON(http.StatusOK, opportunities)
}

// GetOpportunity retrieves an opportunity by ID
func (h *CRMHandler) GetOpportunity(c *gin.Context) {
	opportunityID := c.Param("id")

	opportunities, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": opportunityID},
	)
	if err != nil || len(opportunities) == 0 {
		logger.LogError("Opportunity not found", err, map[string]interface{}{"opportunity_id": opportunityID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Opportunity not found"})
		return
	}

	c.JSON(http.StatusOK, opportunities[0])
}

// UpdateOpportunity updates an existing opportunity
func (h *CRMHandler) UpdateOpportunity(c *gin.Context) {
	opportunityID := c.Param("id")

	var opportunity models.Opportunity
	if err := c.ShouldBindJSON(&opportunity); err != nil {
		logger.LogError("Failed to bind opportunity data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	opportunity.UpdatedAt = time.Now()

	query := `UPDATE $id SET
		opportunity_name = $opportunity_name,
		customer_id = $customer_id,
		customer_name = $customer_name,
		contact_id = $contact_id,
		contact_name = $contact_name,
		stage = $stage,
		probability = $probability,
		expected_revenue = $expected_revenue,
		expected_close_date = $expected_close_date,
		description = $description,
		notes = $notes,
		assigned_to = $assigned_to,
		assigned_to_name = $assigned_to_name,
		updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":                  opportunityID,
		"opportunity_name":    opportunity.OpportunityName,
		"customer_id":         opportunity.CustomerID,
		"customer_name":       opportunity.CustomerName,
		"contact_id":          opportunity.ContactID,
		"contact_name":        opportunity.ContactName,
		"stage":               opportunity.Stage,
		"probability":         opportunity.Probability,
		"expected_revenue":    opportunity.ExpectedRevenue,
		"expected_close_date": opportunity.ExpectedCloseDate,
		"description":         opportunity.Description,
		"notes":               opportunity.Notes,
		"assigned_to":         opportunity.AssignedTo,
		"assigned_to_name":    opportunity.AssignedToName,
		"updated_at":          opportunity.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to update opportunity", err, map[string]interface{}{"opportunity_id": opportunityID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update opportunity"})
		return
	}

	logger.Log.Info().Str("opportunity_id", opportunityID).Msg("Opportunity updated successfully")
	c.JSON(http.StatusOK, result[0])
}

// DeleteOpportunity deletes an opportunity
func (h *CRMHandler) DeleteOpportunity(c *gin.Context) {
	opportunityID := c.Param("id")

	_, err := h.db.QuerySlice(
		"DELETE $id",
		map[string]interface{}{"id": opportunityID},
	)
	if err != nil {
		logger.LogError("Failed to delete opportunity", err, map[string]interface{}{"opportunity_id": opportunityID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete opportunity"})
		return
	}

	logger.Log.Info().Str("opportunity_id", opportunityID).Msg("Opportunity deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Opportunity deleted successfully"})
}

// MoveOpportunityStage moves opportunity to next stage
func (h *CRMHandler) MoveOpportunityStage(c *gin.Context) {
	opportunityID := c.Param("id")

	var req struct {
		Stage       models.OpportunityStage `json:"stage"`
		Probability int                     `json:"probability"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.LogError("Failed to bind stage data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `UPDATE $id SET 
		stage = $stage,
		probability = $probability,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":          opportunityID,
		"stage":       req.Stage,
		"probability": req.Probability,
		"updated_at":  time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to move opportunity stage", err, map[string]interface{}{"opportunity_id": opportunityID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move opportunity stage"})
		return
	}

	logger.Log.Info().Str("opportunity_id", opportunityID).Str("stage", string(req.Stage)).Msg("Opportunity stage moved successfully")
	c.JSON(http.StatusOK, result[0])
}

// WinOpportunity marks opportunity as won
func (h *CRMHandler) WinOpportunity(c *gin.Context) {
	opportunityID := c.Param("id")

	var req struct {
		ActualRevenue decimal.Decimal `json:"actual_revenue"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.LogError("Failed to bind win data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `UPDATE $id SET 
		stage = 'closed_won',
		probability = 100,
		actual_revenue = $actual_revenue,
		actual_close_date = $actual_close_date,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":                opportunityID,
		"actual_revenue":    req.ActualRevenue,
		"actual_close_date": time.Now(),
		"updated_at":        time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to win opportunity", err, map[string]interface{}{"opportunity_id": opportunityID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to win opportunity"})
		return
	}

	logger.Log.Info().Str("opportunity_id", opportunityID).Msg("Opportunity won successfully")
	c.JSON(http.StatusOK, result[0])
}

// LoseOpportunity marks opportunity as lost
func (h *CRMHandler) LoseOpportunity(c *gin.Context) {
	opportunityID := c.Param("id")

	var req struct {
		LossReason string `json:"loss_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.LogError("Failed to bind loss data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `UPDATE $id SET 
		stage = 'closed_lost',
		probability = 0,
		loss_reason = $loss_reason,
		actual_close_date = $actual_close_date,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":                opportunityID,
		"loss_reason":       req.LossReason,
		"actual_close_date": time.Now(),
		"updated_at":        time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to lose opportunity", err, map[string]interface{}{"opportunity_id": opportunityID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lose opportunity"})
		return
	}

	logger.Log.Info().Str("opportunity_id", opportunityID).Msg("Opportunity lost successfully")
	c.JSON(http.StatusOK, result[0])
}

// GetOpportunityPipeline retrieves opportunity pipeline summary
func (h *CRMHandler) GetOpportunityPipeline(c *gin.Context) {
	assignedTo := c.Query("assigned_to")

	query := "SELECT stage, COUNT(*) as count, SUM(expected_revenue) as total_revenue FROM opportunity"
	params := make(map[string]interface{})

	if assignedTo != "" {
		query += " WHERE assigned_to = $assigned_to"
		params["assigned_to"] = assignedTo
	}

	query += " GROUP BY stage"

	pipeline, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get opportunity pipeline", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve opportunity pipeline"})
		return
	}

	c.JSON(http.StatusOK, pipeline)
}

// GetOpportunityForecast retrieves revenue forecast
func (h *CRMHandler) GetOpportunityForecast(c *gin.Context) {
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	query := `SELECT 
		SUM(expected_revenue * probability / 100) as weighted_revenue,
		SUM(expected_revenue) as total_pipeline,
		COUNT(*) as total_opportunities
		FROM opportunity 
		WHERE stage NOT IN ('closed_won', 'closed_lost')`
	
	params := make(map[string]interface{})

	if fromDate != "" && toDate != "" {
		query += " AND expected_close_date >= $from_date AND expected_close_date <= $to_date"
		params["from_date"] = fromDate
		params["to_date"] = toDate
	}

	forecast, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get opportunity forecast", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve opportunity forecast"})
		return
	}

	c.JSON(http.StatusOK, forecast)
}
