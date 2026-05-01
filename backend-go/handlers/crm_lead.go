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
// LEAD MANAGEMENT HANDLERS
// ============================================

// CreateLead creates a new lead
func (h *CRMHandler) CreateLead(c *gin.Context) {
	var lead models.Lead
	if err := c.ShouldBindJSON(&lead); err != nil {
		logger.LogError("Failed to bind lead data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Auto-generate lead number
	year := time.Now().Year()
	leads, err := h.db.QuerySlice(
		fmt.Sprintf("SELECT lead_number FROM lead WHERE lead_number LIKE 'LEAD-%d-%%' ORDER BY lead_number DESC LIMIT 1", year),
		nil,
	)
	if err != nil {
		logger.LogError("Failed to get last lead number", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate lead number"})
		return
	}

	leadNumber := fmt.Sprintf("LEAD-%d-0001", year)
	if len(leads) > 0 {
		lastNumber := leads[0].(map[string]interface{})["lead_number"].(string)
		var lastNum int
		fmt.Sscanf(lastNumber, fmt.Sprintf("LEAD-%d-%%d", year), &lastNum)
		leadNumber = fmt.Sprintf("LEAD-%d-%04d", year, lastNum+1)
	}

	lead.LeadNumber = leadNumber
	lead.CreatedBy = userID.(string)
	lead.CreatedAt = time.Now()
	lead.UpdatedAt = time.Now()
	lead.Status = models.LeadStatusNew
	lead.LeadScore = 0

	query := `CREATE lead CONTENT {
		lead_number: $lead_number,
		lead_name: $lead_name,
		company: $company,
		contact_id: $contact_id,
		contact_name: $contact_name,
		email: $email,
		phone: $phone,
		source: $source,
		status: $status,
		lead_score: $lead_score,
		industry: $industry,
		estimated_value: $estimated_value,
		currency: $currency,
		expected_close_date: $expected_close_date,
		description: $description,
		notes: $notes,
		assigned_to: $assigned_to,
		assigned_to_name: $assigned_to_name,
		created_by: $created_by,
		created_at: $created_at,
		updated_at: $updated_at
	}`

	params := map[string]interface{}{
		"lead_number":         lead.LeadNumber,
		"lead_name":           lead.LeadName,
		"company":             lead.Company,
		"contact_id":          lead.ContactID,
		"contact_name":        lead.ContactName,
		"email":               lead.Email,
		"phone":               lead.Phone,
		"source":              lead.Source,
		"status":              lead.Status,
		"lead_score":          lead.LeadScore,
		"industry":            lead.Industry,
		"estimated_value":     lead.EstimatedValue,
		"currency":            lead.Currency,
		"expected_close_date": lead.ExpectedCloseDate,
		"description":         lead.Description,
		"notes":               lead.Notes,
		"assigned_to":         lead.AssignedTo,
		"assigned_to_name":    lead.AssignedToName,
		"created_by":          lead.CreatedBy,
		"created_at":          lead.CreatedAt,
		"updated_at":          lead.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to create lead", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create lead"})
		return
	}

	logger.Log.Info().Str("lead_number", lead.LeadNumber).Msg("Lead created successfully")
	c.JSON(http.StatusCreated, result[0])
}

// GetLeads retrieves all leads with filters
func (h *CRMHandler) GetLeads(c *gin.Context) {
	status := c.Query("status")
	source := c.Query("source")
	assignedTo := c.Query("assigned_to")

	query := "SELECT * FROM lead"
	params := make(map[string]interface{})

	var conditions []string
	if status != "" {
		conditions = append(conditions, "status = $status")
		params["status"] = status
	}
	if source != "" {
		conditions = append(conditions, "source = $source")
		params["source"] = source
	}
	if assignedTo != "" {
		conditions = append(conditions, "assigned_to = $assigned_to")
		params["assigned_to"] = assignedTo
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY created_at DESC"

	leads, err := h.db.QuerySlice(query, params)
	if err != nil {
		logger.LogError("Failed to get leads", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve leads"})
		return
	}

	c.JSON(http.StatusOK, leads)
}

// GetLead retrieves a lead by ID
func (h *CRMHandler) GetLead(c *gin.Context) {
	leadID := c.Param("id")

	leads, err := h.db.QuerySlice(
		"SELECT * FROM $id",
		map[string]interface{}{"id": leadID},
	)
	if err != nil || len(leads) == 0 {
		logger.LogError("Lead not found", err, map[string]interface{}{"lead_id": leadID})
		c.JSON(http.StatusNotFound, gin.H{"error": "Lead not found"})
		return
	}

	c.JSON(http.StatusOK, leads[0])
}

// UpdateLead updates an existing lead
func (h *CRMHandler) UpdateLead(c *gin.Context) {
	leadID := c.Param("id")

	var lead models.Lead
	if err := c.ShouldBindJSON(&lead); err != nil {
		logger.LogError("Failed to bind lead data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lead.UpdatedAt = time.Now()

	query := `UPDATE $id SET
		lead_name = $lead_name,
		company = $company,
		contact_id = $contact_id,
		contact_name = $contact_name,
		email = $email,
		phone = $phone,
		source = $source,
		status = $status,
		lead_score = $lead_score,
		industry = $industry,
		estimated_value = $estimated_value,
		expected_close_date = $expected_close_date,
		description = $description,
		notes = $notes,
		assigned_to = $assigned_to,
		assigned_to_name = $assigned_to_name,
		updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":                  leadID,
		"lead_name":           lead.LeadName,
		"company":             lead.Company,
		"contact_id":          lead.ContactID,
		"contact_name":        lead.ContactName,
		"email":               lead.Email,
		"phone":               lead.Phone,
		"source":              lead.Source,
		"status":              lead.Status,
		"lead_score":          lead.LeadScore,
		"industry":            lead.Industry,
		"estimated_value":     lead.EstimatedValue,
		"expected_close_date": lead.ExpectedCloseDate,
		"description":         lead.Description,
		"notes":               lead.Notes,
		"assigned_to":         lead.AssignedTo,
		"assigned_to_name":    lead.AssignedToName,
		"updated_at":          lead.UpdatedAt,
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to update lead", err, map[string]interface{}{"lead_id": leadID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lead"})
		return
	}

	logger.Log.Info().Str("lead_id", leadID).Msg("Lead updated successfully")
	c.JSON(http.StatusOK, result[0])
}

// DeleteLead deletes a lead
func (h *CRMHandler) DeleteLead(c *gin.Context) {
	leadID := c.Param("id")

	_, err := h.db.QuerySlice(
		"DELETE $id",
		map[string]interface{}{"id": leadID},
	)
	if err != nil {
		logger.LogError("Failed to delete lead", err, map[string]interface{}{"lead_id": leadID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete lead"})
		return
	}

	logger.Log.Info().Str("lead_id", leadID).Msg("Lead deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Lead deleted successfully"})
}

// QualifyLead qualifies a lead
func (h *CRMHandler) QualifyLead(c *gin.Context) {
	leadID := c.Param("id")

	query := `UPDATE $id SET status = 'qualified', updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         leadID,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to qualify lead", err, map[string]interface{}{"lead_id": leadID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to qualify lead"})
		return
	}

	logger.Log.Info().Str("lead_id", leadID).Msg("Lead qualified successfully")
	c.JSON(http.StatusOK, result[0])
}

// ConvertLead converts a lead to opportunity
func (h *CRMHandler) ConvertLead(c *gin.Context) {
	leadID := c.Param("id")

	var req struct {
		OpportunityID string `json:"opportunity_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.LogError("Failed to bind conversion data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `UPDATE $id SET 
		status = 'converted',
		converted_to_opportunity_id = $opportunity_id,
		converted_at = $converted_at,
		updated_at = $updated_at`
	
	params := map[string]interface{}{
		"id":             leadID,
		"opportunity_id": req.OpportunityID,
		"converted_at":   time.Now(),
		"updated_at":     time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to convert lead", err, map[string]interface{}{"lead_id": leadID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert lead"})
		return
	}

	logger.Log.Info().Str("lead_id", leadID).Msg("Lead converted successfully")
	c.JSON(http.StatusOK, result[0])
}

// UpdateLeadScore updates lead score
func (h *CRMHandler) UpdateLeadScore(c *gin.Context) {
	leadID := c.Param("id")

	var req struct {
		LeadScore int `json:"lead_score"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.LogError("Failed to bind score data", err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `UPDATE $id SET lead_score = $lead_score, updated_at = $updated_at`
	params := map[string]interface{}{
		"id":         leadID,
		"lead_score": req.LeadScore,
		"updated_at": time.Now(),
	}

	result, err := h.db.QuerySlice(query, params)
	if err != nil || len(result) == 0 {
		logger.LogError("Failed to update lead score", err, map[string]interface{}{"lead_id": leadID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lead score"})
		return
	}

	logger.Log.Info().Str("lead_id", leadID).Int("score", req.LeadScore).Msg("Lead score updated successfully")
	c.JSON(http.StatusOK, result[0])
}
