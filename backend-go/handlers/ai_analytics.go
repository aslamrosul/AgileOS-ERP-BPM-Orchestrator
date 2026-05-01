package handlers

import (
	"net/http"
	"strconv"

	"agileos-backend/analytics"
	"agileos-backend/logger"

	"github.com/gin-gonic/gin"
)

// AIAnalyticsHandler handles AI-powered analytics requests
type AIAnalyticsHandler struct {
	pythonClient *analytics.PythonAnalyticsClient
}

// NewAIAnalyticsHandler creates a new AI analytics handler
func NewAIAnalyticsHandler() *AIAnalyticsHandler {
	return &AIAnalyticsHandler{
		pythonClient: analytics.NewPythonAnalyticsClient(),
	}
}

// GetWorkflowPrediction handles workflow completion prediction requests
func (h *AIAnalyticsHandler) GetWorkflowPrediction(c *gin.Context) {
	workflowID := c.Param("workflow_id")
	if workflowID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow_id is required"})
		return
	}

	// Check if Python service is healthy
	if !h.pythonClient.IsHealthy() {
		logger.LogError("Python analytics service is unhealthy", nil, map[string]interface{}{
			"workflow_id": workflowID,
		})
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI analytics service is currently unavailable",
			"fallback": gin.H{
				"message": "Using basic analytics instead of AI predictions",
				"estimated_completion": "2-4 hours (fallback estimate)",
			},
		})
		return
	}

	prediction, err := h.pythonClient.PredictWorkflowCompletion(workflowID)
	if err != nil {
		logger.LogError("Failed to get workflow prediction", err, map[string]interface{}{
			"workflow_id": workflowID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate prediction",
			"details": err.Error(),
		})
		return
	}

	// Log the prediction for audit
	logger.LogAudit("ai_prediction_generated", c.GetString("user_id"), workflowID, map[string]interface{}{
		"predicted_duration": prediction.EstimatedDurationMinutes,
		"confidence_score":   prediction.ConfidenceScore,
		"prediction_method":  prediction.Factors,
	})

	c.JSON(http.StatusOK, gin.H{
		"workflow_id": workflowID,
		"prediction": prediction,
		"ai_powered": true,
		"generated_at": prediction.PredictedCompletionTime,
	})
}

// GetAnomalies handles anomaly detection requests
func (h *AIAnalyticsHandler) GetAnomalies(c *gin.Context) {
	// Optional query parameters for filtering
	severityFilter := c.Query("severity")
	limitStr := c.DefaultQuery("limit", "20")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	// Check if Python service is healthy
	if !h.pythonClient.IsHealthy() {
		logger.LogError("Python analytics service is unhealthy for anomaly detection", nil, nil)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI analytics service is currently unavailable",
			"fallback": gin.H{
				"message": "Anomaly detection requires AI service",
				"anomalies": []interface{}{},
			},
		})
		return
	}

	anomalies, err := h.pythonClient.GetAnomalies()
	if err != nil {
		logger.LogError("Failed to get anomalies", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to detect anomalies",
			"details": err.Error(),
		})
		return
	}

	// Filter by severity if specified
	if severityFilter != "" {
		filteredAnomalies := make([]analytics.AnomalyDetection, 0)
		for _, anomaly := range anomalies {
			if anomaly.Severity == severityFilter {
				filteredAnomalies = append(filteredAnomalies, anomaly)
			}
		}
		anomalies = filteredAnomalies
	}

	// Apply limit
	if len(anomalies) > limit {
		anomalies = anomalies[:limit]
	}

	// Log anomaly detection for audit
	logger.LogAudit("anomaly_detection_performed", c.GetString("user_id"), "", map[string]interface{}{
		"anomalies_found": len(anomalies),
		"severity_filter": severityFilter,
		"limit":          limit,
	})

	c.JSON(http.StatusOK, gin.H{
		"anomalies": anomalies,
		"total_found": len(anomalies),
		"ai_powered": true,
		"filters": gin.H{
			"severity": severityFilter,
			"limit":    limit,
		},
	})
}

// GetComprehensiveAIAnalytics handles comprehensive AI analytics requests
func (h *AIAnalyticsHandler) GetComprehensiveAIAnalytics(c *gin.Context) {
	// Check if Python service is healthy
	if !h.pythonClient.IsHealthy() {
		logger.LogError("Python analytics service is unhealthy for comprehensive analytics", nil, nil)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI analytics service is currently unavailable",
			"fallback": gin.H{
				"message": "Using basic analytics instead of AI-powered insights",
				"basic_analytics_available": true,
			},
		})
		return
	}

	comprehensiveAnalytics, err := h.pythonClient.GetComprehensiveAnalytics()
	if err != nil {
		logger.LogError("Failed to get comprehensive AI analytics", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate comprehensive analytics",
			"details": err.Error(),
		})
		return
	}

	// Enhance response with additional metadata
	response := gin.H{
		"ai_analytics": comprehensiveAnalytics,
		"ai_powered": true,
		"generated_at": gin.H{
			"timestamp": "now",
			"service": "python-fastapi",
		},
		"summary": gin.H{
			"predictions_count": len(comprehensiveAnalytics.Predictions),
			"anomalies_count":   len(comprehensiveAnalytics.Anomalies),
			"insights_count":    len(comprehensiveAnalytics.Insights),
			"has_performance_metrics": len(comprehensiveAnalytics.PerformanceMetrics) > 0,
		},
	}

	// Add severity breakdown for anomalies
	severityBreakdown := make(map[string]int)
	for _, anomaly := range comprehensiveAnalytics.Anomalies {
		severityBreakdown[anomaly.Severity]++
	}
	response["anomaly_severity_breakdown"] = severityBreakdown

	// Add confidence analysis for predictions
	if len(comprehensiveAnalytics.Predictions) > 0 {
		totalConfidence := 0.0
		for _, prediction := range comprehensiveAnalytics.Predictions {
			totalConfidence += prediction.ConfidenceScore
		}
		avgConfidence := totalConfidence / float64(len(comprehensiveAnalytics.Predictions))
		response["prediction_confidence"] = gin.H{
			"average": avgConfidence,
			"quality": func() string {
				if avgConfidence >= 0.8 {
					return "high"
				} else if avgConfidence >= 0.6 {
					return "medium"
				}
				return "low"
			}(),
		}
	}

	// Log comprehensive analytics request for audit
	logger.LogAudit("comprehensive_ai_analytics_generated", c.GetString("user_id"), "", map[string]interface{}{
		"predictions_count": len(comprehensiveAnalytics.Predictions),
		"anomalies_count":   len(comprehensiveAnalytics.Anomalies),
		"insights_count":    len(comprehensiveAnalytics.Insights),
	})

	c.JSON(http.StatusOK, response)
}

// GetWorkflowPerformanceAI handles AI-powered workflow performance analysis
func (h *AIAnalyticsHandler) GetWorkflowPerformanceAI(c *gin.Context) {
	workflowID := c.Param("workflow_id")
	if workflowID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow_id is required"})
		return
	}

	// Check if Python service is healthy
	if !h.pythonClient.IsHealthy() {
		logger.LogError("Python analytics service is unhealthy for workflow performance", nil, map[string]interface{}{
			"workflow_id": workflowID,
		})
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI analytics service is currently unavailable",
			"fallback": gin.H{
				"message": "Using basic workflow analytics instead of AI analysis",
				"workflow_id": workflowID,
			},
		})
		return
	}

	performance, err := h.pythonClient.GetWorkflowPerformance(workflowID)
	if err != nil {
		logger.LogError("Failed to get AI workflow performance", err, map[string]interface{}{
			"workflow_id": workflowID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze workflow performance",
			"details": err.Error(),
		})
		return
	}

	// Enhance response with AI insights
	response := gin.H{
		"workflow_performance": performance,
		"ai_powered": true,
		"ai_insights": gin.H{
			"efficiency_rating": func() string {
				if performance.CompletionRate >= 0.9 {
					return "excellent"
				} else if performance.CompletionRate >= 0.7 {
					return "good"
				} else if performance.CompletionRate >= 0.5 {
					return "fair"
				}
				return "needs_improvement"
			}(),
			"recommendations": generateWorkflowRecommendations(performance),
		},
	}

	// Log workflow performance analysis for audit
	logger.LogAudit("ai_workflow_performance_analyzed", c.GetString("user_id"), workflowID, map[string]interface{}{
		"completion_rate":     performance.CompletionRate,
		"total_instances":     performance.TotalInstances,
		"completed_instances": performance.CompletedInstances,
	})

	c.JSON(http.StatusOK, response)
}

// RefreshAICache handles cache refresh requests for the Python service
func (h *AIAnalyticsHandler) RefreshAICache(c *gin.Context) {
	// Check if Python service is healthy
	if !h.pythonClient.IsHealthy() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI analytics service is currently unavailable",
		})
		return
	}

	err := h.pythonClient.RefreshCache()
	if err != nil {
		logger.LogError("Failed to refresh AI analytics cache", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to refresh AI analytics cache",
			"details": err.Error(),
		})
		return
	}

	// Log cache refresh for audit
	logger.LogAudit("ai_analytics_cache_refreshed", c.GetString("user_id"), "", map[string]interface{}{
		"action": "manual_cache_refresh",
	})

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "AI analytics cache refreshed successfully",
		"timestamp": "now",
	})
}

// GetAIServiceStatus handles AI service health status requests
func (h *AIAnalyticsHandler) GetAIServiceStatus(c *gin.Context) {
	isHealthy := h.pythonClient.IsHealthy()
	
	status := gin.H{
		"service": "python-fastapi-analytics",
		"status": func() string {
			if isHealthy {
				return "healthy"
			}
			return "unhealthy"
		}(),
		"ai_capabilities": gin.H{
			"predictions": isHealthy,
			"anomaly_detection": isHealthy,
			"comprehensive_analytics": isHealthy,
			"workflow_performance": isHealthy,
		},
		"fallback_available": true,
	}

	statusCode := http.StatusOK
	if !isHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, status)
}

// Helper function to generate workflow recommendations
func generateWorkflowRecommendations(performance *analytics.WorkflowPerformance) []string {
	recommendations := make([]string, 0)

	if performance.CompletionRate < 0.8 {
		recommendations = append(recommendations, "Consider reviewing workflow steps to improve completion rate")
	}

	if performance.AvgDurationMinutes != nil && *performance.AvgDurationMinutes > 240 { // More than 4 hours
		recommendations = append(recommendations, "Workflow duration is above average - consider process optimization")
	}

	if performance.TotalInstances < 10 {
		recommendations = append(recommendations, "Limited data available - predictions will improve with more workflow instances")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Workflow performance is within expected parameters")
	}

	return recommendations
}