package handlers

import (
	"net/http"
	"strconv"
	"time"

	"agileos-backend/internal/audit"
	"agileos-backend/logger"

	"github.com/gin-gonic/gin"
)

// AuditHandler handles audit trail API requests
type AuditHandler struct {
	auditService *audit.AuditService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *audit.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetAuditTrails retrieves audit trails with filtering and pagination
func (h *AuditHandler) GetAuditTrails(c *gin.Context) {
	// Parse query parameters
	filters := make(map[string]interface{})

	if actorID := c.Query("actor_id"); actorID != "" {
		filters["actor_id"] = actorID
	}

	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}

	if resourceType := c.Query("resource_type"); resourceType != "" {
		filters["resource_type"] = resourceType
	}

	if resourceID := c.Query("resource_id"); resourceID != "" {
		filters["resource_id"] = resourceID
	}

	if complianceStatus := c.Query("compliance_status"); complianceStatus != "" {
		filters["compliance_status"] = complianceStatus
	}

	// Parse date range
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filters["start_date"] = startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filters["end_date"] = endDate
		}
	}

	// Parse pagination
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Get audit trails
	trails, err := h.auditService.GetAuditTrails(filters, limit, offset)
	if err != nil {
		logger.LogError("Failed to retrieve audit trails", err, filters)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve audit trails",
			"details": err.Error(),
		})
		return
	}

	// Get total count
	totalCount, _ := h.auditService.GetAuditTrailCount(filters)

	// Log audit access
	h.auditService.LogAction(audit.AuditTrail{
		ActorID:      c.GetString("user_id"),
		ActorUsername: c.GetString("username"),
		ActorRole:    c.GetString("role"),
		Action:       "AUDIT_ACCESS",
		ResourceType: "audit_trails",
		ResourceID:   "query",
		IPAddress:    c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
		Metadata: map[string]interface{}{
			"filters": filters,
			"limit":   limit,
			"offset":  offset,
		},
	})

	c.JSON(http.StatusOK, gin.H{
		"audit_trails": trails,
		"pagination": gin.H{
			"total":  totalCount,
			"limit":  limit,
			"offset": offset,
		},
		"filters": filters,
	})
}

// GetComplianceViolations retrieves compliance violations
func (h *AuditHandler) GetComplianceViolations(c *gin.Context) {
	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = parsed
		}
	}

	violations, err := h.auditService.GetComplianceViolations(startDate, endDate)
	if err != nil {
		logger.LogError("Failed to retrieve compliance violations", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve compliance violations",
		})
		return
	}

	// Categorize violations by severity
	critical := 0
	warnings := 0
	for _, v := range violations {
		if v.ComplianceStatus == audit.ComplianceFail {
			critical++
		} else if v.ComplianceStatus == audit.ComplianceWarning {
			warnings++
		}
	}

	// Log compliance report access
	h.auditService.LogAction(audit.AuditTrail{
		ActorID:      c.GetString("user_id"),
		ActorUsername: c.GetString("username"),
		ActorRole:    c.GetString("role"),
		Action:       "COMPLIANCE_REPORT_ACCESS",
		ResourceType: "compliance_violations",
		ResourceID:   "report",
		IPAddress:    c.ClientIP(),
		Metadata: map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
			"violations_count": len(violations),
		},
	})

	c.JSON(http.StatusOK, gin.H{
		"violations": violations,
		"summary": gin.H{
			"total":    len(violations),
			"critical": critical,
			"warnings": warnings,
			"period": gin.H{
				"start": startDate,
				"end":   endDate,
			},
		},
	})
}

// ExportAuditTrails exports audit trails to JSON
func (h *AuditHandler) ExportAuditTrails(c *gin.Context) {
	// Parse filters (same as GetAuditTrails)
	filters := make(map[string]interface{})

	if actorID := c.Query("actor_id"); actorID != "" {
		filters["actor_id"] = actorID
	}

	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filters["start_date"] = startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filters["end_date"] = endDate
		}
	}

	// Export audit trails
	data, err := h.auditService.ExportAuditTrails(filters)
	if err != nil {
		logger.LogError("Failed to export audit trails", err, filters)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to export audit trails",
		})
		return
	}

	// Log export action
	h.auditService.LogAction(audit.AuditTrail{
		ActorID:      c.GetString("user_id"),
		ActorUsername: c.GetString("username"),
		ActorRole:    c.GetString("role"),
		Action:       "AUDIT_EXPORT",
		ResourceType: "audit_trails",
		ResourceID:   "export",
		IPAddress:    c.ClientIP(),
		Metadata: map[string]interface{}{
			"filters": filters,
			"format":  "json",
		},
	})

	// Set headers for file download
	filename := "audit_trails_" + time.Now().Format("20060102_150405") + ".json"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/json", data)
}

// GetWorkflowVersionHistory retrieves version history for a workflow
func (h *AuditHandler) GetWorkflowVersionHistory(c *gin.Context) {
	workflowID := c.Param("workflow_id")
	if workflowID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow_id is required"})
		return
	}

	versions, err := h.auditService.GetWorkflowVersionHistory(workflowID)
	if err != nil {
		logger.LogError("Failed to retrieve workflow version history", err, map[string]interface{}{
			"workflow_id": workflowID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve workflow version history",
		})
		return
	}

	// Log version history access
	h.auditService.LogAction(audit.AuditTrail{
		ActorID:      c.GetString("user_id"),
		ActorUsername: c.GetString("username"),
		ActorRole:    c.GetString("role"),
		Action:       "VERSION_HISTORY_ACCESS",
		ResourceType: "workflow",
		ResourceID:   workflowID,
		IPAddress:    c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"workflow_id": workflowID,
		"versions":    versions,
		"total_versions": len(versions),
	})
}

// CreateWorkflowVersion creates a new workflow version
func (h *AuditHandler) CreateWorkflowVersion(c *gin.Context) {
	var req struct {
		WorkflowID   string                 `json:"workflow_id" binding:"required"`
		Name         string                 `json:"name" binding:"required"`
		Description  string                 `json:"description"`
		Definition   map[string]interface{} `json:"definition" binding:"required"`
		ChangeReason string                 `json:"change_reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	version := audit.WorkflowVersion{
		WorkflowID:   req.WorkflowID,
		Name:         req.Name,
		Description:  req.Description,
		Definition:   req.Definition,
		CreatedBy:    c.GetString("user_id"),
		ChangeReason: req.ChangeReason,
	}

	createdVersion, err := h.auditService.CreateWorkflowVersion(version)
	if err != nil {
		logger.LogError("Failed to create workflow version", err, map[string]interface{}{
			"workflow_id": req.WorkflowID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create workflow version",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"version": createdVersion,
		"message": "Workflow version created successfully",
	})
}

// GetAuditStatistics retrieves audit statistics
func (h *AuditHandler) GetAuditStatistics(c *gin.Context) {
	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = parsed
		}
	}

	filters := map[string]interface{}{
		"start_date": startDate,
		"end_date":   endDate,
	}

	trails, err := h.auditService.GetAuditTrails(filters, 10000, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit statistics"})
		return
	}

	// Calculate statistics
	actionCounts := make(map[string]int)
	userCounts := make(map[string]int)
	complianceCounts := make(map[string]int)
	resourceTypeCounts := make(map[string]int)

	for _, trail := range trails {
		actionCounts[string(trail.Action)]++
		userCounts[trail.ActorUsername]++
		complianceCounts[string(trail.ComplianceStatus)]++
		resourceTypeCounts[trail.ResourceType]++
	}

	c.JSON(http.StatusOK, gin.H{
		"period": gin.H{
			"start": startDate,
			"end":   endDate,
		},
		"total_events": len(trails),
		"statistics": gin.H{
			"by_action":        actionCounts,
			"by_user":          userCounts,
			"by_compliance":    complianceCounts,
			"by_resource_type": resourceTypeCounts,
		},
	})
}