package handlers

import (
	"net/http"
	"strconv"
	"time"

	"agileos-backend/analytics"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	service *analytics.Service
}

func NewAnalyticsHandler(service *analytics.Service) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

// GetOverview handles GET /api/v1/analytics/overview
func (h *AnalyticsHandler) GetOverview(c *gin.Context) {
	// Parse query parameters
	filter := models.AnalyticsFilter{}

	// Days parameter (default: 7)
	if daysStr := c.Query("days"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil {
			filter.Days = days
		}
	} else {
		filter.Days = 7
	}

	// Workflow ID filter
	if workflowID := c.Query("workflow_id"); workflowID != "" {
		filter.WorkflowID = workflowID
	}

	// Department filter
	if department := c.Query("department"); department != "" {
		filter.Department = department
	}

	// Get analytics overview
	overview, err := h.service.GetOverview(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate analytics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetWorkflowEfficiency handles GET /api/v1/analytics/workflows
func (h *AnalyticsHandler) GetWorkflowEfficiency(c *gin.Context) {
	filter := models.AnalyticsFilter{Days: 7}

	efficiency, err := h.service.GetWorkflowEfficiency(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get workflow efficiency",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflows": efficiency,
		"count":     len(efficiency),
	})
}

// GetStepPerformance handles GET /api/v1/analytics/steps
func (h *AnalyticsHandler) GetStepPerformance(c *gin.Context) {
	filter := models.AnalyticsFilter{Days: 7}

	performance, err := h.service.GetStepPerformance(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get step performance",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"steps": performance,
		"count": len(performance),
	})
}

// GetDepartmentMetrics handles GET /api/v1/analytics/departments
func (h *AnalyticsHandler) GetDepartmentMetrics(c *gin.Context) {
	filter := models.AnalyticsFilter{Days: 7}

	metrics, err := h.service.GetDepartmentMetrics(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get department metrics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"departments": metrics,
		"count":       len(metrics),
	})
}

// GetSummary handles GET /api/v1/analytics/summary
func (h *AnalyticsHandler) GetSummary(c *gin.Context) {
	filter := models.AnalyticsFilter{Days: 7}

	summary, err := h.service.GetSummary(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get summary",
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetInsights handles GET /api/v1/analytics/insights
func (h *AnalyticsHandler) GetInsights(c *gin.Context) {
	filter := models.AnalyticsFilter{Days: 7}

	overview, err := h.service.GetOverview(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate insights",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"insights":     overview.Insights,
		"count":        len(overview.Insights),
		"generated_at": time.Now(),
	})
}
